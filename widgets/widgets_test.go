package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

func TestText_Measure(t *testing.T) {
	text := NewText("Hello\nWorld")

	size := text.Measure(runtime.Loose(100, 100))

	if size.Width != 5 {
		t.Errorf("Width = %d, want 5", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2", size.Height)
	}
}

func TestText_Render(t *testing.T) {
	text := NewText("Hi").WithStyle(backend.DefaultStyle())

	text.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})

	buf := runtime.NewBuffer(10, 1)
	ctx := runtime.RenderContext{Buffer: buf, Bounds: runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1}}

	text.Render(ctx)

	if buf.Get(0, 0).Rune != 'H' {
		t.Errorf("Expected 'H' at (0,0), got %c", buf.Get(0, 0).Rune)
	}
	if buf.Get(1, 0).Rune != 'i' {
		t.Errorf("Expected 'i' at (1,0), got %c", buf.Get(1, 0).Rune)
	}
}

func TestLabel_Alignment(t *testing.T) {
	tests := []struct {
		align    Alignment
		expected int // X position of 'H'
	}{
		{AlignLeft, 0},
		{AlignCenter, 4}, // (10-2)/2 = 4
		{AlignRight, 8},  // 10-2 = 8
	}

	for _, tc := range tests {
		label := NewLabel("Hi").WithAlignment(tc.align)
		label.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})

		buf := runtime.NewBuffer(10, 1)
		ctx := runtime.RenderContext{Buffer: buf}

		label.Render(ctx)

		if buf.Get(tc.expected, 0).Rune != 'H' {
			t.Errorf("Align %v: expected 'H' at x=%d, got %c", tc.align, tc.expected, buf.Get(tc.expected, 0).Rune)
		}
	}
}

func TestInput_TextOperations(t *testing.T) {
	input := NewInput()

	if input.Text() != "" {
		t.Error("New input should be empty")
	}

	input.SetText("Hello")
	if input.Text() != "Hello" {
		t.Errorf("SetText failed, got %s", input.Text())
	}

	if input.CursorPos() != 5 {
		t.Errorf("Cursor should be at end, got %d", input.CursorPos())
	}

	input.Clear()
	if input.Text() != "" || input.CursorPos() != 0 {
		t.Error("Clear failed")
	}
}

func TestInput_HandleMessage_Typing(t *testing.T) {
	input := NewInput()
	input.Focus()

	// Type 'H'
	result := input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'H'})
	if !result.Handled {
		t.Error("KeyRune should be handled")
	}
	if input.Text() != "H" {
		t.Errorf("Text = %s, want H", input.Text())
	}
}

func TestInput_HandleMessage_Unfocused(t *testing.T) {
	input := NewInput()
	// Not focused

	result := input.HandleMessage(runtime.KeyMsg{Key: 7, Rune: 'H'})
	if result.Handled {
		t.Error("Unfocused input should not handle messages")
	}
}

func TestPanel_WithBorder(t *testing.T) {
	label := NewLabel("Test")
	panel := NewPanel(label).WithBorder(backend.DefaultStyle())

	size := panel.Measure(runtime.Loose(20, 10))

	// Label is 4 chars wide + 2 for border
	if size.Width != 6 {
		t.Errorf("Width = %d, want 6", size.Width)
	}
	// Label is 1 char tall + 2 for border
	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
}

func TestPanel_Render(t *testing.T) {
	label := NewLabel("Hi")
	panel := NewPanel(label).WithBorder(backend.DefaultStyle())

	panel.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	buf := runtime.NewBuffer(10, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	panel.Render(ctx)

	// Check border corners
	if buf.Get(0, 0).Rune != '╭' {
		t.Errorf("Top-left corner = %c, want ╭", buf.Get(0, 0).Rune)
	}
	if buf.Get(9, 0).Rune != '╮' {
		t.Errorf("Top-right corner = %c, want ╮", buf.Get(9, 0).Rune)
	}
}

func TestBox_PassesThrough(t *testing.T) {
	label := NewLabel("Hi")
	box := NewBox(label)

	size := box.Measure(runtime.Loose(20, 10))
	if size.Width != 2 || size.Height != 1 {
		t.Errorf("Box measure = %v, want {2,1}", size)
	}
}

func TestBase_Focus(t *testing.T) {
	var b Base

	if b.IsFocused() {
		t.Error("New base should not be focused")
	}

	b.Focus()
	if !b.IsFocused() {
		t.Error("After Focus(), should be focused")
	}

	b.Blur()
	if b.IsFocused() {
		t.Error("After Blur(), should not be focused")
	}
}

func TestFocusableBase_CanFocus(t *testing.T) {
	var b Base
	var fb FocusableBase

	if b.CanFocus() {
		t.Error("Base should not be focusable")
	}
	if !fb.CanFocus() {
		t.Error("FocusableBase should be focusable")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxWidth int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 8, "Hello..."},
		{"Hi", 2, "Hi"},
		{"Hello", 3, "Hel"}, // maxWidth <= 3 just truncates without ellipsis
	}

	for _, tc := range tests {
		got := truncateString(tc.input, tc.maxWidth)
		if got != tc.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tc.input, tc.maxWidth, got, tc.expected)
		}
	}
}

func TestInput_SetPlaceholder(t *testing.T) {
	input := NewInput()
	input.SetPlaceholder("Enter text...")

	if input.placeholder != "Enter text..." {
		t.Errorf("expected placeholder 'Enter text...', got '%s'", input.placeholder)
	}
}

func TestInput_SetStyle(t *testing.T) {
	input := NewInput()
	style := backend.DefaultStyle().Bold(true)
	input.SetStyle(style)

	if input.style != style {
		t.Error("style not set correctly")
	}
}

func TestInput_SetFocusStyle(t *testing.T) {
	input := NewInput()
	style := backend.DefaultStyle().Italic(true)
	input.SetFocusStyle(style)

	if input.focusStyle != style {
		t.Error("focusStyle not set correctly")
	}
}

func TestInput_OnSubmit(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("test")

	var submitted string
	input.OnSubmit(func(text string) {
		submitted = text
	})

	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if submitted != "test" {
		t.Errorf("expected submitted 'test', got '%s'", submitted)
	}
}

func TestInput_OnChange(t *testing.T) {
	input := NewInput()
	input.Focus()

	var changed string
	input.OnChange(func(text string) {
		changed = text
	})

	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})

	if changed != "a" {
		t.Errorf("expected changed 'a', got '%s'", changed)
	}
}

func TestInput_Measure(t *testing.T) {
	input := NewInput()

	size := input.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 80 {
		t.Errorf("expected width 80, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("expected height 1, got %d", size.Height)
	}
}

func TestInput_Render(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("Hello")
	input.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	buf := runtime.NewBuffer(20, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	input.Render(ctx)

	// Check first character
	cell := buf.Get(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("expected 'H' at (0,0), got '%c'", cell.Rune)
	}
}

func TestInput_Render_Placeholder(t *testing.T) {
	input := NewInput()
	input.SetPlaceholder("Enter text...")
	input.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	buf := runtime.NewBuffer(20, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	input.Render(ctx)

	// Check placeholder text
	cell := buf.Get(0, 0)
	if cell.Rune != 'E' {
		t.Errorf("expected 'E' at (0,0) for placeholder, got '%c'", cell.Rune)
	}
}

func TestInput_Render_Empty(t *testing.T) {
	input := NewInput()
	input.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})

	buf := runtime.NewBuffer(20, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic with empty bounds
	input.Render(ctx)
}

func TestInput_HandleMessage_Delete(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("test")
	input.cursorPos = 0 // Move cursor to beginning

	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})

	if input.Text() != "est" {
		t.Errorf("expected 'est', got '%s'", input.Text())
	}
}

func TestInput_HandleMessage_CtrlLeft(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("hello world")
	// Cursor is at end (position 11)

	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft, Ctrl: true})

	// Should move to beginning of "world"
	if input.cursorPos != 6 {
		t.Errorf("expected cursor at 6, got %d", input.cursorPos)
	}
}

func TestInput_HandleMessage_CtrlRight(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("hello world")
	input.cursorPos = 0

	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight, Ctrl: true})

	// Should move to end of "hello" and past space
	if input.cursorPos != 6 {
		t.Errorf("expected cursor at 6, got %d", input.cursorPos)
	}
}

func TestInput_HandleMessage_Tab(t *testing.T) {
	input := NewInput()
	input.Focus()

	result := input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})

	if len(result.Commands) == 0 {
		t.Error("expected FocusNext command")
	}
	_, ok := result.Commands[0].(runtime.FocusNext)
	if !ok {
		t.Errorf("expected FocusNext, got %T", result.Commands[0])
	}
}

func TestInput_HandleMessage_ShiftTab(t *testing.T) {
	input := NewInput()
	input.Focus()

	result := input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})

	if len(result.Commands) == 0 {
		t.Error("expected FocusPrev command")
	}
	_, ok := result.Commands[0].(runtime.FocusPrev)
	if !ok {
		t.Errorf("expected FocusPrev, got %T", result.Commands[0])
	}
}

func TestInput_HandleMessage_Escape(t *testing.T) {
	input := NewInput()
	input.Focus()

	result := input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if len(result.Commands) == 0 {
		t.Error("expected Cancel command")
	}
	_, ok := result.Commands[0].(runtime.Cancel)
	if !ok {
		t.Errorf("expected Cancel, got %T", result.Commands[0])
	}
}

func TestMultilineInput_New(t *testing.T) {
	mi := NewMultilineInput()

	if mi == nil {
		t.Fatal("expected non-nil MultilineInput")
	}
	if mi.Text() != "" {
		t.Errorf("expected empty text, got '%s'", mi.Text())
	}
}

func TestMultilineInput_SetText(t *testing.T) {
	mi := NewMultilineInput()
	mi.SetText("line1\nline2\nline3")

	if mi.Text() != "line1\nline2\nline3" {
		t.Errorf("expected 'line1\\nline2\\nline3', got '%s'", mi.Text())
	}
	if len(mi.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(mi.lines))
	}
}

func TestMultilineInput_Clear(t *testing.T) {
	mi := NewMultilineInput()
	mi.SetText("some\ntext")
	mi.Clear()

	if mi.Text() != "" {
		t.Errorf("expected empty text, got '%s'", mi.Text())
	}
}

func TestMultilineInput_OnSubmit(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("test")

	var submitted string
	mi.OnSubmit(func(text string) {
		submitted = text
	})

	// Ctrl+Enter submits
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter, Ctrl: true})

	if submitted != "test" {
		t.Errorf("expected submitted 'test', got '%s'", submitted)
	}
}

func TestMultilineInput_Measure(t *testing.T) {
	mi := NewMultilineInput()
	mi.SetText("line1\nline2\nline3\nline4\nline5")

	size := mi.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 80 {
		t.Errorf("expected width 80, got %d", size.Width)
	}
	if size.Height != 5 {
		t.Errorf("expected height 5, got %d", size.Height)
	}
}

func TestMultilineInput_Measure_Min(t *testing.T) {
	mi := NewMultilineInput()

	size := mi.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	// Minimum height is 3
	if size.Height != 3 {
		t.Errorf("expected minimum height 3, got %d", size.Height)
	}
}

func TestMultilineInput_Render(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("Hello\nWorld")
	mi.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	buf := runtime.NewBuffer(20, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	mi.Render(ctx)

	// Check first line
	cell := buf.Get(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("expected 'H' at (0,0), got '%c'", cell.Rune)
	}
}

func TestMultilineInput_Render_Empty(t *testing.T) {
	mi := NewMultilineInput()
	mi.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})

	buf := runtime.NewBuffer(20, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	mi.Render(ctx)
}

func TestMultilineInput_HandleMessage_Enter(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("hello")
	mi.cursorX = 5
	mi.cursorY = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if len(mi.lines) != 2 {
		t.Errorf("expected 2 lines after Enter, got %d", len(mi.lines))
	}
}

func TestMultilineInput_HandleMessage_Backspace(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("hello")
	mi.cursorX = 5
	mi.cursorY = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	if mi.lines[0] != "hell" {
		t.Errorf("expected 'hell', got '%s'", mi.lines[0])
	}
}

func TestMultilineInput_HandleMessage_BackspaceJoinLines(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("hello\nworld")
	mi.cursorY = 1
	mi.cursorX = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	if len(mi.lines) != 1 {
		t.Errorf("expected 1 line after backspace at line start, got %d", len(mi.lines))
	}
	if mi.lines[0] != "helloworld" {
		t.Errorf("expected 'helloworld', got '%s'", mi.lines[0])
	}
}

func TestMultilineInput_HandleMessage_UpDown(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("line1\nline2\nline3")
	mi.cursorY = 1
	mi.cursorX = 0
	mi.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if mi.cursorY != 0 {
		t.Errorf("expected cursorY 0, got %d", mi.cursorY)
	}

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if mi.cursorY != 1 {
		t.Errorf("expected cursorY 1, got %d", mi.cursorY)
	}
}

func TestMultilineInput_HandleMessage_LeftRight(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("ab")
	mi.cursorX = 1
	mi.cursorY = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if mi.cursorX != 0 {
		t.Errorf("expected cursorX 0, got %d", mi.cursorX)
	}

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if mi.cursorX != 1 {
		t.Errorf("expected cursorX 1, got %d", mi.cursorX)
	}
}

func TestMultilineInput_HandleMessage_LeftAtLineStart(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("hello\nworld")
	mi.cursorY = 1
	mi.cursorX = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})

	if mi.cursorY != 0 {
		t.Errorf("expected cursorY 0, got %d", mi.cursorY)
	}
	if mi.cursorX != 5 {
		t.Errorf("expected cursorX 5, got %d", mi.cursorX)
	}
}

func TestMultilineInput_HandleMessage_RightAtLineEnd(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("hello\nworld")
	mi.cursorY = 0
	mi.cursorX = 5

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})

	if mi.cursorY != 1 {
		t.Errorf("expected cursorY 1, got %d", mi.cursorY)
	}
	if mi.cursorX != 0 {
		t.Errorf("expected cursorX 0, got %d", mi.cursorX)
	}
}

func TestMultilineInput_HandleMessage_Rune(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()
	mi.SetText("test")
	mi.cursorX = 0
	mi.cursorY = 0

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'X'})

	if mi.lines[0] != "Xtest" {
		t.Errorf("expected 'Xtest', got '%s'", mi.lines[0])
	}
}

func TestMultilineInput_HandleMessage_Escape(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if len(result.Commands) == 0 {
		t.Error("expected Cancel command")
	}
}

func TestMultilineInput_HandleMessage_Unfocused(t *testing.T) {
	mi := NewMultilineInput()
	// Not focused

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})

	if result.Handled {
		t.Error("unfocused should not handle messages")
	}
}

func TestMultilineInput_HandleMessage_NonKeyMsg(t *testing.T) {
	mi := NewMultilineInput()
	mi.Focus()

	result := mi.HandleMessage(runtime.ResizeMsg{Width: 80, Height: 24})

	if result.Handled {
		t.Error("non-key message should not be handled")
	}
}

// Test Text widget additional methods
func TestText_SetText(t *testing.T) {
	text := NewText("initial")
	text.SetText("updated")

	if text.Text() != "updated" {
		t.Errorf("expected 'updated', got '%s'", text.Text())
	}
}

func TestText_SetStyle(t *testing.T) {
	text := NewText("test")
	style := backend.DefaultStyle().Bold(true)
	text.SetStyle(style)

	// No panic means success
}

// Test Label additional methods
func TestLabel_SetText(t *testing.T) {
	label := NewLabel("initial")
	label.SetText("updated")

	if label.text != "updated" {
		t.Errorf("expected 'updated', got '%s'", label.text)
	}
}

func TestLabel_SetStyle(t *testing.T) {
	label := NewLabel("test")
	style := backend.DefaultStyle().Bold(true)
	label.SetStyle(style)

	// No panic means success
}

func TestLabel_SetAlignment(t *testing.T) {
	label := NewLabel("test")
	label.SetAlignment(AlignCenter)

	if label.alignment != AlignCenter {
		t.Errorf("expected AlignCenter, got %d", label.alignment)
	}
}

func TestLabel_WithStyle(t *testing.T) {
	label := NewLabel("test")
	style := backend.DefaultStyle().Bold(true)
	result := label.WithStyle(style)

	if result != label {
		t.Error("WithStyle should return same label for chaining")
	}
}

// Test Panel additional methods
func TestPanel_SetStyle(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)
	style := backend.DefaultStyle().Bold(true)
	panel.SetStyle(style)

	// No panic means success
}

func TestPanel_WithStyle(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)
	style := backend.DefaultStyle()
	result := panel.WithStyle(style)

	if result != panel {
		t.Error("WithStyle should return same panel for chaining")
	}
}

func TestPanel_SetBorder(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)
	panel.SetBorder(true)

	if !panel.hasBorder {
		t.Error("expected hasBorder to be true")
	}
}

func TestPanel_SetTitle(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)
	panel.SetTitle("Panel Title")

	if panel.title != "Panel Title" {
		t.Errorf("expected title 'Panel Title', got '%s'", panel.title)
	}
}

func TestPanel_WithTitle(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)
	result := panel.WithTitle("Title")

	if result != panel {
		t.Error("WithTitle should return same panel for chaining")
	}
	if panel.title != "Title" {
		t.Errorf("expected title 'Title', got '%s'", panel.title)
	}
}

func TestPanel_HandleMessage(t *testing.T) {
	label := NewLabel("test")
	panel := NewPanel(label)

	result := panel.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	// Panel forwards to child, but Label doesn't handle messages
	if result.Handled {
		t.Error("expected unhandled")
	}
}

// Test Box additional methods
func TestBox_SetStyle(t *testing.T) {
	label := NewLabel("test")
	box := NewBox(label)
	style := backend.DefaultStyle()
	box.SetStyle(style)

	// No panic means success
}

func TestBox_WithStyle(t *testing.T) {
	label := NewLabel("test")
	box := NewBox(label)
	style := backend.DefaultStyle()
	result := box.WithStyle(style)

	if result != box {
		t.Error("WithStyle should return same box for chaining")
	}
}

func TestBox_Layout(t *testing.T) {
	label := NewLabel("test")
	box := NewBox(label)
	box.Layout(runtime.Rect{X: 5, Y: 5, Width: 20, Height: 10})

	if box.bounds.X != 5 || box.bounds.Y != 5 {
		t.Errorf("expected bounds at (5, 5), got (%d, %d)", box.bounds.X, box.bounds.Y)
	}
}

func TestBox_Render(t *testing.T) {
	label := NewLabel("Hi")
	box := NewBox(label)
	box.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	buf := runtime.NewBuffer(10, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	box.Render(ctx)

	// Check label is rendered
	cell := buf.Get(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("expected 'H' at (0,0), got '%c'", cell.Rune)
	}
}

func TestBox_HandleMessage(t *testing.T) {
	label := NewLabel("test")
	box := NewBox(label)

	result := box.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	// Box forwards to child, but Label doesn't handle
	if result.Handled {
		t.Error("expected unhandled")
	}
}

// Test Base HandleMessage
func TestBase_HandleMessage(t *testing.T) {
	var b Base

	result := b.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if result.Handled {
		t.Error("Base should not handle messages")
	}
}
