// Package storage contains the durable file operations shared by application records.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile replaces a file only after its complete contents have been synced.
// The temporary file is on the same volume. Before rename, errors leave the old
// file intact. A directory-sync failure after rename is an uncertain commit.
func WriteFile(path string, data []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".rivu-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	if err = syncDirectory(filepath.Dir(path)); err != nil {
		return &CommitError{Err: err}
	}
	return nil
}

// CommitError means replacement happened but its durability is uncertain. The
// caller must reread the file before making another durable mutation.
type CommitError struct{ Err error }

func (e *CommitError) Error() string {
	return "replacement committed; directory sync failed: " + e.Err.Error()
}
func (e *CommitError) Unwrap() error { return e.Err }
