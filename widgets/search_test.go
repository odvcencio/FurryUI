package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

func TestSearchWidget_New(t *testing.T) {
	s := NewSearchWidget()
	if s == nil {
		t.Fatal("expected non-nil search widget")
	}
	if s.query != "" {
		t.Errorf("expected empty query, got '%s'", s.query)
	}
}

func TestSearchWidget_Typing(t *testing.T) {
	s := NewSearchWidget()

	var lastQuery string
	s.SetOnSearch(func(query string) {
		lastQuery = query
	})

	// Type "hello"
	for _, r := range "hello" {
		s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}

	if s.Query() != "hello" {
		t.Errorf("expected query 'hello', got '%s'", s.Query())
	}
	if lastQuery != "hello" {
		t.Errorf("expected callback with 'hello', got '%s'", lastQuery)
	}
}

func TestSearchWidget_Backspace(t *testing.T) {
	s := NewSearchWidget()

	// Type "test"
	for _, r := range "test" {
		s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}

	// Backspace twice
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	if s.Query() != "te" {
		t.Errorf("expected query 'te', got '%s'", s.Query())
	}
}

func TestSearchWidget_Escape(t *testing.T) {
	s := NewSearchWidget()

	// Type something
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})

	// Escape should clear and pop
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if s.Query() != "" {
		t.Errorf("expected empty query after escape, got '%s'", s.Query())
	}
	if len(result.Commands) != 1 {
		t.Errorf("expected 1 command (PopOverlay), got %d", len(result.Commands))
	}
}

func TestSearchWidget_Enter(t *testing.T) {
	s := NewSearchWidget()

	// Type something
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})

	// Enter should keep query but pop overlay
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if s.Query() != "a" {
		t.Errorf("expected query 'a' after enter, got '%s'", s.Query())
	}
	if len(result.Commands) != 1 {
		t.Errorf("expected 1 command (PopOverlay), got %d", len(result.Commands))
	}
}

func TestSearchWidget_MatchInfo(t *testing.T) {
	s := NewSearchWidget()

	s.SetMatchInfo(2, 5)

	if s.currentMatch != 2 {
		t.Errorf("expected currentMatch 2, got %d", s.currentMatch)
	}
	if s.matchCount != 5 {
		t.Errorf("expected matchCount 5, got %d", s.matchCount)
	}
}

func TestSearchWidget_Navigate(t *testing.T) {
	s := NewSearchWidget()

	var nextCount, prevCount int
	s.SetOnNavigate(func() { nextCount++ }, func() { prevCount++ })

	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})

	if nextCount != 1 {
		t.Errorf("expected nextCount 1, got %d", nextCount)
	}
	if prevCount != 1 {
		t.Errorf("expected prevCount 1, got %d", prevCount)
	}
}

func TestSearchWidget_Render(t *testing.T) {
	s := NewSearchWidget()
	s.Focus()

	// Type a query
	for _, r := range "test" {
		s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}

	// Layout and render
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
	buf := runtime.NewBuffer(80, 24)
	s.Render(runtime.RenderContext{Buffer: buf})

	// Check that "/ test" appears at the bottom
	cell := buf.Get(0, 23)
	if cell.Rune != '/' {
		t.Errorf("expected '/' at start, got '%c'", cell.Rune)
	}
}
