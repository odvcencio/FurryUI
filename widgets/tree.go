package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// TreeNode represents a node in a tree.
type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
}

// Tree renders a hierarchical tree.
type Tree struct {
	FocusableBase
	Root          *TreeNode
	selectedIndex int
	offset        int
	style         backend.Style
	selectedStyle backend.Style
	indentCache   []string
	flatCache     []treeRow
	flatDirty     bool
	rootRef       *TreeNode
}

// NewTree creates a tree widget.
func NewTree(root *TreeNode) *Tree {
	return &Tree{
		Root:          root,
		selectedIndex: 0,
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
		rootRef:       root,
	}
}

// SetRoot updates the tree root and clears cached rows.
func (t *Tree) SetRoot(root *TreeNode) {
	if t == nil {
		return
	}
	t.Root = root
	t.rootRef = root
	t.flatDirty = true
}

// Measure returns desired size.
func (t *Tree) Measure(constraints runtime.Constraints) runtime.Size {
	count := len(t.flatten())
	height := min(count, constraints.MaxHeight)
	if height <= 0 {
		height = constraints.MinHeight
	}
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: height})
}

// Render draws the tree.
func (t *Tree) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	bounds := t.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.Fill(bounds, ' ', t.style)
	rows := t.flatten()
	if len(rows) == 0 {
		return
	}
	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	}
	if t.selectedIndex >= len(rows) {
		t.selectedIndex = len(rows) - 1
	}
	if t.selectedIndex < t.offset {
		t.offset = t.selectedIndex
	}
	if t.selectedIndex >= t.offset+bounds.Height {
		t.offset = t.selectedIndex - bounds.Height + 1
	}
	for i := 0; i < bounds.Height; i++ {
		rowIndex := t.offset + i
		if rowIndex < 0 || rowIndex >= len(rows) {
			break
		}
		row := rows[rowIndex]
		style := t.style
		if rowIndex == t.selectedIndex {
			style = t.selectedStyle
		}
		prefix := ""
		if len(row.node.Children) > 0 {
			if row.node.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		} else {
			prefix = "  "
		}
		indent := t.indent(row.depth)
		line := indent + prefix + row.node.Label
		line = truncateString(line, bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+i, bounds.Width, line, style)
	}
}

// HandleMessage handles navigation and expansion.
func (t *Tree) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	rows := t.flatten()
	switch key.Key {
	case terminal.KeyUp:
		t.setSelected(t.selectedIndex-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		t.setSelected(t.selectedIndex+1, len(rows))
		return runtime.Handled()
	case terminal.KeyLeft:
		if row := t.selectedRow(rows); row != nil && row.node.Expanded {
			row.node.Expanded = false
			t.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if row := t.selectedRow(rows); row != nil && len(row.node.Children) > 0 {
			row.node.Expanded = true
			t.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyEnter:
		if row := t.selectedRow(rows); row != nil && len(row.node.Children) > 0 {
			row.node.Expanded = !row.node.Expanded
			t.flatDirty = true
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

type treeRow struct {
	node  *TreeNode
	depth int
}

func (t *Tree) flatten() []treeRow {
	if t == nil || t.Root == nil {
		return nil
	}
	if t.rootRef != t.Root {
		t.rootRef = t.Root
		t.flatDirty = true
	}
	if !t.flatDirty {
		return t.flatCache
	}
	rows := t.flatCache[:0]
	var walk func(node *TreeNode, depth int)
	walk = func(node *TreeNode, depth int) {
		if node == nil {
			return
		}
		rows = append(rows, treeRow{node: node, depth: depth})
		if node.Expanded {
			for _, child := range node.Children {
				walk(child, depth+1)
			}
		}
	}
	walk(t.Root, 0)
	t.flatCache = rows
	t.flatDirty = false
	return t.flatCache
}

func (t *Tree) setSelected(index int, count int) {
	if count == 0 {
		t.selectedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	t.selectedIndex = index
}

func (t *Tree) selectedRow(rows []treeRow) *treeRow {
	if t.selectedIndex < 0 || t.selectedIndex >= len(rows) {
		return nil
	}
	return &rows[t.selectedIndex]
}

// ScrollBy scrolls selection by delta.
func (t *Tree) ScrollBy(dx, dy int) {
	if t == nil || dy == 0 {
		return
	}
	rows := t.flatten()
	t.setSelected(t.selectedIndex+dy, len(rows))
	t.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (t *Tree) ScrollTo(x, y int) {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(y, len(rows))
	t.Invalidate()
}

// PageBy scrolls by a number of pages.
func (t *Tree) PageBy(pages int) {
	if t == nil {
		return
	}
	rows := t.flatten()
	pageSize := t.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	t.setSelected(t.selectedIndex+pages*pageSize, len(rows))
	t.Invalidate()
}

func (t *Tree) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if len(t.indentCache) == 0 {
		t.indentCache = []string{""}
	}
	for len(t.indentCache) <= depth {
		t.indentCache = append(t.indentCache, t.indentCache[len(t.indentCache)-1]+"  ")
	}
	return t.indentCache[depth]
}

// ScrollToStart scrolls to the first row.
func (t *Tree) ScrollToStart() {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(0, len(rows))
	t.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (t *Tree) ScrollToEnd() {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(len(rows)-1, len(rows))
	t.Invalidate()
}

var _ scroll.Controller = (*Tree)(nil)
