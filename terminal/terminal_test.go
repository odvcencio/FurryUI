package terminal

import "testing"

func TestKeyConstants(t *testing.T) {
	// Verify key constants are defined
	keys := []Key{
		KeyNone, KeyRune, KeyEnter, KeyBackspace, KeyTab, KeyEscape,
		KeyUp, KeyDown, KeyLeft, KeyRight, KeyHome, KeyEnd,
		KeyPageUp, KeyPageDown, KeyDelete, KeyInsert,
		KeyF1, KeyF2, KeyF3, KeyF4, KeyF5, KeyF6,
		KeyF7, KeyF8, KeyF9, KeyF10, KeyF11, KeyF12,
		KeyCtrlC, KeyCtrlD, KeyCtrlZ,
	}

	// Ensure all are unique
	seen := make(map[Key]bool)
	for _, k := range keys {
		if seen[k] {
			t.Errorf("duplicate key constant: %d", k)
		}
		seen[k] = true
	}
}

func TestEventInterface(t *testing.T) {
	// Verify event types implement Event interface
	var _ Event = KeyEvent{}
	var _ Event = ResizeEvent{}
}

func TestKeyEvent(t *testing.T) {
	ev := KeyEvent{
		Key:   KeyRune,
		Rune:  'a',
		Alt:   true,
		Ctrl:  false,
		Shift: true,
	}

	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
	if !ev.Alt {
		t.Error("expected Alt=true")
	}
	if ev.Ctrl {
		t.Error("expected Ctrl=false")
	}
	if !ev.Shift {
		t.Error("expected Shift=true")
	}
}

func TestResizeEvent(t *testing.T) {
	ev := ResizeEvent{Width: 120, Height: 40}

	if ev.Width != 120 {
		t.Errorf("expected Width=120, got %d", ev.Width)
	}
	if ev.Height != 40 {
		t.Errorf("expected Height=40, got %d", ev.Height)
	}
}
