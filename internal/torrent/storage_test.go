package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/internal/storage"
)

func TestOwnershipPublicationRecoversFromJournalFailure(t *testing.T) {
	for _, uncertain := range []bool{false, true} {
		name := "before commit"
		if uncertain {
			name = "after commit"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			journal := filepath.Join(t.TempDir(), "owners.json")
			first := &rootedStorage{root: root, manifest: journal, saveOwners: func(path string, owners map[string]string) error {
				if uncertain {
					if err := files.SaveOwnership(path, owners); err != nil {
						return err
					}
					return &storage.CommitError{Err: errors.New("directory sync failed")}
				}
				return errors.New("disk full")
			}}
			if f, err := first.openWrite("payload"); err == nil {
				f.Close()
				t.Fatal("expected journal failure")
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("failed claim exposed a payload or leaked a temporary file: %v, %v", entries, err)
			}
			resumed := &rootedStorage{root: root, manifest: journal}
			f, err := resumed.openWrite("payload")
			if err != nil {
				t.Fatal("resume failed after journal recovery:", err)
			}
			if _, err = f.Write([]byte("data")); err != nil {
				t.Fatal(err)
			}
			stat, err := f.Stat()
			if err != nil {
				t.Fatal(err)
			}
			f.Close()
			identity, err := files.FileIdentity(stat)
			if err != nil {
				t.Fatal(err)
			}
			owners, err := files.ReadOwnership(journal)
			if err != nil || owners["payload"] != identity {
				t.Fatalf("payload is not durably owned: %v, %v", owners, err)
			}
		})
	}
}

func TestEmptyFilesAreOwnedOnlyAfterAdmission(t *testing.T) {
	for _, failure := range []string{"none", "journal", "unrelated file"} {
		t.Run(failure, func(t *testing.T) {
			dir := t.TempDir()
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			journal := filepath.Join(t.TempDir(), "owners.json")
			st := &rootedStorage{root: root, manifest: journal}
			info := metainfo.Info{Name: "bundle", PieceLength: 16384, Pieces: make([]byte, 20), Files: []metainfo.FileInfo{
				{Path: []string{"empty"}}, {Path: []string{"skip"}}, {Path: []string{"payload"}, Length: 1},
			}}
			if _, err = st.OpenTorrent(t.Context(), &info, metainfo.Hash{}); err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("storage touched payload before admission: %v, %v", entries, err)
			}
			policy := []session.File{{Priority: 1}, {}, {Priority: 1}}
			if failure == "journal" {
				st.saveOwners = func(string, map[string]string) error { return errors.New("disk full") }
			}
			if failure == "unrelated file" {
				if err = root.MkdirAll("bundle", 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, "bundle/empty"), []byte("unrelated"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if failure != "none" {
				if err = st.admit(policy); err == nil || st.allowed {
					t.Fatalf("unsafe admission succeeded: %v", err)
				}
				if failure == "unrelated file" {
					data, err := os.ReadFile(filepath.Join(dir, "bundle/empty"))
					if err != nil || string(data) != "unrelated" {
						t.Fatalf("unrelated file changed: %q, %v", data, err)
					}
					return
				}
				if _, err = root.Stat("bundle/empty"); !os.IsNotExist(err) {
					t.Fatalf("failed journal exposed an unowned empty file: %v", err)
				}
				st.saveOwners = nil
			}
			if err = st.admit(policy); err != nil {
				t.Fatal(err)
			}
			stat, err := root.Stat("bundle/empty")
			if err != nil {
				t.Fatal(err)
			}
			if stat.Size() != 0 || !stat.Mode().IsRegular() {
				t.Fatal("selected empty file missing or not empty")
			}
			identity, err := files.FileIdentity(stat)
			if err != nil {
				t.Fatal(err)
			}
			owners, err := files.ReadOwnership(journal)
			if err != nil || owners["bundle/empty"] != identity {
				t.Fatalf("empty file not journaled: %v, %v", owners, err)
			}
			for _, name := range []string{"bundle/skip", "bundle/payload"} {
				if _, err = root.Stat(name); !os.IsNotExist(err) {
					t.Fatalf("admission unexpectedly created %s: %v", name, err)
				}
			}
			// A fresh storage handle must recognize the existing owned empty file.
			resumed := &rootedStorage{root: root, manifest: journal}
			if _, err = resumed.OpenTorrent(t.Context(), &info, metainfo.Hash{}); err != nil {
				t.Fatal(err)
			}
			if err = resumed.admit(policy); err != nil {
				t.Fatal("empty-file resume failed:", err)
			}
		})
	}
}

func TestRootedAccessSpansFilesAndSkipsEmptyMembers(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	info := metainfo.Info{Name: "bundle", PieceLength: 16, Pieces: make([]byte, 20), Files: []metainfo.FileInfo{
		{Path: []string{"a"}, Length: 5}, {Path: []string{"empty"}}, {Path: []string{"b"}, Length: 11},
	}}
	st := &rootedStorage{root: root, manifest: filepath.Join(t.TempDir(), "owners.json"), allowed: true}
	impl, err := st.OpenTorrent(t.Context(), &info, metainfo.Hash{})
	if err != nil {
		t.Fatal(err)
	}
	piece := impl.Piece(info.Piece(0))
	payload := []byte("abcdefghijklmnop")
	if n, err := piece.WriteAt(payload, 0); err != nil || n != len(payload) {
		t.Fatalf("write: %d, %v", n, err)
	}
	for _, span := range [][2]int{{0, 16}, {3, 9}, {5, 11}, {15, 1}, {16, 0}} {
		got := make([]byte, span[1])
		if n, err := piece.ReadAt(got, int64(span[0])); err != nil || n != len(got) || !bytes.Equal(got, payload[span[0]:span[0]+span[1]]) {
			t.Fatalf("read %v: %q, %d, %v", span, got, n, err)
		}
	}
}

func BenchmarkRootedRead(b *testing.B) {
	for _, count := range []int{1, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			dir := b.TempDir()
			root, err := os.OpenRoot(dir)
			if err != nil {
				b.Fatal(err)
			}
			defer root.Close()
			info := metainfo.Info{Name: "bundle", PieceLength: 16384, Pieces: make([]byte, 20)}
			for i := range count {
				info.Files = append(info.Files, metainfo.FileInfo{Path: []string{fmt.Sprintf("file%d", i)}})
			}
			info.Files[count-1].Length = 16384
			if err = root.MkdirAll("bundle", 0700); err != nil {
				b.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "bundle", fmt.Sprintf("file%d", count-1)), make([]byte, 16384), 0600); err != nil {
				b.Fatal(err)
			}
			st := &rootedStorage{root: root}
			impl, err := st.OpenTorrent(b.Context(), &info, metainfo.Hash{})
			if err != nil {
				b.Fatal(err)
			}
			piece := impl.Piece(info.Piece(0))
			buf := make([]byte, 16384)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if _, err = piece.ReadAt(buf, 0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestRootedV2AccessRespectsAlignmentGaps(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	info := metainfo.Info{Name: "bundle", MetaVersion: 2, PieceLength: 16384, FileTree: metainfo.FileTree{Dir: map[string]metainfo.FileTree{
		"a": {File: metainfo.FileTreeFile{Length: 3, PiecesRoot: string(make([]byte, 32))}},
		"b": {File: metainfo.FileTreeFile{Length: 5, PiecesRoot: string(make([]byte, 32))}},
	}}}
	st := &rootedStorage{root: root, manifest: filepath.Join(t.TempDir(), "owners.json"), allowed: true}
	impl, err := st.OpenTorrent(t.Context(), &info, metainfo.Hash{})
	if err != nil {
		t.Fatal(err)
	}
	for i, data := range []string{"abc", "defgh"} {
		piece := impl.Piece(info.Piece(i))
		if n, err := piece.WriteAt([]byte(data), 0); err != nil || n != len(data) {
			t.Fatalf("piece %d write: %d, %v", i, n, err)
		}
		got := make([]byte, len(data))
		if n, err := piece.ReadAt(got, 0); err != nil || n != len(got) || string(got) != data {
			t.Fatalf("piece %d read: %q, %d, %v", i, got, n, err)
		}
		if _, err := piece.WriteAt([]byte{1}, int64(len(data))); !errors.Is(err, io.EOF) {
			t.Fatalf("piece %d accepted a write in its padding gap: %v", i, err)
		}
	}
}

func TestOwnershipPublicationDoesNotReplaceConcurrentFile(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	journal := filepath.Join(t.TempDir(), "owners.json")
	s := &rootedStorage{root: root, manifest: journal, saveOwners: func(path string, owners map[string]string) error {
		if err := files.SaveOwnership(path, owners); err != nil {
			return err
		}
		// Simulate another process creating the target after preflight.
		return os.WriteFile(filepath.Join(dir, "payload"), []byte("unrelated"), 0600)
	}}
	if f, err := s.openWrite("payload"); err == nil {
		f.Close()
		t.Fatal("replaced concurrent file")
	}
	got, err := os.ReadFile(filepath.Join(dir, "payload"))
	if err != nil || string(got) != "unrelated" {
		t.Fatalf("changed unrelated payload: %q, %v", got, err)
	}
	s = &rootedStorage{root: root, manifest: journal}
	if f, err := s.openWrite("payload"); err == nil {
		f.Close()
		t.Fatal("claimed unrelated file on retry")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary file leaked: %v, %v", entries, err)
	}
}
