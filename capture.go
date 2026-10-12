package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/egoist/mygo"
)

// An opt-in native smoke check, similar to godiff's capture workflow. It waits
// for initial I/O, captures the actual native content, then exits the app.
func (w *window) captureIfRequested() {
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
		data, err := w.win.CapturePage()
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
