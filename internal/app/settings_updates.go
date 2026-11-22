package app

import (
	"context"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/updater"
	"github.com/ntnyq/rivu/locales"
)

func (w *settingsWindow) updateSettingsView(c *ui.Context, p palette) {
	ui.Row(c).Gap(13).Children(func() {
		ui.Image(c, appMark).Size(48, 48)
		ui.Column(c).Gap(7).Children(func() {
			ui.Text(c, "Rivu").FontSize(16).Bold()
			ui.Text(c, locales.AboutVersion(version)).FontSize(11).TextColor(p.muted)
		})
	})
	s := updater.Snapshot{State: "idle"}
	available := w.app != nil && w.app.updater != nil
	if available {
		s = w.app.updater.Snapshot()
	}
	labels := map[string]string{"idle": locales.UpdatesNotChecked(), "checking": locales.UpdatesChecking(), "unavailable": locales.UpdatesUnavailable(), "upToDate": locales.UpdatesUpToDate(), "available": locales.UpdatesVersionAvailable(s.Version), "installing": locales.UpdatesInstalling(), "restartRequired": locales.UpdatesRestartRequired(), "failed": locales.UpdatesCheckFailed(), "installFailed": locales.UpdatesInstallFailed()}
	settingGroup(c, func() {
		ui.Column(c).Padding(17, 0).Gap(11).Children(func() {
			ui.Text(c, labels[s.State]).FontSize(12).Bold()
			if s.Error != "" && s.State != "installFailed" {
				ui.Text(c, s.Error).FontSize(11).TextColor(p.muted)
			}
			if s.State == "idle" {
				ui.Text(c, locales.UpdatesCheckHint()).FontSize(11).TextColor(p.muted)
			}
			if !s.LastSuccess.IsZero() {
				ui.Text(c, locales.UpdatesLastChecked(s.LastSuccess.Format("2006-01-02 15:04"))).FontSize(10).TextColor(p.muted)
			}
			if hasUpdateDetails(s.State) && actionButton(c, "download", locales.UpdatesView(), true).FontSize(11).Disabled(w.task.Running()).Clicked() {
				w.app.openUpdateWindow()
			}
		})
	})
	settingGroup(c, func() {
		switchRow(c, locales.UpdatesAutoCheck(), locales.UpdatesAutoCheckHint(), &w.settings.AutoCheck, false, true)
		switchRow(c, locales.UpdatesAutoInstall(), locales.UpdatesAutoInstallHint(), &w.settings.AutoInstall, true, w.settings.AutoCheck)
	})
	if !w.settings.AutoCheck {
		w.settings.AutoInstall = false
	}
	ui.Row(c).Gap(8).Wrap().Justify(ui.SpaceBetween).Children(func() {
		if actionButton(c, "refresh", locales.UpdatesCheck(), false).FontSize(11).Disabled(!available || w.task.Running() || s.State == "installing" || s.State == "restartRequired").Clicked() {
			w.run(locales.UpdatesCheck(), func(ctx context.Context) (string, error) {
				return locales.UpdatesCheckComplete(), w.app.updater.Check(ctx)
			}, func() {
				if hasUpdateDetails(w.app.updater.Snapshot().State) {
					w.app.openUpdateWindow()
				}
			})
		}
		ui.Row(c).Gap(8).Shrink(0).Children(func() {
			if actionButton(c, "feedback", locales.CommonFeedback(), false).FontSize(11).Clicked() {
				w.openURL(repository + "/issues/new/choose")
			}
			if actionButton(c, "globe", locales.CommonWebsite(), false).FontSize(11).Clicked() {
				w.openURL(website)
			}
			if actionButton(c, "github", "GitHub", false).FontSize(11).Clicked() {
				w.openURL(repository)
			}
		})
	})
	ui.Text(c, locales.UpdatesInstallHint()).FontSize(11).TextColor(p.muted)
}
