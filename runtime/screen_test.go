package runtime

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/terminal"
)

// mockWidget is a test widget.
type mockWidget struct {
	bounds      Rect
	focused     bool
	handleCalls int
	lastKey     terminal.Key
	commands    []Command
}

func (m *mockWidget) Measure(c Constraints) Size {
	return Size{Width: 10, Height: 5}
}

func (m *mockWidget) Layout(bounds Rect) {
	m.bounds = bounds
}

func (m *mockWidget) Render(ctx RenderContext) {
	// Fill with 'X' to verify rendering
	ctx.Buffer.Fill(m.bounds, 'X', backend.DefaultStyle())
}

func (m *mockWidget) HandleMessage(msg Message) HandleResult {
	m.handleCalls++
	if key, ok := msg.(KeyMsg); ok {
		m.lastKey = key.Key

		// Return commands if set
		if len(m.commands) > 0 {
			return HandleResult{Handled: true, Commands: m.commands}
		}
	}
	return Handled()
}

func (m *mockWidget) CanFocus() bool { return true }
func (m *mockWidget) Focus()         { m.focused = true }
func (m *mockWidget) Blur()          { m.focused = false }
func (m *mockWidget) IsFocused() bool { return m.focused }

func TestScreen_PushPopLayer(t *testing.T) {
	s := NewScreen(80, 24)

	// Set base layer
	root := &mockWidget{}
	s.SetRoot(root)

	if s.LayerCount() != 1 {
		t.Errorf("expected 1 layer, got %d", s.LayerCount())
	}

	// Push overlay
	overlay := &mockWidget{}
	s.PushLayer(overlay, true)

	if s.LayerCount() != 2 {
		t.Errorf("expected 2 layers, got %d", s.LayerCount())
	}

	// Pop overlay
	if !s.PopLayer() {
		t.Error("PopLayer should return true")
	}

	if s.LayerCount() != 1 {
		t.Errorf("expected 1 layer after pop, got %d", s.LayerCount())
	}

	// Can't pop base layer
	if s.PopLayer() {
		t.Error("PopLayer should return false for base layer")
	}
}

func TestScreen_ModalLayerBlocksInput(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	overlay := &mockWidget{}
	s.PushLayer(overlay, true) // Modal = true

	// Send key to screen
	msg := KeyMsg{Key: terminal.KeyEnter}
	s.HandleMessage(msg)

	// Overlay should receive input
	if overlay.handleCalls != 1 {
		t.Errorf("overlay should receive input, got %d calls", overlay.handleCalls)
	}

	// Root should NOT receive input (modal blocks)
	if root.handleCalls != 0 {
		t.Errorf("root should not receive input through modal, got %d calls", root.handleCalls)
	}
}

func TestScreen_LayerStructure(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	overlay := &mockWidget{}
	s.PushLayer(overlay, false) // Modal = false

	// Verify layer structure
	if s.LayerCount() != 2 {
		t.Errorf("expected 2 layers, got %d", s.LayerCount())
	}

	// Top layer should be the overlay
	top := s.TopLayer()
	if top == nil {
		t.Fatal("TopLayer should not be nil")
	}
	if top.Root != overlay {
		t.Error("TopLayer root should be the overlay")
	}
}

func TestScreen_PopOverlayCommand(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	// Push overlay that will emit PopOverlay
	overlay := &mockWidget{}
	overlay.commands = []Command{PopOverlay{}}
	s.PushLayer(overlay, true)

	if s.LayerCount() != 2 {
		t.Errorf("expected 2 layers before, got %d", s.LayerCount())
	}

	// Send any key - overlay will respond with PopOverlay command
	msg := KeyMsg{Key: terminal.KeyEscape}
	s.HandleMessage(msg)

	// Screen should process PopOverlay and remove the layer
	if s.LayerCount() != 1 {
		t.Errorf("expected 1 layer after PopOverlay, got %d", s.LayerCount())
	}
}

func TestScreen_PushOverlayCommand(t *testing.T) {
	s := NewScreen(80, 24)

	newOverlay := &mockWidget{}

	root := &mockWidget{}
	root.commands = []Command{PushOverlay{Widget: newOverlay, Modal: true}}
	s.SetRoot(root)

	if s.LayerCount() != 1 {
		t.Errorf("expected 1 layer before, got %d", s.LayerCount())
	}

	// Send key - root will respond with PushOverlay command
	msg := KeyMsg{Key: terminal.KeyRune, Rune: '@'}
	s.HandleMessage(msg)

	// Screen should process PushOverlay and add the layer
	if s.LayerCount() != 2 {
		t.Errorf("expected 2 layers after PushOverlay, got %d", s.LayerCount())
	}
}

func TestScreen_Render(t *testing.T) {
	s := NewScreen(20, 10)

	root := &mockWidget{}
	root.bounds = Rect{0, 0, 20, 10}
	s.SetRoot(root)

	s.Render()

	// Check that root rendered (filled with 'X')
	buf := s.Buffer()
	cell := buf.Get(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("expected 'X' from root render, got '%c'", cell.Rune)
	}
}

func TestScreen_Resize(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	s.Resize(100, 30)

	w, h := s.Size()
	if w != 100 || h != 30 {
		t.Errorf("expected size 100x30, got %dx%d", w, h)
	}

	// Root should be re-laid out
	if root.bounds.Width != 100 || root.bounds.Height != 30 {
		t.Errorf("expected root bounds 100x30, got %dx%d", root.bounds.Width, root.bounds.Height)
	}
}

func TestScreen_ResizeMultipleLayers(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	overlay := &mockWidget{}
	s.PushLayer(overlay, false)

	s.Resize(100, 30)

	// Both layers should be re-laid out
	if root.bounds.Width != 100 || root.bounds.Height != 30 {
		t.Errorf("root bounds after resize: got %dx%d, want 100x30", root.bounds.Width, root.bounds.Height)
	}
	if overlay.bounds.Width != 100 || overlay.bounds.Height != 30 {
		t.Errorf("overlay bounds after resize: got %dx%d, want 100x30", overlay.bounds.Width, overlay.bounds.Height)
	}
}

func TestScreen_ResizeWithNilRoot(t *testing.T) {
	s := NewScreen(80, 24)

	// Push layer with nil root - should not panic
	s.layers = append(s.layers, &Layer{
		Root:       nil,
		FocusScope: NewFocusScope(),
		Modal:      false,
	})

	// Should not panic
	s.Resize(100, 30)
}

func TestScreen_SetRootNil(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	// Set nil root
	s.SetRoot(nil)

	if s.Root() != nil {
		t.Error("Root() should return nil after SetRoot(nil)")
	}
}

func TestScreen_SetRootReplaces(t *testing.T) {
	s := NewScreen(80, 24)

	root1 := &mockWidget{}
	root2 := &mockWidget{}

	s.SetRoot(root1)
	s.SetRoot(root2)

	if s.Root() != root2 {
		t.Error("Root() should return the new root")
	}
}

func TestScreen_RootEmpty(t *testing.T) {
	s := NewScreen(80, 24)

	if s.Root() != nil {
		t.Error("Root() should return nil for empty screen")
	}
}

func TestScreen_TopLayerEmpty(t *testing.T) {
	s := NewScreen(80, 24)

	if s.TopLayer() != nil {
		t.Error("TopLayer() should return nil for empty screen")
	}
}

func TestScreen_FocusScopeEmpty(t *testing.T) {
	s := NewScreen(80, 24)

	if s.FocusScope() != nil {
		t.Error("FocusScope() should return nil for empty screen")
	}
}

func TestScreen_FocusScopeWithLayer(t *testing.T) {
	s := NewScreen(80, 24)
	s.SetRoot(&mockWidget{})

	fs := s.FocusScope()
	if fs == nil {
		t.Error("FocusScope() should not be nil when layer exists")
	}
}

func TestScreen_NonModalLayerPassesInput(t *testing.T) {
	s := NewScreen(80, 24)

	root := &mockWidget{}
	s.SetRoot(root)

	// Push non-modal overlay that does NOT handle the message
	overlay := &nonHandlingWidget{}
	s.PushLayer(overlay, false) // Modal = false

	// Send key
	msg := KeyMsg{Key: terminal.KeyEnter}
	s.HandleMessage(msg)

	// Overlay receives it
	if overlay.handleCalls != 1 {
		t.Errorf("overlay should receive input, got %d calls", overlay.handleCalls)
	}
	// Root should also receive it (non-modal)
	if root.handleCalls != 1 {
		t.Errorf("root should receive input through non-modal layer, got %d calls", root.handleCalls)
	}
}

// nonHandlingWidget is a widget that doesn't handle messages.
type nonHandlingWidget struct {
	bounds      Rect
	handleCalls int
}

func (m *nonHandlingWidget) Measure(c Constraints) Size { return Size{10, 5} }
func (m *nonHandlingWidget) Layout(bounds Rect)         { m.bounds = bounds }
func (m *nonHandlingWidget) Render(ctx RenderContext)   {}
func (m *nonHandlingWidget) HandleMessage(msg Message) HandleResult {
	m.handleCalls++
	return Unhandled()
}

func TestScreen_HandleMessageEmptyLayers(t *testing.T) {
	s := NewScreen(80, 24)

	// Should not panic and return Unhandled
	result := s.HandleMessage(KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Error("HandleMessage should return unhandled for empty screen")
	}
}

func TestScreen_HandleMessageNilRoot(t *testing.T) {
	s := NewScreen(80, 24)
	s.layers = append(s.layers, &Layer{
		Root:       nil,
		FocusScope: NewFocusScope(),
		Modal:      false,
	})

	// Should not panic
	result := s.HandleMessage(KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Error("HandleMessage should return unhandled when root is nil")
	}
}

func TestScreen_RenderNilRoot(t *testing.T) {
	s := NewScreen(80, 24)
	s.layers = append(s.layers, &Layer{
		Root:       nil,
		FocusScope: NewFocusScope(),
		Modal:      false,
	})

	// Should not panic
	s.Render()
}

func TestScreen_RenderMultipleLayers(t *testing.T) {
	s := NewScreen(20, 10)

	root := &mockWidget{}
	root.bounds = Rect{0, 0, 20, 10}
	s.SetRoot(root)

	overlay := &fillingWidget{char: 'O'}
	overlay.bounds = Rect{0, 0, 20, 10}
	s.PushLayer(overlay, false)

	s.Render()

	// Overlay rendered on top (should see 'O')
	buf := s.Buffer()
	cell := buf.Get(0, 0)
	if cell.Rune != 'O' {
		t.Errorf("expected 'O' from overlay render, got '%c'", cell.Rune)
	}
}

// fillingWidget fills its bounds with a specific character.
type fillingWidget struct {
	bounds Rect
	char   rune
}

func (f *fillingWidget) Measure(c Constraints) Size { return Size{10, 5} }
func (f *fillingWidget) Layout(bounds Rect)         { f.bounds = bounds }
func (f *fillingWidget) Render(ctx RenderContext) {
	ctx.Buffer.Fill(f.bounds, f.char, backend.DefaultStyle())
}
func (f *fillingWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func TestScreen_FocusNextCommand(t *testing.T) {
	s := NewScreen(80, 24)

	// Create widget that emits FocusNext command
	w := &mockWidget{}
	w.commands = []Command{FocusNext{}}
	s.SetRoot(w)

	// Register focusable widget in scope
	focusable := &mockWidget{}
	s.FocusScope().Register(focusable)

	// Should process FocusNext command
	msg := KeyMsg{Key: terminal.KeyTab}
	s.HandleMessage(msg)

	// Verify command was processed (focusable should be focused)
	if !focusable.focused {
		t.Error("FocusNext command should focus the widget")
	}
}

func TestScreen_FocusPrevCommand(t *testing.T) {
	s := NewScreen(80, 24)

	w := &mockWidget{}
	w.commands = []Command{FocusPrev{}}
	s.SetRoot(w)

	// Register focusable widgets
	focusable1 := &mockWidget{}
	focusable2 := &mockWidget{}
	s.FocusScope().Register(focusable1)
	s.FocusScope().Register(focusable2)

	// focusable1 is focused initially
	if !focusable1.focused {
		t.Error("focusable1 should be focused initially")
	}

	// FocusPrev should wrap to focusable2
	msg := KeyMsg{Key: terminal.KeyTab, Shift: true}
	s.HandleMessage(msg)

	if !focusable2.focused {
		t.Error("FocusPrev command should focus focusable2")
	}
}

func TestRenderContext_Sub(t *testing.T) {
	buf := NewBuffer(100, 50)
	ctx := RenderContext{
		Buffer:  buf,
		Focused: true,
		Bounds:  Rect{0, 0, 100, 50},
	}

	subCtx := ctx.Sub(Rect{10, 10, 20, 20})

	if subCtx.Bounds.X != 10 || subCtx.Bounds.Y != 10 {
		t.Error("Sub context should have offset bounds")
	}
	if subCtx.Bounds.Width != 20 || subCtx.Bounds.Height != 20 {
		t.Error("Sub context should have reduced size")
	}
	if subCtx.Buffer != buf {
		t.Error("Sub context should share buffer")
	}
	if !subCtx.Focused {
		t.Error("Sub context should inherit Focused")
	}
}

func TestRenderContext_SubBuffer(t *testing.T) {
	buf := NewBuffer(100, 50)
	ctx := RenderContext{
		Buffer: buf,
		Bounds: Rect{10, 10, 20, 20},
	}

	subBuf := ctx.SubBuffer()

	w, h := subBuf.Size()
	if w != 20 || h != 20 {
		t.Errorf("SubBuffer size = %dx%d, want 20x20", w, h)
	}
}
