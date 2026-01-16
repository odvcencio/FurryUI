package state

import "testing"

func TestSignal_SetAndSubscribe(t *testing.T) {
	sig := NewSignal(1)
	calls := 0

	unsub := sig.Subscribe(func() {
		calls++
	})

	if calls != 0 {
		t.Fatalf("expected no calls before set, got %d", calls)
	}
	if !sig.Set(2) {
		t.Fatalf("expected set to report change")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call after set, got %d", calls)
	}

	unsub()
	sig.Set(3)
	if calls != 1 {
		t.Fatalf("expected no calls after unsubscribe, got %d", calls)
	}
}

func TestSignal_SetEqualFunc(t *testing.T) {
	sig := NewSignal(5)
	sig.SetEqualFunc(EqualComparable[int])

	if sig.Set(5) {
		t.Fatalf("expected set of equal value to report no change")
	}
	if !sig.Set(6) {
		t.Fatalf("expected set of new value to report change")
	}
}

func TestSignal_Update(t *testing.T) {
	sig := NewSignal(1)
	sig.SetEqualFunc(EqualComparable[int])

	if !sig.Update(func(v int) int { return v + 1 }) {
		t.Fatalf("expected update to report change")
	}
	if sig.Get() != 2 {
		t.Fatalf("expected updated value 2, got %d", sig.Get())
	}
	if sig.Update(func(v int) int { return v }) {
		t.Fatalf("expected update of equal value to report no change")
	}
	if sig.Update(nil) {
		t.Fatalf("expected nil update to report no change")
	}
}

func TestSignal_SubscribeWithScheduler(t *testing.T) {
	sig := NewSignal(1)
	queue := NewQueue()
	calls := 0

	sig.SubscribeWithScheduler(queue, func() {
		calls++
	})

	if !sig.Set(2) {
		t.Fatalf("expected set to report change")
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
}
