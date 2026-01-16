package compositor

import (
	"strings"
	"testing"
)

func TestColor(t *testing.T) {
	t.Run("predefined colors", func(t *testing.T) {
		tests := []struct {
			name  string
			color Color
			mode  ColorMode
			value uint32
		}{
			{"none", ColorNone, ColorModeNone, 0},
			{"default", ColorDefault, ColorModeDefault, 0},
			{"black", ColorBlack, ColorMode16, 0},
			{"red", ColorRed, ColorMode16, 1},
			{"bright white", ColorBrightWhite, ColorMode16, 15},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.color.Mode != tt.mode {
					t.Errorf("got mode %v, want %v", tt.color.Mode, tt.mode)
				}
				if tt.color.Value != tt.value {
					t.Errorf("got value %d, want %d", tt.color.Value, tt.value)
				}
			})
		}
	})

	t.Run("Color256", func(t *testing.T) {
		c := Color256(128)
		if c.Mode != ColorMode256 {
			t.Errorf("got mode %v, want ColorMode256", c.Mode)
		}
		if c.Value != 128 {
			t.Errorf("got value %d, want 128", c.Value)
		}
	})

	t.Run("RGB", func(t *testing.T) {
		c := RGB(255, 128, 64)
		if c.Mode != ColorModeRGB {
			t.Errorf("got mode %v, want ColorModeRGB", c.Mode)
		}
		expected := uint32(0xFF8040)
		if c.Value != expected {
			t.Errorf("got value 0x%06X, want 0x%06X", c.Value, expected)
		}
	})

	t.Run("Hex", func(t *testing.T) {
		c := Hex(0xABCDEF)
		if c.Mode != ColorModeRGB {
			t.Errorf("got mode %v, want ColorModeRGB", c.Mode)
		}
		if c.Value != 0xABCDEF {
			t.Errorf("got value 0x%06X, want 0xABCDEF", c.Value)
		}
	})
}

func TestStyle(t *testing.T) {
	t.Run("DefaultStyle", func(t *testing.T) {
		s := DefaultStyle()
		if s.FG != ColorDefault {
			t.Error("default FG should be ColorDefault")
		}
		if s.BG != ColorDefault {
			t.Error("default BG should be ColorDefault")
		}
		if s.Bold || s.Dim || s.Italic || s.Underline {
			t.Error("default style should have no attributes")
		}
	})

	t.Run("WithFG", func(t *testing.T) {
		s := DefaultStyle().WithFG(ColorRed)
		if s.FG != ColorRed {
			t.Errorf("got FG %v, want ColorRed", s.FG)
		}
	})

	t.Run("WithBG", func(t *testing.T) {
		s := DefaultStyle().WithBG(ColorBlue)
		if s.BG != ColorBlue {
			t.Errorf("got BG %v, want ColorBlue", s.BG)
		}
	})

	t.Run("chained modifiers", func(t *testing.T) {
		s := DefaultStyle().
			WithFG(ColorGreen).
			WithBG(ColorBlack).
			WithBold(true).
			WithItalic(true)

		if s.FG != ColorGreen {
			t.Error("FG not set correctly")
		}
		if s.BG != ColorBlack {
			t.Error("BG not set correctly")
		}
		if !s.Bold {
			t.Error("Bold not set")
		}
		if !s.Italic {
			t.Error("Italic not set")
		}
	})

	t.Run("Equal", func(t *testing.T) {
		s1 := DefaultStyle().WithBold(true)
		s2 := DefaultStyle().WithBold(true)
		s3 := DefaultStyle().WithItalic(true)

		if !s1.Equal(s2) {
			t.Error("identical styles should be equal")
		}
		if s1.Equal(s3) {
			t.Error("different styles should not be equal")
		}
	})
}

func TestCell(t *testing.T) {
	t.Run("EmptyCell", func(t *testing.T) {
		c := EmptyCell()
		if c.Rune != ' ' {
			t.Errorf("got rune %q, want ' '", c.Rune)
		}
		if c.Width != 1 {
			t.Errorf("got width %d, want 1", c.Width)
		}
		if !c.Empty() {
			t.Error("empty cell should return Empty() true")
		}
	})

	t.Run("NewCell", func(t *testing.T) {
		style := DefaultStyle().WithFG(ColorRed)
		c := NewCell('X', style)
		if c.Rune != 'X' {
			t.Errorf("got rune %q, want 'X'", c.Rune)
		}
		if !c.Style.Equal(style) {
			t.Error("style not preserved")
		}
	})

	t.Run("non-empty cell", func(t *testing.T) {
		c := NewCell('A', DefaultStyle())
		if c.Empty() {
			t.Error("cell with 'A' should not be empty")
		}
	})

	t.Run("Equal", func(t *testing.T) {
		c1 := NewCell('X', DefaultStyle().WithBold(true))
		c2 := NewCell('X', DefaultStyle().WithBold(true))
		c3 := NewCell('Y', DefaultStyle().WithBold(true))
		c4 := NewCell('X', DefaultStyle())

		if !c1.Equal(c2) {
			t.Error("identical cells should be equal")
		}
		if c1.Equal(c3) {
			t.Error("cells with different runes should not be equal")
		}
		if c1.Equal(c4) {
			t.Error("cells with different styles should not be equal")
		}
	})
}

func TestScreen(t *testing.T) {
	t.Run("NewScreen", func(t *testing.T) {
		s := NewScreen(80, 24)
		w, h := s.Size()
		if w != 80 || h != 24 {
			t.Errorf("got size %dx%d, want 80x24", w, h)
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		s := NewScreen(10, 10)
		style := DefaultStyle().WithFG(ColorRed)
		s.Set(5, 5, 'X', style)

		cell := s.Get(5, 5)
		if cell.Rune != 'X' {
			t.Errorf("got rune %q, want 'X'", cell.Rune)
		}
		if !cell.Style.Equal(style) {
			t.Error("style not preserved")
		}
	})

	t.Run("Set out of bounds", func(t *testing.T) {
		s := NewScreen(10, 10)
		// Should not panic
		s.Set(-1, 0, 'X', DefaultStyle())
		s.Set(0, -1, 'X', DefaultStyle())
		s.Set(100, 0, 'X', DefaultStyle())
		s.Set(0, 100, 'X', DefaultStyle())
	})

	t.Run("Get out of bounds", func(t *testing.T) {
		s := NewScreen(10, 10)
		cell := s.Get(-1, 0)
		if !cell.Empty() {
			t.Error("out of bounds Get should return empty cell")
		}
	})

	t.Run("SetString", func(t *testing.T) {
		s := NewScreen(20, 5)
		style := DefaultStyle()
		n := s.SetString(0, 0, "Hello", style)

		if n != 5 {
			t.Errorf("SetString returned %d, want 5", n)
		}

		expected := "Hello"
		for i, r := range expected {
			cell := s.Get(i, 0)
			if cell.Rune != r {
				t.Errorf("at %d: got %q, want %q", i, cell.Rune, r)
			}
		}
	})

	t.Run("SetString truncation", func(t *testing.T) {
		s := NewScreen(5, 1)
		n := s.SetString(0, 0, "Hello World", DefaultStyle())
		if n != 5 {
			t.Errorf("SetString should stop at screen width, got %d", n)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		s := NewScreen(10, 10)
		s.Set(5, 5, 'X', DefaultStyle())
		s.Clear()

		cell := s.Get(5, 5)
		if !cell.Empty() {
			t.Error("cell should be empty after Clear")
		}
	})

	t.Run("Resize", func(t *testing.T) {
		s := NewScreen(10, 10)
		s.Set(5, 5, 'X', DefaultStyle())
		s.Resize(20, 20)

		w, h := s.Size()
		if w != 20 || h != 20 {
			t.Errorf("got size %dx%d, want 20x20", w, h)
		}

		// Content should be preserved
		cell := s.Get(5, 5)
		if cell.Rune != 'X' {
			t.Error("content should be preserved after resize")
		}
	})

	t.Run("FillRect", func(t *testing.T) {
		s := NewScreen(10, 10)
		style := DefaultStyle().WithBG(ColorBlue)
		s.FillRect(2, 2, 5, 3, '#', style)

		// Check inside
		cell := s.Get(4, 3)
		if cell.Rune != '#' {
			t.Error("fill rect should set rune")
		}
		if !cell.Style.Equal(style) {
			t.Error("fill rect should set style")
		}

		// Check outside
		cell = s.Get(0, 0)
		if cell.Rune == '#' {
			t.Error("outside rect should not be filled")
		}
	})

	t.Run("Box", func(t *testing.T) {
		s := NewScreen(10, 10)
		s.Box(0, 0, 5, 3, DefaultStyle())

		// Check corners
		if s.Get(0, 0).Rune != '┌' {
			t.Error("top-left corner should be ┌")
		}
		if s.Get(4, 0).Rune != '┐' {
			t.Error("top-right corner should be ┐")
		}
		if s.Get(0, 2).Rune != '└' {
			t.Error("bottom-left corner should be └")
		}
		if s.Get(4, 2).Rune != '┘' {
			t.Error("bottom-right corner should be ┘")
		}

		// Check edges
		if s.Get(2, 0).Rune != '─' {
			t.Error("top edge should be ─")
		}
		if s.Get(0, 1).Rune != '│' {
			t.Error("left edge should be │")
		}
	})

	t.Run("Cursor", func(t *testing.T) {
		s := NewScreen(10, 10)
		s.SetCursor(5, 3, true)

		x, y, visible := s.Cursor()
		if x != 5 || y != 3 {
			t.Errorf("got cursor %d,%d, want 5,3", x, y)
		}
		if !visible {
			t.Error("cursor should be visible")
		}
	})

	t.Run("SubScreen", func(t *testing.T) {
		s := NewScreen(20, 10)
		sub := s.Sub(5, 2, 10, 5)

		w, h := sub.Size()
		if w != 10 || h != 5 {
			t.Errorf("got size %dx%d, want 10x5", w, h)
		}

		// Write through subscreen
		sub.SetString(0, 0, "Hello", DefaultStyle())

		// Check main screen
		cell := s.Get(5, 2)
		if cell.Rune != 'H' {
			t.Errorf("got %q, want 'H'", cell.Rune)
		}
	})
}

func TestRegion(t *testing.T) {
	t.Run("Contains", func(t *testing.T) {
		r := Region{X: 5, Y: 5, Width: 10, Height: 10}

		if !r.Contains(10, 10) {
			t.Error("should contain center point")
		}
		if !r.Contains(5, 5) {
			t.Error("should contain top-left")
		}
		if r.Contains(4, 5) {
			t.Error("should not contain point outside left")
		}
		if r.Contains(15, 10) {
			t.Error("should not contain point at right edge")
		}
	})

	t.Run("Intersect", func(t *testing.T) {
		r1 := Region{X: 0, Y: 0, Width: 10, Height: 10}
		r2 := Region{X: 5, Y: 5, Width: 10, Height: 10}

		inter := r1.Intersect(r2)
		if inter.X != 5 || inter.Y != 5 {
			t.Errorf("intersection origin: got %d,%d, want 5,5", inter.X, inter.Y)
		}
		if inter.Width != 5 || inter.Height != 5 {
			t.Errorf("intersection size: got %dx%d, want 5x5", inter.Width, inter.Height)
		}
	})

	t.Run("Intersect no overlap", func(t *testing.T) {
		r1 := Region{X: 0, Y: 0, Width: 5, Height: 5}
		r2 := Region{X: 10, Y: 10, Width: 5, Height: 5}

		inter := r1.Intersect(r2)
		if !inter.IsEmpty() {
			t.Error("non-overlapping regions should have empty intersection")
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		if !(Region{Width: 0, Height: 5}).IsEmpty() {
			t.Error("zero width should be empty")
		}
		if !(Region{Width: 5, Height: 0}).IsEmpty() {
			t.Error("zero height should be empty")
		}
		if (Region{Width: 1, Height: 1}).IsEmpty() {
			t.Error("1x1 should not be empty")
		}
	})
}

func TestANSI(t *testing.T) {
	t.Run("CursorTo", func(t *testing.T) {
		// CursorTo uses 0-indexed input, converts to 1-indexed ANSI
		result := CursorTo(0, 0)
		if result != "\x1b[1;1H" {
			t.Errorf("got %q, want \\x1b[1;1H", result)
		}

		result = CursorTo(9, 4)
		if result != "\x1b[5;10H" {
			t.Errorf("got %q, want \\x1b[5;10H", result)
		}
	})

	t.Run("CursorMovement", func(t *testing.T) {
		if CursorUp(0) != "" {
			t.Error("CursorUp(0) should return empty")
		}
		if CursorUp(3) != "\x1b[3A" {
			t.Errorf("CursorUp(3) got %q", CursorUp(3))
		}
		if CursorDown(2) != "\x1b[2B" {
			t.Errorf("CursorDown(2) got %q", CursorDown(2))
		}
		if CursorForward(5) != "\x1b[5C" {
			t.Errorf("CursorForward(5) got %q", CursorForward(5))
		}
		if CursorBack(1) != "\x1b[1D" {
			t.Errorf("CursorBack(1) got %q", CursorBack(1))
		}
	})

	t.Run("StyleToANSI basic", func(t *testing.T) {
		s := DefaultStyle()
		result := StyleToANSI(s)
		// Should contain reset (0) and default colors (39, 49)
		if !strings.Contains(result, "\x1b[") {
			t.Error("should start with escape sequence")
		}
		if !strings.HasSuffix(result, "m") {
			t.Error("should end with 'm'")
		}
	})

	t.Run("StyleToANSI with attributes", func(t *testing.T) {
		s := DefaultStyle().WithBold(true).WithItalic(true)
		result := StyleToANSI(s)
		if !strings.Contains(result, ";1;") { // Bold
			t.Error("should contain bold code")
		}
		if !strings.Contains(result, ";3;") { // Italic
			t.Error("should contain italic code")
		}
	})

	t.Run("StyleToANSI with 16 colors", func(t *testing.T) {
		s := DefaultStyle().WithFG(ColorRed).WithBG(ColorBlue)
		result := StyleToANSI(s)
		if !strings.Contains(result, ";31;") { // Red FG
			t.Error("should contain red foreground (31)")
		}
		if !strings.Contains(result, ";44") { // Blue BG
			t.Error("should contain blue background (44)")
		}
	})

	t.Run("StyleToANSI with 256 colors", func(t *testing.T) {
		s := DefaultStyle().WithFG(Color256(196))
		result := StyleToANSI(s)
		if !strings.Contains(result, ";38;5;196") {
			t.Error("should contain 256-color FG sequence")
		}
	})

	t.Run("StyleToANSI with RGB", func(t *testing.T) {
		s := DefaultStyle().WithBG(RGB(100, 150, 200))
		result := StyleToANSI(s)
		if !strings.Contains(result, ";48;2;100;150;200") {
			t.Error("should contain RGB BG sequence")
		}
	})
}

func TestANSIWriter(t *testing.T) {
	t.Run("basic writing", func(t *testing.T) {
		w := NewANSIWriter()
		w.WriteString("Hello")
		if !strings.Contains(w.String(), "Hello") {
			t.Error("should contain written text")
		}
	})

	t.Run("cursor optimization", func(t *testing.T) {
		w := NewANSIWriter()
		w.MoveTo(0, 0)
		w.WriteRune('A')
		// Next position should be (1, 0), which is sequential
		// MoveTo should be a no-op
		initialLen := w.Len()
		w.MoveTo(1, 0)
		if w.Len() != initialLen {
			t.Error("sequential cursor position should not emit escape")
		}
	})

	t.Run("style caching", func(t *testing.T) {
		w := NewANSIWriter()
		style := DefaultStyle().WithBold(true)
		w.SetStyle(style)
		initialLen := w.Len()
		w.SetStyle(style) // Same style
		if w.Len() != initialLen {
			t.Error("same style should not emit escape")
		}
	})
}

func TestRenderer(t *testing.T) {
	t.Run("Render empty screen", func(t *testing.T) {
		screen := NewScreen(10, 5)
		r := NewRenderer(screen)
		output := r.Render()
		// Empty screen should produce minimal output
		if len(output) > 100 {
			t.Errorf("empty screen should produce minimal output, got %d bytes", len(output))
		}
	})

	t.Run("RenderFull", func(t *testing.T) {
		screen := NewScreen(10, 5)
		screen.SetString(0, 0, "Hello", DefaultStyle())
		r := NewRenderer(screen)
		output := r.RenderFull()

		if !strings.Contains(output, "Hello") {
			t.Error("full render should contain text")
		}
		if !strings.Contains(output, ANSIClearScreen) {
			t.Error("full render should clear screen")
		}
	})

	t.Run("Render diff", func(t *testing.T) {
		screen := NewScreen(10, 5)
		r := NewRenderer(screen)

		// Initial full render
		screen.SetString(0, 0, "Hello", DefaultStyle())
		_ = r.RenderFull()

		// Now change one character
		screen.SetString(0, 0, "Jello", DefaultStyle())
		output := r.Render()

		// Should only update the changed character
		if strings.Contains(output, ANSIClearScreen) {
			t.Error("diff render should not clear screen")
		}
		// Should contain J but be relatively short
		if !strings.Contains(output, "J") {
			t.Error("should contain changed character")
		}
	})

	t.Run("DiffStats", func(t *testing.T) {
		screen := NewScreen(10, 5)
		r := NewRenderer(screen)

		_ = r.RenderFull()
		screen.SetString(0, 0, "ABC", DefaultStyle())

		stats := r.ComputeDiffStats()
		if stats.TotalCells != 50 {
			t.Errorf("total cells: got %d, want 50", stats.TotalCells)
		}
		if stats.ChangedCells != 3 {
			t.Errorf("changed cells: got %d, want 3", stats.ChangedCells)
		}
	})
}

func TestCompositor(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		c := NewCompositor(80, 24)
		w, h := c.Size()
		if w != 80 || h != 24 {
			t.Errorf("got size %dx%d, want 80x24", w, h)
		}
	})

	t.Run("layer compositing", func(t *testing.T) {
		c := NewCompositor(20, 10)

		// Draw on base layer
		c.Screen().SetString(0, 0, "Background", DefaultStyle())

		// Add overlay
		overlay := c.AddLayer()
		overlay.SetString(0, 0, "Over", DefaultStyle().WithFG(ColorRed))

		// Compose
		c.Compose()

		// Check that overlay takes precedence
		cell := c.Screen().Get(0, 0)
		if cell.Rune != 'O' {
			t.Errorf("overlay should override background, got %q", cell.Rune)
		}
	})

	t.Run("Resize", func(t *testing.T) {
		c := NewCompositor(20, 10)
		c.AddLayer()
		c.Resize(40, 20)

		w, h := c.Size()
		if w != 40 || h != 20 {
			t.Errorf("after resize: got %dx%d, want 40x20", w, h)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		c := NewCompositor(10, 5)
		c.Screen().SetString(0, 0, "Test", DefaultStyle())
		c.Clear()

		if !c.Screen().Get(0, 0).Empty() {
			t.Error("Clear should empty all cells")
		}
	})
}

func TestFrameBuilder(t *testing.T) {
	t.Run("fluent API", func(t *testing.T) {
		screen := NewScreen(20, 10)
		fb := NewFrameBuilder(screen)

		fb.Text(0, 0, "Title", DefaultStyle()).
			HLine(0, 1, 20, '─', DefaultStyle()).
			Box(0, 2, 10, 5, DefaultStyle())

		if screen.Get(0, 0).Rune != 'T' {
			t.Error("Text not rendered")
		}
		if screen.Get(5, 1).Rune != '─' {
			t.Error("HLine not rendered")
		}
		if screen.Get(0, 2).Rune != '┌' {
			t.Error("Box not rendered")
		}
	})
}

func TestRenderToString(t *testing.T) {
	output := RenderToString(10, 3, func(s *Screen) {
		s.SetString(0, 0, "Line 1", DefaultStyle())
		s.SetString(0, 1, "Line 2", DefaultStyle())
	})

	if !strings.Contains(output, "Line 1") {
		t.Error("should contain Line 1")
	}
	if !strings.Contains(output, "Line 2") {
		t.Error("should contain Line 2")
	}
}

func BenchmarkScreenSet(b *testing.B) {
	screen := NewScreen(80, 24)
	style := DefaultStyle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Set(i%80, (i/80)%24, 'X', style)
	}
}

func BenchmarkRender(b *testing.B) {
	screen := NewScreen(80, 24)
	renderer := NewRenderer(screen)
	style := DefaultStyle()

	// Fill screen
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			screen.Set(x, y, 'X', style)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Change a few cells
		screen.Set(i%80, (i/80)%24, 'Y', style)
		_ = renderer.Render()
	}
}

func BenchmarkRenderFull(b *testing.B) {
	screen := NewScreen(80, 24)
	renderer := NewRenderer(screen)
	style := DefaultStyle()

	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			screen.Set(x, y, 'X', style)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.RenderFull()
	}
}

// Bridge tests

func TestStreamRenderer(t *testing.T) {
	t.Run("basic rendering", func(t *testing.T) {
		sr := NewStreamRenderer(40, 5)
		style := DefaultStyle()

		output := sr.Update("Hello World", style)
		if output == "" {
			t.Error("first render should produce output")
		}

		// Same content should return empty
		output = sr.Update("Hello World", style)
		if output != "" {
			t.Error("same content should return empty string")
		}
	})

	t.Run("content change", func(t *testing.T) {
		sr := NewStreamRenderer(40, 5)
		style := DefaultStyle()

		_ = sr.Update("Line 1", style)
		output := sr.Update("Line 1\nLine 2", style)

		if output == "" {
			t.Error("content change should produce output")
		}
	})

	t.Run("resize", func(t *testing.T) {
		sr := NewStreamRenderer(40, 5)
		style := DefaultStyle()

		_ = sr.Update("Test", style)
		sr.Resize(80, 10)

		// After resize, next update should produce full output
		output := sr.Update("Test", style)
		if output == "" {
			t.Error("after resize should produce output")
		}
	})

	t.Run("line overflow", func(t *testing.T) {
		sr := NewStreamRenderer(20, 3)
		style := DefaultStyle()

		// More lines than height
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
		_ = sr.Update(content, style)

		view := sr.View()
		lines := strings.Split(view, "\n")
		if len(lines) > 3 {
			t.Errorf("should limit to height, got %d lines", len(lines))
		}
	})

	t.Run("View fallback", func(t *testing.T) {
		sr := NewStreamRenderer(40, 5)
		style := DefaultStyle()

		_ = sr.Update("Test content", style)
		view := sr.View()
		if view != "Test content" {
			t.Errorf("View() = %q, want %q", view, "Test content")
		}
	})
}

func TestWrapText(t *testing.T) {
	t.Run("basic wrap", func(t *testing.T) {
		lines := wrapText("Hello World Test", 10)
		if len(lines) < 2 {
			t.Error("should wrap text")
		}
	})

	t.Run("preserve paragraphs", func(t *testing.T) {
		lines := wrapText("Para 1\n\nPara 2", 40)
		found := false
		for _, line := range lines {
			if line == "" {
				found = true
				break
			}
		}
		if !found {
			t.Error("should preserve empty lines between paragraphs")
		}
	})

	t.Run("long word", func(t *testing.T) {
		lines := wrapText("supercalifragilisticexpialidocious", 10)
		if len(lines) < 3 {
			t.Error("should break long word across lines")
		}
	})

	t.Run("zero width", func(t *testing.T) {
		lines := wrapText("test", 0)
		if lines != nil {
			t.Error("zero width should return nil")
		}
	})
}

func TestTextBox(t *testing.T) {
	result := TextBox("Hello", 0, 0, 20, 5, DefaultStyle(), DefaultStyle())
	if result == "" {
		t.Error("TextBox should produce output")
	}
	// Should contain the box characters
	if !strings.Contains(result, "┌") {
		t.Error("should have top-left corner")
	}
}

func TestProgressBar(t *testing.T) {
	t.Run("half progress", func(t *testing.T) {
		style := DefaultStyle()
		result := ProgressBar(50, 100, 10, '█', '░', style, style)
		if result == "" {
			t.Error("should produce output")
		}
	})

	t.Run("zero width", func(t *testing.T) {
		result := ProgressBar(50, 100, 0, '█', '░', DefaultStyle(), DefaultStyle())
		if result != "" {
			t.Error("zero width should return empty")
		}
	})

	t.Run("zero total", func(t *testing.T) {
		result := ProgressBar(50, 0, 10, '█', '░', DefaultStyle(), DefaultStyle())
		if result != "" {
			t.Error("zero total should return empty")
		}
	})
}

func TestSpinnerFrame(t *testing.T) {
	// Should cycle through frames
	frames := make(map[rune]bool)
	for i := 0; i < 20; i++ {
		frames[SpinnerFrame(i)] = true
	}
	if len(frames) != len(spinnerFrames) {
		t.Errorf("should cycle through all %d frames, got %d unique", len(spinnerFrames), len(frames))
	}
}

func TestTableRenderer(t *testing.T) {
	tr := NewTableRenderer(40, 10, []int{15, 10, 15})
	style := DefaultStyle()

	tr.AddRow([]string{"Name", "Age", "City"}, style.WithBold(true))
	tr.AddSeparator(style)
	tr.AddRow([]string{"Alice", "30", "NYC"}, style)
	tr.AddRow([]string{"Bob", "25", "LA"}, style)

	result := tr.Render()
	if result == "" {
		t.Error("table should produce output")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("should contain table data")
	}
}

func TestDirectWriter(t *testing.T) {
	var buf strings.Builder
	dw := NewDirectWriter(&buf, 20, 5)

	screen := dw.Screen()
	screen.SetString(0, 0, "Hello", DefaultStyle())

	err := dw.Flush()
	if err != nil {
		t.Errorf("Flush error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("should write to output")
	}
}

func BenchmarkStreamRenderer(b *testing.B) {
	sr := NewStreamRenderer(80, 24)
	style := DefaultStyle()

	// Simulate streaming text
	texts := []string{
		"Starting...",
		"Processing data...",
		"Analyzing results...",
		"Generating output...",
		"Complete!",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sr.Update(texts[i%len(texts)], style)
	}
}
