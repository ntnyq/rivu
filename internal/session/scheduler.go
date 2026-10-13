package session

import (
	"slices"
	"sort"
	"time"
)

// Put the requested group ahead of the existing queue while retaining order
// within each group. Compact ranks avoid ties and repeated-decrement overflow.
func prioritize(records []Record, ids []string) {
	sort.SliceStable(records, func(i, j int) bool {
		left, right := slices.Contains(ids, records[i].ID), slices.Contains(ids, records[j].ID)
		if left != right {
			return left
		}
		return records[i].Order < records[j].Order
	})
	for i := range records {
		records[i].Order = int64(i)
	}
}

// schedule decides the complete admission set before touching handles, so
// reprioritizing work never temporarily exceeds either concurrency limit.
func (m *Manager) schedule(now time.Time) {
	records := append([]Record(nil), m.state.Tasks...)
	sort.SliceStable(records, func(i, j int) bool { return records[i].Order < records[j].Order })
	admitted := make(map[string]bool)
	downloads, metadata := 0, 0
	for _, r := range records {
		if r.Removing || !r.Run || m.holds[r.ID] != "" || now.Before(m.samples[r.ID].retry) {
			continue
		}
		if r.Metadata == "" {
			if metadata >= 2 {
				continue
			}
			metadata++
		} else if !m.active[r.ID] || m.samples[r.ID].task.Phase != "seeding" {
			if downloads >= m.settings.Concurrent {
				continue
			}
			downloads++
		}
		admitted[r.ID] = true
	}
	for id := range m.active {
		if !admitted[id] {
			_ = m.stop(id)
		}
	}
	for _, r := range records {
		if !admitted[r.ID] {
			continue
		}
		if !m.active[r.ID] {
			if err := m.engine.Attach(r); err != nil {
				m.holds[r.ID] = err.Error()
				continue
			}
			m.active[r.ID] = true
			s := m.samples[r.ID]
			s.since = now
			m.samples[r.ID] = s
		}
		if r.Metadata != "" {
			if err := m.engine.Start(r.ID, r.Source.Files); err != nil {
				m.holds[r.ID] = err.Error()
				_ = m.stop(r.ID)
			}
		}
	}
}
