package runtime

import "testing"

func TestConstraints_Tight(t *testing.T) {
	c := Tight(80, 24)

	if c.MinWidth != 80 || c.MaxWidth != 80 {
		t.Errorf("expected width 80-80, got %d-%d", c.MinWidth, c.MaxWidth)
	}
	if c.MinHeight != 24 || c.MaxHeight != 24 {
		t.Errorf("expected height 24-24, got %d-%d", c.MinHeight, c.MaxHeight)
	}
	if !c.IsTight() {
		t.Error("expected IsTight() to be true")
	}
}

func TestConstraints_Loose(t *testing.T) {
	c := Loose(80, 24)

	if c.MinWidth != 0 || c.MaxWidth != 80 {
		t.Errorf("expected width 0-80, got %d-%d", c.MinWidth, c.MaxWidth)
	}
	if c.MinHeight != 0 || c.MaxHeight != 24 {
		t.Errorf("expected height 0-24, got %d-%d", c.MinHeight, c.MaxHeight)
	}
	if c.IsTight() {
		t.Error("expected IsTight() to be false")
	}
}

func TestConstraints_Constrain(t *testing.T) {
	c := Constraints{MinWidth: 10, MaxWidth: 100, MinHeight: 5, MaxHeight: 50}

	tests := []struct {
		input    Size
		expected Size
	}{
		{Size{50, 25}, Size{50, 25}},   // Within bounds
		{Size{5, 25}, Size{10, 25}},    // Below min width
		{Size{150, 25}, Size{100, 25}}, // Above max width
		{Size{50, 2}, Size{50, 5}},     // Below min height
		{Size{50, 100}, Size{50, 50}},  // Above max height
	}

	for _, tc := range tests {
		got := c.Constrain(tc.input)
		if got != tc.expected {
			t.Errorf("Constrain(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestRect_Contains(t *testing.T) {
	r := Rect{X: 10, Y: 10, Width: 20, Height: 20}

	tests := []struct {
		x, y     int
		expected bool
	}{
		{15, 15, true},  // Center
		{10, 10, true},  // Top-left corner
		{29, 29, true},  // Bottom-right (exclusive edge)
		{9, 15, false},  // Left of rect
		{30, 15, false}, // Right of rect
		{15, 9, false},  // Above rect
		{15, 30, false}, // Below rect
	}

	for _, tc := range tests {
		got := r.Contains(tc.x, tc.y)
		if got != tc.expected {
			t.Errorf("Contains(%d, %d) = %v, want %v", tc.x, tc.y, got, tc.expected)
		}
	}
}

func TestRect_Intersection(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, Width: 20, Height: 20}
	r2 := Rect{X: 10, Y: 10, Width: 20, Height: 20}

	intersection := r1.Intersection(r2)
	expected := Rect{X: 10, Y: 10, Width: 10, Height: 10}

	if intersection != expected {
		t.Errorf("Intersection = %v, want %v", intersection, expected)
	}

	// Non-overlapping
	r3 := Rect{X: 100, Y: 100, Width: 10, Height: 10}
	noIntersection := r1.Intersection(r3)
	if noIntersection != ZeroRect {
		t.Errorf("Expected ZeroRect for non-overlapping, got %v", noIntersection)
	}
}

func TestRect_Inset(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 100, Height: 50}
	inset := r.Inset(5, 10, 5, 10)

	expected := Rect{X: 10, Y: 5, Width: 80, Height: 40}
	if inset != expected {
		t.Errorf("Inset = %v, want %v", inset, expected)
	}
}

func TestHandleResult(t *testing.T) {
	h := Handled()
	if !h.Handled || len(h.Commands) != 0 {
		t.Errorf("Handled() = %+v, want {Handled:true, Commands:[]}", h)
	}

	u := Unhandled()
	if u.Handled || len(u.Commands) != 0 {
		t.Errorf("Unhandled() = %+v, want {Handled:false, Commands:[]}", u)
	}

	cmd := Submit{Text: "hello"}
	wc := WithCommand(cmd)
	if !wc.Handled || len(wc.Commands) != 1 {
		t.Errorf("WithCommand() = %+v, want 1 command", wc)
	}
}

func TestWithCommands(t *testing.T) {
	cmd1 := Submit{Text: "hello"}
	cmd2 := Cancel{}
	cmd3 := Quit{}

	result := WithCommands(cmd1, cmd2, cmd3)

	if !result.Handled {
		t.Error("WithCommands should return handled")
	}
	if len(result.Commands) != 3 {
		t.Errorf("WithCommands should have 3 commands, got %d", len(result.Commands))
	}
}

func TestConstraints_TightWidth(t *testing.T) {
	c := TightWidth(50)

	if c.MinWidth != 50 || c.MaxWidth != 50 {
		t.Errorf("TightWidth width = %d-%d, want 50-50", c.MinWidth, c.MaxWidth)
	}
	if c.MinHeight != 0 {
		t.Errorf("TightWidth MinHeight = %d, want 0", c.MinHeight)
	}
	if c.MaxHeight != maxInt {
		t.Errorf("TightWidth MaxHeight = %d, want maxInt", c.MaxHeight)
	}
}

func TestConstraints_TightHeight(t *testing.T) {
	c := TightHeight(30)

	if c.MinHeight != 30 || c.MaxHeight != 30 {
		t.Errorf("TightHeight height = %d-%d, want 30-30", c.MinHeight, c.MaxHeight)
	}
	if c.MinWidth != 0 {
		t.Errorf("TightHeight MinWidth = %d, want 0", c.MinWidth)
	}
	if c.MaxWidth != maxInt {
		t.Errorf("TightHeight MaxWidth = %d, want maxInt", c.MaxWidth)
	}
}

func TestConstraints_Unbounded(t *testing.T) {
	c := Unbounded()

	if c.MinWidth != 0 || c.MinHeight != 0 {
		t.Errorf("Unbounded Min = %d,%d, want 0,0", c.MinWidth, c.MinHeight)
	}
	if c.MaxWidth != maxInt || c.MaxHeight != maxInt {
		t.Errorf("Unbounded Max = %d,%d, want maxInt,maxInt", c.MaxWidth, c.MaxHeight)
	}
}

func TestConstraints_MaxSize(t *testing.T) {
	c := Constraints{MinWidth: 10, MaxWidth: 100, MinHeight: 5, MaxHeight: 50}

	max := c.MaxSize()
	if max.Width != 100 || max.Height != 50 {
		t.Errorf("MaxSize = %v, want {100, 50}", max)
	}
}

func TestSize_Zero(t *testing.T) {
	zero := Size{0, 0}
	if !zero.Zero() {
		t.Error("Size{0,0}.Zero() should be true")
	}

	nonZero := Size{1, 0}
	if nonZero.Zero() {
		t.Error("Size{1,0}.Zero() should be false")
	}

	nonZero2 := Size{0, 1}
	if nonZero2.Zero() {
		t.Error("Size{0,1}.Zero() should be false")
	}
}

func TestNewRect(t *testing.T) {
	r := NewRect(10, 20, 30, 40)

	if r.X != 10 || r.Y != 20 || r.Width != 30 || r.Height != 40 {
		t.Errorf("NewRect = %v, want {10,20,30,40}", r)
	}
}

func TestRectFromSize(t *testing.T) {
	s := Size{30, 40}
	r := RectFromSize(s)

	if r.X != 0 || r.Y != 0 {
		t.Errorf("RectFromSize position = %d,%d, want 0,0", r.X, r.Y)
	}
	if r.Width != 30 || r.Height != 40 {
		t.Errorf("RectFromSize size = %dx%d, want 30x40", r.Width, r.Height)
	}
}

func TestRect_Size(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 30, Height: 40}

	s := r.Size()
	if s.Width != 30 || s.Height != 40 {
		t.Errorf("Rect.Size() = %v, want {30,40}", s)
	}
}

func TestRect_Intersects(t *testing.T) {
	r1 := Rect{X: 0, Y: 0, Width: 20, Height: 20}
	r2 := Rect{X: 10, Y: 10, Width: 20, Height: 20}
	r3 := Rect{X: 30, Y: 30, Width: 10, Height: 10}

	if !r1.Intersects(r2) {
		t.Error("r1 and r2 should intersect")
	}
	if r1.Intersects(r3) {
		t.Error("r1 and r3 should not intersect")
	}
}

func TestRect_InsetNegative(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 10}

	// Large insets should result in 0-size rect, not negative
	inset := r.Inset(10, 10, 10, 10)
	if inset.Width < 0 || inset.Height < 0 {
		t.Error("Inset should not produce negative dimensions")
	}
}

func TestZeroRect(t *testing.T) {
	if ZeroRect.Width != 0 || ZeroRect.Height != 0 {
		t.Errorf("ZeroRect should be {0,0,0,0}, got %v", ZeroRect)
	}
}
