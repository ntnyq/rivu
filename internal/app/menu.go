package app

import (
	"context"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

type command struct {
	id, label, accelerator string
	run                    func()
}

func (a *application) commands() []command {
	return []command{
		{"add", locales.MenuAdd(), "CmdOrCtrl+N", func() { a.openAdd(nil) }},
		{"open", locales.MenuOpenTorrent(), "CmdOrCtrl+O", func() { a.openAdd(nil); a.add.pickTorrents() }},
		{"pauseAll", locales.MenuPauseAll(), "", func() { a.execute(session.Command{Kind: session.CommandPauseAll}) }},
		{"resumeBatch", locales.TransferResumeBatch(), "", func() { a.execute(session.Command{Kind: session.CommandResumeBatch}) }},
		{"slow", locales.MenuToggleSpeedLimit(), "", a.slowMode},
		{"main", locales.MenuOpen(), "CmdOrCtrl+1", func() { a.openWindow() }},
		{"settings", locales.MenuSettings(), "CmdOrCtrl+,", a.openSettings},
		{"reload", locales.MenuReloadAfterRepair(), "", func() {
			w := a.openWindow()
			w.run(locales.CommonReload(), func(context.Context) (string, error) {
				a.settingsMu.Lock()
				defer a.settingsMu.Unlock()
				if err := a.store.Reload(); err != nil {
					return "", err
				}
				if err := a.session.Reload(); err != nil {
					return "", err
				}
				return locales.CommonReloaded(), a.session.Configure(a.store.Snapshot())
			}, func() {
				w.settings = a.store.Snapshot()
				applyTheme(w.settings.Theme)
				a.applyLanguage()
				a.configureTray()
			})
		}},
		{"quit", locales.MenuQuit(), "CmdOrCtrl+Q", mygo.App.Quit},
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
	return mygo.NewMenu(a.menuItems(runtime.GOOS))
}
func (a *application) menuItems(goos string) []*mygo.MenuItem {
	items := []*mygo.MenuItem{
		{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{{Role: mygo.RoleAbout, Label: locales.MenuAbout()}, a.menuItem("settings"), mygo.Separator(), {Role: mygo.RoleServices, Label: locales.MenuServices()}, mygo.Separator(), {Role: mygo.RoleHide, Label: locales.MenuHide()}, {Role: mygo.RoleHideOthers, Label: locales.MenuHideOthers()}, {Role: mygo.RoleUnhide, Label: locales.MenuShowAll()}, mygo.Separator(), a.menuItem("quit")}},
		{Label: locales.MenuFile(), Submenu: []*mygo.MenuItem{a.menuItem("add"), a.menuItem("open"), a.menuItem("reload"), mygo.Separator(), {Role: mygo.RoleClose, Label: locales.MenuCloseWindow()}}},
		{Label: locales.MenuEdit(), Submenu: []*mygo.MenuItem{{Role: mygo.RoleUndo, Label: locales.MenuUndo()}, {Role: mygo.RoleRedo, Label: locales.MenuRedo()}, mygo.Separator(), {Role: mygo.RoleCut, Label: locales.MenuCut()}, {Role: mygo.RoleCopy, Label: locales.CommonCopy()}, {Role: mygo.RolePaste, Label: locales.MenuPaste()}, {Role: mygo.RoleSelectAll, Label: locales.MenuSelectAll()}}},
		{Label: locales.MenuTransfers(), Submenu: []*mygo.MenuItem{a.menuItem("pauseAll"), a.menuItem("resumeBatch"), a.menuItem("slow")}},
		{Label: locales.MenuView(), Submenu: []*mygo.MenuItem{a.menuItem("main"), {Label: locales.MenuCommandPalette(), Accelerator: "CmdOrCtrl+K", Click: func(*mygo.MenuItem, *mygo.Window) { w := a.openWindow(); w.dispatch(func() { w.paletteOpen = true }) }}, {Role: mygo.RoleToggleFullScreen, Label: locales.MenuFullScreen()}}},
		{Role: mygo.RoleWindowMenu, Label: locales.MenuWindow(), Submenu: []*mygo.MenuItem{{Role: mygo.RoleMinimize, Label: locales.MenuMinimize()}, {Role: mygo.RoleZoom, Label: locales.MenuZoom()}, {Role: mygo.RoleFront, Label: locales.MenuBringToFront()}}},
	}
	if goos != "darwin" {
		// RoleAppMenu and RoleAbout are macOS-only and hidden by MyGo elsewhere.
		items = items[1:]
		items[0].Submenu = append(items[0].Submenu, mygo.Separator(), a.menuItem("settings"), &mygo.MenuItem{Label: locales.MenuAbout(), Click: func(*mygo.MenuItem, *mygo.Window) {
			w := a.openWindow()
			go func() {
				_, err := mygo.Dialog.Message(mygo.MessageOptions{Parent: w.win, Type: mygo.MessageInfo, Title: locales.MenuAbout(), Message: appName + " " + version, Detail: website, Buttons: []string{locales.CommonClose()}})
				if err != nil {
					mygo.RunOnMain(func() { a.report(err) })
				}
			}()
		}}, mygo.Separator(), a.menuItem("quit"))
	}
	return items
}
