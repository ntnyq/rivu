package torrent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/anacrolix/torrent/metainfo"
	btstorage "github.com/anacrolix/torrent/storage"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/storage"

	"github.com/ntnyq/rivu/internal/i18n"
)

// Rooted storage resolves every payload open relative to an OS directory
// descriptor. Even a symlink swapped after validation cannot escape this root.
// Completion is deliberately not persisted: every attachment rehashes data.
type rootedStorage struct {
	root       *os.Root
	mu         sync.Mutex
	allowed    bool
	err        error
	info       *metainfo.Info
	names      []string
	complete   map[int]bool
	manifest   string
	owners     map[string]string
	saveOwners func(string, map[string]string) error
}

func (s *rootedStorage) OpenTorrent(_ context.Context, info *metainfo.Info, _ metainfo.Hash) (btstorage.TorrentImpl, error) {
	if err := validateInfo(info); err != nil {
		return btstorage.TorrentImpl{}, err
	}
	var names []string
	for _, f := range info.UpvertedFiles() {
		parts := append([]string{info.BestName()}, f.BestPath()...)
		for _, p := range parts {
			if p == "" || p == "." || p == ".." || strings.ContainsAny(p, "/\\") {
				return btstorage.TorrentImpl{}, fmt.Errorf("unsafe metadata path")
			}
		}
		names = append(names, strings.Join(parts, "/"))
	}
	if err := files.ValidateMembers(names); err != nil {
		return btstorage.TorrentImpl{}, err
	}
	s.info = info
	s.names = names
	s.complete = map[int]bool{}
	return btstorage.TorrentImpl{Piece: func(p metainfo.Piece) btstorage.PieceImpl { return &rootedPiece{s: s, p: p} }, Close: func() error { return nil }}, nil
}
func (s *rootedStorage) setAllowed(v bool) { s.mu.Lock(); defer s.mu.Unlock(); s.allowed = v }
func (s *rootedStorage) failure() error    { s.mu.Lock(); defer s.mu.Unlock(); return s.err }

type rootedPiece struct {
	s *rootedStorage
	p metainfo.Piece
}

func (p *rootedPiece) Completion() btstorage.Completion {
	p.s.mu.Lock()
	defer p.s.mu.Unlock()
	v, ok := p.s.complete[p.p.Index()]
	return btstorage.Completion{Ok: ok, Complete: v}
}
func (p *rootedPiece) MarkComplete() error {
	p.s.mu.Lock()
	defer p.s.mu.Unlock()
	p.s.complete[p.p.Index()] = true
	return nil
}
func (p *rootedPiece) MarkNotComplete() error {
	p.s.mu.Lock()
	defer p.s.mu.Unlock()
	p.s.complete[p.p.Index()] = false
	return nil
}
func (p *rootedPiece) ReadAt(b []byte, off int64) (int, error)  { return p.access(b, off, false) }
func (p *rootedPiece) WriteAt(b []byte, off int64) (int, error) { return p.access(b, off, true) }
func (p *rootedPiece) access(b []byte, off int64, write bool) (int, error) {
	s := p.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if write && !s.allowed {
		return 0, fmt.Errorf("payload is not admitted")
	}
	if off < 0 || off+int64(len(b)) > p.p.Length() {
		return 0, io.EOF
	}
	abs := p.p.Offset() + off
	end := abs + int64(len(b))
	n := 0
	for i, f := range s.info.UpvertedFiles() {
		start := max(abs, f.TorrentOffset)
		stop := min(end, f.TorrentOffset+f.Length)
		if stop <= start {
			continue
		}
		if start > abs+int64(n) {
			gap := int(start-abs) - n
			clear(b[n : n+gap])
			n += gap
		}
		name := s.names[i]
		var file *os.File
		var err error
		if write {
			err = s.root.MkdirAll(path.Dir(name), 0o755)
			if err == nil {
				file, err = s.openWrite(name)
			}
		} else {
			file, err = s.root.Open(name)
		}
		if err != nil {
			if write {
				s.err = err
			}
			return n, err
		}
		st, e := file.Stat()
		if e != nil || !st.Mode().IsRegular() {
			file.Close()
			return n, fmt.Errorf("payload is not a regular file")
		}
		count := int(stop - start)
		var done int
		if write {
			done, err = file.WriteAt(b[n:n+count], start-f.TorrentOffset)
			if err == nil {
				err = file.Sync()
			}
		} else {
			done, err = file.ReadAt(b[n:n+count], start-f.TorrentOffset)
		}
		file.Close()
		n += done
		if err != nil {
			if write {
				s.err = err
			}
			return n, err
		}
	}
	if n < len(b) {
		clear(b[n:])
		n = len(b)
	}
	return n, nil
}

// Claim only files created by this task. Existing files with an unrelated inode
// are never silently overwritten, even if inserted after import validation.
func (s *rootedStorage) openWrite(name string) (*os.File, error) {
	if s.manifest == "" {
		return nil, fmt.Errorf("missing ownership journal")
	}
	if s.owners == nil {
		owners, err := files.ReadOwnership(s.manifest)
		if err != nil {
			return nil, err
		}
		s.owners = owners
	}

	f, err := s.root.OpenFile(name, os.O_RDWR, 0)
	if os.IsNotExist(err) {
		return s.createOwned(name)
	}
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	id, err := files.FileIdentity(stat)
	if err != nil {
		f.Close()
		return nil, err
	}
	if !stat.Mode().IsRegular() || s.owners[name] != id {
		f.Close()
		return nil, fmt.Errorf(i18n.T("文件已被替换或不属于该任务: %s"), name)
	}
	return f, nil
}

// Create under a temporary name and persist its inode before publishing the
// payload name. Failure (or a crash) before publication leaves no unclaimed
// target that would block a later attempt. Publication must never replace a file.
func (s *rootedStorage) createOwned(name string) (_ *os.File, resultErr error) {
	parent, err := s.root.OpenRoot(path.Dir(name))
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	directory, err := parent.Open(".")
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return nil, err
	}
	temporary := ".rivu-file-" + hex.EncodeToString(random[:])
	f, err := parent.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil {
			f.Close()
		}
		_ = parent.Remove(temporary)
	}()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	id, err := files.FileIdentity(stat)
	if err != nil {
		return nil, err
	}
	if err = f.Sync(); err != nil {
		return nil, err
	}
	next := maps.Clone(s.owners)
	if next == nil {
		next = map[string]string{}
	}
	next[name] = id
	save := s.saveOwners
	if save == nil {
		save = files.SaveOwnership
	}
	if err = save(s.manifest, next); err != nil {
		// An uncertain rename may have committed. Never retain a speculative cache.
		s.owners = nil
		var uncertain *storage.CommitError
		if errors.As(err, &uncertain) {
			owners, readErr := files.ReadOwnership(s.manifest)
			if readErr != nil {
				return nil, errors.Join(err, readErr)
			}
			s.owners = owners
		}
		return nil, err
	}
	s.owners = next
	if err = files.RenameExclusive(directory, temporary, path.Base(name)); err != nil {
		return nil, err
	}
	if err = directory.Sync(); err != nil {
		return nil, err
	}
	return f, nil
}
