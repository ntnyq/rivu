package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/i18n"
)

func useTestLanguage(t *testing.T, code string) {
	t.Helper()
	old := i18n.Language()
	i18n.SetLanguage(code, "en-US")
	t.Cleanup(func() { i18n.SetLanguage(old, "") })
}

func TestLocalizedSurfaces(t *testing.T) {
	for _, code := range []string{"zh-Hans", "zh-Hant", "en", "ja", "ko"} {
		t.Run(code, func(t *testing.T) {
			useTestLanguage(t, code)
			store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			w := newWindow(store)
			w.snapshot = fixtureSnapshot()
			w.selectOnly("1")
			app := &application{store: store, snapshot: w.snapshot}
			add := &addWindow{window: newWindow(store), start: true}
			for _, surface := range []struct {
				name          string
				view          func(*ui.Context)
				width, height int
				label         string
			}{
				{"main", w.view, 1180, 760, "全部任务"},
				{"main-minimum", w.view, 880, 560, "全部任务"},
				{"settings", w.settingsView, 750, 630, "界面语言"},
				{"settings-minimum", w.settingsView, 640, 520, "界面语言"},
				{"add", add.view, 550, 660, "添加下载"},
				{"tray", app.trayView, trayWidth, presentTray(app.snapshot).height(), "打开主窗口"},
			} {
				tt := ui.NewTester(surface.view, surface.width, surface.height)
				tt.Frame()
				if _, ok := tt.Find(tr(surface.label)); !ok {
					t.Errorf("%s: missing %q", surface.name, tr(surface.label))
				}
				if surface.name == "settings" && !tt.HasText(tr("跟随系统")) {
					t.Error("default language option missing")
				}
				if surface.name == "tray" {
					for _, key := range []string{"打开主窗口", "添加下载", "设置", "限速模式"} {
						r, ok := tt.Find(tr(key))
						if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > float32(surface.width) || r.Y+r.H > float32(surface.height) {
							t.Errorf("%s clipped: %+v", tr(key), r)
						}
					}
				}
				if dir := os.Getenv("RIVU_SNAPSHOTS"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(dir, fmt.Sprintf("i18n-%s-%s.png", code, surface.name)))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, tt.Image())
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			// Native menus are built without installing them, preserving role shortcuts.
			menu := app.menu()
			if got := menu.Items()[2].Label; got != tr("编辑") {
				t.Errorf("edit menu: %s", got)
			}
			item := app.menuItem("settings")
			if item.Label != tr("设置…") || item.Accelerator != "CmdOrCtrl+," {
				t.Fatal("settings command changed")
			}
			tt := ui.NewTester(w.view, 1180, 760)
			tt.Frame()
			if err := tt.Click(tr("已暂停")); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			if w.filter != 4 || len(w.visible()) != 1 {
				t.Fatal("translated filter failed")
			}
		})
	}
}

func TestLanguageChangeRefreshesExistingViews(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, _ := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	w := newWindow(store)
	tt := ui.NewTester(w.view, 1180, 760)
	tt.Frame()
	if !tt.HasText("全部任务") {
		t.Fatal("initial language")
	}
	i18n.SetLanguage("ja", "zh-CN")
	tt.Frame()
	if !tt.HasText("すべての転送") || tt.HasText("全部任务") {
		t.Fatal("existing view did not change language")
	}
}
