package main

import (
	"context"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/locales"
)

type settingsWindow struct {
	*windowBase
	store                                                *config.Store
	settings                                             config.Settings
	settingsTab                                          int
	loginStatus                                          string
	destination, concurrent, seedRatio, listenPort       string
	downloadLimit, uploadLimit, slowDownload, slowUpload bandwidthInput
	seedMode                                             int
	seedHours, scheduleStart, scheduleEnd                string
}

func newSettingsWindow(store *config.Store) *settingsWindow {
	w := &settingsWindow{windowBase: &windowBase{}, store: store, settings: store.Snapshot()}
	w.resetSettings()
	return w
}
func (a *application) openSettings() {
	if a.settingsWindow != nil && !a.settingsWindow.closed {
		a.settingsWindow.win.Show()
		a.settingsWindow.win.Focus()
		return
	}
	w := newSettingsWindow(a.store)
	w.app = a
	w.loginStatus = locales.SettingsLoginStatusDetails(mygo.App.OpenAtLogin())
	w.win = mygo.NewWindow(mygo.WindowOptions{Title: locales.SettingsTitle(), TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 50, Width: 750, Height: 630, MinWidth: 640, MinHeight: 520, Content: ui.View(w.settingsView)})
	w.desktop.Parent = w.win
	w.win.OnClosed(w.close)
	a.settingsWindow = w
}
func (w *settingsWindow) pickFolder() {
	var folder string
	w.run(locales.FilesChooseDownloadFolder(), func(context.Context) (string, error) {
		var err error
		folder, err = w.desktop.PickFolder()
		return locales.FilesFolderSelected(), err
	}, func() { w.destination = filepath.Clean(folder) })
}
