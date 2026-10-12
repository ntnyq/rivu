package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/internal/updater"
)

type application struct {
	settingsMu                               sync.Mutex
	store                                    *config.Store
	session                                  *session.Manager
	updater                                  *updater.Service
	window, settingsWindow                   *window
	add                                      *addWindow
	tray                                     *mygo.Tray
	panel                                    *mygo.Window
	snapshot                                 session.Snapshot
	pending                                  []string
	ready, quitting, shutdownReady, relaunch bool
	cancel                                   context.CancelFunc
	notifying                                map[string]bool
	errorsSeen                               map[string]string
}

func (a *application) report(err error) {
	if err != nil {
		a.openWindow().report("操作失败", err)
	}
}
func (a *application) receive(sources []string) {
	if len(sources) == 0 {
		return
	}
	if !a.ready {
		a.pending = append(a.pending, sources...)
		return
	}
	a.openAdd(sources)
}
func (a *application) observe() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	ch, off := a.session.Watch()
	go func() {
		defer off()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
				snap := a.session.Snapshot()
				mygo.RunOnMain(func() {
					a.snapshot = snap
					if a.window != nil && !a.window.closed {
						a.window.dispatch(func() { a.window.snapshot = snap })
					}
					if a.panel != nil {
						a.panel.Update(func() {})
					}
					if a.settingsWindow != nil && !a.settingsWindow.closed {
						a.settingsWindow.dispatch(func() {})
					}
					a.notifications(snap)
				})
			case <-ticker.C:
				a.updater.Scheduled(ctx)
			}
		}
	}()
	a.snapshot = a.session.Snapshot()
}
func (a *application) execute(cmd session.Command) {
	cmd.IDs = append([]string(nil), cmd.IDs...)
	go func() {
		err := a.session.Execute(context.Background(), cmd)
		if err != nil {
			mygo.RunOnMain(func() { a.report(err) })
		}
	}()
}
func (a *application) notifications(s session.Snapshot) {
	settings := a.store.Snapshot()
	for _, t := range s.Tasks {
		if t.Epoch > t.Notified && !a.notifying[t.ID] {
			a.notifying[t.ID] = true
			go func() {
				if settings.NotifyComplete {
					n := mygo.NewNotification(mygo.NotificationOptions{ID: fmt.Sprintf("rivu-%s-%d", t.ID, t.Epoch), Title: "下载完成", Body: t.Source.Name})
					n.OnClick(func() { w := a.openWindow(); w.dispatch(func() { w.selectOnly(t.ID) }) })
					if err := n.Show(); err != nil {
						mygo.RunOnMain(func() { a.report(fmt.Errorf("通知未送达: %w", err)) })
					}
				}
				err := a.session.Execute(context.Background(), session.Command{Kind: "notified", IDs: []string{t.ID}})
				mygo.RunOnMain(func() {
					delete(a.notifying, t.ID)
					if err != nil {
						a.report(err)
					}
				})
			}()
		}
		if settings.NotifyError && t.Error != "" && a.errorsSeen[t.ID] != t.Error {
			a.errorsSeen[t.ID] = t.Error
			go func() {
				n := mygo.NewNotification(mygo.NotificationOptions{ID: "rivu-error-" + t.ID, Title: "下载需要处理", Body: t.Source.Name + "：" + t.Error})
				n.OnClick(func() { w := a.openWindow(); w.dispatch(func() { w.filter = 5; w.selectOnly(t.ID) }) })
				_ = n.Show()
			}()
		}
	}
}
func (a *application) beforeQuit(e *mygo.QuitEvent) {
	if a.shutdownReady {
		return
	}
	e.PreventDefault()
	if a.quitting {
		return
	}
	if a.session == nil {
		return
	}
	if a.store.Snapshot().ConfirmQuit && !a.session.Idle() {
		w := a.openWindow()
		w.dispatch(func() { w.quitConfirm = true })
		return
	}
	a.shutdown()
}
func (a *application) shutdown() {
	if a.quitting {
		return
	}
	a.quitting = true
	go func() {
		err := a.session.Close()
		mygo.RunOnMain(func() {
			a.quitting = false
			if err != nil {
				a.report(fmt.Errorf("无法保存会话，已取消退出: %w", err))
				return
			}
			a.shutdownReady = true
			if a.relaunch {
				mygo.App.Relaunch()
			} else {
				mygo.App.Quit()
			}
		})
	}()
}
func (a *application) slowMode() {
	go func() {
		a.settingsMu.Lock()
		defer a.settingsMu.Unlock()
		err := a.store.Update(func(s *config.Settings) { s.SlowMode = !s.SlowMode })
		if err == nil {
			err = a.session.Configure(a.store.Snapshot())
		}
		if err != nil {
			mygo.RunOnMain(func() { a.report(err) })
		}
	}()
}
