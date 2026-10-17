package main

import (
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/locales"
)

func (w *mainWindow) toolbar(c *ui.Context, p palette) {
	left := float32(20)
	if w.sidebarHidden {
		left += c.TitleBar().Left
	}
	ui.Row(c).Height(69).Shrink(0).Padding(0, 20, 0, left).Gap(10).Background(p.alt).BorderWidth(0, 0, 1, 0).BorderColor(p.line).DragWindow().Children(func() {
		if w.sidebarHidden {
			if iconButton(c, "sidebar", locales.SidebarExpand(), true) {
				w.sidebarHidden = false
			}
		}
		ui.Column(c).Gap(4).Children(func() {
			ui.Text(c, filterLabels()[w.filter]).FontSize(15).Bold()
			ui.Text(c, locales.TransferCount(len(w.visible()))).FontSize(10).TextColor(p.muted)
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
			label = locales.TransferAllTags()
		}
		menuControl(c, locales.TransferFilterTag(), label).Width(95).Menu(func(m *ui.Menu) {
			for _, tag := range tags {
				label := tag
				if label == "" {
					label = locales.TransferAllTags()
				}
				if m.Item(label).Checked(w.tagFilter == tag).Chosen() {
					w.tagFilter = tag
				}
			}
		})
		if iconButton(c, "plus", locales.TransferAdd(), true) && w.app != nil {
			w.app.openAdd(nil)
		}
		if iconButton(c, "pause", locales.TransferToggleSelected(), len(w.ids()) > 0) {
			w.toggleSelected()
		}
		if iconButton(c, "trash", locales.TransferRemoveSelected(), len(w.ids()) > 0) {
			w.remove()
		}
		ui.Box(c).Width(1).Height(18).Background(p.line)
		ui.Row(c).Width(164).Height(29).Padding(0, 8).Gap(6).Radius(7).Background(p.track.Alpha(.65)).Children(func() {
			glyph(c, "search", 14, p.muted)
			ui.TextInput(c, &w.search).Label(locales.TransferSearch()).Placeholder(locales.TransferSearch()).Grow(1).MinWidth(0).FontSize(11).Background(ui.Transparent).Border(0, ui.Transparent).Padding(0)
		})
		if iconButton(c, "panel", locales.TransferToggleDetails(), w.chosen() != nil) {
			w.inspector = !w.inspector
		}
	})
}
