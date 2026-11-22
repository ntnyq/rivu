package app

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/ntnyq/rivu/locales"
)

func (w *settingsWindow) saveSettings() {
	s := w.settings
	s.Destination = w.destination
	var err error
	s.Concurrent, err = strconv.Atoi(w.concurrent)
	if err != nil {
		w.report(locales.SettingsInvalidConcurrency(), err)
		return
	}
	s.SeedRatio = 0
	if w.seedMode == 2 {
		s.SeedRatio = -1
	}
	if w.seedMode == 1 {
		s.SeedRatio, err = strconv.ParseFloat(w.seedRatio, 64)
		if err == nil && s.SeedRatio <= 0 {
			err = fmt.Errorf("ratio must be positive")
		}
	}
	if err != nil {
		w.report(locales.SettingsInvalidSeedRatio(), err)
		return
	}
	s.ListenPort, err = strconv.Atoi(w.listenPort)
	if err != nil {
		w.report(locales.SettingsInvalidPort(), err)
		return
	}
	for _, v := range []struct {
		input bandwidthInput
		value *int64
	}{{w.downloadLimit, &s.DownloadLimit}, {w.uploadLimit, &s.UploadLimit}, {w.slowDownload, &s.SlowDownload}, {w.slowUpload, &s.SlowUpload}} {
		*v.value, err = v.input.bytes()
		if err != nil {
			w.report(locales.SettingsInvalidSpeedLimit(), err)
			return
		}
	}
	s.SeedLimitSeconds, err = parseScaled(w.seedHours, 3600)
	if err != nil {
		w.report(locales.SettingsInvalid(), err)
		return
	}
	s.ScheduleStart, err = clockMinute(w.scheduleStart)
	if err != nil {
		w.report(locales.SettingsInvalid(), err)
		return
	}
	s.ScheduleEnd, err = clockMinute(w.scheduleEnd)
	if err != nil {
		w.report(locales.SettingsInvalid(), err)
		return
	}
	if err = s.Validate(); err != nil {
		w.report(locales.SettingsInvalid(), err)
		return
	}
	// Resolve against the user's languages, rather than the currently selected
	// app language, so switching back to the system works from any override.
	language := localePreference(s.Language)
	if language == "" {
		language = i18n.Match(mygo.App.PreferredLanguages(), locales.Catalog.Languages())
	}
	a := w.app
	started := w.runCommitted(locales.SettingsSave(), func() (string, error) {
		committed := false
		if a != nil {
			defer func() {
				mygo.RunOnMain(func() {
					if committed {
						a.applySettings()
					}
					a.settingsSaving = false
					if a.reopenSettings {
						a.reopenSettings = false
						a.openSettings()
					}
				})
			}()
		}
		old := w.store.Snapshot()
		if old.Login != s.Login {
			if err := mygo.App.SetOpenAtLogin(s.Login); err != nil {
				return "", messageError{locales.SettingsLoginRegistrationFailed(err), err}
			}
		}
		if err := w.store.SavePreferences(s); err != nil {
			if old.Login != s.Login {
				_ = mygo.App.SetOpenAtLogin(old.Login)
			}
			return "", err
		}
		committed = true
		if a != nil {
			if err := a.session.Configure(w.store.Snapshot()); err != nil {
				return "", err
			}
		}
		return locales.In(language).SettingsSaved(), nil
	}, func() {
		w.settings = w.store.Snapshot()
		w.loginStatus = locales.SettingsLoginStatus(mygo.App.OpenAtLogin())
	})
	if started && a != nil {
		a.settingsSaving = true
	}
}

func (a *application) applySettings() {
	if a.shutdownReady {
		return
	}
	s := a.store.Snapshot()
	applyTheme(s.Theme)
	a.applyLanguage()
	a.configureTray()
	if w := a.window; w != nil && !w.closed {
		w.dispatch(func() {
			w.settings = s
			w.refreshDisk()
		})
	}
	if !a.quitting && !a.reopenSettings && !canRunInBackground(runtime.GOOS, a.tray != nil) && len(mygo.Windows()) == 0 {
		a.openWindow()
	}
}
