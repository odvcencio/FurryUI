package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

type testInput struct {
	bounds   runtime.Rect
	focused  bool
	label    string
	value    string
	disabled bool
}

func (t *testInput) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: 12, Height: 1})
}

func (t *testInput) Layout(bounds runtime.Rect) {
	t.bounds = bounds
}

func (t *testInput) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	text := t.value
	if text == "" {
		text = " "
	}
	ctx.Buffer.SetString(t.bounds.X, t.bounds.Y, text, backend.DefaultStyle())
}

func (t *testInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t.disabled || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyBackspace:
		if len(t.value) > 0 {
			t.value = t.value[:len(t.value)-1]
		}
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune != 0 {
			t.value += string(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *testInput) Bounds() runtime.Rect { return t.bounds }

func (t *testInput) CanFocus() bool { return !t.disabled }
func (t *testInput) Focus()         { t.focused = true }
func (t *testInput) Blur()          { t.focused = false }
func (t *testInput) IsFocused() bool {
	return t.focused
}

func (t *testInput) AccessibleRole() accessibility.Role { return accessibility.RoleTextbox }
func (t *testInput) AccessibleLabel() string           { return t.label }
func (t *testInput) AccessibleDescription() string     { return "" }
func (t *testInput) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{Disabled: t.disabled}
}
func (t *testInput) AccessibleValue() *accessibility.ValueInfo {
	return &accessibility.ValueInfo{Text: t.value}
}
func (t *testInput) Text() string { return t.value }

type testButton struct {
	bounds   runtime.Rect
	focused  bool
	label    string
	disabled bool
	clicked  bool
}

func (t *testButton) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: len(t.label) + 2, Height: 1})
}

func (t *testButton) Layout(bounds runtime.Rect) {
	t.bounds = bounds
}

func (t *testButton) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	ctx.Buffer.SetString(t.bounds.X, t.bounds.Y, "["+t.label+"]", backend.DefaultStyle())
}

func (t *testButton) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t.disabled || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		t.clicked = true
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *testButton) Bounds() runtime.Rect { return t.bounds }

func (t *testButton) CanFocus() bool { return !t.disabled }
func (t *testButton) Focus()         { t.focused = true }
func (t *testButton) Blur()          { t.focused = false }
func (t *testButton) IsFocused() bool {
	return t.focused
}

func (t *testButton) AccessibleRole() accessibility.Role { return accessibility.RoleButton }
func (t *testButton) AccessibleLabel() string           { return t.label }
func (t *testButton) AccessibleDescription() string     { return "" }
func (t *testButton) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{Disabled: t.disabled}
}
func (t *testButton) AccessibleValue() *accessibility.ValueInfo { return nil }

func TestAgentSnapshotAndActions(t *testing.T) {
	input := &testInput{label: "Name"}
	button := &testButton{label: "Submit"}
	root := runtime.VBox(runtime.Fixed(input), runtime.Fixed(button)).WithGap(1)

	simBackend := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           simBackend,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app, Sim: simBackend, Width: 40, Height: 10})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()
	defer func() {
		app.Post(runtime.Quit{})
		<-done
	}()

	deadline := time.Now().Add(2 * time.Second)
	for app.Screen() == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if app.Screen() == nil {
		t.Fatal("screen not initialized")
	}
	agt.SetScreen(app.Screen())

	if err := agt.WaitForWidget("Name", time.Second); err != nil {
		t.Fatalf("wait for widget: %v", err)
	}

	if err := agt.Focus("Name"); err != nil {
		t.Fatalf("focus name: %v", err)
	}

	if err := agt.Type("Name", "Alice"); err != nil {
		t.Fatalf("type name: %v", err)
	}

	value, err := agt.GetValue("Name")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if value != "Alice" {
		t.Fatalf("value = %q, want %q", value, "Alice")
	}

	if err := agt.Activate("Submit"); err != nil {
		t.Fatalf("activate submit: %v", err)
	}
	if !button.clicked {
		t.Fatal("expected submit to be activated")
	}

	raw, err := agt.SnapshotJSON()
	if err != nil {
		t.Fatalf("snapshot json: %v", err)
	}
	if !strings.Contains(string(raw), "\"widgets\"") {
		t.Fatalf("snapshot json missing widgets: %s", string(raw))
	}
}
