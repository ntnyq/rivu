package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
)

func completionSnapshot(epoch uint64) session.Snapshot {
	return session.Snapshot{Revision: 10, Tasks: []session.Task{{Record: session.Record{ID: "task", Epoch: epoch}}}}
}

func TestCompletionAcknowledgementWaitsForMaintenanceWithoutRedisplay(t *testing.T) {
	var notifications completionNotifications
	s := completionSnapshot(1)
	var finish func(bool, error, error)
	calls, shown := 0, 0
	work := func(_ session.Task, show bool, done func(bool, error, error)) {
		calls++
		if show {
			shown++
		}
		finish = done
	}
	report := func(err error) { t.Errorf("unexpected notification error: %v", err) }
	paused := s
	paused.Quiesced = true
	notifications.update(paused, work, report)
	if calls != 0 {
		t.Fatal("notification started during maintenance")
	}
	notifications.update(s, work, report)
	notifications.update(s, work, report)
	if calls != 1 {
		t.Fatal("duplicate notification started while in flight")
	}
	finish(true, nil, session.ErrQuiesced) // Pause began after the native Show.
	for range 20 {
		notifications.update(paused, work, report)
	}
	if calls != 1 {
		t.Fatal("paused acknowledgements triggered feedback retries")
	}
	notifications.update(s, work, report) // Final maintenance hold released.
	if calls != 2 || shown != 1 {
		t.Fatal("resume redisplayed the completed epoch")
	}
	finish(false, nil, nil)
	notifications.update(s, work, report) // An older queued UI snapshot arrives.
	if calls != 2 {
		t.Fatal("successful acknowledgement was retried from a stale snapshot")
	}
	s.Tasks[0].Notified = 1
	notifications.update(s, work, report)
	notifications.update(completionSnapshot(2), work, report)
	if calls != 3 || shown != 2 {
		t.Fatal("new completion epoch was not displayed")
	}
}

func TestCompletionDefersBeforeShowingAndIgnoresSupersededCallbacks(t *testing.T) {
	var notifications completionNotifications
	s := completionSnapshot(1)
	var callbacks []func(bool, error, error)
	var displays []bool
	work := func(_ session.Task, show bool, done func(bool, error, error)) {
		displays = append(displays, show)
		callbacks = append(callbacks, done)
	}
	report := func(err error) { t.Errorf("unexpected error: %v", err) }
	notifications.update(s, work, report)
	callbacks[0](false, nil, session.ErrQuiesced) // Pause occurred before Show.
	notifications.update(s, work, report)
	if len(displays) != 2 || !displays[1] {
		t.Fatal("unshown notification was lost after maintenance")
	}
	notifications.update(completionSnapshot(2), work, report)
	callbacks[1](true, nil, nil) // A previous epoch finally acknowledged.
	notifications.update(completionSnapshot(2), work, report)
	if len(displays) != 3 {
		t.Fatal("late callback reset the current epoch's in-flight state")
	}
	callbacks[2](true, nil, nil)
	notifications.update(completionSnapshot(2), work, report)
	if len(displays) != 3 {
		t.Fatal("current epoch displayed twice")
	}
}

func TestCompletionSaveFailureDoesNotRepublishOrRedisplay(t *testing.T) {
	var notifications completionNotifications
	s := completionSnapshot(1)
	var finish func(bool, error, error)
	calls, reports := 0, 0
	work := func(_ session.Task, show bool, done func(bool, error, error)) {
		calls++
		if calls > 1 && show {
			t.Error("save retry redisplayed the notification")
		}
		finish = done
	}
	report := func(error) { reports++ }
	notifications.update(s, work, report)
	finish(true, nil, errors.New("save failed"))
	for range 20 {
		notifications.update(s, work, report)
	}
	if calls != 1 || reports != 1 {
		t.Fatal("save failure caused a feedback retry loop")
	}
	s.Revision++ // A successful durable mutation makes retry meaningful.
	notifications.update(s, work, report)
	finish(false, nil, nil)
	if calls != 2 || reports != 1 {
		t.Fatal("acknowledgement did not recover")
	}
}

func TestCompletionRacesRealSessionMaintenance(t *testing.T) {
	dir := t.TempDir()
	data, err := json.Marshal(struct {
		Version int              `json:"version"`
		Tasks   []session.Record `json:"tasks"`
	}{2, []session.Record{{ID: "11111111111111111111111111111111", Source: session.Source{Identity: session.Identity{V1: "fixture"}}, Destination: t.TempDir(), Epoch: 1, Completed: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	m, err := session.Open(dir, lifecycleEngine{}, config.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	var notifications completionNotifications
	var release func()
	calls, shown := 0, 0
	work := func(task session.Task, show bool, done func(bool, error, error)) {
		calls++
		if show {
			shown++
			release, err = m.Quiesce() // Pause exactly between presentation and acknowledgement.
			if err != nil {
				t.Fatal(err)
			}
		}
		err := m.Execute(t.Context(), session.Command{Kind: session.CommandNotified, IDs: []string{task.ID}, Epoch: task.Epoch})
		done(show, nil, err)
	}
	report := func(err error) { t.Errorf("maintenance became a notification failure: %v", err) }
	notifications.update(m.Snapshot(), work, report)
	if release == nil {
		t.Fatal("notification did not start")
	}
	defer release()
	for range 20 {
		notifications.update(m.Snapshot(), work, report)
	}
	if calls != 1 {
		t.Fatal("real quiesced session caused retries")
	}
	release()
	notifications.update(m.Snapshot(), work, report)
	if calls != 2 || shown != 1 || m.Snapshot().Tasks[0].Notified != 1 {
		t.Fatal("resume did not durably acknowledge exactly one display")
	}
}
