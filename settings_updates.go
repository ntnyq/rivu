package main

import (
	"context"
	"fmt"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/updater"
)

func (w *settingsWindow) updateSettingsView(c *ui.Context, p palette) {
	ui.Row(c).Gap(13).Children(func() {
		ui.Image(c, appMark).Size(48, 48)
		ui.Column(c).Gap(7).Children(func() {
			ui.Text(c, "Rivu").FontSize(16).Bold()
			ui.Text(c, fmt.Sprintf(tr("about.version"), version)).FontSize(11).TextColor(p.muted)
		})
	})
	s := updater.Snapshot{State: "idle"}
	available := w.app != nil && w.app.updater != nil
	if available {
		s = w.app.updater.Snapshot()
	}
	labels := map[string]string{"idle": tr("updates.notChecked"), "checking": tr("updates.checking"), "unavailable": tr("updates.unavailable"), "upToDate": tr("updates.upToDate"), "available": fmt.Sprintf(tr("updates.versionAvailable"), s.Version), "installing": tr("updates.installing"), "restartRequired": tr("updates.restartRequired"), "failed": tr("updates.checkFailed"), "installFailed": tr("updates.installFailed")}
	settingGroup(c, func() {
		ui.Column(c).Padding(17, 0).Gap(11).Children(func() {
			ui.Text(c, labels[s.State]).FontSize(12).Bold()
			if s.Error != "" {
				ui.Text(c, s.Error).FontSize(11).TextColor(p.muted)
			}
			if s.Notes != "" {
				ui.Text(c, s.Notes).FontSize(11).TextColor(p.muted).MaxLines(5)
			}
			if s.State == "idle" {
				ui.Text(c, tr("updates.checkHint")).FontSize(11).TextColor(p.muted)
			}
			if !s.LastSuccess.IsZero() {
				ui.Text(c, fmt.Sprintf(tr("updates.lastChecked"), s.LastSuccess.Format("2006-01-02 15:04"))).FontSize(10).TextColor(p.muted)
			}
			if s.State == "installing" && s.Total > 0 {
				progressTrack(c, float64(s.Downloaded)/float64(s.Total), p.blue, p.track, 3)
			}
			if (s.State == "available" || s.State == "installFailed") && actionButton(c, "download", tr("updates.downloadAndInstall"), true).FontSize(11).Disabled(w.task.Running()).Clicked() {
				w.run(tr("updates.install"), func(ctx context.Context) (string, error) {
					return tr("updates.installed"), w.app.updater.Install(ctx, false)
				}, nil)
			}
			if s.State == "restartRequired" && actionButton(c, "refresh", tr("updates.saveAndRestart"), true).FontSize(11).Clicked() {
				w.app.relaunch = true
				mygo.App.Quit()
			}
		})
	})
	settingGroup(c, func() {
		switchRow(c, tr("updates.autoCheck"), tr("updates.autoCheckHint"), &w.settings.AutoCheck, false, true)
		switchRow(c, tr("updates.autoInstall"), tr("updates.autoInstallHint"), &w.settings.AutoInstall, true, w.settings.AutoCheck)
	})
	if !w.settings.AutoCheck {
		w.settings.AutoInstall = false
	}
	ui.Row(c).Gap(8).Children(func() {
		if actionButton(c, "refresh", tr("updates.check"), false).FontSize(11).Disabled(!available || w.task.Running()).Clicked() {
			w.run(tr("updates.check"), func(ctx context.Context) (string, error) {
				return tr("updates.checkComplete"), w.app.updater.Check(ctx)
			}, nil)
		}
		ui.Spacer(c)
		if actionButton(c, "globe", tr("common.website"), false).FontSize(11).Clicked() {
			w.openURL(website)
		}
		if actionButton(c, "github", "GitHub", false).FontSize(11).Clicked() {
			w.openURL(repository)
		}
	})
	ui.Text(c, tr("updates.installHint")).FontSize(11).TextColor(p.muted)
}
