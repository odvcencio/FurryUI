package runtime

import (
	"context"
	"testing"
)

// TestCommands verifies all command types implement the Command interface.
func TestCommands_ImplementsInterface(t *testing.T) {
	// This test verifies that all command types compile correctly
	// and implement the Command interface via Command().

	commands := []Command{
		Quit{},
		Refresh{},
		SendMsg{Message: ResizeMsg{Width: 1, Height: 1}},
		Submit{Text: "test"},
		Cancel{},
		Effect{Run: func(ctx context.Context, post PostFunc) {}},
		FileSelected{Path: "/test"},
		FocusNext{},
		FocusPrev{},
		PushOverlay{Widget: nil, Modal: false},
		PopOverlay{},
		PaletteSelected{ID: "item1", Data: nil},
	}

	for i, cmd := range commands {
		if cmd == nil {
			t.Errorf("Command %d is nil", i)
		}
	}
}

func TestQuit(t *testing.T) {
	cmd := Quit{}
	// Verify it has the Command method
	cmd.Command()
}

func TestRefresh(t *testing.T) {
	cmd := Refresh{}
	cmd.Command()
}

func TestSend(t *testing.T) {
	msg := ResizeMsg{Width: 10, Height: 5}
	cmd := Send(msg)
	cmd.Command()
	if sendMsg, ok := cmd.(SendMsg); ok {
		if sendMsg.Message != msg {
			t.Fatalf("SendMsg.Message mismatch")
		}
	} else {
		t.Fatalf("expected Send to return SendMsg, got %T", cmd)
	}
}

func TestSubmit(t *testing.T) {
	cmd := Submit{Text: "hello world"}
	cmd.Command()

	if cmd.Text != "hello world" {
		t.Errorf("Submit.Text = %q, want %q", cmd.Text, "hello world")
	}
}

func TestCancel(t *testing.T) {
	cmd := Cancel{}
	cmd.Command()
}

func TestEffect(t *testing.T) {
	calls := 0
	cmd := Effect{Run: func(ctx context.Context, post PostFunc) {
		calls++
	}}
	cmd.Command()
	if cmd.Run == nil {
		t.Fatal("expected effect run function")
	}
	cmd.Run(context.Background(), func(Message) bool { return true })
	if calls != 1 {
		t.Fatalf("expected effect run to be called once, got %d", calls)
	}
}

func TestFileSelected(t *testing.T) {
	cmd := FileSelected{Path: "/home/user/file.txt"}
	cmd.Command()

	if cmd.Path != "/home/user/file.txt" {
		t.Errorf("FileSelected.Path = %q, want %q", cmd.Path, "/home/user/file.txt")
	}
}

func TestFocusNext(t *testing.T) {
	cmd := FocusNext{}
	cmd.Command()
}

func TestFocusPrev(t *testing.T) {
	cmd := FocusPrev{}
	cmd.Command()
}

func TestPushOverlay(t *testing.T) {
	w := &testSimpleWidget{}
	cmd := PushOverlay{Widget: w, Modal: true}
	cmd.Command()

	if cmd.Widget != w {
		t.Error("PushOverlay.Widget should be the widget")
	}
	if !cmd.Modal {
		t.Error("PushOverlay.Modal should be true")
	}
}

type testSimpleWidget struct{}

func (t *testSimpleWidget) Measure(c Constraints) Size { return Size{} }
func (t *testSimpleWidget) Layout(bounds Rect)         {}
func (t *testSimpleWidget) Render(ctx RenderContext)   {}
func (t *testSimpleWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func TestPopOverlay(t *testing.T) {
	cmd := PopOverlay{}
	cmd.Command()
}

func TestPaletteSelected(t *testing.T) {
	data := map[string]int{"count": 42}
	cmd := PaletteSelected{ID: "option-1", Data: data}
	cmd.Command()

	if cmd.ID != "option-1" {
		t.Errorf("PaletteSelected.ID = %q, want %q", cmd.ID, "option-1")
	}
	if cmd.Data == nil {
		t.Error("PaletteSelected.Data should not be nil")
	}
}
