package session

import (
	"github.com/ntnyq/rivu/internal/problem"
	"time"
)

type sample struct {
	at            time.Time
	down, up      int64
	task          Task
	since         time.Time
	retry         time.Time
	seedRemainder time.Duration
}

func (m *Manager) tick() {
	if m.closed || m.barrier || m.loadErr != nil {
		return
	}
	changed := false
	now := time.Now()
	d, u := m.settings.LimitsAt(now)
	m.engine.Limits(d, u)
	for i := range m.state.Tasks {
		r := &m.state.Tasks[i]
		if r.Removing {
			continue
		}
		if !m.active[r.ID] {
			continue
		}
		o := m.engine.Observe(r.ID)
		prev := m.samples[r.ID]
		if o.Error != "" || o.Failure != nil {
			m.holds[r.ID] = problem.From(o.Failure)
			if o.Failure == nil {
				m.holds[r.ID] = problem.New(problem.Engine, o.Error, nil)
			}
			_ = m.stop(r.ID)
			continue
		}
		if o.Metadata != nil {
			s := *o.Metadata
			if !s.Identity.Matches(r.Source.Identity) {
				m.holds[r.ID] = problem.New(problem.Metadata, "identity mismatch", nil)
				_ = m.stop(r.ID)
				continue
			}
			for _, other := range m.state.Tasks {
				if other.ID != r.ID && s.Identity.Matches(other.Source.Identity) {
					m.holds[r.ID] = problem.New(problem.Conflict, "duplicate identity", nil)
				}
			}
			if m.holds[r.ID] == nil {
				if err := m.checkDestination(s, r.Destination, r.ID, false); err != nil {
					m.holds[r.ID] = problem.From(err)
				} else if err = m.metadata(r, s); err != nil {
					m.holds[r.ID] = problem.From(err)
				} else {
					changed = true
				}
			}
			_ = m.stop(r.ID)
			continue
		}
		task := Task{Record: *r, Phase: o.Phase, Files: o.Files, Peers: o.Peers, Trackers: o.Trackers, Diagnostics: o.Diagnostics}
		for j, f := range r.Source.Files {
			task.TotalSize += f.Size
			if f.Priority > 0 {
				task.Size += f.Size
				if j < len(o.Files) {
					task.Done += o.Files[j].Done
				}
			}
		}
		r.Uploaded += max(0, o.Uploaded-prev.up)
		task.Uploaded = r.Uploaded
		if !prev.at.IsZero() {
			seconds := now.Sub(prev.at).Seconds()
			task.DownRate = float64(max(0, o.Downloaded-prev.down)) / seconds
			task.UpRate = float64(max(0, o.Uploaded-prev.up)) / seconds
		}
		if r.Completed && len(o.Files) > 0 && task.Done < task.Size && task.Phase != PhaseVerifying {
			r.Completed = false
			changed = true
		}
		if len(o.Files) > 0 && task.Done == task.Size && task.Phase != PhaseVerifying {
			task.Phase = PhaseSeeding
			if prev.task.Phase == PhaseSeeding && !prev.at.IsZero() {
				elapsed := now.Sub(prev.at)
				// Carry subsecond samples without counting paused/suspended time.
				r.SeedSeconds += int64(elapsed / time.Second)
				prev.seedRemainder += elapsed % time.Second
				r.SeedSeconds += int64(prev.seedRemainder / time.Second)
				prev.seedRemainder %= time.Second
			}
			if !r.Completed {
				r.Completed = true
				r.Epoch++
				changed = true
			}
			if !r.SeedOverride && r.SeedLimitSeconds > 0 && r.SeedSeconds >= r.SeedLimitSeconds {
				r.Run = false
				r.Reason = PauseDuration
				changed = true
				_ = m.stop(r.ID)
			}
			ratio := float64(r.Uploaded) / float64(max(1, task.TotalSize))
			if r.Run && !r.SeedOverride && r.SeedRatio >= 0 && (r.SeedRatio == 0 || task.TotalSize > 0 && ratio >= r.SeedRatio) {
				r.Run = false
				r.Reason = PauseRatio
				changed = true
				_ = m.stop(r.ID)
			}
		}
		since := prev.since
		if since.IsZero() || o.Downloaded > prev.down || task.Phase == PhaseVerifying {
			since = now
		}
		m.samples[r.ID] = sample{at: now, down: o.Downloaded, up: o.Uploaded, task: task, since: since, retry: prev.retry, seedRemainder: prev.seedRemainder}
		if !m.active[r.ID] {
			_ = m.stop(r.ID)
		}
		// Yield stalled slots; retry after a minute without changing durable intent.
		stalledPhase := task.Phase == PhaseWaiting || task.Phase == PhaseDownloading || task.Phase == PhaseMetadata
		if stalledPhase && now.Sub(since) > 90*time.Second {
			_ = m.stop(r.ID)
			s := m.samples[r.ID]
			s.retry = now.Add(time.Minute)
			s.since = now
			m.samples[r.ID] = s
		}
	}
	if changed || now.Sub(m.lastSave) >= 15*time.Second {
		if err := m.save(m.state); err != nil {
			m.errorState = problem.New(problem.Persistence, "save session", err)
			m.barrier = true
			for id := range m.active {
				_ = m.stop(id)
			}
			m.publish()
			return
		}
	}
	// Recovery of committed removals is safe and idempotent.
	for i := len(m.state.Tasks) - 1; i >= 0; i-- {
		if m.state.Tasks[i].Removing {
			if err := m.finishRemoval(i); err != nil {
				m.holds[m.state.Tasks[i].ID] = problem.From(err)
				if m.barrier || m.loadErr != nil {
					break
				}
			}
		}
	}
	if m.barrier || m.loadErr != nil {
		m.publish()
		return
	}
	m.schedule(now)
	m.publish()
}
