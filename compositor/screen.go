package compositor

import (
	"sync"

	"github.com/mattn/go-runewidth"
)

// Screen manages a double-buffered virtual terminal.
// It maintains two buffers: current (being built) and previous (last rendered).
// The diff engine compares these to output minimal ANSI escape sequences.
type Screen struct {
	mu       sync.RWMutex
	width    int
	height   int
	current  [][]Cell // Frame being built
	previous [][]Cell // Last rendered frame

	// Cursor position for input
	cursorX, cursorY int
	cursorVisible    bool
}

// NewScreen creates a new screen buffer with the given dimensions.
func NewScreen(width, height int) *Screen {
	s := &Screen{
		width:  width,
		height: height,
	}
	s.current = s.allocBuffer(width, height)
	s.previous = s.allocBuffer(width, height)
	return s
}

// allocBuffer creates a new buffer filled with empty cells.
func (s *Screen) allocBuffer(w, h int) [][]Cell {
	buf := make([][]Cell, h)
	for y := range buf {
		buf[y] = make([]Cell, w)
		for x := range buf[y] {
			buf[y][x] = EmptyCell()
		}
	}
	return buf
}

// Resize changes screen dimensions, preserving content where possible.
// Forces a full redraw by resetting the previous buffer.
func (s *Screen) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if width == s.width && height == s.height {
		return
	}

	newCurrent := s.allocBuffer(width, height)
	newPrevious := s.allocBuffer(width, height)

	// Copy existing content that fits
	for y := 0; y < min(height, s.height); y++ {
		for x := 0; x < min(width, s.width); x++ {
			newCurrent[y][x] = s.current[y][x]
		}
	}

	s.current = newCurrent
	s.previous = newPrevious // Force full redraw on resize
	s.width = width
	s.height = height
}

// Clear resets the current buffer to empty cells.
func (s *Screen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for y := range s.current {
		for x := range s.current[y] {
			s.current[y][x] = EmptyCell()
		}
	}
}

// Set places a rune at the given position with style.
// Handles wide characters (CJK) by placing a continuation cell.
func (s *Screen) Set(x, y int, r rune, style Style) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.setUnsafe(x, y, r, style)
}

// setUnsafe places a rune without locking (caller must hold lock).
func (s *Screen) setUnsafe(x, y int, r rune, style Style) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}

	width := runewidth.RuneWidth(r)
	if width == 0 {
		width = 1 // Control characters, treat as single width
	}

	s.current[y][x] = Cell{Rune: r, Width: uint8(width), Style: style}

	// Wide characters occupy two cells; mark second as continuation
	if width == 2 && x+1 < s.width {
		s.current[y][x+1] = Cell{Rune: 0, Width: 0, Style: style}
	}
}

// SetString writes a string starting at position, returns number of columns written.
func (s *Screen) SetString(x, y int, str string, style Style) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if y < 0 || y >= s.height {
		return 0
	}

	col := x
	for _, r := range str {
		if col >= s.width {
			break
		}
		if col < 0 {
			// Skip characters before visible area, but track position
			col += runewidth.RuneWidth(r)
			continue
		}

		s.setUnsafe(col, y, r, style)
		col += runewidth.RuneWidth(r)
	}

	if col < x {
		return 0
	}
	return col - x
}

// FillRect fills a rectangle with a character and style.
func (s *Screen) FillRect(x, y, w, h int, r rune, style Style) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clamp bounds
	startX := max(0, x)
	startY := max(0, y)
	endX := min(s.width, x+w)
	endY := min(s.height, y+h)

	for row := startY; row < endY; row++ {
		for col := startX; col < endX; col++ {
			s.setUnsafe(col, row, r, style)
		}
	}
}

// FillRectCell fills a rectangle with a specific cell.
func (s *Screen) FillRectCell(x, y, w, h int, cell Cell) {
	s.mu.Lock()
	defer s.mu.Unlock()

	startX := max(0, x)
	startY := max(0, y)
	endX := min(s.width, x+w)
	endY := min(s.height, y+h)

	for row := startY; row < endY; row++ {
		for col := startX; col < endX; col++ {
			s.current[row][col] = cell
		}
	}
}

// HLine draws a horizontal line with a character.
func (s *Screen) HLine(x, y, length int, r rune, style Style) {
	s.FillRect(x, y, length, 1, r, style)
}

// VLine draws a vertical line with a character.
func (s *Screen) VLine(x, y, length int, r rune, style Style) {
	s.FillRect(x, y, 1, length, r, style)
}

// Box draws a box border using Unicode box-drawing characters.
func (s *Screen) Box(x, y, w, h int, style Style) {
	if w < 2 || h < 2 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Corners
	s.setUnsafe(x, y, '┌', style)
	s.setUnsafe(x+w-1, y, '┐', style)
	s.setUnsafe(x, y+h-1, '└', style)
	s.setUnsafe(x+w-1, y+h-1, '┘', style)

	// Top and bottom edges
	for col := x + 1; col < x+w-1; col++ {
		s.setUnsafe(col, y, '─', style)
		s.setUnsafe(col, y+h-1, '─', style)
	}

	// Left and right edges
	for row := y + 1; row < y+h-1; row++ {
		s.setUnsafe(x, row, '│', style)
		s.setUnsafe(x+w-1, row, '│', style)
	}
}

// SetCursor sets the cursor position and visibility.
func (s *Screen) SetCursor(x, y int, visible bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cursorX = x
	s.cursorY = y
	s.cursorVisible = visible
}

// Cursor returns current cursor position and visibility.
func (s *Screen) Cursor() (x, y int, visible bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cursorX, s.cursorY, s.cursorVisible
}

// Size returns current dimensions.
func (s *Screen) Size() (width, height int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.width, s.height
}

// Get returns the cell at the given position.
// Returns EmptyCell if out of bounds.
func (s *Screen) Get(x, y int) Cell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return EmptyCell()
	}
	return s.current[y][x]
}

// GetPrevious returns the cell from the previous frame.
func (s *Screen) GetPrevious(x, y int) Cell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return EmptyCell()
	}
	return s.previous[y][x]
}

// SwapBuffers swaps current and previous buffers, preparing for next frame.
// Returns the previous current buffer for diff computation.
func (s *Screen) SwapBuffers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.previous, s.current = s.current, s.previous
}

// ClearCurrentBuffer resets current buffer to empty cells.
// Call after SwapBuffers to prepare for next frame.
func (s *Screen) ClearCurrentBuffer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for y := range s.current {
		for x := range s.current[y] {
			s.current[y][x] = EmptyCell()
		}
	}
}

// CopyToPrevious copies current buffer to previous.
// Use after RenderFull to sync buffers.
func (s *Screen) CopyToPrevious() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for y := range s.current {
		copy(s.previous[y], s.current[y])
	}
}

// Blit copies a region from another screen onto this one.
// Useful for compositing overlays.
func (s *Screen) Blit(src *Screen, srcX, srcY, dstX, dstY, w, h int) {
	if src == nil {
		return
	}

	s.mu.Lock()
	src.mu.RLock()
	defer s.mu.Unlock()
	defer src.mu.RUnlock()

	for row := 0; row < h; row++ {
		sy := srcY + row
		dy := dstY + row

		if sy < 0 || sy >= src.height || dy < 0 || dy >= s.height {
			continue
		}

		for col := 0; col < w; col++ {
			sx := srcX + col
			dx := dstX + col

			if sx < 0 || sx >= src.width || dx < 0 || dx >= s.width {
				continue
			}

			s.current[dy][dx] = src.current[sy][sx]
		}
	}
}

// Region represents a rectangular area of the screen.
type Region struct {
	X, Y, Width, Height int
}

// Contains checks if a point is within the region.
func (r Region) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Intersect returns the intersection of two regions.
func (r Region) Intersect(other Region) Region {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)

	if x2 <= x1 || y2 <= y1 {
		return Region{} // Empty region
	}

	return Region{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// IsEmpty returns true if the region has zero area.
func (r Region) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// SubScreen creates a view into a region of this screen.
// Writes to the SubScreen are translated to the parent's coordinate space.
type SubScreen struct {
	parent *Screen
	region Region
}

// Sub creates a SubScreen for a region.
func (s *Screen) Sub(x, y, w, h int) *SubScreen {
	return &SubScreen{
		parent: s,
		region: Region{X: x, Y: y, Width: w, Height: h},
	}
}

// Set places a rune in the SubScreen's coordinate space.
func (ss *SubScreen) Set(x, y int, r rune, style Style) {
	ss.parent.Set(ss.region.X+x, ss.region.Y+y, r, style)
}

// SetString writes a string in the SubScreen's coordinate space.
func (ss *SubScreen) SetString(x, y int, str string, style Style) int {
	return ss.parent.SetString(ss.region.X+x, ss.region.Y+y, str, style)
}

// Clear clears the SubScreen region.
func (ss *SubScreen) Clear() {
	ss.parent.FillRect(ss.region.X, ss.region.Y, ss.region.Width, ss.region.Height, ' ', DefaultStyle())
}

// Size returns the SubScreen dimensions.
func (ss *SubScreen) Size() (width, height int) {
	return ss.region.Width, ss.region.Height
}

// Box draws a box in the SubScreen's coordinate space.
func (ss *SubScreen) Box(style Style) {
	ss.parent.Box(ss.region.X, ss.region.Y, ss.region.Width, ss.region.Height, style)
}

// FillRect fills a rectangle in the SubScreen's coordinate space.
func (ss *SubScreen) FillRect(x, y, w, h int, r rune, style Style) {
	ss.parent.FillRect(ss.region.X+x, ss.region.Y+y, w, h, r, style)
}
