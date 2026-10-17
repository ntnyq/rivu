package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func fixtureSnapshot() session.Snapshot {
	now := time.Date(2026, 10, 12, 10, 24, 0, 0, time.Local)
	tasks := []session.Task{
		{Record: session.Record{ID: "1", Source: session.Source{Name: "Ubuntu 24.04.1 Desktop"}, Run: true, Uploaded: 1550000000}, Phase: "downloading", Size: 5900000000, Done: 3680000000, DownRate: 8400000, UpRate: 420000, Peers: 24},
		{Record: session.Record{ID: "2", Source: session.Source{Name: "Blender Open Movies"}, Run: true, Uploaded: 320000000}, Phase: "downloading", Size: 12400000000, Done: 4740000000, DownRate: 4200000, UpRate: 330000, Peers: 18},
		{Record: session.Record{ID: "3", Source: session.Source{Name: "Debian 12.7.0 arm64"}, Run: true, Completed: true, Uploaded: 806000000}, Phase: "seeding", Size: 650000000, Done: 650000000, UpRate: 450000, Peers: 8},
		{Record: session.Record{ID: "4", Source: session.Source{Name: "Sintel · Open Movie"}, Completed: true, Uploaded: 2560000000}, Phase: "completed", Size: 1280000000, Done: 1280000000},
		{Record: session.Record{ID: "5", Source: session.Source{Name: "LibreOffice 24.8 Archive"}}, Phase: "paused", Size: 980000000, Done: 212000000},
		{Record: session.Record{ID: "6", Source: session.Source{Name: "Fedora Workstation 40"}, Run: true}, Phase: "error", Size: 2140000000, Done: 317000000, Error: "下载磁盘空间不足"},
	}
	for i := range tasks {
		t := &tasks[i]
		t.TotalSize = t.Size
		t.Metadata = "fixture"
		t.SeedRatio = 2
		t.Destination = "/Users/demo/Downloads/Rivu"
		t.Added = now.Add(-time.Duration(i+1) * time.Hour)
		t.Source.Files = []session.File{{Path: t.Source.Name, Size: t.Size, Priority: 1}}
		t.Files = []session.File{{Path: t.Source.Name, Size: t.Size, Done: t.Done, Priority: 1}}
	}
	return session.Snapshot{DownRate: 12600000, UpRate: 1200000, Tasks: tasks}
}
func TestFilterSelectionAndEmptyState(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	s, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	w := newWindow(s)
	w.snapshot = fixtureSnapshot()
	w.diskFree = 128000000000
	w.diskTotal = 512000000000
	tt := ui.NewTester(w.view, 1180, 760)
	tt.Frame()
	if err := tt.Click("已暂停"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if w.filter != 4 || len(w.visible()) != 1 {
		t.Fatal("filter failed")
	}
	w.search = "not found"
	tt.Frame()
	if !tt.HasText("没有匹配的任务") {
		t.Fatal("missing empty state")
	}
	w.search = ""
	w.selectOnly("3")
	if len(w.ids()) != 1 {
		t.Fatal("selection failed")
	}
}

func TestDetailsRequireSelection(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{880, 1180} {
		w := newWindow(store)
		tt := ui.NewTester(w.view, width, 760)
		tt.Frame()
		if w.inspector || tt.HasText(locales.TransferSelectHint()) {
			t.Fatal("details opened without a selection")
		}
		if err := tt.Click(locales.TransferToggleDetails()); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if w.inspector {
			t.Fatal("empty library allowed details to open")
		}
		w.snapshot = fixtureSnapshot()
		w.selectOnly("1")
		tt.Frame()
		if tt.HasText(locales.TransferOverview()) {
			t.Fatal("selection opened details before they were requested")
		}
		if err := tt.Click(locales.TransferToggleDetails()); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if !tt.HasText(locales.TransferOverview()) {
			t.Fatal("selected task details did not open")
		}
		w.snapshot = session.Snapshot{}
		tt.Frame()
		if tt.HasText(locales.TransferOverview()) || tt.HasText(locales.TransferSelectHint()) {
			t.Fatal("details remained visible after the selected task was removed")
		}
	}
}

func TestViewSnapshots(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	s, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	for _, dark := range []bool{false, true} {
		for _, size := range [][2]int{{1280, 818}, {1180, 760}, {880, 560}} {
			for _, state := range []string{"tasks", "empty", "collapsed", "fullscreen"} {
				w := newWindow(s)
				w.snapshot = fixtureSnapshot()
				w.diskFree = 128000000000
				w.diskTotal = 512000000000
				w.selectOnly("1")
				w.inspector = state == "tasks"
				if state != "tasks" {
					w.snapshot = session.Snapshot{}
					w.sidebarHidden = state == "collapsed"
				}
				if size[0] < 1060 {
					w.inspector = false
				}
				tt := ui.NewTester(w.view, size[0], size[1])
				if state != "fullscreen" {
					tt.SetTitleBar(ui.TitleBar{Height: 38, Left: 104})
				}
				tt.SetDark(dark)
				tt.Frame()
				if !tt.HasText("全部任务") {
					t.Fatal("missing task screen")
				}
				if dir := os.Getenv("RIVU_SNAPSHOTS"); dir != "" {
					_ = os.MkdirAll(dir, 0o755)
					theme := "light"
					if dark {
						theme = "dark"
					}
					prefix := ""
					if state != "tasks" {
						prefix = state + "-"
					}
					f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s%s-%d.png", prefix, theme, size[0])))
					if err != nil {
						t.Fatal(err)
					}
					if err = png.Encode(f, tt.Image()); err != nil {
						t.Fatal(err)
					}
					f.Close()
				}
			}
		}
	}
}
func TestPresentation(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	s := fixtureSnapshot()
	if !matches(s.Tasks[2], 3) || matches(s.Tasks[4], 1) {
		t.Fatal("incorrect overlapping counts")
	}
	if progress(s.Tasks[0]) <= 0 || eta(s.Tasks[4]) != "—" {
		t.Fatal("invalid progress")
	}
}

func TestAuxiliaryViewSnapshots(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	settings := newSettingsWindow(store)
	app := &application{store: store, snapshot: fixtureSnapshot()}
	emptyApp := &application{store: store}
	add := &addWindow{windowBase: &windowBase{}, settings: store.Snapshot(), destination: store.Snapshot().Destination, start: true, input: "Ubuntu 24.04.1 Desktop.torrent", preparedInput: "Ubuntu 24.04.1 Desktop.torrent", free: 128000000000, items: []importItem{{source: session.Source{Name: "Ubuntu 24.04.1 Desktop", Files: []session.File{{Path: "ubuntu-24.04.1-desktop-arm64.iso", Size: 5900000000, Priority: 1}}}}}}
	for _, dark := range []bool{false, true} {
		for _, v := range []struct {
			name          string
			view          func(*ui.Context)
			width, height int
		}{
			{"settings-general", func(c *ui.Context) { settings.settingsTab = 0; settings.settingsView(c) }, 750, 630},
			{"settings-updates", func(c *ui.Context) { settings.settingsTab = 4; settings.settingsView(c) }, 750, 630},
			{"settings-downloads", func(c *ui.Context) { settings.settingsTab = 1; settings.settingsView(c) }, 750, 630},
			{"settings-automation", func(c *ui.Context) { settings.settingsTab = 5; settings.settingsView(c) }, 750, 630},
			{"add-download", add.view, 550, 660},
			{"menu-bar", app.trayView, trayWidth, presentTray(app.snapshot).height()},
			{"menu-bar-empty", emptyApp.trayView, trayWidth, presentTray(emptyApp.snapshot).height()},
			{"menu-bar-material", func(c *ui.Context) { app.trayView(c) }, trayWidth, presentTray(app.snapshot).height()},
		} {
			tt := ui.NewTester(v.view, v.width, v.height)
			tt.SetDark(dark)
			if v.name == "menu-bar-material" {
				tt.SetVibrancy(true)
			}
			tt.Frame()
			if dir := os.Getenv("RIVU_SNAPSHOTS"); dir != "" {
				theme := "light"
				if dark {
					theme = "dark"
				}
				f, err := os.Create(filepath.Join(dir, v.name+"-"+theme+".png"))
				if err != nil {
					t.Fatal(err)
				}
				if err = png.Encode(f, tt.Image()); err != nil {
					t.Fatal(err)
				}
				f.Close()
			}
		}
	}
}

// Cover the native-material path too: a software-only opaque snapshot previously
// hid the fully transparent background that appeared in the macOS popover.
func TestTrayLayoutAndMaterial(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	app := &application{store: store, snapshot: fixtureSnapshot()}
	for _, dark := range []bool{false, true} {
		for _, snapshot := range []session.Snapshot{fixtureSnapshot(), {}, {Tasks: fixtureSnapshot().Tasks[5:]}} {
			app.snapshot = snapshot
			preview := presentTray(snapshot)
			tt := ui.NewTester(app.trayView, trayWidth, preview.height())
			tt.SetDark(dark)
			tt.SetVibrancy(true)
			if alpha := tt.Image().RGBAAt(4, 40).A; alpha < 240 {
				t.Fatalf("popover material has insufficient tint: alpha=%d", alpha)
			}
			for _, label := range []string{"打开主窗口", "添加下载", "设置", "限速模式"} {
				r, ok := tt.Find(label)
				if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > trayWidth || r.Y+r.H > float32(preview.height()) {
					t.Fatalf("%s clipped or missing at height %d: %+v", label, preview.height(), r)
				}
			}
			if len(snapshot.Tasks) == 6 {
				if !tt.HasText("LibreOffice 24.8 Archive") || !tt.HasText("8.4 MB/s") || !tt.HasText("1 个任务需要处理") {
					t.Fatal("missing paused preview, transfer rate, or attention link")
				}
				if tt.HasText("Sintel") || tt.HasText("Fedora") {
					t.Fatal("completed or error task included in preview")
				}
			}
		}
	}
}
