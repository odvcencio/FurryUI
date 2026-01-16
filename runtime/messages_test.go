package runtime

import (
	"testing"
	"time"

	"github.com/odvcencio/furry-ui/terminal"
)

// TestMessages verifies all message types implement the Message interface.
func TestMessages_ImplementsInterface(t *testing.T) {
	messages := []Message{
		KeyMsg{Key: terminal.KeyEnter},
		ResizeMsg{Width: 80, Height: 24},
		MouseMsg{X: 10, Y: 20, Button: MouseLeft, Action: MousePress},
		PasteMsg{Text: "pasted text"},
		TickMsg{Time: time.Now()},
		QueueFlushMsg{},
		InvalidateMsg{},
	}

	for i, msg := range messages {
		if msg == nil {
			t.Errorf("Message %d is nil", i)
		}
	}
}

func TestKeyMsg(t *testing.T) {
	msg := KeyMsg{
		Key:   terminal.KeyRune,
		Rune:  'a',
		Alt:   true,
		Ctrl:  false,
		Shift: true,
	}
	msg.isMessage()

	if msg.Key != terminal.KeyRune {
		t.Errorf("KeyMsg.Key = %v, want KeyRune", msg.Key)
	}
	if msg.Rune != 'a' {
		t.Errorf("KeyMsg.Rune = %c, want 'a'", msg.Rune)
	}
	if !msg.Alt {
		t.Error("KeyMsg.Alt should be true")
	}
	if msg.Ctrl {
		t.Error("KeyMsg.Ctrl should be false")
	}
	if !msg.Shift {
		t.Error("KeyMsg.Shift should be true")
	}
}

func TestResizeMsg(t *testing.T) {
	msg := ResizeMsg{Width: 120, Height: 40}
	msg.isMessage()

	if msg.Width != 120 {
		t.Errorf("ResizeMsg.Width = %d, want 120", msg.Width)
	}
	if msg.Height != 40 {
		t.Errorf("ResizeMsg.Height = %d, want 40", msg.Height)
	}
}

func TestMouseMsg(t *testing.T) {
	msg := MouseMsg{
		X:      50,
		Y:      25,
		Button: MouseRight,
		Action: MouseRelease,
	}
	msg.isMessage()

	if msg.X != 50 {
		t.Errorf("MouseMsg.X = %d, want 50", msg.X)
	}
	if msg.Y != 25 {
		t.Errorf("MouseMsg.Y = %d, want 25", msg.Y)
	}
	if msg.Button != MouseRight {
		t.Errorf("MouseMsg.Button = %v, want MouseRight", msg.Button)
	}
	if msg.Action != MouseRelease {
		t.Errorf("MouseMsg.Action = %v, want MouseRelease", msg.Action)
	}
}

func TestMouseButtons(t *testing.T) {
	buttons := []MouseButton{
		MouseNone,
		MouseLeft,
		MouseMiddle,
		MouseRight,
		MouseWheelUp,
		MouseWheelDown,
	}

	// Verify they are distinct values
	seen := make(map[MouseButton]bool)
	for _, b := range buttons {
		if seen[b] {
			t.Errorf("Duplicate MouseButton value: %d", b)
		}
		seen[b] = true
	}
}

func TestMouseActions(t *testing.T) {
	actions := []MouseAction{
		MousePress,
		MouseRelease,
		MouseMove,
	}

	// Verify they are distinct values
	seen := make(map[MouseAction]bool)
	for _, a := range actions {
		if seen[a] {
			t.Errorf("Duplicate MouseAction value: %d", a)
		}
		seen[a] = true
	}
}

func TestPasteMsg(t *testing.T) {
	msg := PasteMsg{Text: "Hello, World!"}
	msg.isMessage()

	if msg.Text != "Hello, World!" {
		t.Errorf("PasteMsg.Text = %q, want %q", msg.Text, "Hello, World!")
	}
}

func TestTickMsg(t *testing.T) {
	now := time.Now()
	msg := TickMsg{Time: now}
	msg.isMessage()

	if msg.Time != now {
		t.Errorf("TickMsg.Time = %v, want %v", msg.Time, now)
	}
}
