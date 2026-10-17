package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/locales"
)

func (w *mainWindow) statusBar(c *ui.Context, p palette) {
	ui.Row(c).Height(34).Shrink(0).Padding(0, 15).Gap(22).Background(p.alt).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
		ui.Row(c).Gap(5).Children(func() { glyph(c, "arrow-down", 12, p.blue); ui.Text(c, rate(w.snapshot.DownRate)).FontSize(10) })
		ui.Row(c).Gap(5).Children(func() {
			glyph(c, "arrow-up", 12, p.muted)
			ui.Text(c, rate(w.snapshot.UpRate)).FontSize(10).TextColor(p.muted)
		})
		ui.Spacer(c)
		message := w.status
		if message == locales.CommonReady() {
			message = ""
		}
		if w.snapshot.Error != "" {
			message = problemText(w.snapshot.Issue, w.snapshot.Error)
		}
		col := p.muted
		if w.failed || w.snapshot.Error != "" {
			col = c.Theme().Danger
		}
		if w.snapshot.Error != "" && ui.Button(c, locales.CommonReload()).Clicked() && w.app != nil {
			w.app.command("reload")
		}
		ui.Text(c, message).SingleLine().MinWidth(0).FontSize(10).TextColor(col)
		color := p.muted
		if w.snapshot.SlowMode {
			color = p.blue
		}
		limit := ui.Row(c).Gap(5).Focusable().Label(locales.TransferSpeedLimit())
		limit.Children(func() {
			glyph(c, "gauge", 13, color)
			ui.Text(c, locales.TransferSpeedLimit()).FontSize(10).TextColor(color)
		})
		if limit.Clicked() && w.app != nil {
			w.app.slowMode()
		}
		peers := 0
		for _, task := range w.snapshot.Tasks {
			peers += task.Peers
		}
		network := locales.TrackerWaiting()
		dot := p.muted
		if peers > 0 {
			network = locales.TransferPeersConnected()
			dot = p.green
		}
		ui.Row(c).Gap(5).Children(func() {
			ui.Box(c).Size(5, 5).Radius(3).Background(dot)
			ui.Text(c, network).FontSize(10).TextColor(p.muted)
		})
	})
}
