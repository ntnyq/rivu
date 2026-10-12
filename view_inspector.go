package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
)

func (w *window) inspectorView(c *ui.Context) {
	p := designTheme(c)
	theme := c.Theme()
	ui.Column(c).Width(271).Shrink(0).FillHeight().Padding(22, 18).Gap(0).Background(p.alt).BorderWidth(0, 0, 0, 1).BorderColor(p.line).Children(func() {
		t := w.chosen()
		if t == nil {
			ui.Text(c, tr("任务详情")).FontSize(14).FontWeight(600)
			ui.Text(c, tr("选择一个任务以查看详细信息。")).FontSize(11).TextColor(p.muted).Padding(20, 0)
			return
		}
		ui.Row(c).Height(48).Children(func() {
			fileSymbol(c, *t, true)
			ui.Spacer(c)
			ui.Text(c, tr(phases[t.Phase])).FontSize(10).TextColor(taskColor(c, t.Phase)).Padding(4, 7).Radius(5).Background(p.soft)
		})
		ui.Text(c, t.Source.Name).FontSize(14).FontWeight(600).MaxLines(3).Padding(17, 0, 6, 0)
		subtitle := fmt.Sprintf(tr("文件 · %s"), bytes(t.TotalSize))
		if t.Metadata == "" {
			subtitle = tr("等待元数据")
		}
		ui.Text(c, subtitle).FontSize(10).TextColor(p.muted)
		ui.Row(c).Height(28).Shrink(0).Padding(2).Gap(3).Radius(6).Background(p.track).Margin(22, 0).Children(func() {
			for i, label := range []string{tr("概览"), tr("文件"), tr("连接")} {
				tab := ui.Row(c).Grow(1).Height(24).Radius(4).Justify(ui.Center).Focusable().Label(label)
				if w.tab == i {
					tab.Background(p.surface).Border(1, p.line)
				}
				tab.Children(func() { ui.Text(c, label).FontSize(10) })
				if tab.Clicked() {
					w.tab = i
				}
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(0).Children(func() {
				switch w.tab {
				case 0:
					ui.Text(c, tr("下载进度")).FontSize(10).TextColor(p.muted)
					ui.Row(c).Gap(2).AlignItems(ui.End).Padding(8, 0, 9, 0).Children(func() {
						ui.Text(c, fmt.Sprintf("%.1f", progress(*t)*100)).FontSize(26).FontWeight(600)
						ui.Text(c, "%").FontSize(13).TextColor(p.muted)
					})
					progressTrack(c, progress(*t), taskColor(c, t.Phase), p.track, 4)
					ui.Text(c, bytes(t.Done)+" / "+bytes(t.Size)).FontSize(10).TextColor(p.muted).Padding(9, 0, 0, 0)
					ui.Row(c).Gap(20).Padding(22, 0, 18, 0).Children(func() {
						for i, v := range []struct {
							label string
							value float64
						}{{tr("↓ 下载速度"), t.DownRate}, {tr("↑ 上传速度"), t.UpRate}} {
							ui.Column(c).Grow(1).Gap(7).Children(func() {
								ui.Text(c, v.label).FontSize(10).TextColor(p.muted)
								number, unit := rateParts(v.value)
								color := p.text
								if i == 0 {
									color = p.blue
								}
								ui.Row(c).Gap(3).AlignItems(ui.End).Children(func() {
									ui.Text(c, number).FontSize(16).Bold().TextColor(color)
									ui.Text(c, unit).FontSize(10).TextColor(p.muted).Padding(0, 0, 1, 0)
								})
							})
						}
					})
					ui.Column(c).Padding(14, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Gap(12).Children(func() {
						for _, v := range [][2]string{{tr("预计剩余"), eta(*t)}, {tr("已上传"), bytes(t.Uploaded)}, {tr("分享率"), fmt.Sprintf("%.2f", float64(t.Uploaded)/float64(max(1, t.TotalSize)))}, {tr("连接节点"), fmt.Sprint(t.Peers)}, {tr("添加时间"), t.Added.Format("01-02 15:04")}} {
							ui.Row(c).Gap(8).Children(func() { ui.Text(c, v[0]).FontSize(10).TextColor(p.muted); ui.Spacer(c); ui.Text(c, v[1]).FontSize(10) })
						}
					})
					ui.Column(c).Margin(17, 0, 0, 0).Padding(17, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Gap(10).Children(func() {
						ui.Text(c, tr("保存位置")).FontSize(10).TextColor(p.muted)
						loc := ui.Row(c).Gap(7).Focusable().Label(tr("在 Finder 中显示"))
						loc.Children(func() {
							glyph(c, "folder", 15, p.blue)
							ui.Text(c, compactPath(t.Destination)).FontSize(10).SingleLine().Grow(1).MinWidth(0)
							glyph(c, "external", 15, p.blue)
						})
						if loc.Clicked() {
							w.reveal(t.Destination)
						}
					})
					seed := tr("下载完成后继续做种")
					if t.SeedRatio == 0 {
						seed = tr("下载完成后停止做种")
					}
					note := tr("无限做种")
					if t.SeedRatio == 0 {
						note = ""
					}
					if t.SeedRatio > 0 {
						note = fmt.Sprintf(tr("分享率达到 %.1f 时自动停止"), t.SeedRatio)
					}
					ui.Text(c, seed+"\n"+note).FontSize(10).TextColor(p.muted).Padding(18, 0, 0, 0)
					change := actionButton(c, "settings", tr("调整做种策略"), false).FontSize(10).TextColor(p.blue).Margin(10, 0).Disabled(w.app == nil)
					change.Menu(func(m *ui.Menu) {
						for _, v := range []struct {
							label string
							ratio float64
						}{{tr("完成后停止"), 0}, {tr("分享率 2.0"), 2}, {tr("无限做种"), -1}} {
							if m.Item(v.label).Chosen() {
								w.app.execute(session.Command{Kind: "seed", IDs: []string{t.ID}, Ratio: v.ratio})
							}
						}
					})
					if t.Error != "" {
						ui.Text(c, t.Error).FontSize(11).TextColor(p.orange).Padding(12).Radius(7).Background(p.orange.Alpha(.09))
					}

				case 1:
					if t.Metadata == "" {
						ui.Text(c, tr("等待元数据，文件列表尚不可用。"))
						return
					}
					ui.Text(c, tr("跨文件分片可能占用未选择文件的少量空间。")).FontSize(11).TextColor(theme.TextMuted)
					for i, f := range t.Source.Files {
						ui.Column(c).Key(f.Path).Gap(6).Children(func() {
							ui.Text(c, f.Path).FontSize(12).MaxLines(3)
							done := int64(0)
							if i < len(t.Files) {
								done = t.Files[i].Done
							}
							ui.Text(c, bytes(done)+" / "+bytes(f.Size)).FontSize(11).TextColor(theme.TextMuted)
							labels := []string{tr("不下载"), tr("普通"), tr("优先")}
							if ui.Button(c, labels[f.Priority]).Disabled(w.app == nil).Clicked() {
								w.app.execute(session.Command{Kind: "priority", IDs: []string{t.ID}, FileIndex: i, Priority: (f.Priority + 1) % 3})
							}
						})
					}
				case 2:
					ui.Text(c, fmt.Sprintf(tr("已连接节点：%d"), t.Peers))
					ui.Text(c, "Tracker").Bold()
					if len(t.Trackers) == 0 {
						ui.Text(c, tr("暂无 Tracker 观测")).TextColor(theme.TextMuted)
					}
					for _, tr := range t.Trackers {
						ui.Text(c, tr).FontSize(12).MaxLines(2)
					}
					ui.Text(c, tr("Tracker 地址已隐藏凭据。当前引擎接口未提供逐 Tracker 的公告结果。")).FontSize(11).TextColor(theme.TextMuted)
				}
			})
		})
	})
}
