// Package updater coordinates signed MyGo updates with the session barrier.
package updater

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/locales"
)

type Session interface {
	HasPendingRunIntent() bool
	Quiesce() (func(), error)
}

type installer interface {
	Install(context.Context, func(int64, int64)) error
}

var errBusy = errors.New("update operation already running")
var errRestartRequired = errors.New("restart required")

type Snapshot struct {
	State, Version, Notes, Error string
	Downloaded, Total            int64
	LastSuccess                  time.Time
}
type Service struct {
	mu         sync.Mutex
	store      *config.Store
	session    Session
	snapshot   Snapshot
	queuedWork func() bool
	update     installer
	busy       bool
	stopped    bool
	cancel     context.CancelFunc
	done       chan struct{}
}

func New(store *config.Store, session Session, queuedWork func() bool) *Service {
	return &Service{store: store, session: session, queuedWork: queuedWork, snapshot: Snapshot{State: "idle", LastSuccess: store.Snapshot().LastCheck}}
}
func (s *Service) Snapshot() Snapshot { s.mu.Lock(); defer s.mu.Unlock(); return s.snapshot }

// Stop prevents new checks/installs, cancels the current operation, and waits
// for its cleanup (including context-insensitive replacement) to finish.
func (s *Service) Stop() {
	s.mu.Lock()
	s.stopped = true
	cancel, done := s.cancel, s.done
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// Resume is used when quitting failed and the application remains open.
func (s *Service) Resume() { s.mu.Lock(); s.stopped = false; s.mu.Unlock() }

func (s *Service) begin(ctx context.Context) (context.Context, func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s.stopped {
		return nil, nil, context.Canceled
	}
	if s.busy {
		return nil, nil, errBusy
	}
	if s.snapshot.State == "restartRequired" {
		return nil, nil, errRestartRequired
	}
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	s.busy, s.cancel, s.done = true, cancel, done
	return ctx, func() {
		cancel()
		s.mu.Lock()
		defer s.mu.Unlock()
		s.busy, s.cancel, s.done = false, nil, nil
		close(done)
	}, nil
}
func Due(p config.Settings, now time.Time) bool {
	return p.AutoCheck && now.Sub(p.LastCheck) >= 24*time.Hour && now.Sub(p.LastAttempt) >= 24*time.Hour+7*time.Minute
}
func (s *Service) canAutoInstall() bool {
	return (s.queuedWork == nil || !s.queuedWork()) && !s.session.HasPendingRunIntent()
}

func (s *Service) Scheduled(ctx context.Context) {
	if Due(s.store.Snapshot(), time.Now()) {
		_ = s.Check(ctx)
	}
	if s.store.Snapshot().AutoInstall && s.Snapshot().State == "available" && s.canAutoInstall() {
		_ = s.Install(ctx, true)
	}
}
func (s *Service) Check(ctx context.Context) error {
	ctx, finish, err := s.begin(ctx)
	if errors.Is(err, errBusy) || errors.Is(err, errRestartRequired) {
		return nil
	}
	if err != nil {
		return err
	}
	defer finish()
	s.mu.Lock()
	s.update = nil
	s.snapshot = Snapshot{State: "checking", LastSuccess: s.snapshot.LastSuccess}
	s.mu.Unlock()
	if !mygo.Updater.Enabled() {
		s.mu.Lock()
		s.snapshot.State = "unavailable"
		s.snapshot.Error = locales.UpdatesSourceUnavailable()
		s.mu.Unlock()
		return nil
	}
	err = s.store.Update(func(p *config.Settings) { p.LastAttempt = time.Now() })
	if err != nil {
		return s.fail("failed", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	up, err := checkWithRetry(ctx, 12*time.Second, []time.Duration{500 * time.Millisecond, time.Second}, mygo.Updater.Check)
	if err != nil {
		return s.fail("failed", err)
	}
	now := time.Now()
	if err = s.store.Update(func(p *config.Settings) { p.LastCheck = now }); err != nil {
		return s.fail("failed", err)
	}
	s.mu.Lock()
	s.update = nil
	s.snapshot.LastSuccess = now
	s.snapshot.State = "upToDate"
	if up != nil {
		s.update = up
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
	ctx, finish, err := s.begin(ctx)
	if errors.Is(err, errBusy) || errors.Is(err, errRestartRequired) {
		return fmt.Errorf("%s", locales.UpdatesNothingToInstall())
	}
	if err != nil {
		return err
	}
	defer finish()
	s.mu.Lock()
	up := s.update
	s.mu.Unlock()
	if up == nil {
		return fmt.Errorf("%s", locales.UpdatesNothingToInstall())
	}
	if automatic && !s.canAutoInstall() {
		return nil
	}
	release, err := s.session.Quiesce()
	defer release()
	if err != nil {
		return s.fail("installFailed", err)
	}
	// The barrier prevents admissions; durable queued intent also counts as activity.
	if automatic && !s.canAutoInstall() {
		return nil
	}
	s.mu.Lock()
	s.snapshot.State = "installing"
	s.snapshot.Error = ""
	s.snapshot.Downloaded, s.snapshot.Total = 0, 0
	s.mu.Unlock()
	err = up.Install(ctx, func(done, total int64) {
		s.mu.Lock()
		s.snapshot.Downloaded = done
		s.snapshot.Total = total
		s.mu.Unlock()
	})
	if err != nil {
		return s.fail("installFailed", err)
	}
	s.mu.Lock()
	s.update = nil
	s.snapshot.State = "restartRequired"
	s.mu.Unlock()
	return nil
}
