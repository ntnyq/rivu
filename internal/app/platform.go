package app

import (
	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/locales"
)

func mainWindowOptions(goos string) mygo.WindowOptions {
	opts := mygo.WindowOptions{
		Title: appName, Width: 1180, Height: 760, MinWidth: 880, MinHeight: 560, StateKey: "main",
		TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 38,
		TrafficLightPosition: &mygo.Point{X: 18, Y: 12},
	}
	if goos != "darwin" {
		opts.TitleBarStyle = mygo.TitleBarDefault
		opts.TitleBarHeight = 0
		opts.TrafficLightPosition = nil
	}
	return opts
}

func settingsWindowOptions(goos string) mygo.WindowOptions {
	opts := mygo.WindowOptions{Title: locales.SettingsTitle(), TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 50, Width: 750, Height: 630, MinWidth: 640, MinHeight: 520}
	if goos != "darwin" {
		opts.TitleBarStyle = mygo.TitleBarDefault
		opts.TitleBarHeight = 0
	}
	return opts
}

// macOS has Dock activation; other desktops need a working tray to restore windows.
func canRunInBackground(goos string, trayAvailable bool) bool {
	return goos == "darwin" || trayAvailable
}

func shortcutHint(goos, key string) string {
	if goos == "darwin" {
		if key == "," {
			return "⌘,"
		}
		return "⌘ " + key
	}
	return "Ctrl+" + key
}
