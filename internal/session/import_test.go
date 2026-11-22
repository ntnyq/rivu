package session

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ntnyq/rivu/internal/config"
)

func TestWatchedImportCommitsWithLedger(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing=%t", existing), func(t *testing.T) {
			m, dir := manager(t)
			in := Import{WatchKey: "archive.torrent:fingerprint", Source: Source{Identity: Identity{V1: "archive"}}, Destination: t.TempDir()}
			var manualID string
			if existing {
				manualID = add(t, m, "archive", false)
			}
			before := len(m.Snapshot().Tasks)
			blocked := filepath.Join(dir, "blocked")
			if err := os.WriteFile(blocked, nil, 0600); err != nil {
				t.Fatal(err)
			}
			path := m.path
			m.mu.Lock()
			m.path = filepath.Join(blocked, "session.json")
			m.mu.Unlock()
			if _, err := m.Import(t.Context(), in); err == nil || m.Watched(in.WatchKey) {
				t.Fatal("failed save recorded the watch input")
			}
			m.mu.Lock()
			m.path = path
			m.mu.Unlock()
			if len(m.Snapshot().Tasks) != before {
				t.Fatal("failed save changed the task list")
			}

			id, err := m.Import(t.Context(), in)
			if err != nil || id == "" || existing && id != manualID {
				t.Fatalf("watched import returned %q, %v", id, err)
			}
			exec(t, m, CommandRemove, id)
			if err := m.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(dir, &fakeEngine{running: map[string]bool{}}, config.Default(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			if !reopened.Watched(in.WatchKey) {
				t.Fatal("watch input lost across removal and restart")
			}
			if id, err := reopened.Import(t.Context(), in); id != "" || err != nil || len(reopened.Snapshot().Tasks) != 0 {
				t.Fatalf("removed task was imported again: %q, %v", id, err)
			}
			in.WatchKey += ":changed"
			if id, err := reopened.Import(t.Context(), in); err != nil || id == "" {
				t.Fatalf("changed input was skipped: %q, %v", id, err)
			}
		})
	}
}

func TestWatchJournalMigration(t *testing.T) {
	legacy := filepath.Join(t.TempDir(), "watch-folder.json")
	if err := os.WriteFile(legacy, []byte(`{"older.torrent:fingerprint":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	newSession, _ := manager(t)
	if err := newSession.MigrateWatchJournal(legacy); err != nil || newSession.Watched("older.torrent:fingerprint") {
		t.Fatalf("new session inherited another data directory's legacy journal: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "session.json"), []byte(`{"version":2,"tasks":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	m, err := Open(dir, &fakeEngine{running: map[string]bool{}}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.MigrateWatchJournal(legacy); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if err := m.MigrateWatchJournal(legacy); err != nil || !m.Watched("older.torrent:fingerprint") {
		t.Fatalf("migrated ledger was not retained after reload: %v", err)
	}
}

func TestMetadataWriteFailureKeepsMagnetResolvable(t *testing.T) {
	m := observedManager(t, Observation{})
	id := add(t, m, "legacy", true)
	path := filepath.Join(filepath.Dir(m.path), "metadata", id+".torrent")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	engine := m.engine.(*observedEngine)
	source := Source{Raw: "magnet:fixture", Identity: Identity{V1: "legacy"}, Meta: []byte("fixture"), Files: []File{{Path: "member", Priority: 1}}}
	engine.observation.Metadata = &source
	m.lastSave = time.Time{}
	m.tick()
	if m.holds[id] == nil {
		t.Fatal("metadata write failure did not reach the task")
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := m.Snapshot().Tasks[0].Metadata; got != "" {
		t.Fatalf("checkpoint retained a failed metadata write: %q", got)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	m.state.Tasks[0].Metadata = path // Simulate the bad reference written by an older build.
	if err := m.save(m.state); err != nil {
		t.Fatal(err)
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := m.Snapshot().Tasks[0].Metadata; got != "" {
		t.Fatalf("legacy unresolved magnet was not repaired: %q", got)
	}
	m.tick()
	if got := m.Snapshot().Tasks[0].Metadata; got != path {
		t.Fatalf("magnet did not resolve after repairing storage: %q", got)
	}
}

func TestDestinationConflict(t *testing.T) {
	dir := t.TempDir()
	m := Manager{state: diskState{Tasks: []Record{{ID: "existing", Destination: dir, Source: Source{Files: []File{{Path: "Parent/File"}}}}}}}
	for _, tc := range []struct {
		name     string
		dest     string
		paths    []File
		except   string
		conflict bool
	}{
		{"same path ignoring case", dir, []File{{Path: "parent/file"}}, "", true},
		{"existing path is an ancestor", dir, []File{{Path: "parent/file-other"}, {Path: "parent/file/child"}}, "", true},
		{"incoming path is an ancestor", dir, []File{{Path: "parent"}}, "", true},
		{"nested destination", filepath.Join(dir, "parent"), []File{{Path: "file"}}, "", true},
		{"shared prefix only", dir, []File{{Path: "Parent/Filename"}}, "", false},
		{"relocating the same task", dir, []File{{Path: "Parent/File"}}, "existing", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.destinationConflict(tc.paths, tc.dest, tc.except); got != tc.conflict {
				t.Fatalf("destinationConflict = %t, want %t", got, tc.conflict)
			}
		})
	}
}

func BenchmarkCheckDestinationDisjoint(b *testing.B) {
	dir := b.TempDir()
	var existing, incoming Source
	for i := range 10000 {
		existing.Files = append(existing.Files, File{Path: fmt.Sprintf("existing/%05d", i)})
		incoming.Files = append(incoming.Files, File{Path: fmt.Sprintf("incoming/%05d", i)})
	}
	m := Manager{state: diskState{Tasks: []Record{{ID: "existing", Destination: dir, Source: existing}}}}
	b.ResetTimer()
	for b.Loop() {
		if err := m.checkDestination(incoming, dir, "", false); err != nil {
			b.Fatal(err)
		}
	}
}
