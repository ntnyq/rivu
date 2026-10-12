package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ntnyq/rivu/internal/files"
)

type observedEngine struct {
	fakeEngine
	observation Observation
	peak        int
}

func (e *observedEngine) Attach(r Record) error {
	if err := e.fakeEngine.Attach(r); err != nil {
		return err
	}
	e.peak = max(e.peak, len(e.running))
	return nil
}
func (e *observedEngine) Observe(string) Observation { return e.observation }
func observedManager(t *testing.T, o Observation) *Manager {
	m, _ := manager(t)
	m.cancel()
	<-m.done
	m.engine = &observedEngine{fakeEngine: fakeEngine{running: map[string]bool{}}, observation: o}
	return m
}
func addPayload(t *testing.T, m *Manager, name string) string {
	t.Helper()
	id, err := m.Import(context.Background(), Import{Source: Source{Name: name, Identity: Identity{V1: name}, Meta: []byte("fixture"), Files: []File{{Path: name, Size: 10, Priority: 1}}}, Destination: t.TempDir(), Run: true, SeedRatio: 2})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func TestMetadataQueueYields(t *testing.T) {
	m := observedManager(t, Observation{Phase: "metadata"})
	a := add(t, m, "a", true)
	b := add(t, m, "b", true)
	c := add(t, m, "c", true)
	for _, id := range []string{a, b} {
		s := m.samples[id]
		s.since = time.Now().Add(-10 * time.Minute)
		m.samples[id] = s
	}
	m.tick()
	if !m.active[c] {
		t.Fatal("third magnet remains queued behind two metadata tasks stalled for ten minutes")
	}
}
func TestConnectedStallYields(t *testing.T) {
	m := observedManager(t, Observation{Phase: "downloading", Peers: 1})
	m.settings.Concurrent = 1
	a := addPayload(t, m, "a")
	b := addPayload(t, m, "b")
	s := m.samples[a]
	s.since = time.Now().Add(-10 * time.Minute)
	m.samples[a] = s
	m.tick()
	if !m.active[b] {
		t.Fatal("queued payload starves while first task has a peer but receives zero bytes")
	}
}
func TestContinueHonorsLoadError(t *testing.T) {
	m := observedManager(t, Observation{Phase: "metadata"})
	a := add(t, m, "a", true)
	_ = m.stop(a)
	m.loadErr = errors.New("uncertain commit requires reload")
	m.barrier = true
	m.Continue()
	if m.active[a] || !m.barrier {
		t.Fatal("Continue bypassed persistence recovery barrier")
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if !m.active[a] {
		t.Fatal("explicit reload did not restore admission")
	}
}
func TestFailedRehashClearsCompleted(t *testing.T) {
	m := observedManager(t, Observation{Phase: "waiting", Files: []File{{Path: "a", Size: 10, Done: 5}}})
	id := addPayload(t, m, "a")
	m.state.Tasks[0].Completed = true
	m.tick()
	if m.Snapshot().Tasks[0].Completed {
		t.Error("incomplete rehash retained completed flag")
	}
	exec(t, m, "pause", id)
	if got := m.Snapshot().Tasks[0].Done; got != 5 {
		t.Errorf("pause changed verified progress from 5/10 to %d/10", got)
	}
}
func TestOnlyLatestPauseBatchResumes(t *testing.T) {
	m := observedManager(t, Observation{Phase: "metadata"})
	a := add(t, m, "a", true)
	exec(t, m, "pauseAll")
	_ = add(t, m, "b", true)
	exec(t, m, "pauseAll")
	exec(t, m, "resumeBatch")
	for _, task := range m.Snapshot().Tasks {
		if task.ID == a && task.Run {
			t.Fatal("Resume Batch resumed a task from an older pause-all batch")
		}
	}
}

func TestSkippedFilesDoNotReserveFullSize(t *testing.T) {
	m := observedManager(t, Observation{})
	dir := t.TempDir()
	free, err := files.FreeSpace(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Import(context.Background(), Import{Source: Source{Name: "partial", PieceLength: 16384, Identity: Identity{V1: "partial"}, Meta: []byte("fixture"), Files: []File{{Path: "selected", Size: 1, Priority: 1}, {Path: "skipped", Offset: 1, Size: free + 1, Priority: 0}}}, Destination: dir, SeedRatio: 2})
	if err != nil {
		t.Fatalf("one-byte selection rejected because skipped file exceeds free space: %v", err)
	}
}

func TestMetadataLimitSurvivesPriorityChange(t *testing.T) {
	m := observedManager(t, Observation{Phase: "metadata"})
	_ = add(t, m, "a", true)
	_ = add(t, m, "b", true)
	c := add(t, m, "c", true)
	exec(t, m, "first", c)
	if got := m.engine.(*observedEngine).peak; got > 2 {
		t.Fatalf("priority change left %d active metadata handles; limit is two", got)
	}
}

func TestSchedulerKeepsVerifyingSeedingAndProgressingTasks(t *testing.T) {
	for _, phase := range []string{"verifying", "seeding", "downloading"} {
		t.Run(phase, func(t *testing.T) {
			observation := Observation{Phase: phase, Files: []File{{Path: "a", Size: 10}}}
			if phase == "seeding" {
				observation.Files[0].Done = 10
			}
			m := observedManager(t, observation)
			m.settings.Concurrent = 1
			a := addPayload(t, m, "a")
			_ = addPayload(t, m, "b")
			s := m.samples[a]
			s.since = time.Now().Add(-10 * time.Minute)
			m.samples[a] = s
			if phase == "downloading" {
				m.engine.(*observedEngine).observation.Downloaded = 1
			}
			m.tick()
			if !m.active[a] {
				t.Fatal("active work was treated as stalled")
			}
		})
	}
}
func TestPauseDuringRehashPreservesObservedProgress(t *testing.T) {
	m := observedManager(t, Observation{Phase: "verifying", Files: []File{{Path: "a", Size: 10, Done: 5}}})
	id := addPayload(t, m, "a")
	m.state.Tasks[0].Completed = true
	m.tick()
	exec(t, m, "pause", id)
	task := m.Snapshot().Tasks[0]
	if task.Completed || task.Done != 5 {
		t.Fatalf("published unverified completion: %+v", task)
	}
}
func TestUnchangedRehashDoesNotRepeatCompletionNotification(t *testing.T) {
	m := observedManager(t, Observation{Phase: "verifying", Files: []File{{Path: "a", Size: 10, Done: 5}}})
	_ = addPayload(t, m, "a")
	m.state.Tasks[0].Completed = true
	m.state.Tasks[0].Epoch = 1
	m.state.Tasks[0].Notified = 1
	m.tick()
	m.engine.(*observedEngine).observation = Observation{Phase: "waiting", Files: []File{{Path: "a", Size: 10, Done: 10}}}
	m.tick()
	task := m.Snapshot().Tasks[0]
	if !task.Completed || task.Epoch != 1 {
		t.Fatalf("rehash generated another completion: %+v", task)
	}
}
func TestTickHonorsMaintenanceAndRecovery(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		m := observedManager(t, Observation{Phase: "metadata"})
		id := add(t, m, "a", true)
		_ = m.stop(id)
		if recovery {
			m.loadErr = errors.New("reload required")
		} else {
			m.barrier = true
		}
		m.tick()
		if m.active[id] {
			t.Fatal("tick bypassed admission barrier")
		}
	}
}
