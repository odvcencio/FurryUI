package runtime

import "github.com/odvcencio/furry-ui/state"

// QueueFlushPolicy configures when the app flushes state queues.
type QueueFlushPolicy int

const (
	// FlushOnMessageAndTick flushes on any message or tick.
	FlushOnMessageAndTick QueueFlushPolicy = iota
	// FlushOnMessage flushes on messages except TickMsg.
	FlushOnMessage
	// FlushOnTick flushes only on TickMsg.
	FlushOnTick
	// FlushManual flushes only on QueueFlushMsg.
	FlushManual
)

// WithQueue wraps update to flush queue on TickMsg or QueueFlushMsg.
// If update is nil, DefaultUpdate is used.
func WithQueue(queue *state.Queue, update UpdateFunc) UpdateFunc {
	return WithQueuePolicy(queue, FlushOnTick, update)
}

// WithQueuePolicy wraps update to flush queue based on policy.
// If update is nil, DefaultUpdate is used.
func WithQueuePolicy(queue *state.Queue, policy QueueFlushPolicy, update UpdateFunc) UpdateFunc {
	if update == nil {
		update = DefaultUpdate
	}

	return func(app *App, msg Message) bool {
		dirty := update(app, msg)
		if queue == nil {
			return dirty
		}
		if shouldFlushQueue(policy, msg) {
			if flushed := queue.Flush(); flushed > 0 {
				dirty = true
			}
		}
		return dirty
	}
}

func shouldFlushQueue(policy QueueFlushPolicy, msg Message) bool {
	if _, ok := msg.(QueueFlushMsg); ok {
		return true
	}
	if policy == FlushManual {
		return false
	}
	_, isTick := msg.(TickMsg)
	switch policy {
	case FlushOnMessage:
		return !isTick
	case FlushOnTick:
		return isTick
	default:
		return true
	}
}
