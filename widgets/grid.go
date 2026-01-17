package widgets

import "github.com/odvcencio/fluffy-ui/runtime"

// GridChild positions a widget in the grid.
type GridChild struct {
	Widget  runtime.Widget
	Row     int
	Col     int
	RowSpan int
	ColSpan int
}

// Grid lays out children in rows and columns.
type Grid struct {
	Base
	Rows     int
	Cols     int
	Gap      int
	Children []GridChild
}

// NewGrid creates a grid with the given dimensions.
func NewGrid(rows, cols int) *Grid {
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	return &Grid{Rows: rows, Cols: cols}
}

// Add adds a child at the given cell.
func (g *Grid) Add(child runtime.Widget, row, col, rowSpan, colSpan int) {
	if g == nil || child == nil {
		return
	}
	if rowSpan <= 0 {
		rowSpan = 1
	}
	if colSpan <= 0 {
		colSpan = 1
	}
	g.Children = append(g.Children, GridChild{
		Widget:  child,
		Row:     row,
		Col:     col,
		RowSpan: rowSpan,
		ColSpan: colSpan,
	})
}

// Measure estimates the grid size.
func (g *Grid) Measure(constraints runtime.Constraints) runtime.Size {
	rows := g.Rows
	cols := g.Cols
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	maxW, maxH := 0, 0
	for _, child := range g.Children {
		if child.Widget == nil {
			continue
		}
		size := child.Widget.Measure(runtime.Unbounded())
		if size.Width > maxW {
			maxW = size.Width
		}
		if size.Height > maxH {
			maxH = size.Height
		}
	}
	width := maxW*cols + g.Gap*max(0, cols-1)
	height := maxH*rows + g.Gap*max(0, rows-1)
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

// Layout positions children within the grid.
func (g *Grid) Layout(bounds runtime.Rect) {
	g.Base.Layout(bounds)
	rows := g.Rows
	cols := g.Cols
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	totalGapW := g.Gap * max(0, cols-1)
	totalGapH := g.Gap * max(0, rows-1)
	cellW := 0
	cellH := 0
	if cols > 0 {
		cellW = max(0, (bounds.Width-totalGapW)/cols)
	}
	if rows > 0 {
		cellH = max(0, (bounds.Height-totalGapH)/rows)
	}
	for _, child := range g.Children {
		if child.Widget == nil {
			continue
		}
		rowSpan := child.RowSpan
		colSpan := child.ColSpan
		if rowSpan <= 0 {
			rowSpan = 1
		}
		if colSpan <= 0 {
			colSpan = 1
		}
		x := bounds.X + child.Col*cellW + g.Gap*child.Col
		y := bounds.Y + child.Row*cellH + g.Gap*child.Row
		width := cellW*colSpan + g.Gap*max(0, colSpan-1)
		height := cellH*rowSpan + g.Gap*max(0, rowSpan-1)
		child.Widget.Layout(runtime.Rect{X: x, Y: y, Width: width, Height: height})
	}
}

// Render draws all children.
func (g *Grid) Render(ctx runtime.RenderContext) {
	for _, child := range g.Children {
		if child.Widget != nil {
			child.Widget.Render(ctx)
		}
	}
}

// HandleMessage forwards messages to children.
func (g *Grid) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range g.Children {
		if child.Widget == nil {
			continue
		}
		if result := child.Widget.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns grid children.
func (g *Grid) ChildWidgets() []runtime.Widget {
	if g == nil {
		return nil
	}
	out := make([]runtime.Widget, 0, len(g.Children))
	for _, child := range g.Children {
		if child.Widget != nil {
			out = append(out, child.Widget)
		}
	}
	return out
}
