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
		{"add", tr("添加下载…"), "CmdOrCtrl+N", func() { a.openAdd(nil) }},
		{"open", tr("打开种子文件…"), "CmdOrCtrl+O", func() { a.openAdd(nil); a.add.pickTorrents() }},
		{"pauseAll", tr("暂停全部"), "", func() { a.execute(session.Command{Kind: session.CommandPauseAll}) }},
		{"resumeBatch", tr("恢复本次暂停"), "", func() { a.execute(session.Command{Kind: session.CommandResumeBatch}) }},
		{"slow", tr("切换限速模式"), "", a.slowMode},
		{"main", tr("打开 Rivu"), "CmdOrCtrl+1", func() { a.openWindow() }},
		{"settings", tr("设置…"), "CmdOrCtrl+,", a.openSettings},
		{"reload", tr("修复后重新读取会话与设置"), "", func() {
			w := a.openWindow()
			w.run(tr("重新读取"), func(context.Context) (string, error) {
				a.settingsMu.Lock()
				defer a.settingsMu.Unlock()
				if err := a.store.Reload(); err != nil {
					return "", err
				}
				if err := a.session.Reload(); err != nil {
					return "", err
				}
				return tr("已重新读取"), a.session.Configure(a.store.Snapshot())
			}, func() {
				w.settings = a.store.Snapshot()
				applyTheme(w.settings.Theme)
				a.applyLanguage()
			})
		}},
		{"quit", tr("退出 Rivu"), "CmdOrCtrl+Q", mygo.App.Quit},
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
		{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{{Role: mygo.RoleAbout, Label: tr("关于 Rivu")}, a.menuItem("settings"), mygo.Separator(), {Role: mygo.RoleServices, Label: tr("服务")}, mygo.Separator(), {Role: mygo.RoleHide, Label: tr("隐藏 Rivu")}, {Role: mygo.RoleHideOthers, Label: tr("隐藏其他应用")}, {Role: mygo.RoleUnhide, Label: tr("全部显示")}, mygo.Separator(), a.menuItem("quit")}},
		{Label: tr("文件菜单"), Submenu: []*mygo.MenuItem{a.menuItem("add"), a.menuItem("open"), a.menuItem("reload"), mygo.Separator(), {Role: mygo.RoleClose, Label: tr("关闭窗口")}}},
		{Label: tr("编辑"), Submenu: []*mygo.MenuItem{{Role: mygo.RoleUndo, Label: tr("撤销")}, {Role: mygo.RoleRedo, Label: tr("重做")}, mygo.Separator(), {Role: mygo.RoleCut, Label: tr("剪切")}, {Role: mygo.RoleCopy, Label: tr("复制")}, {Role: mygo.RolePaste, Label: tr("粘贴")}, {Role: mygo.RoleSelectAll, Label: tr("全选")}}},
		{Label: tr("传输"), Submenu: []*mygo.MenuItem{a.menuItem("pauseAll"), a.menuItem("resumeBatch"), a.menuItem("slow")}},
		{Label: tr("显示"), Submenu: []*mygo.MenuItem{a.menuItem("main"), {Label: tr("命令面板…"), Accelerator: "CmdOrCtrl+K", Click: func(*mygo.MenuItem, *mygo.Window) { w := a.openWindow(); w.dispatch(func() { w.paletteOpen = true }) }}, {Role: mygo.RoleToggleFullScreen, Label: tr("切换全屏")}}},
		{Role: mygo.RoleWindowMenu, Label: tr("窗口"), Submenu: []*mygo.MenuItem{{Role: mygo.RoleMinimize, Label: tr("最小化")}, {Role: mygo.RoleZoom, Label: tr("缩放")}, {Role: mygo.RoleFront, Label: tr("全部置于前面")}}},
	})
}
