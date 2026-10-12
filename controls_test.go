package main

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
)

func TestSortMenuSelection(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(store)
	w.snapshot = fixtureSnapshot()
	tt := ui.NewTester(w.view, 1180, 760)
	for i, label := range []string{"添加时间", "名称", "进度", "状态"} {
		before := w.sort
		if err := tt.Click("任务排序"); err != nil {
			t.Fatal(err)
		}
		if w.sort != before {
			t.Fatal("opening the menu changed the sort order")
		}
		if err := tt.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		if w.sort != i || !tt.HasText("排序："+label) {
			t.Fatalf("sort selection %s was not applied", label)
		}
	}
	if err := tt.Click("任务排序"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Key(0, ui.KeySpace)
	if len(tt.Menu()) != 4 {
		t.Fatal("sort menu did not open from the keyboard")
	}
}

func TestSettingsLanguageMenu(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(store)
	tt := ui.NewTester(w.settingsView, 640, 520)
	for _, code := range []string{"en", "ja", "ko", "zh-Hant", "zh-Hans", "system"} {
		if err := tt.Click("界面语言"); err != nil {
			t.Fatal(err)
		}
		for _, option := range []string{"跟随系统", "简体中文", "繁體中文", "English", "日本語", "한국어"} {
			if !slices.Contains(tt.Menu(), option) {
				t.Fatalf("missing language option %s", option)
			}
		}
		if err := tt.ChooseMenuItem(languageName(code)); err != nil {
			t.Fatal(err)
		}
		if w.settings.Language != code || !tt.HasText(languageName(code)) {
			t.Fatalf("language selection %s was not applied", code)
		}
		if store.Snapshot().Language != "system" {
			t.Fatal("language was persisted before saving settings")
		}
	}
}
