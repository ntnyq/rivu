// Package windowtask provides one cancellable job per window. Methods and callbacks
// run on the UI thread; work runs in a goroutine and receives only snapshots.
package windowtask

import "context"

type Task struct {
	cancel     context.CancelFunc
	generation uint64
	closed     bool
}

func (t *Task) Running() bool { return t.cancel != nil }

// Start cancels the previous job. Late progress and completion from replaced
// jobs are discarded, including callbacks already queued on the UI thread.
func (t *Task) Start(dispatch func(func()), work func(context.Context, func(float64)) (string, error), progress func(float64), done func(string, error)) {
	if t.closed {
		return
	}
	t.Cancel()
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	id := t.generation
	post := func(fn func()) {
		dispatch(func() {
			if !t.closed && t.generation == id {
				fn()
			}
		})
	}
	go func() {
		defer cancel()
		result, err := work(ctx, func(value float64) { post(func() { progress(value) }) })
		post(func() { t.cancel = nil; done(result, err) })
	}()
}

func (t *Task) Cancel() {
	t.generation++
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

func (t *Task) Close() { t.closed = true; t.Cancel() }
