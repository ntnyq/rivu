package session

import (
	"context"
	"path/filepath"
	"slices"
	"time"

	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/locales"
)

func (m *Manager) Execute(ctx context.Context, cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.closed || m.barrier {
		return problem.Message(problem.Persistence, locales.Messages.SessionSaving, nil)
	}
	if m.loadErr != nil {
		return m.loadErr
	}
	if err := cmd.validate(m.state.Tasks); err != nil {
		return err
	}
	next := cloneState(m.state)
	stopIDs := []string{}
	selectedIDs := []string{}
	if cmd.Kind == CommandFirst {
		prioritize(next.Tasks, cmd.IDs)
	}
	for i := range next.Tasks {
		r := &next.Tasks[i]
		selected := slices.Contains(cmd.IDs, r.ID)
		if cmd.Kind == CommandPauseAll {
			selected = r.Run
			r.Batch = false // Only the newest Pause All owns batch membership.
		}
		if cmd.Kind == CommandResumeBatch {
			selected = r.Batch
		}
		if !selected {
			continue
		}
		selectedIDs = append(selectedIDs, r.ID)
		switch cmd.Kind {
		case CommandPause, CommandPauseAll:
			r.Run = false
			r.Reason = PauseManual
			r.Batch = cmd.Kind == CommandPauseAll
			stopIDs = append(stopIDs, r.ID)
		case CommandResume, CommandResumeBatch:
			if r.Completed && (r.Reason == PauseRatio || r.Reason == PauseDuration) && cmd.Kind == CommandResume {
				r.SeedOverride = true
			}
			r.Run = true
			r.Reason = ""
			r.Batch = false
		case CommandVerify:
			r.Completed = false
			stopIDs = append(stopIDs, r.ID)
			r.Run = true
			r.Reason = PauseVerify
		case CommandFirst:
			r.Run = true
		case CommandPriority:
			indices := cmd.FileIndices
			if len(indices) == 0 {
				indices = []int{cmd.FileIndex}
			}
			for _, index := range indices {
				previous := r.Source.Files[index].Priority
				r.Source.Files[index].Priority = cmd.Priority
				if (previous == 0) != (cmd.Priority == 0) {
					r.Completed = false
				}
			}
			if err := validateSelection(r.Source.Files); err != nil {
				return err
			}
			stopIDs = append(stopIDs, r.ID)
		case CommandSeed:
			r.SeedRatio = cmd.Ratio
			r.SeedLimitSeconds = cmd.SeedLimitSeconds
			r.SeedOverride = false
		case CommandRemove:
			r.Removing = true
			r.Trash = cmd.Trash
			r.Run = false
			r.Batch = false
			stopIDs = append(stopIDs, r.ID)
		case CommandTags:
			r.Tags = slices.Clone(cmd.Tags)
		case CommandRelocate:
			destination, err := filepath.EvalSymlinks(cmd.Destination)
			if err != nil {
				return err
			}
			if destination == r.Destination {
				return nil // Preserve counters and ownership when the location did not change.
			}
			if err = m.checkDestination(r.Source, destination, r.ID, true); err != nil {
				return err
			}
			adopted, err := files.InspectExisting(destination, names(r.Source.Files), true)
			if err != nil {
				return err
			}
			r.Destination = destination
			r.Adopted = adopted
			r.Completed = false
			r.Reason = PauseVerify
			stopIDs = append(stopIDs, r.ID)
		case CommandNotified:
			r.Notified = r.Epoch
		}
	}
	// A pause must remain safe even when its intent cannot be saved.
	for _, id := range stopIDs {
		if err := m.stop(id); err != nil {
			return err
		}
	}
	if err := m.save(next); err != nil {
		m.errorState = problem.New(problem.Persistence, "save session", err)
		for _, id := range stopIDs {
			m.holds[id] = problem.New(problem.Persistence, "pause", err)
		}
		return err
	}
	for _, id := range selectedIDs {
		if cmd.Kind != CommandTags && cmd.Kind != CommandSeed && cmd.Kind != CommandNotified {
			delete(m.holds, id)
		}
		if cmd.Kind == CommandRelocate {
			delete(m.samples, id)
		}
	}
	m.tick()
	return m.errorState
}

func (m *Manager) stop(id string) error {
	err := m.engine.Stop(id)
	delete(m.active, id)
	s := m.samples[id]
	s.down = 0
	s.up = 0
	s.at = time.Time{}
	s.task.DownRate = 0
	s.task.UpRate = 0
	m.samples[id] = s
	return err
}
