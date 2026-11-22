package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
	"github.com/ntnyq/rivu/internal/updater"
)

func TestCopiedMagnetRetainsExplicitPeer(t *testing.T) {
	raw := "magnet:?xt=urn:btih:" + strings.Repeat("1", 40) + "&x.pe=127.0.0.1%3A6881"
	parser := torrentengine.Parser{}
	source, err := parser.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	copied := magnet(session.Task{Record: session.Record{Source: source}})
	if _, err := parser.Parse(copied); err != nil {
		t.Fatalf("copied magnet no longer bootstraps: %v", err)
	}
	if !strings.Contains(copied, "x.pe=127.0.0.1%3A6881") {
		t.Fatalf("explicit peer missing from %q", copied)
	}
}

func TestScheduledUpdateDoesNotBlockSessionObservations(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	changes := make(chan struct{}, 1)
	ticks := make(chan time.Time, 1)
	started, stopped, published, done := make(chan struct{}), make(chan struct{}), make(chan struct{}, 1), make(chan struct{})
	go func() {
		defer close(done)
		observeSession(ctx, changes, ticks, func(ctx context.Context) {
			close(started)
			<-ctx.Done() // Simulate a slow check or archive download.
			close(stopped)
		}, func() { published <- struct{}{} })
	}()
	ticks <- time.Now()
	waitSignal(t, started)
	changes <- struct{}{}
	waitSignal(t, published)
	cancel()
	waitSignal(t, stopped)
	waitSignal(t, done)
}

func TestObserverWaitsForScheduledCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ticks := make(chan time.Time, 1)
	started, canceled, finish, done := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	finishCleanup := sync.OnceFunc(func() { close(finish) })
	defer finishCleanup()
	go func() {
		defer close(done)
		observeSession(ctx, make(chan struct{}), ticks, func(ctx context.Context) {
			close(started)
			<-ctx.Done()
			close(canceled)
			<-finish // Extraction/replacement may still be finishing after cancel.
		}, func() {})
	}()
	ticks <- time.Now()
	waitSignal(t, started)
	cancel()
	waitSignal(t, canceled)
	select {
	case <-done:
		t.Fatal("observer returned before update cleanup finished")
	case <-time.After(30 * time.Millisecond):
	}
	finishCleanup()
	waitSignal(t, done)
}

func waitSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for worker")
	}
}

type lifecycleEngine struct{}

func (lifecycleEngine) Attach(session.Record) error        { return nil }
func (lifecycleEngine) Start(string, []session.File) error { return nil }
func (lifecycleEngine) Stop(string) error                  { return nil }
func (lifecycleEngine) Observe(string) session.Observation {
	return session.Observation{Phase: session.PhaseMetadata}
}
func (lifecycleEngine) Limits(int64, int64) {}
func (lifecycleEngine) Close() error        { return nil }

func TestPowerEventsRetainUpdateHoldAndArrivalOrder(t *testing.T) {
	m, err := session.Open(t.TempDir(), lifecycleEngine{}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := m.Import(t.Context(), session.Import{Source: session.Source{Raw: "magnet:fixture", Identity: session.Identity{V1: "fixture"}}, Destination: t.TempDir(), Run: true}); err != nil {
		t.Fatal(err)
	}
	update, err := m.Quiesce()
	if err != nil {
		t.Fatal(err)
	}
	defer update()
	a := &application{session: m}
	// Hold the first power worker to force all events to queue before it runs.
	gate := make(chan struct{})
	a.powerDone = gate
	a.setSuspended(true)
	a.setSuspended(true)
	a.setSuspended(false)
	close(gate)
	waitSignal(t, a.powerDone)
	if !m.Snapshot().Quiesced {
		t.Fatal("wake released the update's hold")
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if !m.Snapshot().Quiesced {
		t.Fatal("reload released the update's hold")
	}
	update()
	if m.Snapshot().Quiesced {
		t.Fatal("out-of-order or duplicate power events left a stale hold")
	}
	a.setSuspended(true)
	waitSignal(t, a.powerDone)
	if !m.Snapshot().Quiesced {
		t.Fatal("suspend did not acquire a hold")
	}
	a.setSuspended(false)
	waitSignal(t, a.powerDone)
	if m.Snapshot().Quiesced {
		t.Fatal("wake did not release its hold")
	}
}

func TestExternalSourcesWaitForCurrentDraft(t *testing.T) {
	for _, job := range []string{"review", "preview", "submit"} {
		t.Run(job, func(t *testing.T) {
			d := &addWindow{windowBase: &windowBase{}, input: "A", preparedInput: "A",
				items: []importItem{{source: session.Source{Raw: "A", Files: []session.File{{Path: "chosen", Priority: 2}}}}}}
			if job != "review" {
				d.task.Start(func(func()) {}, func(ctx context.Context, _ func(float64)) (string, error) {
					<-ctx.Done()
					return "", ctx.Err()
				}, func(float64) {}, func(string, error) {})
			}
			t.Cleanup(d.close)
			a := &application{ready: true, add: d, post: func(func()) {}}
			for _, source := range []string{"B", "C"} {
				a.receive([]string{source})
			}
			if !slices.Equal(a.pending, []string{"B", "C"}) {
				t.Fatalf("lost or reordered requests: %v", a.pending)
			}
			if d.input != "A" || d.preparedInput != "A" || len(d.items) != 1 || d.items[0].source.Files[0].Priority != 2 {
				t.Fatal("external sources changed the current draft or file selection")
			}
			// The actual receiver must follow the same queue path without opening
			// a native window or replacing the current draft.
			a.receive([]string{"D"})
			if next := a.nextAddSources(); len(next) != 0 {
				t.Fatalf("opened sources before draft closed: %v", next)
			}
			d.close() // Both successful submission and Cancel close the draft.
			if next := a.nextAddSources(); !slices.Equal(next, []string{"B", "C", "D"}) {
				t.Fatalf("pending batch not delivered after close: %v", next)
			}
			if next := a.nextAddSources(); len(next) != 0 {
				t.Fatalf("pending sources delivered twice: %v", next)
			}
		})
	}
}

func TestPendingSourcesDoNotReopenDuringQuit(t *testing.T) {
	for _, shutdownReady := range []bool{false, true} {
		a := &application{ready: true, quitting: !shutdownReady, shutdownReady: shutdownReady, pending: []string{"B"}}
		if next := a.nextAddSources(); len(next) != 0 {
			t.Fatalf("opened sources during shutdown: %v", next)
		}
	}
}

func TestExternalSourcesStayQueuedDuringQuit(t *testing.T) {
	a := &application{ready: true, quitting: true, pending: []string{"B"}, post: func(func()) {}}
	a.receive([]string{"C"})
	a.quitting = false
	if next := a.nextAddSources(); !slices.Equal(next, []string{"B", "C"}) {
		t.Fatalf("FIFO lost after failed quit: %v", next)
	}
}

func TestSourcesQueueBeforeStartup(t *testing.T) {
	a := &application{post: func(func()) {}}
	a.receive([]string{"A"})
	a.receive([]string{"B"})
	a.ready = true
	if next := a.nextAddSources(); !slices.Equal(next, []string{"A", "B"}) {
		t.Fatalf("startup sources not retained: %v", next)
	}
}

type addQueueFixture struct {
	a       *application
	posted  []func()
	batches [][]string
}

func newAddQueueFixture() *addQueueFixture {
	f := &addQueueFixture{a: newApplication(nil)}
	f.a.ready = true
	f.a.post = func(fn func()) { f.posted = append(f.posted, fn) }
	f.a.showAdd = func(sources []string) {
		f.batches = append(f.batches, slices.Clone(sources))
		f.a.add = &addWindow{windowBase: &windowBase{}}
	}
	return f
}

func (f *addQueueFixture) drain() {
	for len(f.posted) > 0 {
		fn := f.posted[0]
		f.posted = f.posted[1:]
		fn()
	}
}

func TestAddQueueWaitsForParentCloseCascade(t *testing.T) {
	f := newAddQueueFixture()
	parent := &mainWindow{windowBase: &windowBase{}}
	f.a.window = parent
	draft := &addWindow{windowBase: &windowBase{}}
	f.a.add = draft
	f.a.receive([]string{"B"})
	f.drain()            // A is still open; B must remain queued.
	f.a.addClosed(draft) // MyGo closes children before the parent's OnClosed.
	if len(f.batches) != 0 || len(f.posted) != 1 {
		t.Fatal("next modal opened synchronously inside the child close callback")
	}
	f.a.receive([]string{"C"}) // Arrives before deferred delivery of B.
	if len(f.batches) != 0 || len(f.posted) != 1 {
		t.Fatal("new source bypassed deferred FIFO delivery")
	}
	show := f.a.showAdd
	f.a.showAdd = func(sources []string) {
		if !parent.closed {
			t.Error("reused a parent whose native close callback is not finished")
		}
		show(sources)
	}
	parent.close()
	f.drain()
	if len(f.batches) != 1 || !slices.Equal(f.batches[0], []string{"B", "C"}) {
		t.Fatalf("lost or reordered pending batch: %v", f.batches)
	}
	f.a.addClosed(draft) // A late duplicate close cannot reopen the batch.
	f.drain()
	if len(f.batches) != 1 {
		t.Fatal("batch delivered more than once")
	}
}

func TestFailedQuitResumesPendingFIFO(t *testing.T) {
	f := newAddQueueFixture()
	draft := &addWindow{windowBase: &windowBase{}}
	f.a.add = draft
	f.a.receive([]string{"B"})
	f.drain()
	f.a.quitting = true
	f.a.addClosed(draft)
	f.drain()
	f.a.receive([]string{"C"})
	f.drain()
	if len(f.batches) != 0 {
		t.Fatal("opened a batch while quitting")
	}
	f.a.quitFailed()
	f.a.receive([]string{"D"})
	f.drain()
	if len(f.batches) != 1 || !slices.Equal(f.batches[0], []string{"B", "C", "D"}) {
		t.Fatalf("failed quit stranded or reordered requests: %v", f.batches)
	}
}

func TestSuccessfulQuitDoesNotDeliverDeferredBatch(t *testing.T) {
	f := newAddQueueFixture()
	f.a.receive([]string{"B"})
	f.a.shutdownReady = true
	f.drain()
	f.a.receive([]string{"C"})
	f.drain()
	if len(f.batches) != 0 {
		t.Fatal("deferred callback opened a window after shutdown")
	}
}

type shutdownUpdater struct{ stop, resume func() }

func (u shutdownUpdater) Stop()                             { u.stop() }
func (u shutdownUpdater) Resume()                           { u.resume() }
func (shutdownUpdater) Snapshot() updater.Snapshot          { return updater.Snapshot{} }
func (shutdownUpdater) Scheduled(context.Context)           {}
func (shutdownUpdater) Check(context.Context) error         { return nil }
func (shutdownUpdater) Install(context.Context, bool) error { return nil }

type closingEngine struct {
	lifecycleEngine
	close func() error
}

func (e closingEngine) Close() error { return e.close() }

func TestQuitWaitsForUpdaterBeforeClosingSession(t *testing.T) {
	closed := make(chan struct{}, 1)
	m, err := session.Open(t.TempDir(), closingEngine{close: func() error { closed <- struct{}{}; return nil }}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Import(t.Context(), session.Import{Source: session.Source{Identity: session.Identity{V1: "fixture"}}, Destination: t.TempDir(), Run: true}); err != nil {
		t.Fatal(err)
	}
	updateHold, err := m.Quiesce()
	if err != nil {
		t.Fatal(err)
	}
	started, finish := make(chan struct{}), make(chan struct{})
	finishCleanup := sync.OnceFunc(func() { close(finish) })
	defer finishCleanup()
	a := &application{session: m, updater: shutdownUpdater{
		stop: func() {
			close(started)
			<-finish
			updateHold()
			if !m.Snapshot().Quiesced {
				t.Error("updater cleanup released the quit hold and restarted admission")
			}
		},
		resume: func() { t.Error("updates resumed after successful close") },
	}}
	done := make(chan error, 1)
	go func() { done <- a.closeSession() }()
	waitSignal(t, started)
	select {
	case <-closed:
		t.Fatal("session closed while update replacement was still running")
	default:
	}
	finishCleanup()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	waitSignal(t, closed)
}

func TestFailedCloseResumesUpdater(t *testing.T) {
	failure := errors.New("engine close failed")
	closeErr := failure
	m, err := session.Open(t.TempDir(), closingEngine{close: func() error { return closeErr }}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErr = nil; _ = m.Close() })
	stopped, resumed := false, false
	a := &application{session: m, updater: shutdownUpdater{
		stop: func() { stopped = true },
		resume: func() {
			if !stopped {
				t.Error("resume preceded stop")
			}
			resumed = true
		},
	}}
	if err := a.closeSession(); !errors.Is(err, failure) {
		t.Fatalf("lost shutdown failure: %v", err)
	}
	if !stopped || !resumed || !m.Snapshot().Quiesced {
		t.Fatal("failed shutdown lost recovery protection or left updater stopped")
	}
}
