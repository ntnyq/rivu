package main

import (
	"github.com/egoist/mygo/ui"
	"strings"
)

func (w *mainWindow) palette(c *ui.Context) {
	if !w.paletteOpen {
		return
	}
	ui.DialogBase(c, &w.paletteOpen, func(_, p ui.Element) {
		p.Width(460).Padding(22).Gap(12).Children(func() {
			ui.Text(c, tr("命令面板")).FontSize(20).Bold()
			ui.TextInput(c, &w.query).Label(tr("搜索命令")).Placeholder(tr("搜索操作…"))
			count := 0
			if w.app != nil {
				for _, cmd := range w.app.commands() {
					if !strings.Contains(strings.ToLower(cmd.label), strings.ToLower(w.query)) {
						continue
					}
					count++
					if ui.Button(c, cmd.label).Clicked() {
						w.paletteOpen = false
						cmd.run()
					}
				}
			}
			if count == 0 {
				ui.Text(c, tr("没有匹配的命令"))
			}
			if actionButton(c, "x", tr("关闭命令面板"), false).Clicked() {
				w.paletteOpen = false
			}
		})
	})
}
