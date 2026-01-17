package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// TableColumn defines a column in a table.
type TableColumn struct {
	Title string
	Width int
}

// Table is a simple data grid widget.
type Table struct {
	FocusableBase
	Columns       []TableColumn
	Rows          [][]string
	selected      int
	offset        int
	style         backend.Style
	headerStyle   backend.Style
	selectedStyle backend.Style
	cachedWidths  []int
	cachedTotal   int
	cachedSig     uint32
}

// NewTable creates a table with columns.
func NewTable(columns ...TableColumn) *Table {
	return &Table{
		Columns:       columns,
		style:         backend.DefaultStyle(),
		headerStyle:   backend.DefaultStyle().Bold(true),
		selectedStyle: backend.DefaultStyle().Reverse(true),
	}
}

// SetRows updates table rows.
func (t *Table) SetRows(rows [][]string) {
	if t == nil {
		return
	}
	t.Rows = rows
}

// Measure returns the desired size.
func (t *Table) Measure(constraints runtime.Constraints) runtime.Size {
	height := minInt(len(t.Rows)+1, constraints.MaxHeight)
	if height <= 0 {
		height = constraints.MinHeight
	}
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: height})
}

// Render draws the table.
func (t *Table) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	bounds := t.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.Fill(bounds, ' ', t.style)
	widths := t.columnWidths(bounds.Width)
	if len(widths) == 0 {
		return
	}
	// Header
	x := bounds.X
	for i, col := range t.Columns {
		if x >= bounds.X+bounds.Width {
			break
		}
		width := widths[i]
		title := truncateString(col.Title, width)
		writePadded(ctx.Buffer, x, bounds.Y, width, title, t.headerStyle)
		x += width + 1
	}

	// Rows
	rowArea := bounds.Height - 1
	if rowArea <= 0 {
		return
	}
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(t.Rows) {
		t.selected = len(t.Rows) - 1
	}
	if t.selected < t.offset {
		t.offset = t.selected
	}
	if t.selected >= t.offset+rowArea {
		t.offset = t.selected - rowArea + 1
	}
	for row := 0; row < rowArea; row++ {
		rowIndex := t.offset + row
		if rowIndex < 0 || rowIndex >= len(t.Rows) {
			break
		}
		style := t.style
		if rowIndex == t.selected {
			style = t.selectedStyle
		}
		x = bounds.X
		for colIndex, width := range widths {
			if x >= bounds.X+bounds.Width {
				break
			}
			cell := ""
			if colIndex < len(t.Rows[rowIndex]) {
				cell = t.Rows[rowIndex][colIndex]
			}
			cell = truncateString(cell, width)
			writePadded(ctx.Buffer, x, bounds.Y+1+row, width, cell, style)
			x += width + 1
		}
	}
}

// HandleMessage handles row navigation.
func (t *Table) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp:
		t.setSelected(t.selected - 1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.setSelected(t.selected + 1)
		return runtime.Handled()
	case terminal.KeyPageUp:
		t.setSelected(t.selected - t.bounds.Height)
		return runtime.Handled()
	case terminal.KeyPageDown:
		t.setSelected(t.selected + t.bounds.Height)
		return runtime.Handled()
	case terminal.KeyHome:
		t.setSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		t.setSelected(len(t.Rows) - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *Table) setSelected(index int) {
	if t == nil {
		return
	}
	if len(t.Rows) == 0 {
		t.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.Rows) {
		index = len(t.Rows) - 1
	}
	t.selected = index
}

func (t *Table) columnWidths(total int) []int {
	if len(t.Columns) == 0 {
		return nil
	}
	if total == t.cachedTotal && len(t.cachedWidths) == len(t.Columns) && t.cachedSig == t.columnsSignature() {
		return t.cachedWidths
	}
	available := total - (len(t.Columns) - 1)
	if available < 0 {
		available = 0
	}
	fixed := 0
	flexCount := 0
	for _, col := range t.Columns {
		if col.Width > 0 {
			fixed += col.Width
		} else {
			flexCount++
		}
	}
	widths := make([]int, len(t.Columns))
	remaining := available - fixed
	if remaining < 0 {
		remaining = 0
	}
	flexWidth := 0
	if flexCount > 0 {
		flexWidth = remaining / flexCount
		if flexWidth <= 0 {
			flexWidth = 1
		}
	}
	for i, col := range t.Columns {
		if col.Width > 0 {
			widths[i] = col.Width
		} else {
			widths[i] = flexWidth
		}
	}
	t.cachedTotal = total
	t.cachedSig = t.columnsSignature()
	t.cachedWidths = widths
	return widths
}

func (t *Table) columnsSignature() uint32 {
	if t == nil {
		return 0
	}
	var sig uint32 = uint32(len(t.Columns))
	for _, col := range t.Columns {
		sig = sig*31 + uint32(col.Width+1)
	}
	return sig
}

// ScrollBy scrolls selection by delta.
func (t *Table) ScrollBy(dx, dy int) {
	if t == nil || len(t.Rows) == 0 || dy == 0 {
		return
	}
	t.setSelected(t.selected + dy)
	t.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (t *Table) ScrollTo(x, y int) {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(y)
	t.Invalidate()
}

// PageBy scrolls by a number of pages.
func (t *Table) PageBy(pages int) {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	pageSize := t.bounds.Height - 1
	if pageSize < 1 {
		pageSize = 1
	}
	t.setSelected(t.selected + pages*pageSize)
	t.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (t *Table) ScrollToStart() {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(0)
	t.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (t *Table) ScrollToEnd() {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(len(t.Rows) - 1)
	t.Invalidate()
}

var _ scroll.Controller = (*Table)(nil)
