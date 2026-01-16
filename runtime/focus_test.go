package runtime

import (
	"testing"
)

// focusableWidget is a test widget that can receive focus.
type focusableWidget struct {
	canFocus bool
	focused  bool
	id       string
}

func newFocusable(id string) *focusableWidget {
	return &focusableWidget{canFocus: true, id: id}
}

func newNonFocusable(id string) *focusableWidget {
	return &focusableWidget{canFocus: false, id: id}
}

func (f *focusableWidget) Measure(c Constraints) Size   { return Size{10, 5} }
func (f *focusableWidget) Layout(bounds Rect)           {}
func (f *focusableWidget) Render(ctx RenderContext)     {}
func (f *focusableWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}
func (f *focusableWidget) CanFocus() bool   { return f.canFocus }
func (f *focusableWidget) Focus()           { f.focused = true }
func (f *focusableWidget) Blur()            { f.focused = false }
func (f *focusableWidget) IsFocused() bool  { return f.focused }

func TestFocusScope_New(t *testing.T) {
	fs := NewFocusScope()

	if fs.Count() != 0 {
		t.Errorf("Count() = %d, want 0", fs.Count())
	}
	if fs.Current() != nil {
		t.Error("Current() should be nil for empty scope")
	}
}

func TestFocusScope_Register(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)

	if fs.Count() != 1 {
		t.Errorf("Count() = %d, want 1", fs.Count())
	}
	// First focusable widget gets auto-focused
	if fs.Current() != w1 {
		t.Error("Current() should be w1")
	}
	if !w1.focused {
		t.Error("w1 should be focused")
	}

	fs.Register(w2)

	if fs.Count() != 2 {
		t.Errorf("Count() = %d, want 2", fs.Count())
	}
	// First widget should still be focused
	if fs.Current() != w1 {
		t.Error("Current() should still be w1")
	}
}

func TestFocusScope_RegisterDuplicate(t *testing.T) {
	fs := NewFocusScope()
	w := newFocusable("w")

	fs.Register(w)
	fs.Register(w) // Duplicate

	if fs.Count() != 1 {
		t.Errorf("Duplicate register should not add, Count() = %d, want 1", fs.Count())
	}
}

func TestFocusScope_RegisterNonFocusable(t *testing.T) {
	fs := NewFocusScope()
	w1 := newNonFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	// Non-focusable should not auto-focus
	if fs.Current() != nil {
		t.Error("Non-focusable widget should not get focus")
	}
	if w1.focused {
		t.Error("w1 should not be focused")
	}

	fs.Register(w2)
	// Second (focusable) widget should get focus
	if fs.Current() != w2 {
		t.Error("w2 should get focus as first focusable")
	}
}

func TestFocusScope_Unregister(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Unregister non-focused widget
	fs.Unregister(w2)
	if fs.Count() != 2 {
		t.Errorf("Count() = %d, want 2", fs.Count())
	}
	// w1 should still be focused
	if fs.Current() != w1 {
		t.Error("w1 should still be focused")
	}
}

func TestFocusScope_UnregisterFocused(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// Unregister the focused widget
	fs.Unregister(w1)

	if fs.Count() != 1 {
		t.Errorf("Count() = %d, want 1", fs.Count())
	}
	// Focus should move to w2
	if fs.Current() != w2 {
		t.Error("Focus should move to w2")
	}
	if !w2.focused {
		t.Error("w2 should be focused")
	}
	if w1.focused {
		t.Error("w1 should be blurred")
	}
}

func TestFocusScope_UnregisterBeforeFocused(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Focus w3
	fs.SetFocus(w3)

	// Unregister w1 (before the focused index)
	fs.Unregister(w1)

	// w3 should still be focused
	if fs.Current() != w3 {
		t.Error("w3 should still be focused")
	}
}

func TestFocusScope_SetFocus(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// Focus w2
	changed := fs.SetFocus(w2)
	if !changed {
		t.Error("SetFocus should return true when focus changes")
	}
	if fs.Current() != w2 {
		t.Error("w2 should be focused")
	}
	if w1.focused {
		t.Error("w1 should be blurred")
	}
	if !w2.focused {
		t.Error("w2 should be focused")
	}
}

func TestFocusScope_SetFocusSameWidget(t *testing.T) {
	fs := NewFocusScope()
	w := newFocusable("w")

	fs.Register(w)

	// Focus same widget
	changed := fs.SetFocus(w)
	if changed {
		t.Error("SetFocus should return false when focusing already-focused widget")
	}
}

func TestFocusScope_SetFocusNonFocusable(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newNonFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// Try to focus non-focusable widget
	changed := fs.SetFocus(w2)
	if changed {
		t.Error("SetFocus should return false for non-focusable widget")
	}
	if fs.Current() != w1 {
		t.Error("w1 should still be focused")
	}
}

func TestFocusScope_SetFocusUnregistered(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)

	// Try to focus unregistered widget
	changed := fs.SetFocus(w2)
	if changed {
		t.Error("SetFocus should return false for unregistered widget")
	}
}

func TestFocusScope_FocusFirst(t *testing.T) {
	fs := NewFocusScope()
	w1 := newNonFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Focus w3 first
	fs.SetFocus(w3)

	// FocusFirst should skip non-focusable and focus w2
	changed := fs.FocusFirst()
	if !changed {
		t.Error("FocusFirst should return true")
	}
	if fs.Current() != w2 {
		t.Error("FocusFirst should focus w2 (first focusable)")
	}
}

func TestFocusScope_FocusFirstEmpty(t *testing.T) {
	fs := NewFocusScope()

	changed := fs.FocusFirst()
	if changed {
		t.Error("FocusFirst should return false for empty scope")
	}
}

func TestFocusScope_FocusFirstAllNonFocusable(t *testing.T) {
	fs := NewFocusScope()
	w1 := newNonFocusable("w1")
	w2 := newNonFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	changed := fs.FocusFirst()
	if changed {
		t.Error("FocusFirst should return false when all widgets are non-focusable")
	}
}

func TestFocusScope_FocusLast(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newNonFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// FocusLast should skip non-focusable and focus w2
	changed := fs.FocusLast()
	if !changed {
		t.Error("FocusLast should return true")
	}
	if fs.Current() != w2 {
		t.Error("FocusLast should focus w2 (last focusable)")
	}
}

func TestFocusScope_FocusLastEmpty(t *testing.T) {
	fs := NewFocusScope()

	changed := fs.FocusLast()
	if changed {
		t.Error("FocusLast should return false for empty scope")
	}
}

func TestFocusScope_FocusNext(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Initially w1 is focused
	if fs.Current() != w1 {
		t.Fatal("w1 should be initially focused")
	}

	// FocusNext -> w2
	changed := fs.FocusNext()
	if !changed {
		t.Error("FocusNext should return true")
	}
	if fs.Current() != w2 {
		t.Error("FocusNext should focus w2")
	}

	// FocusNext -> w3
	fs.FocusNext()
	if fs.Current() != w3 {
		t.Error("FocusNext should focus w3")
	}

	// FocusNext -> wrap to w1
	fs.FocusNext()
	if fs.Current() != w1 {
		t.Error("FocusNext should wrap to w1")
	}
}

func TestFocusScope_FocusNextSkipsNonFocusable(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newNonFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// FocusNext should skip w2 and go to w3
	fs.FocusNext()
	if fs.Current() != w3 {
		t.Error("FocusNext should skip non-focusable and focus w3")
	}
}

func TestFocusScope_FocusNextEmpty(t *testing.T) {
	fs := NewFocusScope()

	changed := fs.FocusNext()
	if changed {
		t.Error("FocusNext should return false for empty scope")
	}
}

func TestFocusScope_FocusNextNoFocus(t *testing.T) {
	fs := NewFocusScope()
	w1 := newNonFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// Clear focus
	fs.ClearFocus()

	// FocusNext should find first focusable (w2)
	changed := fs.FocusNext()
	if !changed {
		t.Error("FocusNext should return true")
	}
	if fs.Current() != w2 {
		t.Error("FocusNext should focus w2")
	}
}

func TestFocusScope_FocusPrev(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Start at w1 (first registered)
	if fs.Current() != w1 {
		t.Fatal("w1 should be initially focused")
	}

	// FocusPrev -> wrap to w3
	changed := fs.FocusPrev()
	if !changed {
		t.Error("FocusPrev should return true")
	}
	if fs.Current() != w3 {
		t.Error("FocusPrev should wrap to w3")
	}

	// FocusPrev -> w2
	fs.FocusPrev()
	if fs.Current() != w2 {
		t.Error("FocusPrev should focus w2")
	}

	// FocusPrev -> w1
	fs.FocusPrev()
	if fs.Current() != w1 {
		t.Error("FocusPrev should focus w1")
	}
}

func TestFocusScope_FocusPrevSkipsNonFocusable(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newNonFocusable("w2")
	w3 := newFocusable("w3")

	fs.Register(w1)
	fs.Register(w2)
	fs.Register(w3)

	// Focus w3
	fs.SetFocus(w3)

	// FocusPrev should skip w2 and go to w1
	fs.FocusPrev()
	if fs.Current() != w1 {
		t.Error("FocusPrev should skip non-focusable and focus w1")
	}
}

func TestFocusScope_FocusPrevEmpty(t *testing.T) {
	fs := NewFocusScope()

	changed := fs.FocusPrev()
	if changed {
		t.Error("FocusPrev should return false for empty scope")
	}
}

func TestFocusScope_FocusPrevNoFocus(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// Clear focus
	fs.ClearFocus()

	// FocusPrev should find last focusable (w2)
	changed := fs.FocusPrev()
	if !changed {
		t.Error("FocusPrev should return true")
	}
	if fs.Current() != w2 {
		t.Error("FocusPrev should focus w2")
	}
}

func TestFocusScope_ClearFocus(t *testing.T) {
	fs := NewFocusScope()
	w := newFocusable("w")

	fs.Register(w)
	if !w.focused {
		t.Fatal("w should be focused after register")
	}

	fs.ClearFocus()

	if fs.Current() != nil {
		t.Error("Current() should be nil after ClearFocus")
	}
	if w.focused {
		t.Error("w should be blurred after ClearFocus")
	}
}

func TestFocusScope_ClearFocusEmpty(t *testing.T) {
	fs := NewFocusScope()

	// Should not panic
	fs.ClearFocus()

	if fs.Current() != nil {
		t.Error("Current() should be nil")
	}
}

func TestFocusScope_SingleNonFocusableAllOperations(t *testing.T) {
	fs := NewFocusScope()
	w := newNonFocusable("w")

	fs.Register(w)

	// All focus operations should return false
	if fs.FocusFirst() {
		t.Error("FocusFirst should return false")
	}
	if fs.FocusLast() {
		t.Error("FocusLast should return false")
	}
	if fs.FocusNext() {
		t.Error("FocusNext should return false")
	}
	if fs.FocusPrev() {
		t.Error("FocusPrev should return false")
	}
}

func TestFocusScope_ToggleFocusability(t *testing.T) {
	fs := NewFocusScope()
	w1 := newFocusable("w1")
	w2 := newFocusable("w2")

	fs.Register(w1)
	fs.Register(w2)

	// w1 is focused
	if fs.Current() != w1 {
		t.Fatal("w1 should be focused")
	}

	// Make w1 non-focusable
	w1.canFocus = false

	// FocusNext should still move to w2
	fs.FocusNext()
	if fs.Current() != w2 {
		t.Error("FocusNext should focus w2")
	}

	// FocusPrev should skip w1 and stay at w2
	fs.FocusPrev()
	if fs.Current() != w2 {
		t.Error("FocusPrev should stay at w2 (w1 is non-focusable)")
	}
}
