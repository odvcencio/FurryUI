package toast

import (
	"testing"
	"time"
)

func TestToastManagerShowDismiss(t *testing.T) {
	manager := NewToastManager()
	var snapshots [][]*Toast
	manager.SetOnChange(func(items []*Toast) {
		copied := make([]*Toast, len(items))
		copy(copied, items)
		snapshots = append(snapshots, copied)
	})

	id := manager.Show(ToastInfo, "Hello", "World", time.Hour)
	if id == "" {
		t.Fatal("expected non-empty toast id")
	}
	if len(snapshots) < 2 {
		t.Fatalf("expected snapshots for show, got %d", len(snapshots))
	}
	if got := snapshots[1]; len(got) != 1 || got[0].ID != id {
		t.Fatalf("unexpected toast snapshot: %#v", got)
	}

	manager.Dismiss(id)
	if len(snapshots) < 3 {
		t.Fatalf("expected snapshots for dismiss, got %d", len(snapshots))
	}
	last := snapshots[len(snapshots)-1]
	if len(last) != 0 {
		t.Fatalf("expected no toasts after dismiss, got %#v", last)
	}
}

func TestToastManagerMaxCount(t *testing.T) {
	manager := NewToastManager()
	manager.maxCount = 1

	first := manager.Show(ToastInfo, "First", "Toast", time.Hour)
	second := manager.Show(ToastInfo, "Second", "Toast", time.Hour)

	if first == "" || second == "" {
		t.Fatal("expected non-empty toast ids")
	}
	if len(manager.toasts) != 1 {
		t.Fatalf("expected 1 toast after overflow, got %d", len(manager.toasts))
	}
	if manager.toasts[0].ID != second {
		t.Fatalf("expected latest toast retained, got %s", manager.toasts[0].ID)
	}
}
