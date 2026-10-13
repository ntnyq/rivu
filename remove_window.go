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
				ui.Text(c, fmt.Sprintf(tr("移除 %d 个任务？"), len(w.removeIDs))).FontSize(20).Bold()
				ui.Text(c, tr("默认保留已下载文件。移到废纸篓只处理这些任务拥有的文件。"))
				ui.Checkbox(c, &w.trashFiles, tr("同时将下载文件移到废纸篓"))
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "x", tr("取消"), false).Clicked() {
						w.removeOpen = false
					}
					if actionButton(c, "trash", tr("确认移除"), true).Clicked() {
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
				ui.Text(c, tr("退出 Rivu？")).FontSize(20).Bold()
				ui.Text(c, tr("下载和做种将中断。下次启动会按设置恢复。关闭主窗口可以继续在后台下载。"))
				ui.Checkbox(c, &w.rememberQuit, tr("以后退出时不再询问"))
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "play", tr("继续下载"), false).Clicked() {
						w.quitConfirm = false
					}
					if actionButton(c, "check", tr("保存并退出"), true).Clicked() {
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
