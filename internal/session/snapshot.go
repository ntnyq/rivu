package session

import (
	"github.com/ntnyq/rivu/internal/problem"
	"maps"
	"slices"
	"sort"
	"sync"
	"time"
)

func (m *Manager) publish() {
	snap := Snapshot{Revision: m.state.Revision, SlowMode: m.settings.SlowAt(time.Now()), Quiesced: m.barrier}
	if m.errorState != nil {
		snap.Error = m.errorState.Error()
		snap.Issue = problem.From(m.errorState)
	}
	for _, r := range m.state.Tasks {
		t := m.samples[r.ID].task
		t.Record = r
		t.Diagnostics = slices.Clone(t.Diagnostics)
		if t.Phase == PhaseVerifying {
			t.Completed = false
		}
		if len(t.Files) == 0 {
			t.Files = slices.Clone(r.Source.Files)
		}
		t.Size = 0
		t.TotalSize = 0
		for _, f := range r.Source.Files {
			t.TotalSize += f.Size
			if f.Priority > 0 {
				t.Size += f.Size
			}
		}
		if !m.active[r.ID] {
			t.DownRate = 0
			t.UpRate = 0
			t.Peers = 0
			for i := range t.Diagnostics {
				t.Diagnostics[i].Next = time.Time{}
			}
			t.Phase = PhaseQueued
			if !r.Run {
				t.Phase = PhasePaused
				if t.Completed {
					t.Phase = PhaseCompleted
					if len(m.samples[r.ID].task.Files) == 0 {
						t.Done = t.Size
					}
				}
			}
		}
		if r.Removing {
			t.Phase = PhaseRemoving
		}
		if hold := m.holds[r.ID]; hold != nil {
			t.Error = hold.Error()
			t.Issue = problem.From(hold)
			t.Phase = PhaseError
		}
		snap.DownRate += t.DownRate
		snap.UpRate += t.UpRate
		snap.Tasks = append(snap.Tasks, t)
	}
	sort.SliceStable(snap.Tasks, func(i, j int) bool { return snap.Tasks[i].Added.After(snap.Tasks[j].Added) })
	// JSON cloning excludes Source.Meta and transient File.Done, so copy those
	// explicitly instead of sharing mutable file slices across consumers.
	for i := range snap.Tasks {
		t := &snap.Tasks[i]
		t.Source.Files = slices.Clone(t.Source.Files)
		t.Source.Trackers = slices.Clone(t.Source.Trackers)
		t.Files = slices.Clone(t.Files)
		t.Trackers = slices.Clone(t.Trackers)
		t.Trashed = slices.Clone(t.Trashed)
		t.Diagnostics = slices.Clone(t.Diagnostics)
		if t.Issue != nil {
			t.Issue = problem.From(t.Issue)
		}
		t.Tags = slices.Clone(t.Tags)
		t.Adopted = maps.Clone(t.Adopted)
	}
	m.snapshot.Store(&snap)
	for ch := range m.watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (m *Manager) Snapshot() Snapshot {
	p := m.snapshot.Load()
	if p == nil {
		return Snapshot{}
	}
	s := *p
	if s.Issue != nil {
		s.Issue = problem.From(s.Issue)
	}
	s.Tasks = slices.Clone(s.Tasks)
	for i := range s.Tasks {
		t := &s.Tasks[i]
		t.Files = slices.Clone(t.Files)
		t.Source.Files = slices.Clone(t.Source.Files)
		t.Source.Trackers = slices.Clone(t.Source.Trackers)
		t.Trackers = slices.Clone(t.Trackers)
		t.Trashed = slices.Clone(t.Trashed)
		t.Diagnostics = slices.Clone(t.Diagnostics)
		if t.Issue != nil {
			t.Issue = problem.From(t.Issue)
		}
		t.Tags = slices.Clone(t.Tags)
		t.Adopted = maps.Clone(t.Adopted)
	}
	return s
}

func (m *Manager) Watch() (<-chan struct{}, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan struct{}, 1)
	m.watchers[ch] = true
	var once sync.Once
	return ch, func() { once.Do(func() { m.mu.Lock(); delete(m.watchers, ch); close(ch); m.mu.Unlock() }) }
}
