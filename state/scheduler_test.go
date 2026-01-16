package state

import "testing"

func TestQueue_Flush(t *testing.T) {
	queue := NewQueue()
	calls := make([]int, 0, 2)

	queue.Schedule(func() {
		calls = append(calls, 1)
	})
	queue.Schedule(func() {
		calls = append(calls, 2)
	})

	if flushed := queue.Flush(); flushed != 2 {
		t.Fatalf("expected 2 callbacks flushed, got %d", flushed)
	}
	if len(calls) != 2 || calls[0] != 1 || calls[1] != 2 {
		t.Fatalf("unexpected callback order: %v", calls)
	}
	if flushed := queue.Flush(); flushed != 0 {
		t.Fatalf("expected empty flush, got %d", flushed)
	}
}
