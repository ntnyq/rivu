package main

import (
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
)

type mainWindow struct {
	listState ui.ListState
	*windowBase
	store                                                                     *config.Store
	settings                                                                  config.Settings
	snapshot                                                                  session.Snapshot
	diskFree, diskTotal                                                       int64
	sidebarHidden                                                             bool
	selected                                                                  map[string]bool
	anchor                                                                    string
	filter, sort, tab                                                         int
	search, query                                                             string
	tagFilter, tagsDraft, tagsTask                                            string
	fileState                                                                 fileTreeState
	relocateID, relocatePath                                                  string
	relocateOpen                                                              bool
	inspector, paletteOpen, removeOpen, trashFiles, quitConfirm, rememberQuit bool
	removeIDs                                                                 []string
}

func newWindow(store *config.Store) *mainWindow {
	return &mainWindow{windowBase: &windowBase{}, store: store, settings: store.Snapshot(), selected: map[string]bool{}, inspector: true}
}
func (a *application) openWindow() *mainWindow {
	if a.window != nil && !a.window.closed {
		a.window.win.Show()
		a.window.win.Focus()
		return a.window
	}
	w := newWindow(a.store)
	w.app = a
	w.snapshot = a.snapshot
	w.win = mygo.NewWindow(mygo.WindowOptions{Title: appName, Width: 1180, Height: 760, MinWidth: 880, MinHeight: 560, StateKey: "main", TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 58, Content: ui.View(w.view)})
	w.desktop.Parent = w.win
	w.win.OnClosed(w.close)
	a.window = w
	w.refreshDisk()
	return w
}
func (w *mainWindow) refreshDisk() {
	destination := w.settings.Destination
	go func() {
		free, total, err := files.DiskSpace(destination)
		if err != nil {
			free, total, err = files.DiskSpace(filepath.Dir(destination))
		}
		if err == nil {
			w.dispatch(func() { w.diskFree = free; w.diskTotal = total })
		}
	}()
}
