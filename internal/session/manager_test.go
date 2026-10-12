package session

import (
	"context"
	"errors"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"os"
	"path/filepath"
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
func exec(t *testing.T, m *Manager, kind string, ids ...string) {
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
	if m.Execute(context.Background(), Command{Kind: "pause", IDs: []string{id}}) == nil {
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
	if m.Execute(context.Background(), Command{Kind: "pauseAll"}) == nil {
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
	st, _ := os.Stat(filepath.Join(dest, "payload.txt"))
	identity, err := files.FileIdentity(st)
	if err != nil {
		t.Fatal(err)
	}
	if err = files.SaveOwnership(filepath.Join(stateDir, "fixture.torrent.files.json"), map[string]string{"payload.txt": identity}); err != nil {
		t.Fatal(err)
	}
	if err := m.Execute(context.Background(), Command{Kind: "remove", IDs: []string{id}, Trash: true}); err != nil {
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
	m.holds[b] = "disk missing"
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
