package app

import (
	"fmt"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func (a *application) configureTray() {
	if !a.store.Snapshot().Tray {
		if a.panel != nil {
			a.panel.Close()
		}
		if a.tray != nil {
			a.tray.Destroy()
			a.tray = nil
		}
		return
	}
	if a.tray != nil {
		return
	}
	tray, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon(), IconIsTemplate: true, ToolTip: appName})
	if err != nil {
		a.report(err)
		return
	}
	a.tray = tray
	menu := mygo.NewMenu([]*mygo.MenuItem{a.menuItem("main"), a.menuItem("add"), mygo.Separator(), a.menuItem("pauseAll"), a.menuItem("resumeBatch"), a.menuItem("slow"), a.menuItem("settings"), mygo.Separator(), a.menuItem("quit")})
	if runtime.GOOS == "linux" {
		tray.SetMenu(menu)
	} else {
		tray.OnClick(a.togglePanel)
		tray.OnRightClick(func() {
			if a.panel != nil {
				a.panel.Close()
			}
			tray.SetMenu(menu)
			defer tray.SetMenu(nil)
			tray.PopUpMenu()
		})
	}
}

// Panel dimensions are in logical pixels. The panel grows with its preview rows
// instead of reserving a full list for an empty session.
const (
	trayWidth           = 360
	trayRowHeight       = 71
	trayEmptyHeight     = 72
	trayChromeHeight    = 232
	trayAttentionHeight = 40
)

type trayPresentation struct {
	tasks             []session.Task
	active, attention int
	hasBatch          bool
}

func presentTray(snapshot session.Snapshot) trayPresentation {
	var p trayPresentation
	for _, t := range snapshot.Tasks {
		p.hasBatch = p.hasBatch || t.Batch
		if t.Error != "" {
			p.attention++
		}
		if t.Run && t.Error == "" && !t.Removing {
			p.active++
		}
		if len(p.tasks) < 4 && t.Phase != "completed" && t.Error == "" && !t.Removing {
			p.tasks = append(p.tasks, t)
		}
	}
	return p
}

func (p trayPresentation) height() int {
	height := trayChromeHeight + len(p.tasks)*trayRowHeight
	if len(p.tasks) == 0 {
		height += trayEmptyHeight
	}
	if p.attention > 0 {
		height += trayAttentionHeight
	}
	return height
}

func (a *application) togglePanel() {
	if a.panel != nil {
		a.panel.Close()
		return
	}
	a.panel = mygo.NewFlyout(mygo.FlyoutOptions{Tray: a.tray, Placement: mygo.PlacementBottom, Gap: 4, Width: trayWidth, Height: presentTray(a.snapshot).height(), Focusable: true, Shadow: true, Popover: true, Content: ui.View(a.trayView)})
	a.panel.OnClosed(func() { a.panel = nil })
}

func (a *application) closePanel() {
	if a.panel != nil {
		a.panel.Close()
	}
}

func (a *application) trayView(c *ui.Context) {
	p := designTheme(c)
	preview := presentTray(a.snapshot)
	c.Root().Background(ui.Transparent)
	// AppKit supplies the arrow and outer shape. Tint its material so desktop
	// wallpaper does not swamp the UI through fully transparent content.
	background := p.sidebar
	if c.Vibrancy() {
		background = background.Alpha(.96)
	}
	ui.Column(c).Fill().Background(background).Padding(18, 16, 11, 16).Gap(0).Children(func() {
		ui.Row(c).Height(17).Shrink(0).Children(func() {
			ui.Text(c, "Rivu").FontSize(14).FontWeight(600)
			ui.Spacer(c)
			ui.Text(c, locales.TrayActiveCount(preview.active)).FontSize(10).TextColor(p.muted)
		})
		ui.Row(c).Height(94).Shrink(0).Padding(22, 0, 19, 0).Gap(20).Children(func() {
			for i, v := range []struct {
				label, icon string
				value       float64
			}{{locales.TransferDownload(), "arrow-down", a.snapshot.DownRate}, {locales.TransferUpload(), "arrow-up", a.snapshot.UpRate}} {
				col := ui.Column(c).Width(154).Shrink(0).Gap(7)
				if i == 1 {
					col.Padding(0, 0, 0, 19).BorderWidth(0, 0, 0, 1).BorderColor(p.line)
				}
				col.Children(func() {
					ui.Row(c).Gap(4).Children(func() { glyph(c, v.icon, 12, p.muted); ui.Text(c, v.label).FontSize(10).TextColor(p.muted) })
					number, unit := rateParts(v.value)
					ui.Row(c).Gap(4).AlignItems(ui.End).Children(func() {
						ui.Text(c, number).FontSize(28).FontWeight(550)
						ui.Text(c, unit).FontSize(11).TextColor(p.muted).Padding(0, 0, 4, 0)
					})
				})
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(0).Children(func() {
				for _, t := range preview.tasks {
					ui.Row(c).Key(t.ID).Height(trayRowHeight).Shrink(0).AlignItems(ui.Start).Padding(12, 0).Gap(8).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
						body := ui.Column(c).Grow(1).MinWidth(0).Gap(8).Focusable().Label(t.Source.Name)
						body.Children(func() {
							ui.Text(c, t.Source.Name).FontSize(11).FontWeight(550).SingleLine()
							progressTrack(c, progress(t), p.blue, p.track, 3)
							ui.Row(c).Children(func() {
								ui.Text(c, fmt.Sprintf("%s · %s", taskPercent(t), phaseLabel(t.Phase))).FontSize(10).TextColor(p.muted)
								ui.Spacer(c)
								speed := "—"
								if t.Run {
									speed = rate(t.DownRate)
									if t.Phase == "seeding" {
										speed = "↑ " + rate(t.UpRate)
									}
								}
								ui.Text(c, speed).FontSize(10).TextColor(p.muted)
							})
						})
						if body.Clicked() {
							a.closePanel()
							w := a.openWindow()
							w.dispatch(func() { w.selectOnly(t.ID) })
						}
						kind, icon := session.CommandResume, "play"
						if t.Run {
							kind, icon = session.CommandPause, "pause"
						}
						if compactIconButton(c, icon, locales.TrayControl(t.ID), 23, 15, p.muted) {
							a.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
						}
					})
				}
				if len(preview.tasks) == 0 {
					message := locales.TrayEmpty()
					if len(a.snapshot.Tasks) > 0 {
						message = locales.TrayNoActive()
					}
					ui.Row(c).Height(trayEmptyHeight).Shrink(0).Justify(ui.Center).Children(func() {
						ui.Text(c, message).FontSize(12).TextColor(p.muted).TextAlign(ui.Center)
					})
				}
				if preview.attention > 0 {
					row := ui.Row(c).Height(trayAttentionHeight).Shrink(0).Padding(12, 0, 5, 0).Gap(5).Focusable().Label(locales.TrayViewAttention())
					row.Children(func() {
						glyph(c, "alert", 13, p.orange)
						ui.Text(c, locales.TrayAttentionCount(preview.attention)).FontSize(10).TextColor(p.orange)
					})
					if row.Clicked() {
						a.closePanel()
						w := a.openWindow()
						w.dispatch(func() { w.filter = 5 })
					}
				}
			})
		})
		ui.Row(c).Height(58).Shrink(0).Gap(8).Padding(15, 0, 13, 0).Children(func() {
			label, kind, icon := locales.TransferPauseAll(), session.CommandPauseAll, "pause"
			if preview.hasBatch {
				label, kind, icon = locales.TransferResumeBatch(), session.CommandResumeBatch, "play"
			}
			if actionButton(c, icon, label, false).FontSize(10).Height(30).Grow(1).Disabled(preview.active == 0 && !preview.hasBatch).Clicked() {
				a.execute(session.Command{Kind: kind})
			}
			limit := actionButton(c, "gauge", locales.TransferSpeedLimit(), false).FontSize(10).Height(30).Grow(1)
			if a.snapshot.SlowMode {
				limit.TextColor(p.blue).Background(p.soft)
			}
			if limit.Clicked() {
				a.slowMode()
			}
		})
		ui.Row(c).Height(34).Shrink(0).Gap(7).Padding(10, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
			open := ui.Row(c).Grow(1).Gap(4).Focusable().Label(locales.TrayOpenWindow())
			open.Children(func() {
				ui.Text(c, locales.TrayOpenWindow()).FontSize(11).TextColor(p.blue)
				ui.Text(c, shortcutHint(runtime.GOOS, "1")).FontSize(10).TextColor(p.muted)
			})
			if open.Clicked() {
				a.closePanel()
				a.openWindow()
			}
			if compactIconButton(c, "plus", locales.TransferAdd(), 25, 15, p.muted) {
				a.closePanel()
				a.openAdd(nil)
			}
			if compactIconButton(c, "settings", locales.SettingsLabel(), 25, 15, p.muted) {
				a.closePanel()
				a.openSettings()
			}
		})
	})
}
func rateParts(v float64) (string, string) {
	if v >= 1000000 {
		return fmt.Sprintf("%.1f", v/1000000), "MB/s"
	}
	return fmt.Sprintf("%.0f", v/1000), "KB/s"
}
