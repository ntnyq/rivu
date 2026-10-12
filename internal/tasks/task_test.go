package tasks

import (
	"context"
	"testing"
	"time"
)

func receive(t *testing.T, queue <-chan func()) func() {
	t.Helper()
	select {
	case fn := <-queue:
		return fn
	case <-time.After(time.Second):
		t.Fatal("callback never dispatched")
		return nil
	}
}

func TestReplacementDropsQueuedCallbacks(t *testing.T) {
	queue := make(chan func(), 8)
	dispatch := func(fn func()) { queue <- fn }
	var task Task
	var values []string
	done := func(result string, err error) { values = append(values, result) }
	task.Start(dispatch, func(ctx context.Context, progress func(float64)) (string, error) { progress(.5); return "old", nil }, func(float64) { values = append(values, "old progress") }, done)
	oldProgress, oldDone := receive(t, queue), receive(t, queue)
	task.Start(dispatch, func(context.Context, func(float64)) (string, error) { return "new", nil }, func(float64) {}, done)
	oldProgress()
	oldDone()
	if !task.Running() {
		t.Fatal("old callback cleared the new task")
	}
	receive(t, queue)()
	if len(values) != 1 || values[0] != "new" || task.Running() {
		t.Fatalf("unexpected results: %v", values)
	}
}

func TestCloseCancelsWorkAndDiscardsCompletion(t *testing.T) {
	queue := make(chan func(), 1)
	started := make(chan struct{})
	var task Task
	task.Start(func(fn func()) { queue <- fn }, func(ctx context.Context, _ func(float64)) (string, error) {
		close(started)
		<-ctx.Done()
		return "", ctx.Err()
	}, func(float64) {}, func(string, error) { t.Error("callback ran after close") })
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("task did not start")
	}
	task.Close()
	receive(t, queue)()
	if task.Running() {
		t.Fatal("closed task is running")
	}
}
