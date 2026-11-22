package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/internal/updater"
	"github.com/ntnyq/rivu/locales"
)

type application struct {
	work                           workQueue
	store                          *config.Store
	session                        *session.Manager
	updater                        updateService
	window                         *mainWindow
	settingsWindow                 *settingsWindow
	settingsSaving, reopenSettings bool
	add                            *addWindow
	tray                           *mygo.Tray
	cancelTrayClick                func()
	panel                          *mygo.Window
	snapshot                       session.Snapshot
	pending                        []string
	addScheduled                   bool
	post                           func(func())
	showAdd                        func([]string)
	ready, quitting, shutdownReady bool
	cancel                         context.CancelFunc
	powerDone                      <-chan struct{}
	powerResume                    func()
	completions                    completionNotifications
	errorsSeen                     map[string]string
}

func newApplication(store *config.Store) *application {
	a := &application{store: store, errorsSeen: map[string]string{}}
	// RunOnMain runs inline on the UI thread. Enter it from a worker so queue
	// delivery waits until the current native close cascade has returned.
	a.post = func(fn func()) { go mygo.RunOnMain(fn) }
	a.showAdd = a.openAdd
	return a
}

type updateService interface {
	Snapshot() updater.Snapshot
	Scheduled(context.Context)
	Check(context.Context) error
	Install(context.Context, bool) error
	Stop()
	Resume()
}

func (a *application) report(err error) {
	if err != nil && !a.shutdownReady {
		a.openWindow().report(locales.CommonOperationFailed(), err)
	}
}
func (a *application) receive(sources []string) {
	if len(sources) == 0 {
		return
	}
	a.pending = append(a.pending, sources...)
	a.scheduleAdd()
}

// Sources opened externally wait for the current draft to finish, preserving
// its file choices even when a preview or submission is still running.
func (a *application) nextAddSources() []string {
	if !a.ready || a.quitting || a.shutdownReady || a.add != nil && !a.add.closed {
		return nil
	}
	sources := a.pending
	a.pending = nil
	return sources
}

func (a *application) scheduleAdd() {
	if a.addScheduled || len(a.pending) == 0 || !a.ready || a.quitting || a.shutdownReady || a.add != nil && !a.add.closed {
		return
	}
	a.addScheduled = true
	a.post(func() {
		a.addScheduled = false
		if sources := a.nextAddSources(); len(sources) > 0 {
			a.showAdd(sources)
		}
	})
}

func (a *application) quitFailed() {
	a.quitting = false
	a.scheduleAdd()
}
func (a *application) observe() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	go a.watchFolder(ctx)
	ch, off := a.session.Watch()
	go func() {
		defer off()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		observeSession(ctx, ch, ticker.C, a.updater.Scheduled, func() {
			snap := a.session.Snapshot()
			mygo.RunOnMain(func() {
				if a.shutdownReady {
					return
				}
				previousPanelHeight := presentTray(a.snapshot).height()
				a.snapshot = snap
				if a.window != nil && !a.window.closed {
					a.window.dispatch(func() { a.window.snapshot = snap })
				}
				if a.panel != nil {
					if height := presentTray(snap).height(); height != previousPanelHeight {
						a.panel.SetSize(trayWidth, height)
					}
					a.panel.Update(func() {})
				}
				if a.settingsWindow != nil && !a.settingsWindow.closed {
					a.settingsWindow.dispatch(func() {})
				}
				a.notifications(snap)
			})
		})
	}()
	a.snapshot = a.session.Snapshot()
}

func observeSession(ctx context.Context, changes <-chan struct{}, ticks <-chan time.Time, scheduled func(context.Context), changed func()) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	defer func() { cancel(); <-done }()
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ticks:
				if !ok {
					return
				}
				if ctx.Err() != nil {
					return
				}
				scheduled(ctx)
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-changes:
			if !ok {
				return
			}
			changed()
		}
	}
}

// setSuspended is called on the UI thread. Chain workers in event order so a
// quick wake cannot overtake a suspend that is still saving on another worker.
func (a *application) setSuspended(suspended bool) {
	previous := a.powerDone
	done := make(chan struct{})
	a.powerDone = done
	go func() {
		defer close(done)
		if previous != nil {
			<-previous
		}
		if suspended && a.powerResume == nil {
			release, err := a.session.Quiesce()
			a.powerResume = release
			if err != nil {
				mygo.RunOnMain(func() { a.report(err) })
			}
		} else if !suspended && a.powerResume != nil {
			a.powerResume()
			a.powerResume = nil
		}
	}()
}
func (a *application) execute(cmd session.Command) {
	if a.quitting || a.shutdownReady {
		return
	}
	cmd.IDs = append([]string(nil), cmd.IDs...)
	cmd.FileIndices = append([]int(nil), cmd.FileIndices...)
	cmd.Tags = append([]string(nil), cmd.Tags...)
	a.work.enqueue(func() {
		err := a.session.Execute(context.Background(), cmd)
		if err != nil {
			mygo.RunOnMain(func() { a.report(err) })
		}
	})
}
func (a *application) beforeQuit(e *mygo.QuitEvent) {
	if a.shutdownReady {
		return
	}
	if e.SessionEnding {
		a.shutdownReady = true
		if a.cancel != nil {
			a.cancel()
		}
		if a.session != nil {
			if err := a.session.TryCheckpoint(); err != nil {
				slog.Error("save session at system shutdown", "error", err)
			}
		}
		return
	}
	e.PreventDefault()
	a.requestQuit(false)
}

func (a *application) requestQuit(restart bool) {
	if a.quitting || a.shutdownReady || a.session == nil {
		return
	}
	if a.store.Snapshot().ConfirmQuit && (a.work.drained() != nil || a.session.HasPendingRunIntent()) {
		w := a.openWindow()
		w.dispatch(func() {
			if !w.quitConfirm || w.quitRestart != restart {
				w.rememberQuit = false
			}
			w.quitRestart = restart
			w.quitConfirm = true
		})
		return
	}
	a.shutdown(restart, false)
}
func (a *application) shutdown(restart, remember bool) {
	if a.quitting || a.shutdownReady {
		return
	}
	a.quitting = true
	go func() {
		if done := a.work.drained(); done != nil {
			<-done
		}
		if remember {
			if err := a.store.Update(func(p *config.Settings) { p.ConfirmQuit = false }); err != nil {
				mygo.RunOnMain(func() {
					a.finishShutdown(restart, messageError{locales.SettingsSaveQuitPreference() + " · " + err.Error(), err})
				})
				return
			}
		}
		err := a.closeSession()
		mygo.RunOnMain(func() {
			if err != nil {
				err = messageError{locales.QuitSaveFailed(err), err}
			}
			a.finishShutdown(restart, err)
		})
	}()
}

func (a *application) finishShutdown(restart bool, err error) {
	if a.shutdownReady {
		return
	}
	if err != nil {
		a.report(err)
		a.quitFailed()
		return
	}
	a.quitting = false
	a.shutdownReady = true
	if a.cancel != nil {
		a.cancel()
	}
	if restart {
		mygo.App.Relaunch()
	} else {
		mygo.App.Quit()
	}
}

func (a *application) closeSession() error {
	// Retain an outer hold while updater cleanup releases its own hold, so
	// transfers cannot restart between cancellation and the final close.
	release, err := a.session.Quiesce()
	defer release()
	if err != nil {
		return err
	}
	if a.updater != nil {
		a.updater.Stop()
	}
	err = a.session.Close()
	if err != nil && a.updater != nil {
		a.updater.Resume()
	}
	return err
}
func (a *application) slowMode() {
	if a.quitting || a.shutdownReady {
		return
	}
	a.work.enqueue(func() {
		err := a.store.Update(func(s *config.Settings) { s.SlowMode = !s.SlowMode })
		if err == nil {
			err = a.session.Configure(a.store.Snapshot())
		}
		if err != nil {
			mygo.RunOnMain(func() { a.report(err) })
		}
	})
}
