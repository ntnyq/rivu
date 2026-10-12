package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
)

func (w *window) resetSettings() {
	s := w.settings
	w.destination = s.Destination
	w.concurrent = strconv.Itoa(s.Concurrent)
	w.seedRatio = strconv.FormatFloat(s.SeedRatio, 'f', 1, 64)
	w.downloadLimit = strconv.FormatInt(s.DownloadLimit, 10)
	w.uploadLimit = strconv.FormatInt(s.UploadLimit, 10)
	w.slowDownload = strconv.FormatInt(s.SlowDownload, 10)
	w.slowUpload = strconv.FormatInt(s.SlowUpload, 10)
	w.listenPort = strconv.Itoa(s.ListenPort)
}
func field(c *ui.Context, label string, v *string) {
	ui.Column(c).Gap(5).Children(func() { ui.Text(c, label).FontSize(12); ui.TextInput(c, v).Label(label) })
}
func settingGroup(c *ui.Context, body func()) {
	ui.Column(c).Padding(0, 14).Radius(9).Border(1, c.Theme().Border).Background(c.Theme().Surface).Children(body)
}
func switchRow(c *ui.Context, title, caption string, value *bool, line bool, enabled bool) {
	row := ui.Row(c).Key(title).Padding(15, 0).Gap(20)
	if line {
		row.BorderWidth(1, 0, 0, 0).BorderColor(c.Theme().Border)
	}
	row.Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
			ui.Text(c, title).FontSize(12)
			if caption != "" {
				ui.Text(c, caption).FontSize(10).TextColor(c.Theme().TextMuted)
			}
		})
		ui.Switch(c, value).Label(title).Disabled(!enabled).Size(29, 17)
	})
}
func (w *window) settingsView(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(p.alt)
	labels := []string{"通用", "下载与做种", "连接", "通知", "更新与关于"}
	icons := []string{"settings", "download", "globe", "bell", "refresh"}
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Height(50).Shrink(0).Justify(ui.Center).Background(p.sidebar).BorderWidth(0, 0, 1, 0).BorderColor(p.line).DragWindow().Children(func() { ui.Text(c, "设置").FontSize(13).Bold() })
		ui.Row(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(168).Shrink(0).Padding(18, 10).Gap(3).Background(p.sidebar).BorderWidth(0, 1, 0, 0).BorderColor(p.line).Children(func() {
				for i, label := range labels {
					selected := i == w.settingsTab
					ink, ic := p.text, p.blue
					row := ui.Row(c).Key(label).Height(37).Padding(0, 10).Gap(9).Radius(6).Focusable().Label(label)
					if selected {
						row.Background(p.blue)
						ink = ui.Hex("#ffffff")
						ic = ink
					}
					row.Children(func() { glyph(c, icons[i], 17, ic); ui.Text(c, label).FontSize(12).TextColor(ink) })
					if row.Clicked() {
						w.settingsTab = i
					}
				}
			})
			ui.Column(c).Grow(1).MinWidth(0).Padding(27, 29, 16, 29).Gap(20).Children(func() {
				ui.Text(c, labels[w.settingsTab]).FontSize(19).Bold()
				ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
					ui.Column(c).Gap(17).Children(func() {
						switch w.settingsTab {
						case 0:
							settingGroup(c, func() {
								ui.Row(c).Padding(15, 0).Children(func() {
									ui.Text(c, "外观").FontSize(12)
									ui.Spacer(c)
									labels := map[string]string{"system": "跟随系统", "light": "浅色", "dark": "深色"}
									ui.Button(c, labels[w.settings.Theme]+"  ⌄").FontSize(11).Menu(func(m *ui.Menu) {
										for _, v := range []string{"system", "light", "dark"} {
											if m.Item(labels[v]).Chosen() {
												w.settings.Theme = v
											}
										}
									})
								})
								switchRow(c, "在菜单栏显示 Rivu", "随时查看下载进度", &w.settings.Tray, true, true)
							})
							settingGroup(c, func() {
								switchRow(c, "登录时启动 Rivu", "", &w.settings.Login, false, true)
								switchRow(c, "仅在后台启动", "登录后继续下载，不打开主窗口", &w.settings.BackgroundLogin, true, w.settings.Login)
								switchRow(c, "恢复上次运行的任务", "手动暂停的任务保持暂停", &w.settings.Restore, true, true)
							})
							settingGroup(c, func() {
								switchRow(c, "退出时确认", "避免意外中断下载和做种", &w.settings.ConfirmQuit, false, true)
							})
							ui.Text(c, "关闭主窗口后，下载会继续。\n从 Rivu 菜单选择“退出 Rivu”可结束应用。").FontSize(11).TextColor(p.muted)
							if w.loginStatus != "" {
								ui.Text(c, w.loginStatus).FontSize(10).TextColor(p.muted)
							}

						case 1:
							settingGroup(c, func() {
								ui.Column(c).Padding(15, 0).Gap(10).Children(func() {
									ui.Text(c, "默认下载位置").FontSize(12)
									ui.Row(c).Gap(8).Children(func() {
										ui.TextInput(c, &w.destination).Label("默认下载目录").FontSize(11).Grow(1).MinWidth(0)
										if actionButton(c, "folder", "选择…", false).FontSize(11).Clicked() {
											w.pickFolder()
										}
									})
								})
								settingField(c, "并行下载数", &w.concurrent, true)
								settingField(c, "默认分享率", &w.seedRatio, true)
							})
							settingGroup(c, func() {
								settingField(c, "普通下载上限", &w.downloadLimit, false)
								settingField(c, "普通上传上限", &w.uploadLimit, true)
								settingField(c, "临时下载上限", &w.slowDownload, true)
								settingField(c, "临时上传上限", &w.slowUpload, true)
							})
							ui.Text(c, "限速单位为字节/秒，0 表示不限速。\n分享率 -1 表示无限做种，0 表示完成后停止。").FontSize(10).TextColor(p.muted)

						case 2:
							settingGroup(c, func() {
								ui.Column(c).Padding(15, 0).Gap(10).Children(func() { field(c, "监听端口（0 自动分配）", &w.listenPort) })
								switchRow(c, "启用 DHT", "仅用于公开种子", &w.settings.DHT, true, true)
								switchRow(c, "启用节点交换 PEX", "仅用于公开种子", &w.settings.PEX, true, true)
							})
							ui.Text(c, "连接设置在下次启动生效。未知磁力链接仅使用指定 Tracker 或节点获取元数据。").FontSize(11).TextColor(p.muted)
						case 3:
							settingGroup(c, func() {
								switchRow(c, "下载完成时通知", "完成后，轻轻提醒你", &w.settings.NotifyComplete, false, true)
								switchRow(c, "任务异常时通知", "需要处理时及时提醒", &w.settings.NotifyError, true, true)
							})
							ui.Text(c, "可在系统设置 → 通知中管理 Rivu。通知是否送达取决于系统授权和专注模式。").FontSize(11).TextColor(p.muted)

						case 4:
							w.updateSettingsView(c, p)

						}
					})
				})
				ui.Row(c).Gap(12).Children(func() {
					ui.Text(c, w.status).FontSize(10).TextColor(p.muted).Grow(1).MinWidth(0).MaxLines(2)
					if actionButton(c, "check", "保存设置", true).FontSize(12).Disabled(w.task.Running()).Clicked() {
						w.saveSettings()
					}
				})
			})
		})
	})
}
func (w *window) saveSettings() {
	s := w.settings
	s.Destination = w.destination
	var err error
	s.Concurrent, err = strconv.Atoi(w.concurrent)
	if err != nil {
		w.report("并行数量无效", err)
		return
	}
	s.SeedRatio, err = strconv.ParseFloat(w.seedRatio, 64)
	if err != nil {
		w.report("分享率无效", err)
		return
	}
	s.ListenPort, err = strconv.Atoi(w.listenPort)
	if err != nil {
		w.report("端口无效", err)
		return
	}
	for _, v := range []struct {
		text  string
		value *int64
	}{{w.downloadLimit, &s.DownloadLimit}, {w.uploadLimit, &s.UploadLimit}, {w.slowDownload, &s.SlowDownload}, {w.slowUpload, &s.SlowUpload}} {
		*v.value, err = strconv.ParseInt(v.text, 10, 64)
		if err != nil {
			w.report("限速数值无效", err)
			return
		}
	}
	if err = s.Validate(); err != nil {
		w.report("设置无效", err)
		return
	}
	w.run("保存设置", func(context.Context) (string, error) {
		if w.app != nil {
			w.app.settingsMu.Lock()
			defer w.app.settingsMu.Unlock()
		}
		old := w.store.Snapshot()
		if old.Login != s.Login {
			if err := mygo.App.SetOpenAtLogin(s.Login); err != nil {
				return "", fmt.Errorf("登录项注册失败: %w", err)
			}
		}
		if err := w.store.SavePreferences(s); err != nil {
			if old.Login != s.Login {
				_ = mygo.App.SetOpenAtLogin(old.Login)
			}
			return "", err
		}
		if w.app != nil {
			if err := w.app.session.Configure(w.store.Snapshot()); err != nil {
				return "", err
			}
		}
		return "设置已保存；连接设置将在下次启动生效", nil
	}, func() {
		w.settings = w.store.Snapshot()
		applyTheme(w.settings.Theme)
		if w.app != nil {
			w.app.configureTray()
			if main := w.app.window; main != nil && !main.closed {
				main.dispatch(func() { main.settings = w.settings; main.refreshDisk() })
			}
		}
		w.loginStatus = fmt.Sprintf("系统当前启用：%t", mygo.App.OpenAtLogin())
	})
}
func (w *window) saveQuitPreference() {
	w.run("保存退出偏好", func(context.Context) (string, error) {
		return "已保存", w.store.Update(func(p *config.Settings) { p.ConfirmQuit = false })
	}, w.app.shutdown)
}

func settingField(c *ui.Context, label string, value *string, line bool) {
	row := ui.Row(c).Padding(12, 0).Gap(10)
	if line {
		row.BorderWidth(1, 0, 0, 0).BorderColor(c.Theme().Border)
	}
	row.Children(func() {
		ui.Text(c, label).FontSize(12)
		ui.Spacer(c)
		ui.TextInput(c, value).Label(label).Width(145).FontSize(11)
	})
}
