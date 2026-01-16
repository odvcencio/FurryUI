package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/state"
)

func TestSignalLabel_LifecycleQueue(t *testing.T) {
	sig := state.NewSignal("start")
	queue := state.NewQueue()
	label := NewSignalLabel(sig, queue)

	label.Mount()
	if label.text != "start" {
		t.Fatalf("expected initial text start, got %q", label.text)
	}

	sig.Set("next")
	if label.text != "start" {
		t.Fatalf("expected text to update after flush, got %q", label.text)
	}
	if flushed := queue.Flush(); flushed != 1 {
		t.Fatalf("expected 1 queued callback, got %d", flushed)
	}
	if label.text != "next" {
		t.Fatalf("expected updated text next, got %q", label.text)
	}

	label.Unmount()
	sig.Set("final")
	if flushed := queue.Flush(); flushed != 0 {
		t.Fatalf("expected no queued callbacks after unmount, got %d", flushed)
	}
	if label.text != "next" {
		t.Fatalf("expected text to remain next after unmount, got %q", label.text)
	}
}
