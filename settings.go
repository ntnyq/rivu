package main

import (
	"fmt"
	"strconv"

	"github.com/egoist/mygo/ui"
)

func (w *settingsWindow) resetSettings() {
	s := w.settings
	w.destination = s.Destination
	w.concurrent = strconv.Itoa(s.Concurrent)
	w.seedRatio = strconv.FormatFloat(s.SeedRatio, 'f', 1, 64)
	w.downloadLimit = bandwidthValue(s.DownloadLimit)
	w.uploadLimit = bandwidthValue(s.UploadLimit)
	w.slowDownload = bandwidthValue(s.SlowDownload)
	w.slowUpload = bandwidthValue(s.SlowUpload)
	w.listenPort = strconv.Itoa(s.ListenPort)
	w.seedMode = 1
	if s.SeedRatio == 0 {
		w.seedMode = 0
	}
	if s.SeedRatio < 0 {
		w.seedMode = 2
	}
	w.seedHours = strconv.FormatFloat(float64(s.SeedLimitSeconds)/3600, 'f', -1, 64)
	w.scheduleStart = fmt.Sprintf("%02d:%02d", s.ScheduleStart/60, s.ScheduleStart%60)
	w.scheduleEnd = fmt.Sprintf("%02d:%02d", s.ScheduleEnd/60, s.ScheduleEnd%60)
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
func (w *settingsWindow) settingsView(c *ui.Context) {
	p := designTheme(c)
	c.Root().Background(p.alt)
	labels := []string{tr("settings.general"), tr("settings.downloads"), tr("common.connections"), tr("settings.notifications"), tr("settings.updates"), tr("settings.automation")}
	icons := []string{"settings", "download", "globe", "bell", "refresh", "settings"}
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Height(50).Shrink(0).Justify(ui.Center).Background(p.sidebar).BorderWidth(0, 0, 1, 0).BorderColor(p.line).DragWindow().Children(func() { ui.Text(c, tr("settings.label")).FontSize(13).Bold() })
		ui.Row(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(168).Shrink(0).Padding(18, 10).Gap(3).Background(p.sidebar).BorderWidth(0, 1, 0, 0).BorderColor(p.line).Children(func() {
				for i, label := range labels {
					selected := i == w.settingsTab
					ink, ic := p.text, p.blue
					row := navigationButton(c, label, selected, p).Height(37)
					if selected {
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
									ui.Text(c, tr("settings.appearance")).FontSize(12)
									ui.Spacer(c)
									labels := map[string]string{"system": tr("common.followSystem"), "light": tr("settings.theme.light"), "dark": tr("settings.theme.dark")}
									menuControl(c, tr("settings.appearance"), labels[w.settings.Theme]).Width(145).Menu(func(m *ui.Menu) {
										for _, v := range []string{"system", "light", "dark"} {
											if m.Item(labels[v]).Checked(w.settings.Theme == v).Chosen() {
												w.settings.Theme = v
											}
										}
									})
								})
								ui.Row(c).Padding(15, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Children(func() {
									ui.Text(c, tr("settings.language")).FontSize(12)
									ui.Spacer(c)
									menuControl(c, tr("settings.languageLabel"), languageName(w.settings.Language)).Width(145).Menu(func(m *ui.Menu) {
										for _, code := range []string{"system", "zh-Hans", "zh-Hant", "en", "ja", "ko"} {
											if m.Item(languageName(code)).Checked(w.settings.Language == code).Chosen() {
												w.settings.Language = code
											}
										}
									})
								})
								switchRow(c, tr("settings.tray"), tr("settings.trayHint"), &w.settings.Tray, true, true)
							})
							settingGroup(c, func() {
								switchRow(c, tr("settings.login.enable"), "", &w.settings.Login, false, true)
								switchRow(c, tr("settings.login.background"), tr("settings.login.backgroundHint"), &w.settings.BackgroundLogin, true, w.settings.Login)
								switchRow(c, tr("settings.restore"), tr("settings.restoreHint"), &w.settings.Restore, true, true)
							})
							settingGroup(c, func() {
								switchRow(c, tr("settings.confirmQuit"), tr("settings.confirmQuitHint"), &w.settings.ConfirmQuit, false, true)
							})
							ui.Text(c, tr("settings.languageHint")).FontSize(11).TextColor(p.muted)
							ui.Text(c, tr("settings.backgroundHint")).FontSize(11).TextColor(p.muted)
							if w.loginStatus != "" {
								ui.Text(c, w.loginStatus).FontSize(10).TextColor(p.muted)
							}

						case 1:
							settingGroup(c, func() {
								ui.Column(c).Padding(15, 0).Gap(10).Children(func() {
									ui.Text(c, tr("settings.destination")).FontSize(12)
									ui.Row(c).Gap(8).Children(func() {
										ui.TextInput(c, &w.destination).Label(tr("settings.destinationLabel")).FontSize(11).Grow(1).MinWidth(0)
										if actionButton(c, "folder", tr("common.choose"), false).FontSize(11).Clicked() {
											w.pickFolder()
										}
									})
								})
								settingField(c, tr("settings.concurrency"), &w.concurrent, true)
								seedingControl(c, &w.seedMode)
								if w.seedMode == 1 {
									settingField(c, tr("settings.seedRatio"), &w.seedRatio, true)
								}
								settingField(c, tr("settings.seedHours"), &w.seedHours, true)
							})
							settingGroup(c, func() {
								bandwidthField(c, tr("settings.downloadLimit"), &w.downloadLimit, false)
								bandwidthField(c, tr("settings.uploadLimit"), &w.uploadLimit, true)
								bandwidthField(c, tr("settings.temporaryDownloadLimit"), &w.slowDownload, true)
								bandwidthField(c, tr("settings.temporaryUploadLimit"), &w.slowUpload, true)
							})
							ui.Text(c, tr("settings.bandwidthHint")).FontSize(10).TextColor(p.muted)

						case 2:
							settingGroup(c, func() {
								ui.Column(c).Padding(15, 0).Gap(10).Children(func() { field(c, tr("settings.listenPort"), &w.listenPort) })
								switchRow(c, tr("settings.dht"), tr("settings.publicOnly"), &w.settings.DHT, true, true)
								switchRow(c, tr("settings.pex"), tr("settings.publicOnly"), &w.settings.PEX, true, true)
							})
							ui.Text(c, tr("settings.connectionsHint")).FontSize(11).TextColor(p.muted)
						case 3:
							settingGroup(c, func() {
								switchRow(c, tr("settings.notifyComplete"), tr("settings.notifyCompleteHint"), &w.settings.NotifyComplete, false, true)
								switchRow(c, tr("settings.notifyError"), tr("settings.notifyErrorHint"), &w.settings.NotifyError, true, true)
							})
							ui.Text(c, tr("settings.notificationsHint")).FontSize(11).TextColor(p.muted)

						case 4:
							w.updateSettingsView(c, p)
						case 5:
							w.automationView(c)

						}
					})
				})
				ui.Row(c).Gap(12).Children(func() {
					ui.Text(c, w.status).FontSize(10).TextColor(p.muted).Grow(1).MinWidth(0).MaxLines(2)
					if actionButton(c, "check", tr("settings.save"), true).FontSize(12).Disabled(w.task.Running()).Clicked() {
						w.saveSettings()
					}
				})
			})
		})
	})
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
