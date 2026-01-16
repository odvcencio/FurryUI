package runtime

import (
	"sync/atomic"

	"github.com/odvcencio/furry-ui/state"
)

// QueueScheduler enqueues callbacks and wakes the app to flush.
type QueueScheduler struct {
	queue   *state.Queue
	post    func(Message) bool
	pending atomic.Bool
}

// NewQueueScheduler wires a queue to a post function.
func NewQueueScheduler(queue *state.Queue, post func(Message) bool) *QueueScheduler {
	if queue == nil {
		queue = state.NewQueue()
	}
	return &QueueScheduler{
		queue: queue,
		post:  post,
	}
}

// Schedule enqueues the callback and posts a flush message.
func (s *QueueScheduler) Schedule(fn func()) {
	if s == nil || s.queue == nil || fn == nil {
		return
	}
	s.queue.Schedule(fn)
	if s.post == nil {
		return
	}
	if s.pending.CompareAndSwap(false, true) {
		if !s.post(QueueFlushMsg{}) {
			s.pending.Store(false)
		}
	}
}

func (s *QueueScheduler) resetPending() {
	if s == nil {
		return
	}
	s.pending.Store(false)
}
