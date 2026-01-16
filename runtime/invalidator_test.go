package runtime

import "testing"

func TestInvalidator_PostsInvalidate(t *testing.T) {
	posted := 0
	invalidator := NewInvalidator(func(msg Message) bool {
		if _, ok := msg.(InvalidateMsg); ok {
			posted++
			return true
		}
		return false
	})

	invalidator.Invalidate()
	invalidator.Invalidate()
	if posted != 1 {
		t.Fatalf("expected 1 invalidate post, got %d", posted)
	}

	invalidator.resetPending()
	invalidator.Invalidate()
	if posted != 2 {
		t.Fatalf("expected 2 invalidate posts after reset, got %d", posted)
	}
}

func TestInvalidator_RepostsOnFailedSend(t *testing.T) {
	attempts := 0
	invalidator := NewInvalidator(func(msg Message) bool {
		attempts++
		return false
	})

	invalidator.Invalidate()
	invalidator.Invalidate()
	if attempts != 2 {
		t.Fatalf("expected 2 post attempts, got %d", attempts)
	}
}

func TestInvalidator_Schedule(t *testing.T) {
	posted := 0
	calls := 0
	invalidator := NewInvalidator(func(msg Message) bool {
		if _, ok := msg.(InvalidateMsg); ok {
			posted++
			return true
		}
		return false
	})

	invalidator.Schedule(func() { calls++ })
	if calls != 1 {
		t.Fatalf("expected schedule to run callback, got %d", calls)
	}
	if posted != 1 {
		t.Fatalf("expected invalidate post after schedule, got %d", posted)
	}
}
