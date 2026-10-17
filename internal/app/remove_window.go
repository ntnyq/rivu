package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func (w *mainWindow) confirmations(c *ui.Context) {
	w.relocationConfirmation(c)
	if w.removeOpen {
		ui.DialogBase(c, &w.removeOpen, func(_, p ui.Element) {
			p.Width(450).Padding(24).Gap(16).Children(func() {
				ui.Text(c, locales.RemovalTitle(len(w.removeIDs))).FontSize(20).Bold()
				ui.Text(c, locales.RemovalHint())
				ui.Checkbox(c, &w.trashFiles, locales.RemovalTrashFiles())
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "x", locales.CommonCancel(), false).Clicked() {
						w.removeOpen = false
					}
					if actionButton(c, "trash", locales.CommonRemove(), true).Clicked() {
						w.app.execute(session.Command{Kind: session.CommandRemove, IDs: w.removeIDs, Trash: w.trashFiles})
						w.removeOpen = false
					}
				})
			})
		})
	}
	if w.quitConfirm {
		ui.DialogBase(c, &w.quitConfirm, func(_, p ui.Element) {
			p.Width(430).Padding(24).Gap(16).Children(func() {
				ui.Text(c, locales.QuitTitle()).FontSize(20).Bold()
				ui.Text(c, locales.QuitHint())
				ui.Checkbox(c, &w.rememberQuit, locales.QuitDontAsk())
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "play", locales.QuitKeepDownloading(), false).Clicked() {
						w.quitConfirm = false
					}
					if actionButton(c, "check", locales.QuitSaveAndQuit(), true).Clicked() {
						w.quitConfirm = false
						if w.rememberQuit {
							w.saveQuitPreference()
						} else {
							w.app.shutdown()
						}
					}
				})
			})
		})
	}
}
