package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
)

func (w *mainWindow) view(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(p.surface)
	root := ui.Column(c).Fill()
	if paths := root.DroppedFiles(); len(paths) > 0 && w.app != nil {
		w.app.openAdd(paths)
	}
	root.Children(func() {
		ui.Row(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
			if !w.sidebarHidden {
				w.sidebar(c, p)
			}
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				w.toolbar(c, p)
				ui.Row(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
					ui.Column(c).Grow(1).MinWidth(0).Children(func() {
						ui.Row(c).Height(38).Shrink(0).Padding(0, 19).BorderWidth(0, 0, 1, 0).BorderColor(p.line).Children(func() {
							ui.Text(c, tr("名称与进度")).FontSize(10).TextColor(p.muted)
							ui.Spacer(c)
							labels := []string{tr("添加时间"), tr("名称"), tr("进度"), tr("状态")}
							menuControl(c, tr("任务排序"), fmt.Sprintf(tr("排序：%s"), labels[w.sort])).Menu(func(m *ui.Menu) {
								for i, label := range labels {
									if m.Item(label).Checked(w.sort == i).Chosen() {
										w.sort = i
									}
								}
							})
						})
						tasks := w.visible()
						if len(tasks) == 0 {
							w.emptyDownloads(c, p)
						} else {
							list := ui.List(c.Key("downloads"), &w.listState, len(tasks)).Grow(1).MinHeight(0).
								ItemKey(func(i int) any { return tasks[i].ID }).Rows(func(row ui.ListRow) {
								w.taskRow(row.Context, tasks[row.Index], tasks)
							})
							if list.Shortcut(0, ui.KeySpace) {
								w.toggleSelected()
							}
							ui.Row(c).Height(64).Shrink(0).Justify(ui.Center).Gap(4).Children(func() {
								glyph(c, "plus", 13, p.muted)
								ui.Text(c, tr("将种子文件拖到这里，或粘贴磁力链接")).FontSize(10).TextColor(p.muted)
							})
						}
					})
					width, _ := c.Size()
					if w.inspector && width >= 1060 {
						w.inspectorView(c)
					}
				})
			})
		})
		w.statusBar(c, p)
	})
	w.palette(c)
	w.confirmations(c)
	width, _ := c.Size()
	if w.inspector && width < 1060 && w.chosen() != nil {
		open := true
		ui.DialogBase(c, &open, func(_, panel ui.Element) {
			panel.Width(310).Height(550).Children(func() {
				w.inspectorView(c)
				if actionButton(c, "x", tr("关闭详情"), false).Clicked() {
					w.inspector = false
				}
			})
		})
		if !open {
			w.inspector = false
		}
	}
}

func (w *mainWindow) emptyDownloads(c *ui.Context, p palette) {
	ui.Column(c).Key("empty-downloads").Grow(1).MinHeight(0).FillWidth().Padding(24).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		ui.Column(c).FillWidth().Gap(14).AlignItems(ui.Center).Children(func() {
			glyph(c, "download", 44, p.blue)
			title, detail := tr("让下载安静地流动"), tr("添加种子文件或磁力链接，Rivu 会在后台继续下载。")
			if len(w.snapshot.Tasks) > 0 {
				title, detail = tr("没有匹配的任务"), tr("试试其他分类，或清除搜索内容。")
			}
			ui.Text(c, title).FontSize(17).Bold().TextAlign(ui.Center)
			ui.Text(c, detail).FontSize(12).TextColor(p.muted).TextAlign(ui.Center)
			if actionButton(c, "plus", tr("添加下载"), true).Clicked() && w.app != nil {
				w.app.openAdd(nil)
			}
			ui.Text(c, tr("将种子文件拖到这里，或粘贴磁力链接")).FontSize(10).TextColor(p.muted).TextAlign(ui.Center)
		})
	})
}
