package state

import "sync"

// Scheduler dispatches subscription callbacks.
type Scheduler interface {
	Schedule(fn func())
}

// SchedulerFunc adapts a function into a Scheduler.
type SchedulerFunc func(func())

// Schedule dispatches fn using the wrapped function.
func (f SchedulerFunc) Schedule(fn func()) {
	if f == nil || fn == nil {
		return
	}
	f(fn)
}

// DirectScheduler runs callbacks immediately in the caller goroutine.
var DirectScheduler Scheduler = SchedulerFunc(func(fn func()) {
	if fn != nil {
		fn()
	}
})

// AsyncScheduler runs callbacks in a new goroutine.
type AsyncScheduler struct{}

// Schedule dispatches fn asynchronously.
func (AsyncScheduler) Schedule(fn func()) {
	if fn == nil {
		return
	}
	go fn()
}

// Queue batches callbacks for explicit flushing.
type Queue struct {
	mu      sync.Mutex
	pending []func()
}

// NewQueue creates an empty queue.
func NewQueue() *Queue {
	return &Queue{}
}

// Schedule enqueues a callback for later flushing.
func (q *Queue) Schedule(fn func()) {
	if q == nil || fn == nil {
		return
	}
	q.mu.Lock()
	q.pending = append(q.pending, fn)
	q.mu.Unlock()
}

// Flush executes queued callbacks and returns the count.
func (q *Queue) Flush() int {
	if q == nil {
		return 0
	}
	q.mu.Lock()
	pending := q.pending
	q.pending = nil
	q.mu.Unlock()
	for _, fn := range pending {
		fn()
	}
	return len(pending)
}
