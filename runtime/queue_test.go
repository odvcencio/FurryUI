package runtime

import (
	"testing"
	"time"

	"github.com/odvcencio/furry-ui/state"
)

func TestWithQueue_FlushOnTick(t *testing.T) {
	queue := state.NewQueue()
	calls := 0
	queue.Schedule(func() {
		calls++
	})

	update := WithQueue(queue, func(app *App, msg Message) bool { return false })
	if dirty := update(nil, TickMsg{Time: time.Now()}); !dirty {
		t.Fatalf("expected dirty after tick flush")
	}
	if calls != 1 {
		t.Fatalf("expected 1 callback after tick flush, got %d", calls)
	}
}

func TestWithQueue_FlushOnMessage(t *testing.T) {
	queue := state.NewQueue()
	calls := 0
	queue.Schedule(func() {
		calls++
	})

	update := WithQueue(queue, func(app *App, msg Message) bool { return false })
	if dirty := update(nil, QueueFlushMsg{}); !dirty {
		t.Fatalf("expected dirty after queue flush message")
	}
	if calls != 1 {
		t.Fatalf("expected 1 callback after queue flush message, got %d", calls)
	}
}

func TestWithQueue_NoFlushOnOtherMessages(t *testing.T) {
	queue := state.NewQueue()
	calls := 0
	queue.Schedule(func() {
		calls++
	})

	update := WithQueue(queue, func(app *App, msg Message) bool { return true })
	if dirty := update(nil, ResizeMsg{Width: 10, Height: 5}); !dirty {
		t.Fatalf("expected dirty from update")
	}
	if calls != 0 {
		t.Fatalf("expected no queue flush, got %d callbacks", calls)
	}
}
