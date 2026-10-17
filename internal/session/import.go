package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/storage"
	"github.com/ntnyq/rivu/locales"
)

func names(fs []File) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.Path)
	}
	return out
}

func validateSelection(fs []File) error {
	selected := false
	for _, f := range fs {
		if f.Size < 0 || f.Size > 1<<50 || f.Priority < 0 || f.Priority > 2 {
			return fmt.Errorf("invalid priority")
		}
		selected = selected || f.Priority > 0
	}
	if !selected {
		return problem.Message(problem.Engine, locales.Messages.FilesSelectionRequired, nil)
	}
	return nil
}

func (m *Manager) checkDestination(s Source, dir string, except string, adopt bool) error {
	if _, err := files.InspectExisting(dir, names(s.Files), adopt); err != nil {
		return err
	}
	for _, r := range m.state.Tasks {
		if r.ID == except {
			continue
		}
		for _, a := range r.Source.Files {
			for _, b := range s.Files {
				left := strings.ToLower(filepath.Join(r.Destination, filepath.FromSlash(a.Path)))
				right := strings.ToLower(filepath.Join(dir, filepath.FromSlash(b.Path)))
				if left == right || strings.HasPrefix(left, right+string(filepath.Separator)) || strings.HasPrefix(right, left+string(filepath.Separator)) {
					return problem.Message(problem.Engine, locales.Messages.FilesDestinationConflict, nil)
				}
			}
		}
	}
	free, err := files.FreeSpace(dir)
	if err != nil {
		return err
	}
	size := s.requiredSpace()
	if adopt {
		for _, file := range s.Files {
			stat, err := os.Stat(filepath.Join(dir, filepath.FromSlash(file.Path)))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			if stat.Size() > file.Size {
				return fmt.Errorf("existing file exceeds torrent size: %s", file.Path)
			}
			if file.Priority > 0 {
				size -= min(stat.Size(), file.Size)
			}
		}
	}
	if size > free {
		return problem.New(problem.DiskFull, fmt.Sprintf("%d bytes required", size), nil)
	}
	return nil
}

func (m *Manager) metadata(r *Record, s Source) error {
	dir := filepath.Join(filepath.Dir(m.path), "metadata")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	r.Metadata = filepath.Join(dir, r.ID+".torrent")
	if err := storage.WriteFile(r.Metadata, s.Meta, 0o600); err != nil {
		return err
	}
	s.Meta = nil
	r.Source = s
	return nil
}

func (m *Manager) Import(ctx context.Context, in Import) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if m.closed || m.barrier {
		return "", problem.Message(problem.Engine, locales.Messages.SessionSavingRetry, nil)
	}
	if m.loadErr != nil {
		return "", m.loadErr
	}
	for _, r := range m.state.Tasks {
		if r.Source.Identity.Matches(in.Source.Identity) {
			return r.ID, nil
		}
	}
	in.Source.Files = slices.Clone(in.Source.Files)
	in.Source.Trackers = slices.Clone(in.Source.Trackers)
	in.Source.Meta = slices.Clone(in.Source.Meta)
	if in.Source.Identity == (Identity{}) {
		return "", problem.Message(problem.Engine, locales.Messages.ImportInvalidSource, nil)
	}
	if len(in.Source.Files) > 0 {
		if err := validateSelection(in.Source.Files); err != nil {
			return "", err
		}
	}
	absolute, err := filepath.Abs(in.Destination)
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	in.Destination = canonical
	if err := m.checkDestination(in.Source, in.Destination, "", in.AdoptExisting); err != nil {
		return "", err
	}
	if math.IsNaN(in.SeedRatio) || math.IsInf(in.SeedRatio, 0) || in.SeedRatio < -1 {
		return "", fmt.Errorf("invalid seed ratio")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	var adopted map[string]string
	if in.AdoptExisting {
		if len(in.Source.Files) == 0 {
			return "", fmt.Errorf("metadata required to adopt existing data")
		}
		adopted, err = files.InspectExisting(in.Destination, names(in.Source.Files), true)
		if err != nil {
			return "", err
		}
	}
	if err := validateTags(in.Tags); err != nil {
		return "", err
	}
	if in.SeedLimitSeconds < 0 {
		return "", fmt.Errorf("invalid seeding duration")
	}
	r := Record{Adopted: adopted, Tags: slices.Clone(in.Tags), SeedLimitSeconds: in.SeedLimitSeconds, ID: hex.EncodeToString(id), Source: in.Source, Destination: in.Destination, Run: in.Run, SeedRatio: in.SeedRatio, Added: time.Now(), Order: time.Now().UnixNano()}
	if len(in.Source.Meta) > 0 {
		if err := m.metadata(&r, in.Source); err != nil {
			return "", err
		}
	}
	next := cloneState(m.state)
	next.Tasks = append(next.Tasks, r)
	if err := m.save(next); err != nil {
		return "", err
	}
	m.tick()
	return r.ID, nil
}
