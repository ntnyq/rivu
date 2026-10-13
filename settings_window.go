package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
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
	w.loginStatus = fmt.Sprintf(tr("系统当前启用：%t（详细授权状态由系统设置显示）"), mygo.App.OpenAtLogin())
	w.win = mygo.NewWindow(mygo.WindowOptions{Title: tr("Rivu 设置"), TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 50, Width: 750, Height: 630, MinWidth: 640, MinHeight: 520, Content: ui.View(w.settingsView)})
	w.desktop.Parent = w.win
	w.win.OnClosed(w.close)
	a.settingsWindow = w
}
func (w *settingsWindow) pickFolder() {
	var folder string
	w.run(tr("选择下载目录"), func(context.Context) (string, error) {
		var err error
		folder, err = w.desktop.PickFolder()
		return tr("已选择目录"), err
	}, func() { w.destination = filepath.Clean(folder) })
}
