package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
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
func TestViewSnapshots(t *testing.T) {
	s, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	for _, dark := range []bool{false, true} {
		for _, size := range [][2]int{{1180, 760}, {880, 560}} {
			w := newWindow(s)
			w.snapshot = fixtureSnapshot()
			w.diskFree = 128000000000
			w.diskTotal = 512000000000
			w.selectOnly("1")
			if size[0] < 1060 {
				w.inspector = false
			}
			tt := ui.NewTester(w.view, size[0], size[1])
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
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s-%d.png", theme, size[0])))
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
func TestPresentation(t *testing.T) {
	s := fixtureSnapshot()
	if !matches(s.Tasks[2], 3) || matches(s.Tasks[4], 1) {
		t.Fatal("incorrect overlapping counts")
	}
	if progress(s.Tasks[0]) <= 0 || eta(s.Tasks[4]) != "—" {
		t.Fatal("invalid progress")
	}
}

func TestAuxiliaryViewSnapshots(t *testing.T) {
	store, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	w := newWindow(store)
	app := &application{store: store, snapshot: fixtureSnapshot()}
	add := &addWindow{window: newWindow(store), start: true, input: "Ubuntu 24.04.1 Desktop.torrent", preparedInput: "Ubuntu 24.04.1 Desktop.torrent", free: 128000000000, items: []importItem{{source: session.Source{Name: "Ubuntu 24.04.1 Desktop", Files: []session.File{{Path: "ubuntu-24.04.1-desktop-arm64.iso", Size: 5900000000, Priority: 1}}}}}}
	for _, dark := range []bool{false, true} {
		for _, v := range []struct {
			name          string
			view          func(*ui.Context)
			width, height int
		}{
			{"settings-general", func(c *ui.Context) { w.settingsTab = 0; w.settingsView(c) }, 750, 630},
			{"settings-updates", func(c *ui.Context) { w.settingsTab = 4; w.settingsView(c) }, 750, 630},
			{"add-download", add.view, 550, 660},
			{"menu-bar", app.trayView, 360, 490},
		} {
			tt := ui.NewTester(v.view, v.width, v.height)
			tt.SetDark(dark)
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
