package session

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ntnyq/rivu/internal/problem"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/storage"
)

type diskState struct {
	Version  int      `json:"version"`
	Revision uint64   `json:"revision"`
	Tasks    []Record `json:"tasks"`
}

func (m *Manager) validate() error {
	if m.state.Version == 1 {
		m.state.Version = 2
	}
	if m.state.Version != 2 {
		return fmt.Errorf("unsupported session schema %d", m.state.Version)
	}
	seen := map[string]bool{}
	for _, r := range m.state.Tasks {
		id, err := hex.DecodeString(r.ID)
		if err != nil || len(id) != 16 || seen[r.ID] {
			return fmt.Errorf("invalid task ID")
		}
		seen[r.ID] = true
		if !filepath.IsAbs(r.Destination) || r.Source.Identity == (Identity{}) || math.IsNaN(r.SeedRatio) || math.IsInf(r.SeedRatio, 0) || r.SeedRatio < -1 || r.SeedSeconds < 0 || r.SeedLimitSeconds < 0 {
			return fmt.Errorf("invalid task record")
		}
		if r.Metadata != "" && r.Metadata != filepath.Join(filepath.Dir(m.path), "metadata", r.ID+".torrent") {
			return fmt.Errorf("invalid metadata reference")
		}
		if err := files.ValidateMembers(names(r.Source.Files)); err != nil {
			return err
		}
		if len(r.Source.Files) > 0 {
			if err := validateSelection(r.Source.Files); err != nil {
				return err
			}
		}
		if err := validateTags(r.Tags); err != nil {
			return err
		}
		for name, identity := range r.Adopted {
			if identity == "" || !slices.Contains(names(r.Source.Files), name) {
				return fmt.Errorf("invalid adopted file record")
			}
		}
		for _, p := range r.Trashed {
			if !slices.Contains(names(r.Source.Files), p) {
				return fmt.Errorf("invalid trash journal")
			}
		}
	}
	for i, r := range m.state.Tasks {
		for _, other := range m.state.Tasks[:i] {
			if r.Source.Identity.Matches(other.Source.Identity) {
				return fmt.Errorf("duplicate torrent identity")
			}
		}
	}
	return nil
}

func (m *Manager) save(next diskState) error {
	if m.loadErr != nil {
		return m.loadErr
	}
	next.Revision = m.state.Revision + 1
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return err
	}
	if err = storage.WriteFile(m.path, append(data, '\n'), 0o600); err != nil {
		var uncertain *storage.CommitError
		if errors.As(err, &uncertain) {
			m.loadErr = err
			m.errorState = problem.New(problem.Persistence, "save session", err)
			m.barrier = true
			for id := range m.active {
				_ = m.stop(id)
			}
		}
		return err
	}
	m.state = next
	m.lastSave = time.Now()
	m.errorState = nil
	return nil
}

func cloneState(s diskState) diskState {
	data, _ := json.Marshal(s)
	var out diskState
	_ = json.Unmarshal(data, &out)
	return out
}
