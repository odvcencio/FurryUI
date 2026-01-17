package clipboard

import "testing"

func TestMemoryClipboardReadWrite(t *testing.T) {
	cb := &MemoryClipboard{}
	if !cb.Available() {
		t.Fatal("expected memory clipboard to be available")
	}
	if err := cb.Write("hello"); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	got, err := cb.Read()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if got != "hello" {
		t.Fatalf("read = %q, want %q", got, "hello")
	}
}

func TestMemoryClipboardNilReceiver(t *testing.T) {
	var cb *MemoryClipboard
	if err := cb.Write("noop"); err != nil {
		t.Fatalf("nil write failed: %v", err)
	}
	got, err := cb.Read()
	if err != nil {
		t.Fatalf("nil read failed: %v", err)
	}
	if got != "" {
		t.Fatalf("nil read = %q, want empty", got)
	}
	if !cb.Available() {
		t.Fatal("nil memory clipboard should report available")
	}
}

func TestUnavailableClipboard(t *testing.T) {
	cb := UnavailableClipboard{}
	if cb.Available() {
		t.Fatal("unavailable clipboard should report unavailable")
	}
	got, err := cb.Read()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if got != "" {
		t.Fatalf("read = %q, want empty", got)
	}
	if err := cb.Write("noop"); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
