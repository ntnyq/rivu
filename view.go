package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"path/filepath"
	"strings"
)

func (w *window) view(c *ui.Context) {
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
							ui.Text(c, "名称与进度").FontSize(10).TextColor(p.muted)
							ui.Spacer(c)
							labels := []string{"添加时间", "名称", "进度", "状态"}
							sortButton := ui.Row(c).Gap(8).Focusable().Label("任务排序")
							sortButton.Children(func() {
								ui.Text(c, "排序：  "+labels[w.sort]).FontSize(10).TextColor(p.muted)
								ui.Text(c, "⌄").TextColor(p.muted)
							})
							if sortButton.Clicked() {
								w.sort = (w.sort + 1) % 4
							}
						})
						tasks := w.visible()
						if len(tasks) == 0 {
							w.emptyDownloads(c, p)
						} else {
							list := ui.Scroll(c).Key("downloads").Grow(1).MinHeight(0)
							list.Children(func() {
								ui.Column(c).Children(func() {
									for _, task := range tasks {
										w.taskRow(c, task, tasks)
									}
								})
							})
							if list.Shortcut(0, ui.KeySpace) {
								w.toggleSelected()
							}
							ui.Row(c).Height(64).Shrink(0).Justify(ui.Center).Gap(4).Children(func() {
								glyph(c, "plus", 13, p.muted)
								ui.Text(c, "将种子文件拖到这里，或粘贴磁力链接").FontSize(10).TextColor(p.muted)
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
				if actionButton(c, "x", "关闭详情", false).Clicked() {
					w.inspector = false
				}
			})
		})
		if !open {
			w.inspector = false
		}
	}
}
func (w *window) emptyDownloads(c *ui.Context, p palette) {
	ui.Column(c).Key("empty-downloads").Grow(1).MinHeight(0).FillWidth().Padding(24).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		ui.Column(c).FillWidth().Gap(14).AlignItems(ui.Center).Children(func() {
			glyph(c, "download", 44, p.blue)
			title, detail := "让下载安静地流动", "添加种子文件或磁力链接，Rivu 会在后台继续下载。"
			if len(w.snapshot.Tasks) > 0 {
				title, detail = "没有匹配的任务", "试试其他分类，或清除搜索内容。"
			}
			ui.Text(c, title).FontSize(17).Bold().TextAlign(ui.Center)
			ui.Text(c, detail).FontSize(12).TextColor(p.muted).TextAlign(ui.Center)
			if actionButton(c, "plus", "添加下载", true).Clicked() && w.app != nil {
				w.app.openAdd(nil)
			}
			ui.Text(c, "将种子文件拖到这里，或粘贴磁力链接").FontSize(10).TextColor(p.muted).TextAlign(ui.Center)
		})
	})
}

func (w *window) sidebar(c *ui.Context, p palette) {
	ui.Column(c).Width(191).Shrink(0).Background(p.sidebar).BorderWidth(0, 1, 0, 0).BorderColor(p.line).Children(func() {
		ui.Row(c).Height(58).Shrink(0).Padding(0, 18).Children(func() {
			ui.Spacer(c).DragWindow()
			if iconButton(c, "sidebar", "收起侧栏", true) {
				w.sidebarHidden = true
			}
		})
		ui.Row(c).Height(41).Shrink(0).Padding(10, 14, 0, 14).Gap(7).Children(func() { ui.Image(c, appMark).Size(38, 38); ui.Text(c, "Rivu").FontSize(20).Bold() })
		ui.Column(c).Padding(29, 12, 0, 12).Gap(3).Children(func() {
			ui.Text(c, "任务库").FontSize(11).Bold().TextColor(p.muted).Padding(0, 10, 8, 10)
			icons := []string{"grid", "arrow-down", "arrow-up", "check", "pause", "alert"}
			for i, label := range filters {
				count := 0
				for _, task := range w.snapshot.Tasks {
					if matches(task, i) {
						count++
					}
				}
				selected := w.filter == i
				bg, ink, iconColor := ui.Transparent, p.text, p.blue
				if i == 5 {
					iconColor = p.orange
				}
				if selected {
					bg = p.blue
					if c.Theme().Dark {
						bg = ui.Hex("#176dcf")
					}
					ink = ui.Hex("#ffffff")
					iconColor = ink
				}
				row := ui.Row(c).Key(label).Height(35).FillWidth().Padding(0, 11).Gap(10).Radius(7).Background(bg).Focusable().Label(label)
				if !selected && row.Hovered() {
					row.Background(p.track)
				}
				row.Children(func() {
					glyph(c, icons[i], 17, iconColor)
					ui.Text(c, label).FontSize(12).TextColor(ink)
					ui.Spacer(c)
					color := p.muted
					if selected {
						color = ink
					}
					ui.Text(c, fmt.Sprint(count)).FontSize(11).TextColor(color)
				})
				if row.Clicked() || row.Shortcut(0, ui.KeyEnter) {
					w.filter = i
				}
			}
		})
		ui.Spacer(c)
		ui.Column(c).Padding(20, 18, 18, 18).Gap(12).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				glyph(c, "drive", 17, p.muted)
				ui.Column(c).Gap(5).MinWidth(0).Children(func() {
					ui.Text(c, "下载位置").FontSize(11)
					ui.Text(c, compactPath(w.settings.Destination)).FontSize(9).SingleLine().TextColor(p.muted)
				})
			})
			if w.diskTotal > 0 {
				ui.Text(c, "可用 "+bytes(w.diskFree)).FontSize(9).TextColor(p.muted)
				progressTrack(c, 1-float64(w.diskFree)/float64(w.diskTotal), p.muted.Alpha(.5), p.track, 3)
			} else {
				ui.Box(c).Height(3).Background(p.track).Radius(3)
			}
			ui.Row(c).Padding(12, 0, 0, 0).Gap(9).Focusable().Label("设置").OnClick(func() {
				if w.app != nil {
					w.app.openSettings()
				}
			}).Children(func() {
				glyph(c, "settings", 16, p.text)
				ui.Text(c, "设置").FontSize(12)
				ui.Spacer(c)
				ui.Text(c, "⌘ ,").FontSize(11).TextColor(p.muted)
			})
		})
	})
}
func (w *window) toolbar(c *ui.Context, p palette) {
	left := float32(20)
	if w.sidebarHidden {
		left += c.TitleBar().Left
	}
	ui.Row(c).Height(69).Shrink(0).Padding(0, 20, 0, left).Gap(10).Background(p.alt).BorderWidth(0, 0, 1, 0).BorderColor(p.line).DragWindow().Children(func() {
		if w.sidebarHidden {
			if iconButton(c, "sidebar", "展开侧栏", true) {
				w.sidebarHidden = false
			}
		}
		ui.Column(c).Gap(4).Children(func() {
			ui.Text(c, filters[w.filter]).FontSize(15).Bold()
			ui.Text(c, fmt.Sprintf("%d 个任务", len(w.visible()))).FontSize(10).TextColor(p.muted)
		})
		ui.Spacer(c)
		if iconButton(c, "plus", "添加下载", true) && w.app != nil {
			w.app.openAdd(nil)
		}
		if iconButton(c, "pause", "暂停或继续所选任务", len(w.ids()) > 0) {
			w.toggleSelected()
		}
		if iconButton(c, "trash", "移除所选任务", len(w.ids()) > 0) {
			w.remove()
		}
		ui.Box(c).Width(1).Height(18).Background(p.line)
		ui.Row(c).Width(164).Height(29).Padding(0, 8).Gap(6).Radius(7).Background(p.track.Alpha(.65)).Children(func() {
			glyph(c, "search", 14, p.muted)
			ui.TextInput(c, &w.search).Label("搜索任务").Placeholder("搜索任务").Grow(1).MinWidth(0).FontSize(11).Background(ui.Transparent).Border(0, ui.Transparent).Padding(0)
		})
		if iconButton(c, "panel", "切换任务详情", true) {
			w.inspector = !w.inspector
		}
	})
}
func (w *window) statusBar(c *ui.Context, p palette) {
	ui.Row(c).Height(34).Shrink(0).Padding(0, 15).Gap(22).Background(p.alt).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
		ui.Row(c).Gap(5).Children(func() { glyph(c, "arrow-down", 12, p.blue); ui.Text(c, rate(w.snapshot.DownRate)).FontSize(10) })
		ui.Row(c).Gap(5).Children(func() {
			glyph(c, "arrow-up", 12, p.muted)
			ui.Text(c, rate(w.snapshot.UpRate)).FontSize(10).TextColor(p.muted)
		})
		ui.Spacer(c)
		message := w.status
		if message == "准备就绪" {
			message = ""
		}
		if w.snapshot.Error != "" {
			message = w.snapshot.Error
		}
		col := p.muted
		if w.failed || w.snapshot.Error != "" {
			col = c.Theme().Danger
		}
		ui.Text(c, message).SingleLine().MinWidth(0).FontSize(10).TextColor(col)
		color := p.muted
		if w.snapshot.SlowMode {
			color = p.blue
		}
		limit := ui.Row(c).Gap(5).Focusable().Label("限速模式")
		limit.Children(func() { glyph(c, "gauge", 13, color); ui.Text(c, "限速模式").FontSize(10).TextColor(color) })
		if limit.Clicked() && w.app != nil {
			w.app.slowMode()
		}
		peers := 0
		for _, task := range w.snapshot.Tasks {
			peers += task.Peers
		}
		network := "等待连接"
		dot := p.muted
		if peers > 0 {
			network = "已连接节点"
			dot = p.green
		}
		ui.Row(c).Gap(5).Children(func() {
			ui.Box(c).Size(5, 5).Radius(3).Background(dot)
			ui.Text(c, network).FontSize(10).TextColor(p.muted)
		})
	})
}
func fileKind(t session.Task) string {
	name := strings.ToLower(t.Source.Name)
	if strings.Contains(name, "movie") || strings.HasSuffix(name, ".mp4") || strings.HasSuffix(name, ".mkv") {
		return "movie"
	}
	if strings.Contains(name, "archive") || strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar") {
		return "archive"
	}
	return "file"
}
func fileSymbol(c *ui.Context, t session.Task, large bool) {
	kind := fileKind(t)
	color := c.Theme().Accent
	if kind == "movie" {
		color = ui.Hex("#9263b7")
	}
	if kind == "archive" {
		color = c.Theme().TextMuted
	}
	width, height := float32(32), float32(39)
	if large {
		width, height = 40, 48
	}
	ui.Row(c).Size(width, height).Shrink(0).Radius(5).Border(1, color.Alpha(.17)).Background(color.Alpha(.06)).Justify(ui.Center).Children(func() { glyph(c, kind, 20, color) })
}
func (w *window) taskRow(c *ui.Context, t session.Task, visible []session.Task) {
	p := designTheme(c)
	row := ui.Row(c).Key(t.ID).Padding(17, 17, 15, 17).Gap(12).AlignItems(ui.Start).Height(102).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(p.line).Label(t.Source.Name).Focusable()
	if w.selected[t.ID] {
		row.Background(p.soft)
	} else if row.Hovered() {
		row.Background(p.alt)
	}
	row.Children(func() {
		ui.Box(c).Margin(3, 0, 0, 0).Shrink(0).Children(func() { fileSymbol(c, t, false) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
			ui.Row(c).Gap(10).Children(func() {
				ui.Text(c, t.Source.Name).FontSize(12).FontWeight(600).Grow(1).MinWidth(0).SingleLine()
				ui.Text(c, taskPercent(t)).FontSize(11).TextColor(p.muted)
			})
			ui.Row(c).Gap(10).Children(func() {
				text := bytes(t.Done) + " / " + bytes(t.Size)
				if t.Metadata == "" {
					text = "文件列表和大小待获取"
				}
				ui.Text(c, text).FontSize(10).TextColor(p.muted)
				ui.Spacer(c)
				color := p.muted
				if t.Completed || t.Error != "" {
					color = taskColor(c, t.Phase)
				}
				ui.Text(c, phases[t.Phase]).FontSize(10).TextColor(color)
			})
			ui.Box(c).Margin(4, 0, 1, 0).Children(func() {
				progressTrack(c, progress(t), taskColor(c, t.Phase), p.track, 3)
			})
			ui.Row(c).Gap(10).Children(func() {
				left, right := "↓ "+rate(t.DownRate)+"   ↑ "+rate(t.UpRate), eta(t)
				color := p.blue
				switch t.Phase {
				case "paused":
					left, right, color = "任务已暂停", "等待继续", p.muted
				case "completed":
					left, right, color = "文件已准备就绪", "已达到做种目标", p.muted
				case "seeding":
					left = "↑ " + rate(t.UpRate)
					right = fmt.Sprintf("分享率 %.2f / %.1f", float64(t.Uploaded)/float64(max(1, t.TotalSize)), t.SeedRatio)
					color = p.muted
				case "error":
					left, right, color = t.Error, "请处理后重试", p.muted
				case "metadata":
					left, right, color = "正在获取元数据", "", p.muted
				}
				ui.Text(c, left).FontSize(10).TextColor(color).SingleLine().Grow(1).MinWidth(0)
				ui.Text(c, right).FontSize(10).TextColor(p.muted)
			})
		})
		icon := "pause"
		if !t.Run {
			icon = "play"
		}
		if t.Phase == "completed" {
			icon = "folder"
		}
		if t.Error != "" {
			icon = "alert"
		}
		if compactIconButton(c, icon, "操作 "+t.ID, 23, 15, p.muted) && w.app != nil {
			if t.Phase == "completed" {
				w.reveal(t.Destination)
			} else {
				kind := "resume"
				if t.Run && t.Error == "" {
					kind = "pause"
				}
				w.app.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
			}
		}
	})
	if row.Clicked() {
		mods := row.ClickModifiers()
		if mods&ui.Shift != 0 && w.anchor != "" {
			start, end := -1, -1
			for i, v := range visible {
				if v.ID == w.anchor {
					start = i
				}
				if v.ID == t.ID {
					end = i
				}
			}
			if start >= 0 && end >= 0 {
				for _, v := range visible[min(start, end) : max(start, end)+1] {
					w.selected[v.ID] = true
				}
			}
		} else if mods&ui.Cmd != 0 {
			w.selected[t.ID] = !w.selected[t.ID]
			w.anchor = t.ID
		} else {
			w.selectOnly(t.ID)
		}
	}
	if row.DoubleClicked() {
		w.selectOnly(t.ID)
		w.inspector = true
	}
	if row.Shortcut(0, ui.KeySpace) {
		w.toggleSelected()
	}
	row.ContextMenu(func(m *ui.Menu) {
		kind, label := "resume", "继续"
		if t.Run {
			kind, label = "pause", "暂停"
		}
		if m.Item(label).Chosen() {
			w.app.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
		}
		if m.Item("在 Finder 中显示").Chosen() {
			w.reveal(t.Destination)
		}
		if m.Item("复制磁力链接").Chosen() {
			w.copy(magnet(t))
		}
		if m.Item("校验数据").Chosen() {
			w.app.execute(session.Command{Kind: "verify", IDs: []string{t.ID}})
		}
		if m.Item("优先开始").Chosen() {
			w.app.execute(session.Command{Kind: "first", IDs: []string{t.ID}})
		}
		m.Separator()
		if m.Item("移除…").Chosen() {
			w.selectOnly(t.ID)
			w.remove()
		}
	})
}
func compactPath(path string) string {
	if path == "" {
		return "尚未选择"
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) > 3 && parts[1] == "Users" {
		return "~/" + strings.Join(parts[3:], "/")
	}
	return path
}
func (w *window) palette(c *ui.Context) {
	if !w.paletteOpen {
		return
	}
	ui.DialogBase(c, &w.paletteOpen, func(_, p ui.Element) {
		p.Width(460).Padding(22).Gap(12).Children(func() {
			ui.Text(c, "命令面板").FontSize(20).Bold()
			ui.TextInput(c, &w.query).Label("搜索命令").Placeholder("搜索操作…")
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
				ui.Text(c, "没有匹配的命令")
			}
			if actionButton(c, "x", "关闭命令面板", false).Clicked() {
				w.paletteOpen = false
			}
		})
	})
}
func (w *window) confirmations(c *ui.Context) {
	if w.removeOpen {
		ui.DialogBase(c, &w.removeOpen, func(_, p ui.Element) {
			p.Width(450).Padding(24).Gap(16).Children(func() {
				ui.Text(c, fmt.Sprintf("移除 %d 个任务？", len(w.removeIDs))).FontSize(20).Bold()
				ui.Text(c, "默认保留已下载文件。移到废纸篓只处理这些任务拥有的文件。")
				ui.Checkbox(c, &w.trashFiles, "同时将下载文件移到废纸篓")
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "x", "取消", false).Clicked() {
						w.removeOpen = false
					}
					if actionButton(c, "trash", "确认移除", true).Clicked() {
						w.app.execute(session.Command{Kind: "remove", IDs: w.removeIDs, Trash: w.trashFiles})
						w.removeOpen = false
					}
				})
			})
		})
	}
	if w.quitConfirm {
		ui.DialogBase(c, &w.quitConfirm, func(_, p ui.Element) {
			p.Width(430).Padding(24).Gap(16).Children(func() {
				ui.Text(c, "退出 Rivu？").FontSize(20).Bold()
				ui.Text(c, "下载和做种将中断。下次启动会按设置恢复。关闭主窗口可以继续在后台下载。")
				ui.Checkbox(c, &w.rememberQuit, "以后退出时不再询问")
				ui.Row(c).Gap(10).Children(func() {
					if actionButton(c, "play", "继续下载", false).Clicked() {
						w.quitConfirm = false
					}
					if actionButton(c, "check", "保存并退出", true).Clicked() {
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
