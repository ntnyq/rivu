package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/internal/tasks"
)

type window struct {
	diskFree, diskTotal                                                                                     int64
	app                                                                                                     *application
	win                                                                                                     *mygo.Window
	desktop                                                                                                 desktop.Service
	store                                                                                                   *config.Store
	settings                                                                                                config.Settings
	snapshot                                                                                                session.Snapshot
	sidebarHidden                                                                                           bool
	selected                                                                                                map[string]bool
	anchor                                                                                                  string
	filter, sort, tab, settingsTab                                                                          int
	search, status, query                                                                                   string
	failed, closed, inspector, paletteOpen, settingsOnly, removeOpen, trashFiles, quitConfirm, rememberQuit bool
	task                                                                                                    tasks.Task
	job                                                                                                     string
	removeIDs                                                                                               []string
	loginStatus                                                                                             string
	destination, concurrent, seedRatio, downloadLimit, uploadLimit, slowDownload, slowUpload, listenPort    string
}

func newWindow(store *config.Store) *window {
	w := &window{store: store, settings: store.Snapshot(), selected: map[string]bool{}, inspector: true, status: "准备就绪"}
	w.resetSettings()
	return w
}
func (a *application) openWindow() *window {
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
func (a *application) openSettings() {
	if a.settingsWindow != nil && !a.settingsWindow.closed {
		a.settingsWindow.win.Show()
		a.settingsWindow.win.Focus()
		return
	}
	w := newWindow(a.store)
	w.app = a
	w.settingsOnly = true
	w.loginStatus = fmt.Sprintf("系统当前启用：%t（详细授权状态由系统设置显示）", mygo.App.OpenAtLogin())
	w.win = mygo.NewWindow(mygo.WindowOptions{Title: "Rivu 设置", TitleBarStyle: mygo.TitleBarHiddenInset, TitleBarHeight: 50, Width: 750, Height: 630, MinWidth: 640, MinHeight: 520, Content: ui.View(w.settingsView)})
	w.desktop.Parent = w.win
	w.win.OnClosed(w.close)
	a.settingsWindow = w
}
func (w *window) close() {
	if w.closed {
		return
	}
	w.closed = true
	w.task.Close()
}
func (w *window) dispatch(fn func()) {
	if w.win == nil {
		if !w.closed {
			fn()
		}
		return
	}
	w.win.Update(func() {
		if !w.closed {
			fn()
		}
	})
}
func (w *window) report(message string, err error) {
	w.failed = err != nil
	w.status = message
	if err != nil {
		w.status += "：" + err.Error()
	}
}
func (w *window) run(label string, work func(context.Context) (string, error), apply func()) {
	if w.task.Running() {
		return
	}
	w.job = label
	w.report(label+"…", nil)
	w.task.Start(w.dispatch, func(ctx context.Context, _ func(float64)) (string, error) { return work(ctx) }, func(float64) {}, func(message string, err error) {
		w.job = ""
		if errors.Is(err, desktop.ErrCanceled) || errors.Is(err, context.Canceled) {
			w.report("已取消", nil)
			return
		}
		if err != nil {
			w.report(label, err)
			return
		}
		if apply != nil {
			apply()
		}
		w.report(message, nil)
	})
}
func (w *window) selectOnly(id string) { w.selected = map[string]bool{id: true}; w.anchor = id }
func (w *window) ids() []string {
	var ids []string
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] {
			ids = append(ids, t.ID)
		}
	}
	return ids
}
func (w *window) chosen() *session.Task {
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] {
			return &t
		}
	}
	return nil
}
func (w *window) toggleSelected() {
	ids := w.ids()
	if len(ids) == 0 || w.app == nil {
		return
	}
	kind := "resume"
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] && t.Run {
			kind = "pause"
		}
	}
	w.app.execute(session.Command{Kind: kind, IDs: ids})
}
func (w *window) remove() {
	w.removeIDs = w.ids()
	w.trashFiles = false
	w.removeOpen = len(w.removeIDs) > 0
}
func (w *window) reveal(path string) {
	w.run("在文件管理器中显示", func(context.Context) (string, error) { return "已显示文件", w.desktop.Reveal(path) }, nil)
}
func (w *window) copy(text string) {
	w.run("复制", func(context.Context) (string, error) { return "已复制", w.desktop.Copy(text) }, nil)
}
func (w *window) openURL(raw string) {
	w.run("打开网页", func(context.Context) (string, error) { return "已打开浏览器", w.desktop.OpenURL(raw) }, nil)
}
func applyTheme(theme string) { mygo.Theme.SetSource(mygo.ThemeSource(theme)) }
func (w *window) pickFolder() {
	var folder string
	w.run("选择下载目录", func(context.Context) (string, error) {
		var err error
		folder, err = w.desktop.PickFolder()
		return "已选择目录", err
	}, func() { w.destination = filepath.Clean(folder) })
}

func (w *window) refreshDisk() {
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
