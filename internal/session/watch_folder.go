package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/locales"
)

// MigrateWatchJournal preserves the old config-adjacent ledger in this session.
func (m *Manager) MigrateWatchJournal(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state.WatchJournalMigrated {
		return nil
	}
	if m.closed || m.quiesced() {
		return problem.Message(problem.Persistence, locales.Messages.SessionSaving, ErrQuiesced)
	}
	if m.loadErr != nil {
		return m.loadErr
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read watch journal: %w", err)
	}
	next := cloneState(m.state)
	if err == nil {
		var journal map[string]bool
		if err := json.Unmarshal(data, &journal); err != nil {
			return fmt.Errorf("read watch journal: %w", err)
		}
		if journal == nil {
			return fmt.Errorf("invalid watch journal")
		}
		for key, processed := range journal {
			if processed && key != "" {
				next.markWatched(key)
			}
		}
	}
	next.WatchJournalMigrated = true
	defer m.publish()
	return m.save(next)
}

func (m *Manager) Watched(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.Watched[key]
}

func (s *diskState) markWatched(key string) {
	if key == "" {
		return
	}
	if s.Watched == nil {
		s.Watched = map[string]bool{}
	}
	s.Watched[key] = true
}
