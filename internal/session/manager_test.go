package session

import (
	"context"
	"errors"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

type fakeEngine struct {
	mu      sync.Mutex
	running map[string]bool
}

func (e *fakeEngine) Attach(r Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running[r.ID] = true
	return nil
}
func (e *fakeEngine) Start(string, []File) error { return nil }
func (e *fakeEngine) Stop(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.running, id)
	return nil
}
func (e *fakeEngine) Observe(string) Observation { return Observation{Phase: "waiting"} }
func (e *fakeEngine) Limits(int64, int64)        {}
func (e *fakeEngine) Close() error               { return nil }
func manager(t *testing.T) (*Manager, string) {
	t.Helper()
	dir := t.TempDir()
	e := &fakeEngine{running: map[string]bool{}}
	m, err := Open(dir, e, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.cancel(); <-m.done })
	return m, dir
}
func add(t *testing.T, m *Manager, name string, run bool) string {
	t.Helper()
	id, err := m.Import(context.Background(), Import{Source: Source{Raw: "magnet:fixture", Name: name, Identity: Identity{V1: name}}, Destination: t.TempDir(), Run: run, SeedRatio: 2})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func exec(t *testing.T, m *Manager, kind CommandKind, ids ...string) {
	t.Helper()
	if err := m.Execute(context.Background(), Command{Kind: kind, IDs: ids}); err != nil {
		t.Fatal(err)
	}
}
func TestBatchPauseRecoveryAndManualIntent(t *testing.T) {
	m, dir := manager(t)
	a := add(t, m, "a", true)
	b := add(t, m, "b", false)
	c := add(t, m, "c", true)
	exec(t, m, "pauseAll")
	exec(t, m, "pause", c)
	exec(t, m, "resumeBatch")
	s := m.Snapshot()
	for _, task := range s.Tasks {
		want := task.ID == a
		if task.Run != want {
			t.Fatalf("wrong run for %s", task.ID)
		}
	}
	if err := m.Quiesce(); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(dir, &fakeEngine{running: map[string]bool{}}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	for _, task := range restored.Snapshot().Tasks {
		if task.ID == b && task.Run {
			t.Fatal("manual pause lost")
		}
	}
}
func TestDuplicateAndSnapshotIsolation(t *testing.T) {
	m, _ := manager(t)
	id := add(t, m, "same", false)
	if got := add(t, m, "same", true); got != id {
		t.Fatal("duplicate imported")
	}
	s := m.Snapshot()
	s.Tasks[0].Source.Name = "changed"
	if m.Snapshot().Tasks[0].Source.Name == "changed" {
		t.Fatal("snapshot aliases manager")
	}
}
func TestFailedPauseStaysStopped(t *testing.T) {
	m, dir := manager(t)
	id := add(t, m, "a", true)
	m.mu.Lock()
	m.path = filepath.Join(dir, "blocked", "session.json")
	_ = os.WriteFile(filepath.Join(dir, "blocked"), []byte("file"), 0o600)
	m.mu.Unlock()
	if m.Execute(context.Background(), Command{Kind: CommandPause, IDs: []string{id}}) == nil {
		t.Fatal("save should fail")
	}
	s := m.Snapshot()
	if s.Tasks[0].Phase != "error" || s.Error == "" {
		t.Fatal("missing unsaved-state error")
	}
	m.mu.Lock()
	active := m.active[id]
	m.mu.Unlock()
	if active {
		t.Fatal("failed pause left payload running")
	}
	if m.Close() == nil {
		t.Fatal("quit allowed after failed save")
	}
}
func TestCorruptSessionPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	_ = os.WriteFile(path, []byte("broken"), 0o600)
	m, err := Open(dir, &fakeEngine{running: map[string]bool{}}, config.Default(), nil)
	if err == nil {
		t.Fatal("accepted corrupt session")
	}
	defer func() { m.cancel(); <-m.done }()
	if m.Execute(context.Background(), Command{Kind: CommandPauseAll}) == nil {
		t.Fatal("mutation allowed")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "broken" {
		t.Fatal("overwritten")
	}
}
func TestRemoveRetainsFilesAndTrashFailureIsRecoverable(t *testing.T) {
	m, stateDir := manager(t)
	id := add(t, m, "a", false)
	var dest string
	m.mu.Lock()
	for i := range m.state.Tasks {
		r := &m.state.Tasks[i]
		if r.ID == id {
			dest = r.Destination
			r.Metadata = filepath.Join(stateDir, "fixture.torrent")
			r.Source.Files = []File{{Path: "payload.txt", Size: 4, Priority: 1}}
		}
	}
	m.trash = func(string) error { return errors.New("denied") }
	m.mu.Unlock()
	_ = os.WriteFile(filepath.Join(dest, "payload.txt"), []byte("data"), 0o644)
	file, err := os.Open(filepath.Join(dest, "payload.txt"))
	if err != nil {
		t.Fatal(err)
	}
	identity, err := files.FileIdentity(file)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err = files.SaveOwnership(filepath.Join(stateDir, "fixture.torrent.files.json"), map[string]string{"payload.txt": identity}); err != nil {
		t.Fatal(err)
	}
	if err := m.Execute(context.Background(), Command{Kind: CommandRemove, IDs: []string{id}, Trash: true}); err != nil {
		t.Fatal(err)
	}
	if len(m.Snapshot().Tasks) != 1 || m.Snapshot().Tasks[0].Error == "" {
		t.Fatal("failed trash disappeared")
	}
	m.mu.Lock()
	m.trash = func(path string) error { return os.Remove(path) }
	m.tick()
	m.mu.Unlock()
	if len(m.Snapshot().Tasks) != 0 {
		t.Fatal("removal not retried")
	}
}

func TestResumeIncompleteTaskKeepsSeedingLimit(t *testing.T) {
	m, _ := manager(t)
	id := add(t, m, "a", false)
	exec(t, m, "resume", id)
	if m.Snapshot().Tasks[0].SeedOverride {
		t.Fatal("ordinary resume disabled seed ratio")
	}
}
func TestBatchResumeDoesNotClearOtherError(t *testing.T) {
	m, _ := manager(t)
	a := add(t, m, "a", true)
	b := add(t, m, "b", false)
	exec(t, m, "pauseAll")
	m.mu.Lock()
	for i := range m.state.Tasks {
		if m.state.Tasks[i].ID == b {
			m.state.Tasks[i].Run = true
		}
	}
	m.holds[b] = errors.New("disk missing")
	m.mu.Unlock()
	exec(t, m, "resumeBatch")
	for _, v := range m.Snapshot().Tasks {
		if v.ID == a && !v.Run {
			t.Fatal("batch not resumed")
		}
		if v.ID == b && v.Error != "disk missing" {
			t.Fatal("unrelated hold cleared")
		}
	}
}

func TestTrashValidatesOwnedPathComponents(t *testing.T) {
	for _, kind := range []string{"regular", "case alias", "symlink parent"} {
		t.Run(kind, func(t *testing.T) {
			if kind == "case alias" && runtime.GOOS != "windows" {
				t.Skip("Windows path casing")
			}
			m, stateDir := manager(t)
			id := add(t, m, "owned", false)
			record := m.Snapshot().Tasks[0]
			name, actual := "payload.txt", "payload.txt"
			if kind == "case alias" {
				actual = "Payload.txt"
			}
			if kind == "symlink parent" {
				if err := os.Mkdir(filepath.Join(record.Destination, "real"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("real", filepath.Join(record.Destination, "link")); err != nil {
					if runtime.GOOS == "windows" {
						t.Skipf("symlink privilege unavailable: %v", err)
					}
					t.Fatal(err)
				}
				name, actual = "link/payload.txt", "real/payload.txt"
			}
			full := filepath.Join(record.Destination, filepath.FromSlash(actual))
			if err := os.WriteFile(full, []byte("keep"), 0600); err != nil {
				t.Fatal(err)
			}
			file, err := os.Open(full)
			if err != nil {
				t.Fatal(err)
			}
			identity, err := files.FileIdentity(file)
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
			metadata := filepath.Join(stateDir, "fixture.torrent")
			if err := files.SaveOwnership(metadata+".files.json", map[string]string{name: identity}); err != nil {
				t.Fatal(err)
			}
			called := 0
			m.mu.Lock()
			m.state.Tasks[0].Metadata = metadata
			m.state.Tasks[0].Source.Files = []File{{Path: name, Size: 4, Priority: 1}}
			m.trash = func(path string) error { called++; return os.Remove(path) }
			m.mu.Unlock()
			if err := m.Execute(t.Context(), Command{Kind: CommandRemove, IDs: []string{id}, Trash: true}); err != nil {
				t.Fatal(err)
			}
			tasks := m.Snapshot().Tasks
			m.mu.Lock()
			calls := called
			m.mu.Unlock()
			if kind == "symlink parent" {
				if len(tasks) != 1 || tasks[0].Error == "" || calls != 0 {
					t.Fatalf("unsafe trash accepted: tasks=%v calls=%d", tasks, calls)
				}
				if data, err := os.ReadFile(full); err != nil || string(data) != "keep" {
					t.Fatalf("linked payload changed: %q, %v", data, err)
				}
			} else if len(tasks) != 0 || calls != 1 {
				t.Fatalf("owned payload removal failed: tasks=%v calls=%d", tasks, calls)
			}
		})
	}
}
