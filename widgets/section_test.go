package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

func TestSection_New(t *testing.T) {
	s := NewSection("Test Section")

	if s.title != "Test Section" {
		t.Errorf("expected title 'Test Section', got '%s'", s.title)
	}
	if !s.expanded {
		t.Error("expected expanded=true by default")
	}
}

func TestSection_SetItems(t *testing.T) {
	s := NewSection("Tasks")

	items := []SectionItem{
		{Icon: '✓', Text: "Done task"},
		{Icon: '→', Text: "In progress", Active: true},
		{Icon: '○', Text: "Pending"},
	}
	s.SetItems(items)

	if len(s.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(s.items))
	}
}

func TestSection_Toggle(t *testing.T) {
	s := NewSection("Test")

	if !s.IsExpanded() {
		t.Error("should be expanded by default")
	}

	s.Toggle()
	if s.IsExpanded() {
		t.Error("should be collapsed after toggle")
	}

	s.Toggle()
	if !s.IsExpanded() {
		t.Error("should be expanded after second toggle")
	}
}

func TestSection_Measure_Expanded(t *testing.T) {
	s := NewSection("Test")
	s.SetItems([]SectionItem{
		{Icon: '✓', Text: "Item 1"},
		{Icon: '→', Text: "Item 2"},
		{Icon: '○', Text: "Item 3"},
	})
	s.SetExpanded(true)

	size := s.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 20})

	// 1 header + 3 items = 4
	if size.Height != 4 {
		t.Errorf("expected height 4 when expanded, got %d", size.Height)
	}
}

func TestSection_Measure_Collapsed(t *testing.T) {
	s := NewSection("Test")
	s.SetItems([]SectionItem{
		{Icon: '✓', Text: "Item 1"},
		{Icon: '→', Text: "Item 2"},
	})
	s.SetExpanded(false)

	size := s.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 20})

	// Only header when collapsed
	if size.Height != 1 {
		t.Errorf("expected height 1 when collapsed, got %d", size.Height)
	}
}

func TestSection_Measure_MaxItems(t *testing.T) {
	s := NewSection("Test")
	s.SetItems([]SectionItem{
		{Icon: '✓', Text: "Item 1"},
		{Icon: '✓', Text: "Item 2"},
		{Icon: '✓', Text: "Item 3"},
		{Icon: '✓', Text: "Item 4"},
		{Icon: '✓', Text: "Item 5"},
	})
	s.SetMaxItems(3)
	s.SetExpanded(true)

	size := s.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 20})

	// 1 header + 3 items (limited by maxItems) = 4
	if size.Height != 4 {
		t.Errorf("expected height 4 with maxItems=3, got %d", size.Height)
	}
}

func TestSection_HandleMessage_Toggle(t *testing.T) {
	s := NewSection("Test")

	// Space toggles
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if !result.Handled {
		t.Error("space should be handled")
	}
	if s.expanded {
		t.Error("should be collapsed after space")
	}

	// Enter also toggles
	result = s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("enter should be handled")
	}
	if !s.expanded {
		t.Error("should be expanded after enter")
	}
}

func TestSection_Render(t *testing.T) {
	s := NewSection("Tasks")
	s.SetItems([]SectionItem{
		{Icon: '✓', Text: "Done"},
		{Icon: '→', Text: "Active", Active: true},
	})
	s.SetExpanded(true)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 30, Height: 10})

	buf := runtime.NewBuffer(30, 10)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	s.Render(ctx)

	// Check expand icon
	cell := buf.Get(0, 0)
	if cell.Rune != '▼' {
		t.Errorf("expected expand icon '▼', got '%c'", cell.Rune)
	}
}

func TestSection_ContentHeight(t *testing.T) {
	s := NewSection("Test")
	s.SetItems([]SectionItem{
		{Icon: '✓', Text: "1"},
		{Icon: '✓', Text: "2"},
		{Icon: '✓', Text: "3"},
	})

	// Expanded
	s.SetExpanded(true)
	if s.ContentHeight() != 3 {
		t.Errorf("expected content height 3 when expanded, got %d", s.ContentHeight())
	}

	// Collapsed
	s.SetExpanded(false)
	if s.ContentHeight() != 0 {
		t.Errorf("expected content height 0 when collapsed, got %d", s.ContentHeight())
	}
}
