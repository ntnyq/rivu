package session

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/locales"
)

func (m *Manager) finishRemoval(i int) error {
	r := m.state.Tasks[i]
	if err := m.stop(r.ID); err != nil {
		return err
	}
	if r.Trash {
		if m.trash == nil {
			return problem.Message(problem.Conflict, locales.Messages.FilesTrashUnavailable, nil)
		}
		if err := files.ValidateDestination(r.Destination); err != nil {
			return err
		}
		for _, f := range r.Source.Files {
			if r.Adopted[f.Path] != "" || slices.Contains(r.Trashed, f.Path) {
				continue
			}
			full := filepath.Join(r.Destination, filepath.FromSlash(f.Path))
			shared := false
			for _, other := range m.state.Tasks {
				if other.ID == r.ID {
					continue
				}
				for _, of := range other.Source.Files {
					if strings.EqualFold(full, filepath.Join(other.Destination, filepath.FromSlash(of.Path))) {
						shared = true
					}
				}
			}
			if shared {
				return problem.Message(problem.Conflict, locales.Messages.FilesShared, nil)
			}
			identities, err := files.InspectExisting(r.Destination, []string{f.Path}, true)
			if err != nil {
				return err
			}
			if identity, exists := identities[f.Path]; exists {
				owners, err := files.ReadOwnership(r.Metadata + ".files.json")
				if err != nil {
					return err
				}
				if owners[f.Path] != identity {
					return problem.Message(problem.Conflict, func(m locales.Messages) string { return m.FilesOwnershipChanged(f.Path) }, nil)
				}
				if err := m.trash(full); err != nil {
					return err
				}
			}
			r.Trashed = append(r.Trashed, f.Path)
			next := cloneState(m.state)
			next.Tasks[i] = r
			if err := m.save(next); err != nil {
				return err
			}
		}
	}
	next := cloneState(m.state)
	next.Tasks = slices.Delete(next.Tasks, i, i+1)
	if err := m.save(next); err != nil {
		return err
	}
	delete(m.samples, r.ID)
	delete(m.holds, r.ID)
	return nil
}
