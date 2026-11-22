package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ntnyq/rivu/internal/problem"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ntnyq/rivu/internal/config"
)

// ErrQuiesced allows background bookkeeping to defer work during maintenance
// without reporting an expected pause as an application failure.
var ErrQuiesced = errors.New("session is quiesced")

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
	maintenance         int
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
		// An old journal shared through --config belongs to an existing
		// session. A new data directory starts with a clean ledger.
		m.state.WatchJournalMigrated = true
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
			if !m.closed && !m.quiesced() && m.loadErr == nil {
				m.tick()
			}
			m.mu.Unlock()
		}
	}
}

func (m *Manager) Configure(s config.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	return m.configure(s)
}

func (m *Manager) configure(s config.Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	m.settings = s
	d, u := s.LimitsAt(time.Now())
	m.engine.Limits(d, u)
	return nil
}

// TryCheckpoint is used when the OS can end the process as soon as the native
// callback returns. If busy, the last durable checkpoint remains available.
func (m *Manager) TryCheckpoint() error {
	if !m.mu.TryLock() {
		return nil
	}
	defer m.mu.Unlock()
	if m.closed || m.quiesced() || m.loadErr != nil {
		return nil
	}
	return m.save(m.state)
}

// quiesced is called under mu; persistence recovery and maintenance have
// independent lifetimes. Reload may clear only the former.
func (m *Manager) quiesced() bool { return m.barrier || m.maintenance > 0 }

// Quiesce acquires an independent maintenance hold and saves/stops the session.
// The caller must release its hold even on error. Release is idempotent and
// never clears a persistence recovery barrier or another caller's hold.
func (m *Manager) Quiesce() (release func(), err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maintenance++
	var once sync.Once
	release = func() {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.maintenance--
			m.tick()
			m.publish()
		})
	}
	defer m.publish()
	for id := range m.active {
		if stopErr := m.stop(id); stopErr != nil {
			err = errors.Join(err, stopErr)
		}
	}
	if err == nil {
		err = m.save(m.state)
	}
	if err != nil {
		m.barrier = true
		m.errorState = problem.New(problem.Persistence, "save session", err)
	}
	return release, err
}

func (m *Manager) Close() error {
	release, err := m.Quiesce()
	defer release()
	if err != nil {
		return err
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	err = m.engine.Close()
	if err == nil {
		m.closed = true
		m.cancel()
	} else {
		m.barrier = true
	}
	m.mu.Unlock()
	if err == nil {
		<-m.done
	}
	return err
}

func (m *Manager) HasPendingRunIntent() bool {
	for _, t := range m.Snapshot().Tasks {
		if t.Run && !t.Removing {
			return true
		}
	}
	return false
}
