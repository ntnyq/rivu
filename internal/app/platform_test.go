package app

import (
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/locales"
)

func TestPlatformWindowDecorations(t *testing.T) {
	for _, goos := range []string{"darwin", "windows", "linux"} {
		t.Run(goos, func(t *testing.T) {
			main, settings := mainWindowOptions(goos), settingsWindowOptions(goos)
			if goos == "darwin" {
				if main.TitleBarStyle != mygo.TitleBarHiddenInset || main.TitleBarHeight != 38 || main.TrafficLightPosition == nil || *main.TrafficLightPosition != (mygo.Point{X: 18, Y: 12}) || settings.TitleBarStyle != mygo.TitleBarHiddenInset || settings.TitleBarHeight != 50 {
					t.Fatal("macOS window decoration changed")
				}
			} else {
				for _, opts := range []mygo.WindowOptions{main, settings} {
					if opts.TitleBarStyle != mygo.TitleBarDefault || opts.TitleBarHeight != 0 || opts.TrafficLightPosition != nil {
						t.Fatalf("native controls would overlay application content: %+v", opts)
					}
				}
			}
		})
	}
}

func TestPlatformMenusAndBackgroundPolicy(t *testing.T) {
	useTestLanguage(t, "en")
	app := &application{}
	for _, goos := range []string{"darwin", "windows", "linux"} {
		items := app.menuItems(goos)
		if goos == "darwin" {
			if items[0].Role != mygo.RoleAppMenu || len(items) != 6 {
				t.Fatal("macOS menu changed")
			}
		} else {
			if items[0].Role == mygo.RoleAppMenu {
				t.Fatal("application commands hidden by a macOS-only menu")
			}
			for _, label := range []string{locales.MenuSettings(), locales.MenuQuit(), locales.MenuAbout()} {
				found := false
				for _, item := range items[0].Submenu {
					if item.Label == label && item.Role == "" && item.Click != nil {
						found = true
					}
				}
				if !found {
					t.Errorf("%s: %s has no visible action", goos, label)
				}
			}
		}
		if !canRunInBackground(goos, true) {
			t.Errorf("%s cannot restore from working tray", goos)
		}
		if canRunInBackground(goos, false) != (goos == "darwin") {
			t.Errorf("%s can become unreachable without a tray", goos)
		}
		want := "Ctrl+,"
		if goos == "darwin" {
			want = "⌘,"
		}
		if shortcutHint(goos, ",") != want {
			t.Errorf("%s settings shortcut", goos)
		}
	}
}

func TestNativeTitleBarContentBounds(t *testing.T) {
	for _, code := range []string{"zh-Hans", "zh-Hant", "en", "ja", "ko"} {
		t.Run(code, func(t *testing.T) {
			useTestLanguage(t, code)
			store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			for _, width := range []int{880, 1180} {
				for _, collapsed := range []bool{false, true} {
					w := newWindow(store)
					w.snapshot = fixtureSnapshot()
					w.sidebarHidden = collapsed
					tt := ui.NewTester(w.view, width, 560)
					// Default native decorations live outside the content area, on either side.
					tt.SetTitleBar(ui.TitleBar{})
					tt.Frame()
					for _, label := range []string{locales.TransferAdd(), locales.TransferSearch(), locales.TransferToggleDetails()} {
						bounds, ok := tt.Find(label)
						if !ok || bounds.X < 0 || bounds.Y < 0 || bounds.X+bounds.W > float32(width) || bounds.Y+bounds.H > 69 {
							t.Errorf("%s %d collapsed=%v: toolbar control clipped: %s %+v", code, width, collapsed, label, bounds)
						}
					}
					if !collapsed {
						bounds, ok := tt.Find(locales.SidebarCollapse())
						if !ok || bounds.Y+bounds.H > 69 {
							t.Errorf("sidebar kept an extra titlebar row: %+v", bounds)
						}
					}
				}
			}
		})
	}
}
