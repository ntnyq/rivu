package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistenceAndMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"theme":"dark","recentFiles":["demo"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Snapshot().Version != 2 || s.Snapshot().Theme != "dark" || s.Snapshot().Concurrent != 3 {
		t.Fatal(s.Snapshot())
	}
	if err = s.Update(func(p *Settings) { p.SlowMode = true }); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil || !s.Snapshot().SlowMode {
		t.Fatal(err)
	}
}
func TestInvalidSettingsArePreserved(t *testing.T) {
	for _, data := range []string{`{broken`, `{"version":3}`, `{"version":2,"theme":"neon"}`, `{"version":2,"concurrent":0}`} {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			_ = os.WriteFile(path, []byte(data), 0o600)
			s, err := Open(path)
			if err == nil {
				t.Fatal("accepted invalid settings")
			}
			if s.Update(func(p *Settings) { p.Theme = "dark" }) == nil {
				t.Fatal("overwrote invalid file")
			}
			got, _ := os.ReadFile(path)
			if string(got) != data {
				t.Fatal("file changed")
			}
		})
	}
}
func TestFailedSaveDoesNotPublish(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(filepath.Join(dir, "child", "settings.json"))
	_ = os.WriteFile(filepath.Join(dir, "child"), []byte("file"), 0o600)
	if s.Update(func(p *Settings) { p.Theme = "dark" }) == nil {
		t.Fatal("expected failure")
	}
	if s.Snapshot().Theme != "system" {
		t.Fatal("published failed write")
	}
}

func TestSavePreferencesPreservesConcurrentRuntimeChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	draft := store.Snapshot()
	draft.Theme = "dark"
	if err = store.Update(func(s *Settings) {
		s.SlowMode = true
		s.LastCheck = time.Now().UTC()
		s.LastAttempt = s.LastCheck.Add(-time.Minute)
	}); err != nil {
		t.Fatal(err)
	}
	current := store.Snapshot()
	if err = store.SavePreferences(draft); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.Snapshot()
	if got.Theme != "dark" || !got.SlowMode || !got.LastCheck.Equal(current.LastCheck) || !got.LastAttempt.Equal(current.LastAttempt) {
		t.Fatalf("draft overwrote runtime preferences: %+v", got)
	}
	down, up := got.Limits()
	if down != got.SlowDownload || up != got.SlowUpload {
		t.Fatal("saved settings disabled slow limits")
	}
}

func TestLanguagePreferencePersistence(t *testing.T) {
	for _, code := range []string{"system", "zh-Hans", "zh-Hant", "en", "ja", "ko"} {
		t.Run(code, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			// Existing v2 files without a language remain valid and follow the system.
			if err := os.WriteFile(path, []byte(`{"version":2,"theme":"dark"}`), 0600); err != nil {
				t.Fatal(err)
			}
			store, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if store.Snapshot().Language != "system" {
				t.Fatal("old preferences did not follow system")
			}
			draft := store.Snapshot()
			draft.Language = code
			if err := store.SavePreferences(draft); err != nil {
				t.Fatal(err)
			}
			reloaded, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if reloaded.Snapshot().Language != code {
				t.Fatal("language not persisted")
			}
			if err := store.Update(func(s *Settings) { s.Language = "unknown" }); err == nil {
				t.Fatal("accepted invalid language")
			}
			if store.Snapshot().Language != code {
				t.Fatal("failed write changed language")
			}
		})
	}
}
