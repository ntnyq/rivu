package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

type completionNotice struct {
	epoch, failedRevision                uint64
	shown, running, acknowledged, failed bool
}

// Presentation and acknowledgement have separate lifetimes. All state and
// completion callbacks run on the UI thread; work runs on a worker.
type completionNotifications struct{ notices map[string]*completionNotice }

func (n *completionNotifications) update(s session.Snapshot, work func(session.Task, bool, func(bool, error, error)), report func(error)) {
	if s.Quiesced {
		return
	}
	if n.notices == nil {
		n.notices = map[string]*completionNotice{}
	}
	present := map[string]bool{}
	for _, t := range s.Tasks {
		present[t.ID] = true
		if t.Removing || t.Epoch <= t.Notified {
			delete(n.notices, t.ID)
			continue
		}
		notice := n.notices[t.ID]
		if notice == nil || notice.epoch != t.Epoch {
			notice = &completionNotice{epoch: t.Epoch}
			n.notices[t.ID] = notice
		}
		if notice.running || notice.acknowledged || notice.failed && notice.failedRevision == s.Revision {
			continue
		}
		notice.running = true
		work(t, !notice.shown, func(shown bool, showErr, ackErr error) {
			if n.notices[t.ID] != notice {
				return // A newer completion or removal superseded this callback.
			}
			notice.running = false
			notice.shown = notice.shown || shown
			if showErr != nil {
				report(showErr)
			}
			if ackErr == nil {
				notice.acknowledged = true
			} else if !errors.Is(ackErr, session.ErrQuiesced) {
				notice.failed, notice.failedRevision = true, s.Revision
				report(ackErr)
			}
		})
	}
	for id := range n.notices {
		if !present[id] {
			delete(n.notices, id)
		}
	}
}

func (a *application) notifications(s session.Snapshot) {
	if s.Quiesced || a.quitting || a.shutdownReady {
		return
	}
	settings := a.store.Snapshot()
	a.completions.update(s, func(t session.Task, show bool, done func(bool, error, error)) {
		go func() {
			if a.session.Snapshot().Quiesced {
				mygo.RunOnMain(func() { done(false, nil, session.ErrQuiesced) })
				return
			}
			var showErr error
			if show && settings.NotifyComplete {
				n := mygo.NewNotification(mygo.NotificationOptions{ID: fmt.Sprintf("rivu-%s-%d", t.ID, t.Epoch), Title: locales.NotificationComplete(), Body: t.Source.Name})
				n.OnClick(func() { w := a.openWindow(); w.dispatch(func() { w.selectOnly(t.ID) }) })
				if err := n.Show(); err != nil {
					showErr = messageError{locales.NotificationDeliveryFailed(err), err}
				}
			}
			err := a.session.Execute(context.Background(), session.Command{Kind: session.CommandNotified, IDs: []string{t.ID}, Epoch: t.Epoch})
			mygo.RunOnMain(func() { done(show, showErr, err) })
		}()
	}, func(err error) {
		if !a.quitting && !a.shutdownReady {
			a.report(err)
		}
	})
	for _, t := range s.Tasks {
		if settings.NotifyError && t.Error != "" && a.errorsSeen[t.ID] != t.Error {
			a.errorsSeen[t.ID] = t.Error
			go func() {
				n := mygo.NewNotification(mygo.NotificationOptions{ID: "rivu-error-" + t.ID, Title: locales.NotificationAttention(), Body: t.Source.Name + "：" + taskError(t)})
				n.OnClick(func() { w := a.openWindow(); w.dispatch(func() { w.filter = 5; w.selectOnly(t.ID) }) })
				_ = n.Show()
			}()
		}
	}
}
