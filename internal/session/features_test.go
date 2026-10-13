package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTagsDoNotReleaseRecoveryHold(t *testing.T) {
	m, _ := manager(t)
	id := add(t, m, "hold", true)
	m.mu.Lock()
	_ = m.stop(id)
	m.holds[id] = errors.New("disk missing")
	m.publish()
	m.mu.Unlock()
	if err := m.Execute(t.Context(), Command{Kind: CommandTags, IDs: []string{id}, Tags: []string{"work"}}); err != nil {
		t.Fatal(err)
	}
	if m.Snapshot().Tasks[0].Error != "disk missing" {
		t.Fatal("tag edit released recovery hold")
	}
}

func TestSessionV1MigrationPreservesIntent(t *testing.T) {
	m, dir := manager(t)
	_ = add(t, m, "migrate", false)
	m.mu.Lock()
	m.state.Version = 1
	if err := m.save(m.state); err != nil {
		t.Fatal(err)
	}
	m.mu.Unlock()
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if m.state.Version != 2 || m.Snapshot().Tasks[0].Run {
		t.Fatal("migration lost pause intent")
	}
	if _, err := os.Stat(filepath.Join(dir, "session.json")); err != nil {
		t.Fatal(err)
	}
}

func TestAdoptionRelocationAndTrashOwnership(t *testing.T) {
	m, _ := manager(t)
	dest := t.TempDir()
	payload := filepath.Join(dest, "payload")
	if err := os.WriteFile(payload, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	in := Import{Source: Source{Identity: Identity{V1: "adopt"}, Files: []File{{Path: "payload", Size: 4, Priority: 1}}, Meta: []byte("fixture")}, Destination: dest, Tags: []string{"linux"}}
	if _, err := m.Import(t.Context(), in); err == nil {
		t.Fatal("silently adopted existing data")
	}
	in.AdoptExisting = true
	id, err := m.Import(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if s.Tasks[0].Adopted["payload"] == "" {
		t.Fatal("missing adopted identity")
	}
	s.Tasks[0].Adopted["payload"] = "mutated"
	s.Tasks[0].Tags[0] = "mutated"
	if m.Snapshot().Tasks[0].Tags[0] != "linux" || m.Snapshot().Tasks[0].Adopted["payload"] == "mutated" {
		t.Fatal("snapshot aliases state")
	}
	newDest := t.TempDir()
	newDest, err = filepath.EvalSymlinks(newDest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDest, "payload"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Execute(t.Context(), Command{Kind: CommandRelocate, IDs: []string{id}, Destination: newDest}); err != nil {
		t.Fatal(err)
	}
	if m.Snapshot().Tasks[0].Run {
		t.Fatal("relocation resumed a manually paused task")
	}
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if m.Snapshot().Tasks[0].Destination != newDest {
		t.Fatal("relocation not durable")
	}
	m.mu.Lock()
	m.trash = func(string) error { t.Error("adopted file sent to Trash"); return nil }
	m.mu.Unlock()
	if err := m.Execute(t.Context(), Command{Kind: CommandRemove, IDs: []string{id}, Trash: true}); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{dest, newDest} {
		if data, err := os.ReadFile(filepath.Join(dir, "payload")); err != nil || string(data) != "data" {
			t.Fatal("existing data changed", err)
		}
	}
	if len(m.Snapshot().Tasks) != 0 {
		t.Fatal("task not removed")
	}
}

func TestAtomicFilePriorityAndCommandValidation(t *testing.T) {
	m, _ := manager(t)
	id, err := m.Import(t.Context(), Import{Source: Source{Identity: Identity{V1: "files"}, Files: []File{{Path: "a", Priority: 1}, {Path: "b", Priority: 1}}}, Destination: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []Command{
		{Kind: CommandPriority, IDs: []string{id}, FileIndices: []int{0, 1}, Priority: 0},
		{Kind: CommandPriority, IDs: []string{id}, FileIndices: []int{0, 2}, Priority: 2},
		{Kind: CommandTags, IDs: []string{id}, Tags: []string{"bad,tag"}},
		{Kind: CommandPause, IDs: []string{id, "missing"}},
	} {
		before := m.Snapshot().Revision
		if err := m.Execute(t.Context(), cmd); err == nil {
			t.Fatal("invalid command accepted", cmd)
		}
		if m.Snapshot().Revision != before {
			t.Fatal("failed command partially committed")
		}
	}
	if err := m.Execute(t.Context(), Command{Kind: CommandPriority, IDs: []string{id}, FileIndices: []int{0, 1}, Priority: 2}); err != nil {
		t.Fatal(err)
	}
	for _, file := range m.Snapshot().Tasks[0].Source.Files {
		if file.Priority != 2 {
			t.Fatal("batch priority missing")
		}
	}
}

func TestSeedDurationStopsAndExplicitResumeOverrides(t *testing.T) {
	m := observedManager(t, Observation{Phase: PhaseSeeding, Files: []File{{Path: "a", Size: 10, Done: 10}}})
	id := add(t, m, "duration", true)
	m.mu.Lock()
	r := &m.state.Tasks[0]
	r.Source.Files = []File{{Path: "a", Size: 10, Priority: 1}}
	r.Metadata = "fixture"
	r.SeedRatio = -1
	r.SeedLimitSeconds = 2
	m.samples[id] = sample{at: time.Now().Add(-3 * time.Second), task: Task{Phase: PhaseSeeding}}
	m.tick()
	m.mu.Unlock()
	task := m.Snapshot().Tasks[0]
	if task.Run || task.Reason != PauseDuration || task.SeedSeconds < 2 {
		t.Fatalf("duration policy not applied: %+v", task.Record)
	}
	exec(t, m, CommandResume, id)
	if !m.Snapshot().Tasks[0].SeedOverride {
		t.Fatal("explicit resume did not override duration")
	}
}

func BenchmarkSnapshot1000Tasks10000Files(b *testing.B) {
	m := &Manager{samples: map[string]sample{}, active: map[string]bool{}, holds: map[string]error{}}
	for i := 0; i < 1000; i++ {
		r := Record{ID: fmt.Sprint(i)}
		for j := 0; j < 10; j++ {
			r.Source.Files = append(r.Source.Files, File{Path: fmt.Sprintf("folder/%d", j), Size: 1024, Priority: 1})
		}
		m.state.Tasks = append(m.state.Tasks, r)
	}
	m.publish()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = m.Snapshot()
	}
}

func TestCanceledCommandHasNoSideEffects(t *testing.T) {
	m, _ := manager(t)
	id := add(t, m, "cancel", true)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	before := m.Snapshot().Revision
	if err := m.Execute(ctx, Command{Kind: CommandPause, IDs: []string{id}}); err == nil {
		t.Fatal("cancellation ignored")
	}
	if m.Snapshot().Revision != before || !m.Snapshot().Tasks[0].Run {
		t.Fatal("canceled command changed state")
	}
}
