package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// MenuItem describes a menu entry.
type MenuItem struct {
	ID       string
	Title    string
	Shortcut string
	Children []*MenuItem
	Expanded bool
	Disabled bool
	OnSelect func()
}

// Menu renders a vertical menu.
type Menu struct {
	FocusableBase
	Items         []*MenuItem
	selectedIndex int
	offset        int
	style         backend.Style
	selectedStyle backend.Style
	indentCache   []string
	flatCache     []menuRow
	flatDirty     bool
	itemsLen      int
	itemsFirst    *MenuItem
}

// NewMenu creates a new menu.
func NewMenu(items ...*MenuItem) *Menu {
	return &Menu{
		Items:         items,
		selectedIndex: 0,
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
		itemsLen:      len(items),
		itemsFirst:    firstItem(items),
	}
}

// SetItems replaces the menu items and clears cached rows.
func (m *Menu) SetItems(items ...*MenuItem) {
	if m == nil {
		return
	}
	m.Items = items
	m.itemsLen = len(items)
	m.itemsFirst = firstItem(items)
	m.flatDirty = true
}

// Measure returns desired size.
func (m *Menu) Measure(constraints runtime.Constraints) runtime.Size {
	count := len(m.flatten())
	height := minInt(count, constraints.MaxHeight)
	if height <= 0 {
		height = constraints.MinHeight
	}
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: height})
}

// Render draws the menu.
func (m *Menu) Render(ctx runtime.RenderContext) {
	if m == nil {
		return
	}
	bounds := m.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.Fill(bounds, ' ', m.style)
	rows := m.flatten()
	if len(rows) == 0 {
		return
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
	if m.selectedIndex >= len(rows) {
		m.selectedIndex = len(rows) - 1
	}
	if m.selectedIndex < m.offset {
		m.offset = m.selectedIndex
	}
	if m.selectedIndex >= m.offset+bounds.Height {
		m.offset = m.selectedIndex - bounds.Height + 1
	}
	for i := 0; i < bounds.Height; i++ {
		rowIndex := m.offset + i
		if rowIndex < 0 || rowIndex >= len(rows) {
			break
		}
		row := rows[rowIndex]
		style := m.style
		if rowIndex == m.selectedIndex {
			style = m.selectedStyle
		}
		prefix := "  "
		if len(row.item.Children) > 0 {
			if row.item.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		}
		indent := m.indent(row.depth)
		line := indent + prefix + row.item.Title
		if row.item.Shortcut != "" {
			line += " (" + row.item.Shortcut + ")"
		}
		line = truncateString(line, bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+i, bounds.Width, line, style)
	}
}

// HandleMessage handles navigation and selection.
func (m *Menu) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if m == nil || !m.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	rows := m.flatten()
	switch key.Key {
	case terminal.KeyUp:
		m.setSelected(m.selectedIndex-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		m.setSelected(m.selectedIndex+1, len(rows))
		return runtime.Handled()
	case terminal.KeyLeft:
		if row := m.selectedRow(rows); row != nil && row.item.Expanded {
			row.item.Expanded = false
			m.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if row := m.selectedRow(rows); row != nil && len(row.item.Children) > 0 {
			row.item.Expanded = true
			m.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyEnter:
		if row := m.selectedRow(rows); row != nil && !row.item.Disabled {
			if len(row.item.Children) > 0 {
				row.item.Expanded = !row.item.Expanded
				m.flatDirty = true
			}
			if row.item.OnSelect != nil {
				row.item.OnSelect()
			}
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

type menuRow struct {
	item  *MenuItem
	depth int
}

func (m *Menu) flatten() []menuRow {
	currentFirst := firstItem(m.Items)
	if m.itemsLen != len(m.Items) || m.itemsFirst != currentFirst {
		m.itemsLen = len(m.Items)
		m.itemsFirst = currentFirst
		m.flatDirty = true
	}
	if !m.flatDirty {
		return m.flatCache
	}
	rows := m.flatCache[:0]
	var walk func(items []*MenuItem, depth int)
	walk = func(items []*MenuItem, depth int) {
		for _, item := range items {
			if item == nil {
				continue
			}
			rows = append(rows, menuRow{item: item, depth: depth})
			if item.Expanded {
				walk(item.Children, depth+1)
			}
		}
	}
	walk(m.Items, 0)
	m.flatCache = rows
	m.flatDirty = false
	return m.flatCache
}

func (m *Menu) setSelected(index int, count int) {
	if count == 0 {
		m.selectedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	m.selectedIndex = index
}

func (m *Menu) selectedRow(rows []menuRow) *menuRow {
	if m.selectedIndex < 0 || m.selectedIndex >= len(rows) {
		return nil
	}
	return &rows[m.selectedIndex]
}

// ScrollBy scrolls selection by delta.
func (m *Menu) ScrollBy(dx, dy int) {
	if m == nil || dy == 0 {
		return
	}
	rows := m.flatten()
	m.setSelected(m.selectedIndex+dy, len(rows))
	m.Invalidate()
}

func (m *Menu) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if len(m.indentCache) == 0 {
		m.indentCache = []string{""}
	}
	for len(m.indentCache) <= depth {
		m.indentCache = append(m.indentCache, m.indentCache[len(m.indentCache)-1]+"  ")
	}
	return m.indentCache[depth]
}

func firstItem(items []*MenuItem) *MenuItem {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

// ScrollTo scrolls to an absolute row index.
func (m *Menu) ScrollTo(x, y int) {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(y, len(rows))
	m.Invalidate()
}

// PageBy scrolls by a number of pages.
func (m *Menu) PageBy(pages int) {
	if m == nil {
		return
	}
	rows := m.flatten()
	pageSize := m.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	m.setSelected(m.selectedIndex+pages*pageSize, len(rows))
	m.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (m *Menu) ScrollToStart() {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(0, len(rows))
	m.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (m *Menu) ScrollToEnd() {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(len(rows)-1, len(rows))
	m.Invalidate()
}

var _ scroll.Controller = (*Menu)(nil)
