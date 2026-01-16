package runtime

// HitGrid maps screen cells to widgets for mouse hit testing.
type HitGrid struct {
	width   int
	height  int
	cells   []int
	widgets []Widget
}

// NewHitGrid creates a new hit grid with the given dimensions.
func NewHitGrid(width, height int) *HitGrid {
	grid := &HitGrid{}
	grid.Resize(width, height)
	return grid
}

// Resize updates the hit grid dimensions.
func (g *HitGrid) Resize(width, height int) {
	if width == g.width && height == g.height {
		return
	}
	g.width = width
	g.height = height
	size := width * height
	if size <= 0 {
		g.cells = nil
		g.widgets = nil
		return
	}
	g.cells = make([]int, size)
	g.Clear()
}

// Clear resets the grid contents.
func (g *HitGrid) Clear() {
	for i := range g.cells {
		g.cells[i] = -1
	}
	g.widgets = g.widgets[:0]
}

// Add records a widget occupying the specified bounds.
func (g *HitGrid) Add(widget Widget, bounds Rect) {
	if widget == nil || g.width <= 0 || g.height <= 0 {
		return
	}
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	bounds = bounds.Intersection(Rect{X: 0, Y: 0, Width: g.width, Height: g.height})
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	id := len(g.widgets)
	g.widgets = append(g.widgets, widget)

	for y := bounds.Y; y < bounds.Y+bounds.Height; y++ {
		row := y * g.width
		for x := bounds.X; x < bounds.X+bounds.Width; x++ {
			g.cells[row+x] = id
		}
	}
}

// WidgetAt returns the widget at the given screen position.
func (g *HitGrid) WidgetAt(x, y int) Widget {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return nil
	}
	idx := g.cells[y*g.width+x]
	if idx < 0 || idx >= len(g.widgets) {
		return nil
	}
	return g.widgets[idx]
}
