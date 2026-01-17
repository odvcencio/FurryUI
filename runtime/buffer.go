package runtime

// Widget Rendering Architecture:
//
// Widgets render to runtime.Buffer via RenderContext. The Buffer tracks
// which cells changed (dirty tracking). app_widget.render() iterates
// dirty cells and writes them to the backend.
//
// compositor.Screen exists as an alternative for pure-ANSI output but
// is not used in the tcell backend path.

import "github.com/odvcencio/fluffy-ui/backend"

// Cell represents a single character cell in the buffer.
type Cell = backend.Cell

// Buffer is a 2D grid of cells for rendering widgets.
// Widgets render to the buffer, then the buffer is flushed to the backend.
// Supports dirty-region tracking for partial redraws.
type Buffer struct {
	cells  []Cell
	width  int
	height int

	// Dirty tracking - tracks which cells have changed
	dirtyStamp       []uint32 // Generation marker per cell
	dirtyGen         uint32
	dirtyAll         bool
	dirtyCount       int   // Number of dirty cells (fast check)
	dirtyRect        Rect  // Bounding box of dirty region
	dirtyIndices     []int // Sparse dirty list for fast iteration
	dirtyListCap     int   // Max indices to collect before disabling list
	dirtyListEnabled bool
}

// NewBuffer creates a buffer with the given dimensions.
func NewBuffer(w, h int) *Buffer {
	total := w * h
	dirtyListCap := calcDirtyListCap(total)
	return &Buffer{
		cells:            make([]Cell, total),
		dirtyStamp:       make([]uint32, total),
		dirtyGen:         1,
		width:            w,
		height:           h,
		dirtyIndices:     make([]int, 0, min(total, dirtyListCap)),
		dirtyListCap:     dirtyListCap,
		dirtyListEnabled: true,
	}
}

// Size returns the buffer dimensions.
func (b *Buffer) Size() (w, h int) {
	return b.width, b.height
}

// Resize changes the buffer dimensions, preserving content where possible.
func (b *Buffer) Resize(w, h int) {
	if w == b.width && h == b.height {
		return
	}
	newCells := make([]Cell, w*h)
	// Copy existing content
	minW := min(w, b.width)
	minH := min(h, b.height)
	for y := 0; y < minH; y++ {
		copy(newCells[y*w:y*w+minW], b.cells[y*b.width:y*b.width+minW])
	}
	b.cells = newCells
	b.dirtyStamp = make([]uint32, w*h)
	b.dirtyGen = 1
	b.dirtyAll = false
	b.dirtyCount = 0
	b.dirtyRect = Rect{}
	b.width = w
	b.height = h
	b.dirtyListCap = calcDirtyListCap(w * h)
	b.dirtyListEnabled = true
	b.dirtyIndices = b.dirtyIndices[:0]
	if cap(b.dirtyIndices) < min(w*h, b.dirtyListCap) {
		b.dirtyIndices = make([]int, 0, min(w*h, b.dirtyListCap))
	}
	// Mark entire buffer dirty on resize
	b.MarkAllDirty()
}

// Clear fills the buffer with spaces and default style.
func (b *Buffer) Clear() {
	b.Fill(Rect{0, 0, b.width, b.height}, ' ', backend.DefaultStyle())
}

// ClearRect fills a rectangular region with spaces and default style.
func (b *Buffer) ClearRect(r Rect) {
	b.Fill(r, ' ', backend.DefaultStyle())
}

// Get returns the cell at position (x, y).
// Returns empty cell if out of bounds.
func (b *Buffer) Get(x, y int) Cell {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return Cell{Rune: ' '}
	}
	return b.cells[y*b.width+x]
}

// Set writes a rune with style at position (x, y).
// No-op if out of bounds. Marks the cell as dirty if changed.
func (b *Buffer) Set(x, y int, r rune, s backend.Style) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	idx := y*b.width + x
	old := b.cells[idx]
	// Only mark dirty if content actually changed
	if old.Rune != r || old.Style != s {
		b.cells[idx] = Cell{Rune: r, Style: s}
		b.markCellDirty(x, y, idx)
	}
}

// SetString writes a string starting at (x, y).
// Clips to buffer bounds. Marks changed cells as dirty.
func (b *Buffer) SetString(x, y int, s string, style backend.Style) {
	if y < 0 || y >= b.height {
		return
	}
	if x >= b.width {
		return
	}
	if x >= 0 {
		px := x
		i := 0
		for i < len(s) && px < b.width {
			ch := s[i]
			if ch >= 0x80 {
				break
			}
			idx := y*b.width + px
			old := b.cells[idx]
			r := rune(ch)
			if old.Rune != r || old.Style != style {
				b.cells[idx] = Cell{Rune: r, Style: style}
				b.markCellDirty(px, y, idx)
			}
			i++
			px++
		}
		if i >= len(s) {
			return
		}
		offset := x + i
		for j, r := range s[i:] {
			px := offset + j
			if px < 0 {
				continue
			}
			if px >= b.width {
				break
			}
			idx := y*b.width + px
			old := b.cells[idx]
			if old.Rune != r || old.Style != style {
				b.cells[idx] = Cell{Rune: r, Style: style}
				b.markCellDirty(px, y, idx)
			}
		}
		return
	}
	for i, r := range s {
		px := x + i
		if px < 0 {
			continue
		}
		if px >= b.width {
			break
		}
		idx := y*b.width + px
		old := b.cells[idx]
		if old.Rune != r || old.Style != style {
			b.cells[idx] = Cell{Rune: r, Style: style}
			b.markCellDirty(px, y, idx)
		}
	}
}

// Fill fills a rectangular region with a rune and style.
// Marks changed cells as dirty.
func (b *Buffer) Fill(r Rect, ch rune, s backend.Style) {
	// Clip to buffer bounds
	x0 := max(0, r.X)
	y0 := max(0, r.Y)
	x1 := min(b.width, r.X+r.Width)
	y1 := min(b.height, r.Y+r.Height)

	cell := Cell{Rune: ch, Style: s}
	for y := y0; y < y1; y++ {
		idx := y*b.width + x0
		for x := x0; x < x1; x++ {
			if b.cells[idx] != cell {
				b.cells[idx] = cell
				b.markCellDirty(x, y, idx)
			}
			idx++
		}
	}
}

// DrawBox draws a border around a rect using box-drawing characters.
func (b *Buffer) DrawBox(r Rect, s backend.Style) {
	if r.Width < 2 || r.Height < 2 {
		return
	}

	// Corners
	b.Set(r.X, r.Y, '┌', s)
	b.Set(r.X+r.Width-1, r.Y, '┐', s)
	b.Set(r.X, r.Y+r.Height-1, '└', s)
	b.Set(r.X+r.Width-1, r.Y+r.Height-1, '┘', s)

	// Horizontal edges
	for x := r.X + 1; x < r.X+r.Width-1; x++ {
		b.Set(x, r.Y, '─', s)
		b.Set(x, r.Y+r.Height-1, '─', s)
	}

	// Vertical edges
	for y := r.Y + 1; y < r.Y+r.Height-1; y++ {
		b.Set(r.X, y, '│', s)
		b.Set(r.X+r.Width-1, y, '│', s)
	}
}

// DrawRoundedBox draws a border with rounded corners.
func (b *Buffer) DrawRoundedBox(r Rect, s backend.Style) {
	if r.Width < 2 || r.Height < 2 {
		return
	}

	// Rounded corners
	b.Set(r.X, r.Y, '╭', s)
	b.Set(r.X+r.Width-1, r.Y, '╮', s)
	b.Set(r.X, r.Y+r.Height-1, '╰', s)
	b.Set(r.X+r.Width-1, r.Y+r.Height-1, '╯', s)

	// Horizontal edges
	for x := r.X + 1; x < r.X+r.Width-1; x++ {
		b.Set(x, r.Y, '─', s)
		b.Set(x, r.Y+r.Height-1, '─', s)
	}

	// Vertical edges
	for y := r.Y + 1; y < r.Y+r.Height-1; y++ {
		b.Set(r.X, y, '│', s)
		b.Set(r.X+r.Width-1, y, '│', s)
	}
}

// SubBuffer returns a view into a rectangular region of the buffer.
// Writes to the SubBuffer are translated and clipped to the region.
type SubBuffer struct {
	parent *Buffer
	bounds Rect
}

// Sub creates a SubBuffer for the given region.
func (b *Buffer) Sub(r Rect) *SubBuffer {
	return &SubBuffer{parent: b, bounds: r}
}

// Size returns the sub-buffer dimensions.
func (s *SubBuffer) Size() (w, h int) {
	return s.bounds.Width, s.bounds.Height
}

// Set writes a rune at position relative to the sub-buffer.
func (s *SubBuffer) Set(x, y int, r rune, style backend.Style) {
	if x < 0 || x >= s.bounds.Width || y < 0 || y >= s.bounds.Height {
		return
	}
	s.parent.Set(s.bounds.X+x, s.bounds.Y+y, r, style)
}

// SetString writes a string at position relative to the sub-buffer.
func (s *SubBuffer) SetString(x, y int, str string, style backend.Style) {
	if y < 0 || y >= s.bounds.Height {
		return
	}
	for i, r := range str {
		px := x + i
		if px < 0 {
			continue
		}
		if px >= s.bounds.Width {
			break
		}
		s.parent.Set(s.bounds.X+px, s.bounds.Y+y, r, style)
	}
}

// Fill fills a region relative to the sub-buffer.
func (s *SubBuffer) Fill(r Rect, ch rune, style backend.Style) {
	// Translate and clip to sub-buffer bounds
	clipped := r.Intersection(Rect{0, 0, s.bounds.Width, s.bounds.Height})
	if clipped.Width == 0 || clipped.Height == 0 {
		return
	}
	s.parent.Fill(Rect{
		X:      s.bounds.X + clipped.X,
		Y:      s.bounds.Y + clipped.Y,
		Width:  clipped.Width,
		Height: clipped.Height,
	}, ch, style)
}

// Clear fills the sub-buffer region with spaces.
func (s *SubBuffer) Clear() {
	s.Fill(Rect{0, 0, s.bounds.Width, s.bounds.Height}, ' ', backend.DefaultStyle())
}

// --- Dirty Tracking Methods ---

// markCellDirty marks a single cell as dirty and updates the bounding box.
func (b *Buffer) markCellDirty(x, y, idx int) {
	if b.dirtyAll {
		return
	}
	if b.dirtyStamp[idx] != b.dirtyGen {
		b.dirtyStamp[idx] = b.dirtyGen
		b.dirtyCount++

		// Expand dirty rect
		if b.dirtyCount == 1 {
			// First dirty cell - initialize rect
			b.dirtyRect = Rect{X: x, Y: y, Width: 1, Height: 1}
		} else {
			// Expand to include this cell
			if x < b.dirtyRect.X {
				b.dirtyRect.Width += b.dirtyRect.X - x
				b.dirtyRect.X = x
			} else if x >= b.dirtyRect.X+b.dirtyRect.Width {
				b.dirtyRect.Width = x - b.dirtyRect.X + 1
			}
			if y < b.dirtyRect.Y {
				b.dirtyRect.Height += b.dirtyRect.Y - y
				b.dirtyRect.Y = y
			} else if y >= b.dirtyRect.Y+b.dirtyRect.Height {
				b.dirtyRect.Height = y - b.dirtyRect.Y + 1
			}
		}
		if b.dirtyListEnabled {
			if b.dirtyCount <= b.dirtyListCap {
				b.dirtyIndices = append(b.dirtyIndices, idx)
			} else {
				b.dirtyListEnabled = false
				b.dirtyIndices = b.dirtyIndices[:0]
			}
		}
	}
}

// MarkAllDirty marks the entire buffer as dirty.
func (b *Buffer) MarkAllDirty() {
	b.dirtyAll = true
	b.dirtyCount = b.width * b.height
	b.dirtyRect = Rect{X: 0, Y: 0, Width: b.width, Height: b.height}
	b.dirtyListEnabled = false
	b.dirtyIndices = b.dirtyIndices[:0]
}

// ClearDirty resets all dirty flags.
func (b *Buffer) ClearDirty() {
	b.dirtyAll = false
	b.dirtyCount = 0
	b.dirtyRect = Rect{}
	b.dirtyListEnabled = true
	b.dirtyIndices = b.dirtyIndices[:0]
	b.advanceDirtyGen()
}

// IsDirty returns true if any cells have changed.
func (b *Buffer) IsDirty() bool {
	return b.dirtyAll || b.dirtyCount > 0
}

// DirtyCount returns the number of dirty cells.
func (b *Buffer) DirtyCount() int {
	if b.dirtyAll {
		return b.width * b.height
	}
	return b.dirtyCount
}

// DirtyRect returns the bounding box of dirty cells.
// Returns empty rect if nothing is dirty.
func (b *Buffer) DirtyRect() Rect {
	if b.dirtyAll {
		return Rect{X: 0, Y: 0, Width: b.width, Height: b.height}
	}
	return b.dirtyRect
}

// IsCellDirty returns true if the cell at (x, y) is dirty.
func (b *Buffer) IsCellDirty(x, y int) bool {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return false
	}
	if b.dirtyAll {
		return true
	}
	return b.dirtyStamp[y*b.width+x] == b.dirtyGen
}

// ForEachDirtyCell calls fn for each dirty cell.
// More efficient than iterating all cells when few are dirty.
func (b *Buffer) ForEachDirtyCell(fn func(x, y int, cell Cell)) {
	if b.dirtyAll {
		for y := 0; y < b.height; y++ {
			rowStart := y * b.width
			for x := 0; x < b.width; x++ {
				idx := rowStart + x
				fn(x, y, b.cells[idx])
			}
		}
		return
	}
	if b.dirtyCount == 0 {
		return
	}
	// If most cells are dirty, iterate linearly
	if b.dirtyCount > b.width*b.height/2 {
		for y := 0; y < b.height; y++ {
			for x := 0; x < b.width; x++ {
				idx := y*b.width + x
				if b.dirtyStamp[idx] == b.dirtyGen {
					fn(x, y, b.cells[idx])
				}
			}
		}
		return
	}
	if b.dirtyListEnabled && len(b.dirtyIndices) == b.dirtyCount {
		rectArea := b.dirtyRect.Width * b.dirtyRect.Height
		if rectArea > b.dirtyCount*2 {
			for _, idx := range b.dirtyIndices {
				if b.dirtyStamp[idx] != b.dirtyGen {
					continue
				}
				y := idx / b.width
				x := idx - y*b.width
				fn(x, y, b.cells[idx])
			}
			return
		}
	}
	// Otherwise, iterate only within dirty rect
	for y := b.dirtyRect.Y; y < b.dirtyRect.Y+b.dirtyRect.Height && y < b.height; y++ {
		start := y*b.width + b.dirtyRect.X
		end := y*b.width + min(b.width, b.dirtyRect.X+b.dirtyRect.Width)
		for idx := start; idx < end; idx++ {
			if b.dirtyStamp[idx] == b.dirtyGen {
				x := idx - y*b.width
				fn(x, y, b.cells[idx])
			}
		}
	}
}

// ForEachDirtySpan calls fn for each contiguous dirty span per row.
func (b *Buffer) ForEachDirtySpan(fn func(y, startX, endX int)) {
	if b.dirtyAll {
		for y := 0; y < b.height; y++ {
			fn(y, 0, b.width)
		}
		return
	}
	if b.dirtyCount == 0 {
		return
	}
	rect := b.dirtyRect
	if rect.Width <= 0 || rect.Height <= 0 {
		return
	}
	xEnd := min(b.width, rect.X+rect.Width)
	yEnd := min(b.height, rect.Y+rect.Height)
	for y := rect.Y; y < yEnd; y++ {
		rowStart := y * b.width
		x := rect.X
		for x < xEnd {
			if b.dirtyStamp[rowStart+x] != b.dirtyGen {
				x++
				continue
			}
			start := x
			x++
			for x < xEnd && b.dirtyStamp[rowStart+x] == b.dirtyGen {
				x++
			}
			fn(y, start, x)
		}
	}
}

// Cells returns the underlying cell slice.
func (b *Buffer) Cells() []Cell {
	return b.cells
}

func calcDirtyListCap(total int) int {
	if total <= 0 {
		return 0
	}
	capacity := total / 4
	if capacity < 256 {
		capacity = 256
	}
	if capacity > 8192 {
		capacity = 8192
	}
	if capacity > total {
		capacity = total
	}
	return capacity
}

func (b *Buffer) advanceDirtyGen() {
	b.dirtyGen++
	if b.dirtyGen == 0 {
		clear(b.dirtyStamp)
		b.dirtyGen = 1
	}
}
