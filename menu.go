package main

import "github.com/egoist/mygo"
import "context"
import "github.com/ntnyq/rivu/internal/session"

type command struct {
	id, label, accelerator string
	run                    func()
}

func (a *application) commands() []command {
	return []command{
		{"add", tr("menu.add"), "CmdOrCtrl+N", func() { a.openAdd(nil) }},
		{"open", tr("menu.openTorrent"), "CmdOrCtrl+O", func() { a.openAdd(nil); a.add.pickTorrents() }},
		{"pauseAll", tr("menu.pauseAll"), "", func() { a.execute(session.Command{Kind: session.CommandPauseAll}) }},
		{"resumeBatch", tr("transfer.resumeBatch"), "", func() { a.execute(session.Command{Kind: session.CommandResumeBatch}) }},
		{"slow", tr("menu.toggleSpeedLimit"), "", a.slowMode},
		{"main", tr("menu.open"), "CmdOrCtrl+1", func() { a.openWindow() }},
		{"settings", tr("menu.settings"), "CmdOrCtrl+,", a.openSettings},
		{"reload", tr("menu.reloadAfterRepair"), "", func() {
			w := a.openWindow()
			w.run(tr("common.reload"), func(context.Context) (string, error) {
				a.settingsMu.Lock()
				defer a.settingsMu.Unlock()
				if err := a.store.Reload(); err != nil {
					return "", err
				}
				if err := a.session.Reload(); err != nil {
					return "", err
				}
				return tr("common.reloaded"), a.session.Configure(a.store.Snapshot())
			}, func() {
				w.settings = a.store.Snapshot()
				applyTheme(w.settings.Theme)
				a.applyLanguage()
			})
		}},
		{"quit", tr("menu.quit"), "CmdOrCtrl+Q", mygo.App.Quit},
	}
}
func (a *application) command(id string) {
	for _, c := range a.commands() {
		if c.id == id {
			c.run()
			return
		}
	}
}
func (a *application) menuItem(id string) *mygo.MenuItem {
	for _, c := range a.commands() {
		if c.id == id {
			return &mygo.MenuItem{Label: c.label, Accelerator: c.accelerator, Click: func(*mygo.MenuItem, *mygo.Window) { a.command(id) }}
		}
	}
	return mygo.Separator()
}
func (a *application) menu() *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{{Role: mygo.RoleAbout, Label: tr("menu.about")}, a.menuItem("settings"), mygo.Separator(), {Role: mygo.RoleServices, Label: tr("menu.services")}, mygo.Separator(), {Role: mygo.RoleHide, Label: tr("menu.hide")}, {Role: mygo.RoleHideOthers, Label: tr("menu.hideOthers")}, {Role: mygo.RoleUnhide, Label: tr("menu.showAll")}, mygo.Separator(), a.menuItem("quit")}},
		{Label: tr("menu.file"), Submenu: []*mygo.MenuItem{a.menuItem("add"), a.menuItem("open"), a.menuItem("reload"), mygo.Separator(), {Role: mygo.RoleClose, Label: tr("menu.closeWindow")}}},
		{Label: tr("menu.edit"), Submenu: []*mygo.MenuItem{{Role: mygo.RoleUndo, Label: tr("menu.undo")}, {Role: mygo.RoleRedo, Label: tr("menu.redo")}, mygo.Separator(), {Role: mygo.RoleCut, Label: tr("menu.cut")}, {Role: mygo.RoleCopy, Label: tr("common.copy")}, {Role: mygo.RolePaste, Label: tr("menu.paste")}, {Role: mygo.RoleSelectAll, Label: tr("menu.selectAll")}}},
		{Label: tr("menu.transfers"), Submenu: []*mygo.MenuItem{a.menuItem("pauseAll"), a.menuItem("resumeBatch"), a.menuItem("slow")}},
		{Label: tr("menu.view"), Submenu: []*mygo.MenuItem{a.menuItem("main"), {Label: tr("menu.commandPalette"), Accelerator: "CmdOrCtrl+K", Click: func(*mygo.MenuItem, *mygo.Window) { w := a.openWindow(); w.dispatch(func() { w.paletteOpen = true }) }}, {Role: mygo.RoleToggleFullScreen, Label: tr("menu.fullScreen")}}},
		{Role: mygo.RoleWindowMenu, Label: tr("menu.window"), Submenu: []*mygo.MenuItem{{Role: mygo.RoleMinimize, Label: tr("menu.minimize")}, {Role: mygo.RoleZoom, Label: tr("menu.zoom")}, {Role: mygo.RoleFront, Label: tr("menu.bringToFront")}}},
	})
}
