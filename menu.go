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
		{"add", "添加下载…", "CmdOrCtrl+N", func() { a.openAdd(nil) }},
		{"open", "打开种子文件…", "CmdOrCtrl+O", func() { a.openAdd(nil); a.add.pickTorrents() }},
		{"pauseAll", "暂停全部", "", func() { a.execute(session.Command{Kind: "pauseAll"}) }},
		{"resumeBatch", "恢复本次暂停", "", func() { a.execute(session.Command{Kind: "resumeBatch"}) }},
		{"slow", "切换限速模式", "", a.slowMode},
		{"main", "打开 Rivu", "CmdOrCtrl+1", func() { a.openWindow() }},
		{"settings", "设置…", "CmdOrCtrl+,", a.openSettings},
		{"reload", "修复后重新读取会话与设置", "", func() {
			w := a.openWindow()
			w.run("重新读取", func(context.Context) (string, error) {
				a.settingsMu.Lock()
				defer a.settingsMu.Unlock()
				if err := a.store.Reload(); err != nil {
					return "", err
				}
				if err := a.session.Reload(); err != nil {
					return "", err
				}
				return "已重新读取", a.session.Configure(a.store.Snapshot())
			}, func() { w.settings = a.store.Snapshot(); w.resetSettings(); applyTheme(w.settings.Theme) })
		}},
		{"quit", "退出 Rivu", "CmdOrCtrl+Q", mygo.App.Quit},
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
		{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{{Role: mygo.RoleAbout}, a.menuItem("settings"), mygo.Separator(), {Role: mygo.RoleServices}, mygo.Separator(), {Role: mygo.RoleHide}, {Role: mygo.RoleHideOthers}, {Role: mygo.RoleUnhide}, mygo.Separator(), a.menuItem("quit")}},
		{Label: "文件", Submenu: []*mygo.MenuItem{a.menuItem("add"), a.menuItem("open"), a.menuItem("reload"), mygo.Separator(), {Role: mygo.RoleClose}}},
		{Label: "编辑", Submenu: []*mygo.MenuItem{{Role: mygo.RoleUndo}, {Role: mygo.RoleRedo}, mygo.Separator(), {Role: mygo.RoleCut}, {Role: mygo.RoleCopy}, {Role: mygo.RolePaste}, {Role: mygo.RoleSelectAll}}},
		{Label: "传输", Submenu: []*mygo.MenuItem{a.menuItem("pauseAll"), a.menuItem("resumeBatch"), a.menuItem("slow")}},
		{Label: "显示", Submenu: []*mygo.MenuItem{a.menuItem("main"), {Label: "命令面板…", Accelerator: "CmdOrCtrl+K", Click: func(*mygo.MenuItem, *mygo.Window) { w := a.openWindow(); w.dispatch(func() { w.paletteOpen = true }) }}, {Role: mygo.RoleToggleFullScreen}}},
		{Role: mygo.RoleWindowMenu},
	})
}
