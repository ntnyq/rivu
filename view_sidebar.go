package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
)

func (w *mainWindow) sidebar(c *ui.Context, p palette) {
	ui.Column(c).Width(191).Shrink(0).Background(p.sidebar).BorderWidth(0, 1, 0, 0).BorderColor(p.line).Children(func() {
		ui.Row(c).Height(58).Shrink(0).Padding(0, 18).Children(func() {
			ui.Spacer(c).DragWindow()
			if iconButton(c, "sidebar", tr("sidebar.collapse"), true) {
				w.sidebarHidden = true
			}
		})
		ui.Row(c).Height(41).Shrink(0).Padding(10, 14, 0, 14).Gap(7).Children(func() { ui.Image(c, appMark).Size(38, 38); ui.Text(c, "Rivu").FontSize(20).Bold() })
		ui.Column(c).Padding(29, 12, 0, 12).Gap(3).Children(func() {
			ui.Text(c, tr("sidebar.library")).FontSize(11).Bold().TextColor(p.muted).Padding(0, 10, 8, 10)
			icons := []string{"grid", "arrow-down", "arrow-up", "check", "pause", "alert"}
			for i, label := range filterLabels() {
				count := 0
				for _, task := range w.snapshot.Tasks {
					if matches(task, i) {
						count++
					}
				}
				selected := w.filter == i
				bg, ink, iconColor := ui.Transparent, p.text, p.blue
				if i == 5 {
					iconColor = p.orange
				}
				if selected {
					bg = p.blue
					if c.Theme().Dark {
						bg = ui.Hex("#176dcf")
					}
					ink = ui.Hex("#ffffff")
					iconColor = ink
				}
				row := ui.Row(c).Key(label).Height(35).FillWidth().Padding(0, 11).Gap(10).Radius(7).Background(bg).Focusable().Label(label)
				if !selected && row.Hovered() {
					row.Background(p.track)
				}
				row.Children(func() {
					glyph(c, icons[i], 17, iconColor)
					ui.Text(c, label).FontSize(12).TextColor(ink)
					ui.Spacer(c)
					color := p.muted
					if selected {
						color = ink
					}
					ui.Text(c, fmt.Sprint(count)).FontSize(11).TextColor(color)
				})
				if row.Clicked() || row.Shortcut(0, ui.KeyEnter) {
					w.filter = i
				}
			}
		})
		ui.Spacer(c)
		ui.Column(c).Padding(20, 18, 18, 18).Gap(12).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				glyph(c, "drive", 17, p.muted)
				ui.Column(c).Gap(5).MinWidth(0).Children(func() {
					ui.Text(c, tr("sidebar.downloadLocation")).FontSize(11)
					ui.Text(c, compactPath(w.settings.Destination)).FontSize(9).SingleLine().TextColor(p.muted)
				})
			})
			if w.diskTotal > 0 {
				ui.Text(c, fmt.Sprintf(tr("sidebar.availableSpace"), bytes(w.diskFree))).FontSize(9).TextColor(p.muted)
				progressTrack(c, 1-float64(w.diskFree)/float64(w.diskTotal), p.muted.Alpha(.5), p.track, 3)
			} else {
				ui.Box(c).Height(3).Background(p.track).Radius(3)
			}
			settings := navigationButton(c, tr("settings.label"), false, p).Margin(0, -6).Children(func() {
				glyph(c, "settings", 17, p.text)
				ui.Text(c, tr("settings.label")).FontSize(12)
				ui.Spacer(c)
				ui.Text(c, "⌘,").FontSize(11).TextColor(p.muted)
			})
			if settings.Clicked() && w.app != nil {
				w.app.openSettings()
			}
		})
	})
}
