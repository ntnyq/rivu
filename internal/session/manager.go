package session

import (
	"context"
	"encoding/json"
	"github.com/ntnyq/rivu/internal/problem"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ntnyq/rivu/internal/config"
)

// Manager serializes all mutations and engine calls; rendering reads atomic snapshots.
type Manager struct {
	mu                  sync.Mutex
	path                string
	state               diskState
	engine              Engine
	settings            config.Settings
	active              map[string]bool
	holds               map[string]error
	samples             map[string]sample
	watchers            map[chan struct{}]bool
	snapshot            atomic.Pointer[Snapshot]
	loadErr, errorState error
	barrier, closed     bool
	cancel              context.CancelFunc
	done                chan struct{}
	lastSave            time.Time
	trash               func(string) error
}

func Open(dir string, e Engine, s config.Settings, trash func(string) error) (*Manager, error) {
	m := &Manager{path: filepath.Join(dir, "session.json"), engine: e, settings: s, state: diskState{Version: 2}, active: map[string]bool{}, holds: map[string]error{}, samples: map[string]sample{}, watchers: map[chan struct{}]bool{}, done: make(chan struct{}), trash: trash, lastSave: time.Now()}
	data, err := os.ReadFile(m.path)
	if err == nil {
		err = json.Unmarshal(data, &m.state)
		if err == nil {
			err = m.validate()
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		m.state = diskState{Version: 2}
		m.loadErr = err
		m.errorState = problem.New(problem.Persistence, "read session", err)
	}
	for _, r := range m.state.Tasks {
		if !s.Restore && r.Run {
			m.holds[r.ID] = problem.New(problem.RestoreDisabled, "", nil)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.publish()
	go m.loop(ctx)
	return m, err
}

func (m *Manager) loop(ctx context.Context) {
	defer close(m.done)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if !m.closed && !m.barrier && m.loadErr == nil {
				m.tick()
			}
			m.mu.Unlock()
		}
	}
}

func (m *Manager) Configure(s config.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := s.Validate(); err != nil {
		return err
	}
	m.settings = s
	d, u := s.LimitsAt(time.Now())
	m.engine.Limits(d, u)
	m.publish()
	return nil
}

func (m *Manager) Quiesce() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.barrier = true
	for id := range m.active {
		if err := m.stop(id); err != nil {
			return err
		}
	}
	err := m.save(m.state)
	if err != nil {
		m.errorState = problem.New(problem.Persistence, "save session", err)
	}
	m.publish()
	return err
}

func (m *Manager) Continue() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed && m.loadErr == nil {
		m.barrier = false
		m.tick()
	}
}

func (m *Manager) Close() error {
	if err := m.Quiesce(); err != nil {
		return err
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	err := m.engine.Close()
	if err == nil {
		m.closed = true
		m.cancel()
	}
	m.mu.Unlock()
	if err == nil {
		<-m.done
	}
	return err
}

func (m *Manager) Idle() bool {
	for _, t := range m.Snapshot().Tasks {
		if t.Run && !t.Removing {
			return false
		}
	}
	return true
}
