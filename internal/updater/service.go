// Package updater coordinates signed MyGo updates with the session barrier.
package updater

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"

	"github.com/ntnyq/rivu/internal/i18n"
)

type Session interface {
	Idle() bool
	Quiesce() error
	Continue()
}
type Snapshot struct {
	State, Version, Notes, Error string
	Downloaded, Total            int64
	LastSuccess                  time.Time
}
type Service struct {
	mu       sync.Mutex
	store    *config.Store
	session  Session
	snapshot Snapshot
	update   *mygo.Update
	busy     bool
}

func New(store *config.Store, session Session) *Service {
	return &Service{store: store, session: session, snapshot: Snapshot{State: "idle", LastSuccess: store.Snapshot().LastCheck}}
}
func (s *Service) Snapshot() Snapshot { s.mu.Lock(); defer s.mu.Unlock(); return s.snapshot }
func Due(p config.Settings, now time.Time) bool {
	return p.AutoCheck && now.Sub(p.LastCheck) >= 24*time.Hour && now.Sub(p.LastAttempt) >= 24*time.Hour+7*time.Minute
}
func (s *Service) Scheduled(ctx context.Context) {
	if Due(s.store.Snapshot(), time.Now()) {
		_ = s.Check(ctx)
	}
	if s.store.Snapshot().AutoInstall && s.Snapshot().State == "available" && s.session.Idle() {
		_ = s.Install(ctx, true)
	}
}
func (s *Service) Check(ctx context.Context) error {
	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		return nil
	}
	s.busy = true
	s.snapshot.State = "checking"
	s.snapshot.Error = ""
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.busy = false; s.mu.Unlock() }()
	if !mygo.Updater.Enabled() {
		s.mu.Lock()
		s.snapshot.State = "unavailable"
		s.snapshot.Error = i18n.T("updates.sourceUnavailable")
		s.mu.Unlock()
		return nil
	}
	err := s.store.Update(func(p *config.Settings) { p.LastAttempt = time.Now() })
	if err != nil {
		return s.fail("failed", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	up, err := mygo.Updater.Check(ctx)
	if err != nil {
		return s.fail("failed", err)
	}
	now := time.Now()
	if err = s.store.Update(func(p *config.Settings) { p.LastCheck = now }); err != nil {
		return s.fail("failed", err)
	}
	s.mu.Lock()
	s.update = up
	s.snapshot.LastSuccess = now
	s.snapshot.State = "upToDate"
	if up != nil {
		s.snapshot.State = "available"
		s.snapshot.Version = up.Version
		s.snapshot.Notes = up.Notes
	}
	s.mu.Unlock()
	return nil
}
func (s *Service) fail(state string, err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.State = state
	s.snapshot.Error = err.Error()
	return err
}
func (s *Service) Install(ctx context.Context, automatic bool) error {
	s.mu.Lock()
	if s.busy || s.update == nil {
		s.mu.Unlock()
		return fmt.Errorf("%s", i18n.T("updates.nothingToInstall"))
	}
	s.busy = true
	up := s.update
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.busy = false; s.mu.Unlock() }()
	if automatic && !s.session.Idle() {
		return nil
	}
	if err := s.session.Quiesce(); err != nil {
		return s.fail("installFailed", err)
	}
	defer s.session.Continue()
	// The barrier prevents admissions; durable queued intent also counts as activity.
	if automatic && !s.session.Idle() {
		return nil
	}
	s.mu.Lock()
	s.snapshot.State = "installing"
	s.mu.Unlock()
	err := up.Install(ctx, func(done, total int64) {
		s.mu.Lock()
		s.snapshot.Downloaded = done
		s.snapshot.Total = total
		s.mu.Unlock()
	})
	if err != nil {
		return s.fail("installFailed", err)
	}
	s.mu.Lock()
	s.snapshot.State = "restartRequired"
	s.mu.Unlock()
	return nil
}
