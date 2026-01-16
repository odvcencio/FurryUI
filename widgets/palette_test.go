package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

func TestPaletteWidget_New(t *testing.T) {
	p := NewPaletteWidget("Test Palette")

	if p.title != "Test Palette" {
		t.Errorf("expected title 'Test Palette', got '%s'", p.title)
	}
	if p.maxVisible != 10 {
		t.Errorf("expected maxVisible 10, got %d", p.maxVisible)
	}
}

func TestPaletteWidget_SetItems(t *testing.T) {
	p := NewPaletteWidget("Test")

	items := []PaletteItem{
		{ID: "1", Label: "Item One"},
		{ID: "2", Label: "Item Two"},
		{ID: "3", Label: "Item Three"},
	}
	p.SetItems(items)

	if len(p.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(p.items))
	}
	if len(p.filtered) != 3 {
		t.Errorf("expected 3 filtered items, got %d", len(p.filtered))
	}
}

func TestPaletteWidget_Filter(t *testing.T) {
	p := NewPaletteWidget("Test")

	items := []PaletteItem{
		{ID: "1", Label: "Apple"},
		{ID: "2", Label: "Banana"},
		{ID: "3", Label: "Cherry"},
	}
	p.SetItems(items)

	// Type 'a' to filter
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})

	if p.query != "a" {
		t.Errorf("expected query 'a', got '%s'", p.query)
	}
	if len(p.filtered) != 2 { // Apple and Banana contain 'a'
		t.Errorf("expected 2 filtered items (Apple, Banana), got %d", len(p.filtered))
	}
}

func TestPaletteWidget_CaseInsensitiveFilter(t *testing.T) {
	p := NewPaletteWidget("Test")

	items := []PaletteItem{
		{ID: "1", Label: "UPPER"},
		{ID: "2", Label: "lower"},
		{ID: "3", Label: "Mixed"},
	}
	p.SetItems(items)

	// Filter with "ppe" - should match "UPPER" case-insensitively
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'p'})
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'p'})
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'e'})

	if len(p.filtered) != 1 { // Only UPPER contains "ppe"
		t.Errorf("expected 1 filtered item, got %d", len(p.filtered))
	}
	if p.filtered[0].Label != "UPPER" {
		t.Errorf("expected 'UPPER', got '%s'", p.filtered[0].Label)
	}
}

func TestPaletteWidget_Navigation(t *testing.T) {
	p := NewPaletteWidget("Test")

	items := []PaletteItem{
		{ID: "1", Label: "First"},
		{ID: "2", Label: "Second"},
		{ID: "3", Label: "Third"},
	}
	p.SetItems(items)

	// Initial selection is 0
	if p.selected != 0 {
		t.Errorf("expected selected 0, got %d", p.selected)
	}

	// Down
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if p.selected != 1 {
		t.Errorf("expected selected 1 after down, got %d", p.selected)
	}

	// Down again
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if p.selected != 2 {
		t.Errorf("expected selected 2 after down, got %d", p.selected)
	}

	// Down at bottom shouldn't exceed
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if p.selected != 2 {
		t.Errorf("expected selected 2 (clamped), got %d", p.selected)
	}

	// Up
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if p.selected != 1 {
		t.Errorf("expected selected 1 after up, got %d", p.selected)
	}
}

func TestPaletteWidget_HandleEnter(t *testing.T) {
	p := NewPaletteWidget("Test")

	var selectedItem *PaletteItem
	p.SetOnSelect(func(item PaletteItem) {
		selectedItem = &item
	})

	items := []PaletteItem{
		{ID: "action-1", Label: "Do Something", Data: "custom-data"},
	}
	p.SetItems(items)

	result := p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if !result.Handled {
		t.Error("Enter should be handled")
	}
	if selectedItem == nil {
		t.Error("onSelect callback should be called")
	}
	if selectedItem.ID != "action-1" {
		t.Errorf("expected ID 'action-1', got '%s'", selectedItem.ID)
	}

	// Should emit PaletteSelected and PopOverlay commands
	if len(result.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(result.Commands))
	}
}

func TestPaletteWidget_HandleEscape(t *testing.T) {
	p := NewPaletteWidget("Test")

	result := p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if !result.Handled {
		t.Error("Escape should be handled")
	}

	// Should emit PopOverlay
	if len(result.Commands) == 0 {
		t.Error("expected PopOverlay command")
	}

	_, ok := result.Commands[0].(runtime.PopOverlay)
	if !ok {
		t.Errorf("expected PopOverlay, got %T", result.Commands[0])
	}
}

func TestPaletteWidget_HandleBackspace(t *testing.T) {
	p := NewPaletteWidget("Test")

	// Type some characters
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'b'})

	if p.query != "ab" {
		t.Errorf("expected query 'ab', got '%s'", p.query)
	}

	// Backspace
	result := p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if !result.Handled {
		t.Error("Backspace should be handled")
	}
	if p.query != "a" {
		t.Errorf("expected query 'a', got '%s'", p.query)
	}

	// Another backspace
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if p.query != "" {
		t.Errorf("expected empty query, got '%s'", p.query)
	}

	// Backspace on empty should pop overlay
	result = p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if len(result.Commands) == 0 {
		t.Error("expected PopOverlay on empty backspace")
	}
}

func TestPaletteWidget_Measure(t *testing.T) {
	p := NewPaletteWidget("Test")
	p.SetItems([]PaletteItem{
		{ID: "1", Label: "Item 1"},
		{ID: "2", Label: "Item 2"},
	})

	size := p.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 60 {
		t.Errorf("expected width 60, got %d", size.Width)
	}
	// title(1) + query(1) + separator(1) + items(2) + border(2) = 7
	if size.Height != 7 {
		t.Errorf("expected height 7, got %d", size.Height)
	}
}

func TestPaletteWidget_Layout_Centers(t *testing.T) {
	p := NewPaletteWidget("Test")

	p.Layout(runtime.Rect{X: 0, Y: 0, Width: 100, Height: 40})

	bounds := p.Bounds()

	// Should be centered
	if bounds.X < 10 {
		t.Errorf("expected X > 10 (centered), got %d", bounds.X)
	}
	if bounds.Y < 10 {
		t.Errorf("expected Y > 10 (centered), got %d", bounds.Y)
	}
}

func TestPaletteWidget_Render(t *testing.T) {
	p := NewPaletteWidget("Commands")
	p.SetItems([]PaletteItem{
		{ID: "1", Label: "New file", Shortcut: "Ctrl+N"},
		{ID: "2", Label: "Save", Shortcut: "Ctrl+S"},
	})
	p.Focus()
	p.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})

	buf := runtime.NewBuffer(80, 24)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	p.Render(ctx)

	// Check for border
	bounds := p.Bounds()
	cell := buf.Get(bounds.X, bounds.Y)
	if cell.Rune != '╭' {
		t.Errorf("expected top-left corner, got '%c'", cell.Rune)
	}
}

func TestPaletteWidget_SelectedItem(t *testing.T) {
	p := NewPaletteWidget("Test")

	// No items
	if p.SelectedItem() != nil {
		t.Error("expected nil when no items")
	}

	// Add items
	p.SetItems([]PaletteItem{
		{ID: "1", Label: "First"},
		{ID: "2", Label: "Second"},
	})

	item := p.SelectedItem()
	if item == nil {
		t.Error("expected non-nil selected item")
	}
	if item.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", item.ID)
	}

	// Select second
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	item = p.SelectedItem()
	if item.ID != "2" {
		t.Errorf("expected ID '2', got '%s'", item.ID)
	}
}

func TestPaletteWidget_CustomFilter(t *testing.T) {
	p := NewPaletteWidget("Test")

	// Custom filter that only matches exact prefix
	p.SetFilterFn(func(item PaletteItem, query string) bool {
		return len(item.Label) >= len(query) && item.Label[:len(query)] == query
	})

	p.SetItems([]PaletteItem{
		{ID: "1", Label: "Apple"},
		{ID: "2", Label: "Apricot"},
		{ID: "3", Label: "Banana"},
	})

	// Filter with "Ap"
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'A'})
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'p'})

	// Should match Apple and Apricot only
	if len(p.filtered) != 2 {
		t.Errorf("expected 2 filtered items, got %d", len(p.filtered))
	}
}

func TestPaletteWidget_Categories(t *testing.T) {
	p := NewPaletteWidget("Test")

	items := []PaletteItem{
		{ID: "1", Category: "Recent", Label: "Open project"},
		{ID: "2", Category: "Recent", Label: "Save all"},
		{ID: "3", Category: "Actions", Label: "New file"},
	}
	p.SetItems(items)

	// Items should preserve categories
	if len(p.filtered) != 3 {
		t.Errorf("expected 3 items, got %d", len(p.filtered))
	}
	if p.filtered[0].Category != "Recent" {
		t.Errorf("expected category 'Recent', got '%s'", p.filtered[0].Category)
	}
}
