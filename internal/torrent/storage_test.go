package torrent

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ntnyq/rivu/internal/files"
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
