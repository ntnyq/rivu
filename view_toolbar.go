package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"slices"
)

func (w *mainWindow) toolbar(c *ui.Context, p palette) {
	left := float32(20)
	if w.sidebarHidden {
		left += c.TitleBar().Left
	}
	ui.Row(c).Height(69).Shrink(0).Padding(0, 20, 0, left).Gap(10).Background(p.alt).BorderWidth(0, 0, 1, 0).BorderColor(p.line).DragWindow().Children(func() {
		if w.sidebarHidden {
			if iconButton(c, "sidebar", tr("展开侧栏"), true) {
				w.sidebarHidden = false
			}
		}
		ui.Column(c).Gap(4).Children(func() {
			ui.Text(c, filterLabels()[w.filter]).FontSize(15).Bold()
			ui.Text(c, fmt.Sprintf(tr("%d 个任务"), len(w.visible()))).FontSize(10).TextColor(p.muted)
		})
		ui.Spacer(c)
		tags := []string{""}
		for _, task := range w.snapshot.Tasks {
			for _, tag := range task.Tags {
				if !slices.Contains(tags, tag) {
					tags = append(tags, tag)
				}
			}
		}
		label := w.tagFilter
		if label == "" {
			label = tr("全部标签")
		}
		menuControl(c, tr("按标签筛选"), label).Width(95).Menu(func(m *ui.Menu) {
			for _, tag := range tags {
				label := tag
				if label == "" {
					label = tr("全部标签")
				}
				if m.Item(label).Checked(w.tagFilter == tag).Chosen() {
					w.tagFilter = tag
				}
			}
		})
		if iconButton(c, "plus", tr("添加下载"), true) && w.app != nil {
			w.app.openAdd(nil)
		}
		if iconButton(c, "pause", tr("暂停或继续所选任务"), len(w.ids()) > 0) {
			w.toggleSelected()
		}
		if iconButton(c, "trash", tr("移除所选任务"), len(w.ids()) > 0) {
			w.remove()
		}
		ui.Box(c).Width(1).Height(18).Background(p.line)
		ui.Row(c).Width(164).Height(29).Padding(0, 8).Gap(6).Radius(7).Background(p.track.Alpha(.65)).Children(func() {
			glyph(c, "search", 14, p.muted)
			ui.TextInput(c, &w.search).Label(tr("搜索任务")).Placeholder(tr("搜索任务")).Grow(1).MinWidth(0).FontSize(11).Background(ui.Transparent).Border(0, ui.Transparent).Padding(0)
		})
		if iconButton(c, "panel", tr("切换任务详情"), true) {
			w.inspector = !w.inspector
		}
	})
}
