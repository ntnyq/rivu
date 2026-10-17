package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func (d *addWindow) view(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(p.alt)
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Padding(25, 26, 20, 26).Gap(13).BorderWidth(0, 0, 1, 0).BorderColor(p.line).Children(func() {
			ui.Row(c).Size(43, 43).Radius(11).Background(p.soft).Justify(ui.Center).Children(func() { glyph(c, "download", 24, p.blue) })
			ui.Column(c).Gap(5).Children(func() {
				ui.Text(c, locales.TransferAdd()).FontSize(17).Bold()
				ui.Text(c, locales.ImportSubtitle()).FontSize(11).TextColor(p.muted)
			})
			ui.Spacer(c)
			if iconButton(c, "x", locales.ImportClose(), true) {
				d.win.Close()
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Padding(23, 26).Gap(12).Children(func() {
				ui.Text(c, locales.ImportSources()).FontSize(12).Bold()
				ui.TextArea(c, &d.input).Label(locales.ImportSourcesLabel()).Placeholder(locales.ImportSourcesPlaceholder()).Height(62).FontSize(12).Disabled(d.task.Running())
				ui.Text(c, locales.ImportLocalFilesHint()).FontSize(10).TextColor(p.muted)
				ui.Row(c).Gap(8).Children(func() {
					if actionButton(c, "folder", locales.ImportChooseFilesMenu(), false).FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.pickTorrents()
					}
					if actionButton(c, "clipboard", locales.ImportPaste(), false).FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.paste()
					}
					ui.Spacer(c)
					if actionButton(c, "search", locales.ImportReview(), false).FontSize(11).TextColor(p.blue).Disabled(d.task.Running()).Clicked() {
						d.prepare()
					}
				})
				if d.task.Running() && ui.Button(c, locales.ImportCancelPreview()).Clicked() {
					d.task.Cancel()
					d.job = ""
					d.report(locales.CommonCanceled(), nil)
				}
				for i := range d.items {
					item := &d.items[i]
					ui.Column(c).Key(i).Gap(12).Padding(12, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
						ui.Row(c).Gap(10).Children(func() {
							fileSymbol(c, session.Task{Record: session.Record{Source: item.source}}, false)
							ui.Column(c).Gap(6).Grow(1).MinWidth(0).Children(func() {
								ui.Text(c, item.source.Name).FontSize(12).Bold().MaxLines(2)
								ui.Text(c, locales.ImportFileCount(len(item.source.Files))).FontSize(10).TextColor(p.muted)
							})
						})
						if item.error != "" {
							ui.Text(c, problemText(item.issue, item.error)).TextColor(c.Theme().Danger).FontSize(11)
							return
						}
						if item.submitted {
							ui.Text(c, locales.ImportAdded()).FontSize(11).TextColor(p.green)
							return
						}
						if len(item.source.Files) == 0 {
							ui.Text(c, locales.ImportMetadataHint()).FontSize(11).TextColor(p.muted)
						} else {
							ui.Column(c).Padding(4, 11).Background(p.surface).Border(1, p.line).Radius(7).Children(func() {
								fileTree(c, item.source.Files, &item.tree, !d.task.Running(), func(indices []int, priority int) {
									for _, index := range indices {
										item.source.Files[index].Priority = priority
									}
								})
							})
						}
					})
				}
				ui.Text(c, locales.ImportSaveTo()).FontSize(12).Bold().Padding(11, 0, 0, 0)
				ui.Row(c).Gap(10).Children(func() {
					ui.TextInput(c, &d.destination).Label(locales.FilesSaveLocation()).FontSize(12).Grow(1)
					if actionButton(c, "folder", locales.CommonChoose(), false).FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.pickFolder()
					}
				})
				if d.spaceError != "" {
					ui.Text(c, d.spaceError).FontSize(10).TextColor(c.Theme().Danger)
				} else if d.free > 0 {
					ui.Text(c, locales.ImportAvailableSpace(bytes(d.free))).FontSize(10).TextColor(p.muted)
				}
				ui.Checkbox(c, &d.adoptExisting, locales.ImportAdoptExisting()).FontSize(11).Disabled(d.task.Running())
				field(c, locales.TransferTags(), &d.tags)
				ui.Checkbox(c, &d.start, locales.ImportStartImmediately()).FontSize(12).Disabled(d.task.Running())
				ui.Text(c, seedingSummary(d.settings.SeedRatio, d.settings.SeedLimitSeconds)).FontSize(10).TextColor(p.muted)
				if d.status != "" {
					color := p.muted
					if d.failed {
						color = c.Theme().Danger
					}
					ui.Text(c, d.status).FontSize(11).TextColor(color).MaxLines(3)
				}
			})
		})
		valid := false
		for _, item := range d.items {
			if item.error == "" && !item.submitted {
				valid = true
			}
		}
		ui.Row(c).Padding(15, 25).Gap(9).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
			ui.Spacer(c)
			if actionButton(c, "x", locales.CommonCancel(), false).FontSize(12).Clicked() {
				d.win.Close()
			}
			if actionButton(c, "plus", locales.TransferAdd(), true).FontSize(12).Disabled(!valid || d.task.Running() || d.input != d.preparedInput).Clicked() {
				d.submit()
			}
		})
	})
}
