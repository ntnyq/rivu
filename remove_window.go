package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
)

func (w *mainWindow) confirmations(c *ui.Context) {
	w.relocationConfirmation(c)
	if w.removeOpen {
		ui.DialogBase(c, &w.removeOpen, func(_, p ui.Element) {
			p.Width(450).Padding(24).Gap(16).Children(func() {
				ui.Text(c, fmt.Sprintf(tr("removal.title"), len(w.removeIDs))).FontSize(20).Bold()
				ui.Text(c, tr("removal.hint"))
				ui.Checkbox(c, &w.trashFiles, tr("removal.trashFiles"))
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "x", tr("common.cancel"), false).Clicked() {
						w.removeOpen = false
					}
					if actionButton(c, "trash", tr("common.remove"), true).Clicked() {
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
				ui.Text(c, tr("quit.title")).FontSize(20).Bold()
				ui.Text(c, tr("quit.hint"))
				ui.Checkbox(c, &w.rememberQuit, tr("quit.dontAsk"))
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "play", tr("quit.keepDownloading"), false).Clicked() {
						w.quitConfirm = false
					}
					if actionButton(c, "check", tr("quit.saveAndQuit"), true).Clicked() {
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
