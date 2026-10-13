package session

import (
	"encoding/json"
	"fmt"
	"github.com/ntnyq/rivu/internal/problem"
	"os"
)

func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if m.closed {
		return fmt.Errorf("session closed")
	}
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	var next diskState
	if err = json.Unmarshal(data, &next); err != nil {
		return err
	}
	old := m.state
	m.state = next
	if err = m.validate(); err != nil {
		m.state = old
		return err
	}
	for id := range m.active {
		if err = m.stop(id); err != nil {
			m.state = old
			return err
		}
	}
	m.loadErr = nil
	m.errorState = nil
	m.barrier = false
	m.holds = map[string]error{}
	m.samples = map[string]sample{}
	if !m.settings.Restore {
		for _, r := range m.state.Tasks {
			if r.Run {
				m.holds[r.ID] = problem.New(problem.RestoreDisabled, "", nil)
			}
		}
	}
	m.tick()
	return nil
}
