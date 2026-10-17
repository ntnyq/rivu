package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/ntnyq/rivu/locales"
)

// messageError keeps the original error available to errors.Is/As while its
// display text is formatted by a typed catalog accessor.
type messageError struct {
	message string
	cause   error
}

func (e messageError) Error() string { return e.message }
func (e messageError) Unwrap() error { return e.cause }

// Autonyms remain recognizable even if the user selects an unfamiliar language.
func languageName(code string) string {
	if code == "system" {
		return locales.CommonFollowSystem()
	}
	return i18n.Autonym(code)
}

// applyLanguage runs on the main thread, after preferences have been persisted.
func (a *application) applyLanguage() {
	mygo.App.SetLocale(localePreference(a.store.Snapshot().Language))
}

func localePreference(preference string) string {
	if preference == "system" {
		return ""
	}
	return preference
}

// refreshLanguage also handles changes to the system's preferred languages.
func (a *application) refreshLanguage() {
	mygo.App.SetMenu(a.menu())
	if w := a.window; w != nil && !w.closed {
		w.dispatch(func() {
			if !w.failed && !w.task.Running() {
				w.status = ""
			}
		})
	}
	if w := a.settingsWindow; w != nil && !w.closed {
		if w.win != nil {
			w.win.SetTitle(locales.SettingsTitle())
		}
		w.dispatch(func() {
			w.loginStatus = locales.SettingsLoginStatusDetails(mygo.App.OpenAtLogin())
		})
	}
	if d := a.add; d != nil && !d.closed {
		if d.win != nil {
			d.win.SetTitle(locales.TransferAdd())
		}
		d.dispatch(func() {})
	}
	// A tray menu is captured by its native right-click handler. Recreate it so
	// every entry, including Linux's persistent menu, uses the new language.
	a.closePanel()
	if a.tray != nil {
		a.tray.Destroy()
		a.tray = nil
	}
	a.configureTray()
}
