package app

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/locales"
)

func TestLocalePreference(t *testing.T) {
	for _, tc := range []struct{ preference, want string }{
		{"system", ""}, {"en", "en"}, {"zh-Hant", "zh-Hant"},
	} {
		if got := localePreference(tc.preference); got != tc.want {
			t.Errorf("%q override = %q; want %q", tc.preference, got, tc.want)
		}
	}
}

func TestLocalizedErrorPreservesCause(t *testing.T) {
	useTestLanguage(t, "ja")
	cause := errors.New("permission denied")
	err := messageError{locales.SettingsReadFailed(cause), cause}
	if !errors.Is(err, cause) || err.Error() != locales.In("ja").SettingsReadFailed(cause) {
		t.Fatalf("message or cause was lost: %v", err)
	}
}

func useTestLanguage(t *testing.T, code string) {
	t.Helper()
	old := i18n.Locale()
	i18n.SetLocaleFunc(func() string { return code })
	t.Cleanup(func() { i18n.SetLocaleFunc(func() string { return old }) })
}

func TestLocalizedSurfaces(t *testing.T) {
	for _, code := range []string{"zh-Hans", "zh-Hant", "en", "ja", "ko"} {
		t.Run(code, func(t *testing.T) {
			useTestLanguage(t, code)
			store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			settings := newSettingsWindow(store)
			w := newWindow(store)
			w.snapshot = fixtureSnapshot()
			w.selectOnly("1")
			app := &application{store: store, snapshot: w.snapshot}
			add := &addWindow{windowBase: &windowBase{}, settings: store.Snapshot(), destination: store.Snapshot().Destination, start: true}
			for _, surface := range []struct {
				name          string
				view          func(*ui.Context)
				width, height int
				label         string
			}{
				{"main", w.view, 1180, 760, locales.FilterAll()},
				{"main-minimum", w.view, 880, 560, locales.FilterAll()},
				{"settings", settings.settingsView, 750, 630, locales.SettingsLanguageLabel()},
				{"settings-minimum", settings.settingsView, 640, 520, locales.SettingsLanguageLabel()},
				{"add", add.view, 550, 660, locales.TransferAdd()},
				{"tray", app.trayView, trayWidth, presentTray(app.snapshot).height(), locales.TrayOpenWindow()},
			} {
				tt := ui.NewTester(surface.view, surface.width, surface.height)
				tt.SetLocale(code)
				tt.Frame()
				if tt.HasText("Translation unavailable") {
					t.Errorf("%s: missing translation reached the interface", surface.name)
				}
				if _, ok := tt.Find(surface.label); !ok {
					t.Errorf("%s: missing %q", surface.name, surface.label)
				}
				if surface.name == "settings" && !tt.HasText(locales.CommonFollowSystem()) {
					t.Error("default language option missing")
				}
				if surface.name == "tray" {
					for _, label := range []string{locales.TrayOpenWindow(), locales.TransferAdd(), locales.SettingsLabel(), locales.TransferSpeedLimit()} {
						r, ok := tt.Find(label)
						if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > float32(surface.width) || r.Y+r.H > float32(surface.height) {
							t.Errorf("%s clipped: %+v", label, r)
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
			if got := menu.Items()[2].Label; got != locales.MenuEdit() {
				t.Errorf("edit menu: %s", got)
			}
			item := app.menuItem("settings")
			if item.Label != locales.MenuSettings() || item.Accelerator != "CmdOrCtrl+," {
				t.Fatal("settings command changed")
			}
			tt := ui.NewTester(w.view, 1180, 760)
			tt.SetLocale(code)
			tt.Frame()
			if err := tt.Click(locales.PhasePaused()); err != nil {
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
	tt.SetLocale("zh-Hans")
	tt.Frame()
	if !tt.HasText("全部任务") {
		t.Fatal("initial language")
	}
	i18n.SetLocaleFunc(func() string { return "ja" })
	tt.SetLocale("ja")
	tt.Frame()
	if !tt.HasText("すべての転送") || tt.HasText("全部任务") {
		t.Fatal("existing view did not change language")
	}
}
