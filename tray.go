package main

import (
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"runtime"
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
func (a *application) togglePanel() {
	if a.panel != nil {
		a.panel.Close()
		return
	}
	a.panel = mygo.NewFlyout(mygo.FlyoutOptions{Tray: a.tray, Placement: mygo.PlacementBottom, Gap: 4, Width: 360, Height: 490, Focusable: true, Shadow: true, Popover: true, Content: ui.View(a.trayView)})
	a.panel.OnClosed(func() { a.panel = nil })
}
func (a *application) trayView(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(ui.Transparent)
	box := ui.Column(c).Fill().Padding(18, 16, 11, 16).Gap(0)
	if !c.Vibrancy() {
		box.Background(p.sidebar)
	}
	active := 0
	for _, t := range a.snapshot.Tasks {
		if t.Run && t.Error == "" && !t.Removing {
			active++
		}
	}
	box.Children(func() {
		ui.Row(c).Children(func() {
			ui.Text(c, "Rivu").FontSize(14).Bold()
			ui.Spacer(c)
			ui.Text(c, fmt.Sprintf("%d 个活跃任务", active)).FontSize(10).TextColor(p.muted)
		})
		ui.Row(c).Padding(22, 0, 19, 0).Gap(20).Children(func() {
			for i, v := range []struct {
				label, icon string
				value       float64
			}{{"下载速度", "arrow-down", a.snapshot.DownRate}, {"上传速度", "arrow-up", a.snapshot.UpRate}} {
				col := ui.Column(c).Grow(1).Gap(7)
				if i == 1 {
					col.Padding(0, 0, 0, 19).BorderWidth(0, 0, 0, 1).BorderColor(p.line)
				}
				col.Children(func() {
					ui.Row(c).Gap(4).Children(func() { glyph(c, v.icon, 12, p.muted); ui.Text(c, v.label).FontSize(10).TextColor(p.muted) })
					number, unit := rateParts(v.value)
					ui.Row(c).Gap(4).AlignItems(ui.End).Children(func() {
						ui.Text(c, number).FontSize(28)
						ui.Text(c, unit).FontSize(11).TextColor(p.muted).Padding(0, 0, 4, 0)
					})
				})
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(0).Children(func() {
				count, attention := 0, 0
				for _, t := range a.snapshot.Tasks {
					if t.Error != "" {
						attention++
					}
					if count >= 4 || t.Phase == "completed" || t.Error != "" || (active > 0 && !t.Run) {
						continue
					}
					count++
					ui.Row(c).Key(t.ID).Padding(12, 0).Gap(8).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
						body := ui.Column(c).Grow(1).MinWidth(0).Gap(8).Focusable().Label(t.Source.Name)
						body.Children(func() {
							ui.Text(c, t.Source.Name).FontSize(11).Bold().SingleLine()
							progressTrack(c, progress(t), taskColor(c, t.Phase), p.track, 3)
							ui.Row(c).Children(func() {
								ui.Text(c, phases[t.Phase]).FontSize(10).TextColor(p.muted)
								ui.Spacer(c)
								ui.Text(c, fmt.Sprintf("%.1f%%", progress(t)*100)).FontSize(10).TextColor(p.muted)
							})
						})
						if body.Clicked() {
							w := a.openWindow()
							w.dispatch(func() { w.selectOnly(t.ID) })
							if a.panel != nil {
								a.panel.Close()
							}
						}
						kind, icon := "resume", "play"
						if t.Run && t.Error == "" && !t.Removing {
							kind, icon = "pause", "pause"
						}
						if iconButton(c, icon, "控制 "+t.ID, true) {
							a.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
						}
					})
				}
				if count == 0 {
					ui.Text(c, "暂无下载，添加一个种子开始吧。").FontSize(12).TextColor(p.muted).Padding(28, 0)
				}
				if attention > 0 {
					row := ui.Row(c).Padding(12, 0, 5, 0).Gap(5).Focusable().Label("查看需要处理的任务")
					row.Children(func() {
						glyph(c, "alert", 13, p.orange)
						ui.Text(c, fmt.Sprintf("%d 个任务需要处理", attention)).FontSize(10).TextColor(p.orange)
					})
					if row.Clicked() {
						w := a.openWindow()
						w.dispatch(func() { w.filter = 5 })
						if a.panel != nil {
							a.panel.Close()
						}
					}
				}
			})
		})
		ui.Row(c).Gap(8).Padding(15, 0, 13, 0).Children(func() {
			if ui.Button(c, "＋ 添加下载").FontSize(10).Grow(1).Clicked() {
				a.openAdd(nil)
			}
			hasBatch := false
			for _, t := range a.snapshot.Tasks {
				hasBatch = hasBatch || t.Batch
			}
			label, kind := "暂停全部", "pauseAll"
			if hasBatch {
				label, kind = "恢复本次暂停", "resumeBatch"
			}
			if ui.Button(c, label).FontSize(10).Grow(1).Clicked() {
				a.execute(session.Command{Kind: kind})
			}
		})
		ui.Row(c).Gap(7).Padding(10, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
			open := ui.Text(c, "打开主窗口").FontSize(11).TextColor(p.blue).Grow(1).Focusable()
			if open.Clicked() {
				a.openWindow()
				if a.panel != nil {
					a.panel.Close()
				}
			}
			if iconButton(c, "gauge", "限速模式", true) {
				a.slowMode()
			}
			if iconButton(c, "settings", "设置", true) {
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
