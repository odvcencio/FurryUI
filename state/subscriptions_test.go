package state

import "testing"

func TestSubscriptions_Clear(t *testing.T) {
	subs := &Subscriptions{}
	calls := 0

	subs.Add(func() { calls++ })
	subs.Add(func() { calls++ })

	subs.Clear()
	if calls != 2 {
		t.Fatalf("expected 2 unsubscribe calls, got %d", calls)
	}

	subs.Clear()
	if calls != 2 {
		t.Fatalf("expected no extra calls after clear, got %d", calls)
	}
}

func TestSubscriptions_Scheduler(t *testing.T) {
	sig := NewSignal(1)
	queue := NewQueue()
	subs := NewSubscriptions(queue)
	calls := 0

	subs.Observe(sig, func() {
		calls++
	})

	if !sig.Set(2) {
		t.Fatalf("expected signal to change")
	}
	if calls != 0 {
		t.Fatalf("expected callback to be queued, got %d", calls)
	}
	if flushed := queue.Flush(); flushed != 1 {
		t.Fatalf("expected 1 callback flushed, got %d", flushed)
	}
	if calls != 1 {
		t.Fatalf("expected callback after flush, got %d", calls)
	}

	subs.Clear()
	sig.Set(3)
	queue.Flush()
	if calls != 1 {
		t.Fatalf("expected no callbacks after clear, got %d", calls)
	}
}

func TestSubscriptions_SetScheduler(t *testing.T) {
	sig := NewSignal(1)
	queue := NewQueue()
	subs := &Subscriptions{}
	calls := 0

	subs.SetScheduler(queue)
	subs.Observe(sig, func() {
		calls++
	})

	sig.Set(2)
	queue.Flush()
	if calls != 1 {
		t.Fatalf("expected callback with scheduler, got %d", calls)
	}
}
