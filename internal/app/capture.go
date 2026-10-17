package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/egoist/mygo"
)

// An opt-in native smoke check, similar to godiff's capture workflow. It waits
// for initial I/O, captures the actual native content, then exits the app.
func (w *mainWindow) captureIfRequested() {
	path := os.Getenv("RIVU_CAPTURE")
	if path == "" {
		return
	}
	go func() {
		time.Sleep(time.Second)
		for deadline := time.Now().Add(15 * time.Second); ; {
			var busy bool
			mygo.RunOnMain(func() { busy = w.task.Running() })
			if !busy {
				break
			}
			if time.Now().After(deadline) {
				slog.Error("capture timed out waiting for initial I/O")
				mygo.App.Exit(1)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		target := w.win
		switch os.Getenv("RIVU_CAPTURE_SURFACE") {
		case "settings-downloads", "settings-automation":
			mygo.RunOnMain(func() {
				w.app.openSettings()
				settings := w.app.settingsWindow
				settings.dispatch(func() {
					settings.settingsTab = 1
					if os.Getenv("RIVU_CAPTURE_SURFACE") == "settings-automation" {
						settings.settingsTab = 5
					}
				})
				target = settings.win
			})
			time.Sleep(500 * time.Millisecond)
		}
		if os.Getenv("RIVU_CAPTURE_SURFACE") == "tray" {
			mygo.RunOnMain(func() {
				target = nil
				if w.app.tray != nil {
					if w.app.panel == nil {
						w.app.togglePanel()
					}
					target = w.app.panel
				}
			})
			if target == nil {
				slog.Error("tray capture requires the menu bar icon to be enabled")
				mygo.App.Exit(1)
				return
			}
			// Wait for AppKit's popover opening animation and its first GPU frame.
			time.Sleep(500 * time.Millisecond)
		}
		data, err := target.CapturePage()
		if err == nil {
			err = os.WriteFile(path, data, 0o600)
		}
		if err != nil {
			slog.Error("capture failed", "error", err)
			mygo.App.Exit(1)
			return
		}
		mygo.App.Quit()
	}()
}
