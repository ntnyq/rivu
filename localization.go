package main

import (
	"fmt"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/i18n"
)

func tr(message string) string { return i18n.T(message) }

// Autonyms remain recognizable even if the user selects an unfamiliar language.
func languageName(code string) string {
	switch code {
	case "zh-Hans":
		return "简体中文"
	case "zh-Hant":
		return "繁體中文"
	case "en":
		return "English"
	case "ja":
		return "日本語"
	case "ko":
		return "한국어"
	default:
		return tr("跟随系统")
	}
}

// applyLanguage runs on the main thread, after preferences have been persisted.
func (a *application) applyLanguage() {
	i18n.SetLanguage(a.store.Snapshot().Language, mygo.App.Locale())
	mygo.App.SetMenu(a.menu())
	if w := a.window; w != nil && !w.closed {
		w.dispatch(func() {
			if !w.failed && !w.task.Running() {
				w.status = ""
			}
		})
	}
	if w := a.settingsWindow; w != nil && !w.closed {
		w.win.SetTitle(tr("Rivu 设置"))
		w.dispatch(func() {
			w.loginStatus = fmt.Sprintf(tr("系统当前启用：%t（详细授权状态由系统设置显示）"), mygo.App.OpenAtLogin())
		})
	}
	if d := a.add; d != nil && !d.closed {
		d.win.SetTitle(tr("添加下载"))
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
