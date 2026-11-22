package updater

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ntnyq/rivu/internal/config"
)

type installFunc func(context.Context, func(int64, int64)) error

func (f installFunc) Install(ctx context.Context, progress func(int64, int64)) error {
	return f(ctx, progress)
}

type heldSession struct {
	holds    atomic.Int32
	released atomic.Int32
	err      error
	pending  func() bool
}

func (s *heldSession) HasPendingRunIntent() bool {
	if s.pending != nil {
		return s.pending()
	}
	return false
}
func (s *heldSession) Quiesce() (func(), error) {
	s.holds.Add(1)
	return func() { s.holds.Add(-1); s.released.Add(1) }, s.err
}

func await(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for update operation")
	}
}

func TestStopWaitsForInstallCleanupAndPreventsNewWork(t *testing.T) {
	for _, installErr := range []error{nil, context.Canceled} {
		t.Run(fmt.Sprint(installErr), func(t *testing.T) {
			session := &heldSession{}
			session.holds.Store(1) // An independent sleep/quit hold must survive.
			started, canceled, cleanup, stopped := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			finishCleanup := sync.OnceFunc(func() { close(cleanup) })
			defer finishCleanup()
			service := &Service{session: session, update: installFunc(func(ctx context.Context, _ func(int64, int64)) error {
				close(started)
				<-ctx.Done()
				close(canceled)
				<-cleanup // Model replacement that cannot stop mid-rename.
				return installErr
			})}
			installed := make(chan error, 1)
			go func() { installed <- service.Install(t.Context(), false) }()
			await(t, started)
			go func() { service.Stop(); close(stopped) }()
			await(t, canceled)
			if err := service.Check(t.Context()); !errors.Is(err, context.Canceled) {
				t.Fatalf("check admitted after stop: %v", err)
			}
			if err := service.Install(t.Context(), false); !errors.Is(err, context.Canceled) {
				t.Fatalf("install admitted after stop: %v", err)
			}
			select {
			case <-stopped:
				t.Fatal("stop returned before replacement/cleanup completed")
			case <-time.After(30 * time.Millisecond):
			}
			finishCleanup()
			await(t, stopped)
			if err := <-installed; !errors.Is(err, installErr) {
				t.Fatalf("install result: %v", err)
			}
			if session.holds.Load() != 1 || session.released.Load() != 1 {
				t.Fatalf("incorrect hold release: holds=%d releases=%d", session.holds.Load(), session.released.Load())
			}
			service.Stop() // Stopping an idle service remains safe.
			service.Resume()
			if installErr != nil {
				service.update = installFunc(func(context.Context, func(int64, int64)) error { return nil })
				if err := service.Install(t.Context(), false); err != nil {
					t.Fatalf("failed-quit recovery did not reopen updates: %v", err)
				}
			}
		})
	}
}

func TestInstallQuiesceFailureReleasesOwnHold(t *testing.T) {
	failure := errors.New("save failed")
	session := &heldSession{err: failure}
	session.holds.Store(1)
	service := &Service{session: session, update: installFunc(func(context.Context, func(int64, int64)) error {
		t.Error("installation started after save failure")
		return nil
	})}
	if err := service.Install(t.Context(), false); !errors.Is(err, failure) {
		t.Fatalf("lost save failure: %v", err)
	}
	service.Stop()
	if session.holds.Load() != 1 || session.released.Load() != 1 || service.Snapshot().State != "installFailed" {
		t.Fatal("quiesce failure left an update hold or lost the failure state")
	}
}

func TestCanceledInstallDoesNotAcquireHold(t *testing.T) {
	session := &heldSession{}
	service := &Service{session: session}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := service.Install(ctx, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if session.holds.Load() != 0 || session.released.Load() != 0 {
		t.Fatal("canceled operation touched session maintenance")
	}
}

func TestAutomaticInstallRechecksIdleAfterQuiescing(t *testing.T) {
	checks := 0
	session := &heldSession{pending: func() bool { checks++; return checks > 1 }}
	session.holds.Store(1)
	service := &Service{session: session, update: installFunc(func(context.Context, func(int64, int64)) error {
		t.Error("automatic install ignored new durable run intent")
		return nil
	})}
	if err := service.Install(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	if checks != 2 || session.holds.Load() != 1 || session.released.Load() != 1 {
		t.Fatal("idle recheck did not release only its own maintenance hold")
	}
}

func TestAutomaticInstallWaitsForQueuedWork(t *testing.T) {
	session := &heldSession{}
	service := &Service{
		session:    session,
		queuedWork: func() bool { return true },
		update: installFunc(func(context.Context, func(int64, int64)) error {
			t.Error("automatic install began while user work was still queued")
			return nil
		}),
	}
	if err := service.Install(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	if session.released.Load() != 0 {
		t.Fatal("automatic install acquired a maintenance hold before the queue drained")
	}
}

func TestSuccessfulInstallStaysReadyToRestart(t *testing.T) {
	store, err := config.Open(t.TempDir() + "/settings.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Update(func(p *config.Settings) { p.AutoCheck, p.AutoInstall = true, true }); err != nil {
		t.Fatal(err)
	}
	service := New(store, &heldSession{}, nil)
	service.update = installFunc(func(context.Context, func(int64, int64)) error { return nil })
	if err := service.Install(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	if err := service.Check(t.Context()); err != nil {
		t.Fatal(err)
	}
	service.Scheduled(t.Context())
	if got := service.Snapshot().State; got != "restartRequired" {
		t.Fatalf("check replaced pending restart: %s", got)
	}
	if service.update != nil || service.Install(t.Context(), false) == nil {
		t.Fatal("installed update remained eligible for installation")
	}
	if !store.Snapshot().LastAttempt.IsZero() {
		t.Fatal("pending restart performed a network check")
	}
}
