// Package config owns validated, versioned preferences independently of the UI.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ntnyq/rivu/internal/storage"
)

type Store struct {
	mu       sync.Mutex
	path     string
	settings Settings
	loadErr  error
}

// Open returns a usable default snapshot on failure, but refuses later writes
// until Reload succeeds. A corrupt or future-version file is never overwritten.
func Open(path string) (*Store, error) {
	s := &Store{path: path, settings: Default()}
	err := s.Reload()
	return s, err
}

func (s *Store) Path() string { return s.path }

func (s *Store) Snapshot() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.settings
	return v
}

func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := Default()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	} else if err == nil {
		err = json.Unmarshal(data, &next)
		if err == nil && next.Version == 1 {
			next.Version = 2
		}
	}
	if err == nil {
		err = next.Validate()
	}
	s.loadErr = err
	if err == nil {
		s.settings = next
	}
	return err
}

func (s *Store) Update(change func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loadErr != nil {
		return fmt.Errorf("repair settings and reload before saving: %w", s.loadErr)
	}
	next := s.settings
	change(&next)
	if err := next.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	if err = storage.WriteFile(s.path, append(data, '\n'), 0o600); err != nil {
		var uncertain *storage.CommitError
		if errors.As(err, &uncertain) {
			s.loadErr = err
		}
		return err
	}
	s.settings = next
	return nil
}

// SavePreferences applies a settings draft without overwriting values controlled
// by the tray or updater while that draft was open.
func (s *Store) SavePreferences(draft Settings) error {
	return s.Update(func(current *Settings) {
		draft.SlowMode = current.SlowMode
		draft.LastCheck = current.LastCheck
		draft.LastAttempt = current.LastAttempt
		*current = draft
	})
}
