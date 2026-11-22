package app

import "sync"

// workQueue runs user actions in admission order without blocking the UI.
// The application stops admitting actions before waiting for a drain.
type workQueue struct {
	mu      sync.Mutex
	pending []func()
	done    chan struct{}
}

func (q *workQueue) enqueue(work func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending = append(q.pending, work)
	if q.done == nil {
		q.done = make(chan struct{})
		go q.run()
	}
}

func (q *workQueue) run() {
	for {
		q.mu.Lock()
		if len(q.pending) == 0 {
			close(q.done)
			q.done = nil
			q.mu.Unlock()
			return
		}
		work := q.pending[0]
		q.pending[0] = nil
		q.pending = q.pending[1:]
		q.mu.Unlock()
		work()
	}
}

func (q *workQueue) drained() <-chan struct{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.done
}
