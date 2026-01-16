package runtime

import (
	"testing"

	"github.com/odvcencio/furry-ui/state"
)

func TestQueueScheduler_PostsFlush(t *testing.T) {
	queue := state.NewQueue()
	posted := 0
	scheduler := NewQueueScheduler(queue, func(msg Message) bool {
		if _, ok := msg.(QueueFlushMsg); ok {
			posted++
			return true
		}
		return false
	})

	scheduler.Schedule(func() {})
	if posted != 1 {
		t.Fatalf("expected 1 flush post, got %d", posted)
	}
}

func TestQueueScheduler_CoalescesPosts(t *testing.T) {
	queue := state.NewQueue()
	posted := 0
	scheduler := NewQueueScheduler(queue, func(msg Message) bool {
		if _, ok := msg.(QueueFlushMsg); ok {
			posted++
			return true
		}
		return false
	})

	scheduler.Schedule(func() {})
	scheduler.Schedule(func() {})
	if posted != 1 {
		t.Fatalf("expected 1 flush post, got %d", posted)
	}

	scheduler.resetPending()
	scheduler.Schedule(func() {})
	if posted != 2 {
		t.Fatalf("expected 2 flush posts after reset, got %d", posted)
	}
}

func TestQueueScheduler_RepostsOnFailedSend(t *testing.T) {
	queue := state.NewQueue()
	attempts := 0
	scheduler := NewQueueScheduler(queue, func(msg Message) bool {
		attempts++
		return false
	})

	scheduler.Schedule(func() {})
	scheduler.Schedule(func() {})
	if attempts != 2 {
		t.Fatalf("expected 2 post attempts, got %d", attempts)
	}
}
