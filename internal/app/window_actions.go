package app

import (
	"github.com/ntnyq/rivu/internal/session"
)

func (w *mainWindow) selectOnly(id string) {
	w.selected = map[string]bool{id: true}
	w.anchor = id
	w.fileState = fileTreeState{}
}

func (w *mainWindow) ids() []string {
	var ids []string
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

func (w *mainWindow) chosen() *session.Task {
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] {
			return &t
		}
	}
	return nil
}

func (w *mainWindow) toggleSelected() {
	ids := w.ids()
	if len(ids) == 0 || w.app == nil {
		return
	}
	kind := session.CommandResume
	for _, t := range w.snapshot.Tasks {
		if w.selected[t.ID] && t.Run {
			kind = session.CommandPause
		}
	}
	w.app.execute(session.Command{Kind: kind, IDs: ids})
}

func (w *mainWindow) remove() {
	w.removeIDs = w.ids()
	w.trashFiles = false
	w.removeOpen = len(w.removeIDs) > 0
}
