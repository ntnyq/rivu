package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
)

type importItem struct {
	source    session.Source
	error     string
	submitted bool
}
type addWindow struct {
	*window
	input         string
	preparedInput string
	items         []importItem
	start         bool
	free          int64
	spaceError    string
	generation    uint64
}

func (a *application) openAdd(sources []string) {
	if a.add != nil && !a.add.closed {
		a.add.win.Show()
		a.add.win.Focus()
		if len(sources) > 0 {
			a.add.input = strings.Join(sources, "\n")
			a.add.prepare()
		}
		return
	}
	parent := a.openWindow()
	d := &addWindow{window: newWindow(a.store), start: true}
	d.app = a
	d.win = mygo.NewWindow(mygo.WindowOptions{Title: "添加下载", Width: 550, Height: 660, MinWidth: 500, MinHeight: 500, Parent: parent.win, Modal: true, Content: ui.View(d.view)})
	d.desktop.Parent = d.win
	d.win.OnClosed(d.close)
	a.add = d
	d.input = strings.Join(sources, "\n")
	if len(sources) > 0 {
		d.prepare()
	}
}
func (d *addWindow) prepare() {
	if d.task.Running() {
		return
	}
	input := d.input
	destination := d.destination
	d.generation++
	gen := d.generation
	var items []importItem
	var free int64
	var spaceError string
	d.run("读取下载来源", func(ctx context.Context) (string, error) {
		home, _ := os.UserHomeDir()
		if destination == filepath.Join(home, "Downloads", "Rivu") {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				spaceError = err.Error()
			}
		}
		if err := files.ValidateDestination(destination); err != nil {
			spaceError = err.Error()
		} else if n, err := files.FreeSpace(destination); err != nil {
			spaceError = err.Error()
		} else {
			free = n
		}
		for _, line := range strings.Split(input, "\n") {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			s, err := (torrentengine.Parser{}).Parse(line)
			item := importItem{source: s}
			if err != nil {
				item.source.Name = filepath.Base(line)
				if strings.HasPrefix(line, "magnet:") {
					item.source.Name = "磁力链接"
				}
				item.error = err.Error()
			}
			items = append(items, item)
		}
		if len(items) == 0 {
			return "", fmt.Errorf("请添加种子文件或磁力链接")
		}
		return "请确认保存位置和文件选择", nil
	}, func() {
		if gen == d.generation {
			d.items = items
			d.preparedInput = input
			d.free = free
			d.spaceError = spaceError
		}
	})
}
func (d *addWindow) pickTorrents() {
	var paths []string
	d.run("选择种子文件", func(context.Context) (string, error) {
		var err error
		paths, err = mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: d.win, Title: "选择种子文件", Multiple: true})
		if err != nil {
			return "", err
		}
		if len(paths) == 0 {
			return "", desktop.ErrCanceled
		}
		return "已选择种子文件，请读取并审阅", nil
	}, func() { d.input = strings.Join(paths, "\n"); d.items = nil })
}
func (d *addWindow) paste() {
	var value string
	d.run("粘贴磁力链接", func(context.Context) (string, error) {
		var err error
		value, err = d.desktop.Paste()
		return "已粘贴，请读取并审阅", err
	}, func() { d.input = value; d.items = nil })
}
func (d *addWindow) submit() {
	if d.input != d.preparedInput {
		d.report("请重新读取已修改的下载来源", fmt.Errorf("来源已更改"))
		return
	}
	items := make([]importItem, len(d.items))
	copy(items, d.items)
	for i := range items {
		items[i].source.Files = append([]session.File(nil), items[i].source.Files...)
	}
	destination, start, ratio := d.destination, d.start, d.settings.SeedRatio
	var selected string
	var failed bool
	d.run("添加下载", func(context.Context) (string, error) {
		for i := range items {
			item := &items[i]
			if item.error != "" || item.submitted {
				failed = failed || item.error != ""
				continue
			}
			id, err := d.app.session.Import(context.Background(), session.Import{Source: item.source, Destination: destination, Run: start, SeedRatio: ratio})
			if err != nil {
				item.error = err.Error()
				failed = true
			} else {
				item.submitted = true
				selected = id
			}
		}
		return "导入已处理；错误项目可修改后重新读取", nil
	}, func() {
		d.items = items
		if selected != "" {
			w := d.app.openWindow()
			w.dispatch(func() { w.selectOnly(selected) })
		}
		if !failed {
			d.win.Close()
		}
	})
}
func (d *addWindow) view(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(p.alt)
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Padding(25, 26, 20, 26).Gap(13).BorderWidth(0, 0, 1, 0).BorderColor(p.line).Children(func() {
			ui.Row(c).Size(43, 43).Radius(11).Background(p.soft).Justify(ui.Center).Children(func() { glyph(c, "download", 24, p.blue) })
			ui.Column(c).Gap(5).Children(func() {
				ui.Text(c, "添加下载").FontSize(17).Bold()
				ui.Text(c, "将新的内容，带到你的 Mac。").FontSize(11).TextColor(p.muted)
			})
			ui.Spacer(c)
			if iconButton(c, "x", "关闭添加下载", true) {
				d.win.Close()
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Padding(23, 26).Gap(12).Children(func() {
				ui.Text(c, "磁力链接或种子路径").FontSize(12).Bold()
				ui.TextArea(c, &d.input).Label("下载来源，每行一个").Placeholder("magnet:?xt=urn:btih:… 每行一个来源").Height(62).FontSize(12).Disabled(d.task.Running())
				ui.Text(c, "也可以选择本地 .torrent 文件。").FontSize(10).TextColor(p.muted)
				ui.Row(c).Gap(8).Children(func() {
					if ui.Button(c, "选择种子文件…").FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.pickTorrents()
					}
					if ui.Button(c, "粘贴链接").FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.paste()
					}
					ui.Spacer(c)
					if ui.Button(c, "读取并审阅").FontSize(11).TextColor(p.blue).Disabled(d.task.Running()).Clicked() {
						d.prepare()
					}
				})
				for i := range d.items {
					item := &d.items[i]
					ui.Column(c).Key(i).Gap(12).Padding(12, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
						ui.Row(c).Gap(10).Children(func() {
							fileSymbol(c, session.Task{Record: session.Record{Source: item.source}}, false)
							ui.Column(c).Gap(6).Grow(1).MinWidth(0).Children(func() {
								ui.Text(c, item.source.Name).FontSize(12).Bold().MaxLines(2)
								ui.Text(c, fmt.Sprintf("%d 个文件", len(item.source.Files))).FontSize(10).TextColor(p.muted)
							})
						})
						if item.error != "" {
							ui.Text(c, item.error).TextColor(c.Theme().Danger).FontSize(11)
							return
						}
						if item.submitted {
							ui.Text(c, "已添加").FontSize(11).TextColor(p.green)
							return
						}
						if len(item.source.Files) == 0 {
							ui.Text(c, "元数据待获取。开始后将下载全部文件；取消立即开始可先保存为暂停任务。").FontSize(11).TextColor(p.muted)
						} else {
							ui.Column(c).Padding(4, 11).Background(p.surface).Border(1, p.line).Radius(7).Children(func() {
								for j := range item.source.Files {
									f := &item.source.Files[j]
									row := ui.Row(c).Key(j).Padding(9, 0).Gap(8)
									if j > 0 {
										row.BorderWidth(1, 0, 0, 0).BorderColor(p.line)
									}
									row.Children(func() {
										selected := f.Priority > 0
										if ui.Checkbox(c, &selected, f.Path).FontSize(11).Grow(1).MinWidth(0).Disabled(d.task.Running()).Changed() {
											if selected {
												f.Priority = 1
											} else {
												f.Priority = 0
											}
										}
										ui.Text(c, bytes(f.Size)).FontSize(10).TextColor(p.muted)
									})
								}
							})
						}
					})
				}
				ui.Text(c, "保存到").FontSize(12).Bold().Padding(11, 0, 0, 0)
				ui.Row(c).Gap(10).Children(func() {
					ui.TextInput(c, &d.destination).Label("保存位置").FontSize(12).Grow(1)
					if ui.Button(c, "选择…").FontSize(11).Disabled(d.task.Running()).Clicked() {
						d.pickFolder()
					}
				})
				if d.spaceError != "" {
					ui.Text(c, d.spaceError).FontSize(10).TextColor(c.Theme().Danger)
				} else if d.free > 0 {
					ui.Text(c, "可用空间 "+bytes(d.free)+" · 提交时重新检查").FontSize(10).TextColor(p.muted)
				}
				ui.Checkbox(c, &d.start, "添加后立即开始下载").FontSize(12).Disabled(d.task.Running())
				ui.Text(c, fmt.Sprintf("下载完成后继续做种，分享率达到 %.1f 时停止。", d.settings.SeedRatio)).FontSize(10).TextColor(p.muted)
				if d.status != "准备就绪" {
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
			if ui.Button(c, "取消").FontSize(12).Clicked() {
				d.win.Close()
			}
			if ui.PrimaryButton(c, "添加下载").FontSize(12).Disabled(!valid || d.task.Running() || d.input != d.preparedInput).Clicked() {
				d.submit()
			}
		})
	})
}
