package app

import (
	"context"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/updater"
	"github.com/ntnyq/rivu/locales"
)

type updateWindow struct {
	*windowBase
}

func hasUpdateDetails(state string) bool {
	switch state {
	case "available", "installing", "installFailed", "restartRequired":
		return true
	}
	return false
}

func (a *application) openUpdateWindow() {
	if a.quitting || a.shutdownReady {
		return
	}
	if w := a.updateWindow; w != nil && !w.closed {
		w.win.Show()
		w.win.Focus()
		return
	}

	w := &updateWindow{windowBase: &windowBase{app: a}}
	w.win = mygo.NewWindow(mygo.WindowOptions{
		Title: locales.UpdatesInstall(), Width: 460, Height: 310,
		DisableResize: true, DisableMaximize: true, Content: ui.View(w.view),
	})
	a.updateWindow = w
	ctx, cancel := context.WithCancel(context.Background())
	w.win.OnClosed(func() {
		cancel()
		w.close()
		a.updateWindow = nil
	})
	previous := a.updater.Snapshot()
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current := a.updater.Snapshot()
				if current != previous {
					w.dispatch(func() {})
					previous = current
				}
			}
		}
	}()
}

func (w *updateWindow) view(c *ui.Context) {
	w.render(c, w.app.updater.Snapshot())
}

func (w *updateWindow) render(c *ui.Context, s updater.Snapshot) {
	p := designTheme(c)
	c.Root().Background(p.surface)
	busy := w.task.Running() || s.State == "installing"
	title := locales.UpdatesVersionAvailable(s.Version)
	if busy {
		title = locales.UpdatesInstalling()
	} else if s.State == "restartRequired" {
		title = locales.UpdatesRestartRequired()
	} else if s.State == "installFailed" {
		title = locales.UpdatesInstallFailed()
	} else if !hasUpdateDetails(s.State) {
		title = locales.UpdatesNothingToInstall()
	}
	ui.Column(c).Fill().Padding(20, 22).Gap(13).Children(func() {
		ui.Row(c).Gap(12).Children(func() {
			ui.Image(c, appMark).Size(35, 35)
			ui.Text(c, title).FontSize(16).Bold().Grow(1).MinWidth(0)
		})
		if busy && s.Total > 0 {
			progressTrack(c, float64(s.Downloaded)/float64(s.Total), p.blue, p.track, 6)
			ui.Text(c, bytes(s.Downloaded)+" / "+bytes(s.Total)).FontSize(11).TextColor(p.muted)
		}
		if !busy && s.State != "restartRequired" && (s.Error != "" || w.failed || s.Notes != "") {
			ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
				if s.Error != "" {
					ui.Text(c, s.Error).FontSize(11).TextColor(c.Theme().Danger)
				} else if w.failed {
					ui.Text(c, w.status).FontSize(11).TextColor(c.Theme().Danger)
				} else {
					ui.Text(c, s.Notes).FontSize(11).TextColor(p.muted)
				}
			})
		} else {
			ui.Spacer(c)
		}
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			label := locales.CommonClose()
			if w.task.Running() {
				label = locales.CommonCancel()
			}
			if actionButton(c, "x", label, false).FontSize(11).Clicked() {
				w.win.Close()
			}
			if !busy && (s.State == "available" || s.State == "installFailed") &&
				actionButton(c, "download", locales.UpdatesDownloadAndInstall(), true).FontSize(11).Clicked() {
				w.run(locales.UpdatesInstall(), func(ctx context.Context) (string, error) {
					return locales.UpdatesInstalled(), w.app.updater.Install(ctx, false)
				}, nil)
			}
			if s.State == "restartRequired" && actionButton(c, "refresh", locales.UpdatesSaveAndRestart(), true).FontSize(11).Clicked() {
				w.app.requestQuit(true)
			}
		})
	})
}
