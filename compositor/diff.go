package compositor

import (
	"strings"
)

// Renderer computes minimal ANSI output from screen buffer changes.
type Renderer struct {
	screen *Screen
	writer *ANSIWriter
}

// NewRenderer creates a renderer for the given screen.
func NewRenderer(screen *Screen) *Renderer {
	return &Renderer{
		screen: screen,
		writer: NewANSIWriter(),
	}
}

// Render computes the diff between current and previous buffers,
// returns minimal ANSI escape sequences to update the terminal.
func (r *Renderer) Render() string {
	r.screen.mu.Lock()
	defer r.screen.mu.Unlock()

	r.writer = NewANSIWriter()
	r.writer.Grow(r.screen.width * r.screen.height / 4) // ~25% change estimate

	r.writer.HideCursor()

	lastX, lastY := -1, -1
	var lastStyle Style
	styleSet := false

	for y := 0; y < r.screen.height; y++ {
		for x := 0; x < r.screen.width; x++ {
			curr := r.screen.current[y][x]
			prev := r.screen.previous[y][x]

			// Skip continuation cells from wide characters
			if curr.Width == 0 {
				continue
			}

			// Skip unchanged cells
			if curr.Equal(prev) {
				continue
			}

			// Position cursor if not sequential
			if y != lastY || x != lastX+1 {
				r.writer.MoveTo(x, y)
			}

			// Update style if different
			if !styleSet || !curr.Style.Equal(lastStyle) {
				r.writer.SetStyle(curr.Style)
				lastStyle = curr.Style
				styleSet = true
			}

			// Write the character
			if curr.Rune == 0 {
				r.writer.WriteRune(' ')
			} else {
				r.writer.WriteRune(curr.Rune)
			}

			lastX = x + int(curr.Width) - 1
			lastY = y
		}
	}

	// Reset style
	r.writer.Reset()

	// Position cursor for input if visible
	if r.screen.cursorVisible {
		r.writer.buf.WriteString(CursorTo(r.screen.cursorX, r.screen.cursorY))
		r.writer.ShowCursor()
	}

	// Swap buffers for next frame
	r.screen.previous, r.screen.current = r.screen.current, r.screen.previous

	// Clear the new current buffer for next frame
	for y := range r.screen.current {
		for x := range r.screen.current[y] {
			r.screen.current[y][x] = EmptyCell()
		}
	}

	return r.writer.String()
}

// RenderFull forces a complete redraw (e.g., after resize or initial paint).
func (r *Renderer) RenderFull() string {
	r.screen.mu.Lock()
	defer r.screen.mu.Unlock()

	r.writer = NewANSIWriter()
	r.writer.Grow(r.screen.width * r.screen.height * 2)

	// Clear screen and home cursor
	r.writer.buf.WriteString(ANSIClearScreen)
	r.writer.buf.WriteString(ANSICursorHome)
	r.writer.HideCursor()

	var lastStyle Style
	styleSet := false

	for y := 0; y < r.screen.height; y++ {
		if y > 0 {
			r.writer.buf.WriteString("\r\n")
		}

		for x := 0; x < r.screen.width; x++ {
			cell := r.screen.current[y][x]

			// Skip continuation cells
			if cell.Width == 0 {
				continue
			}

			// Update style if different
			if !styleSet || !cell.Style.Equal(lastStyle) {
				r.writer.SetStyle(cell.Style)
				lastStyle = cell.Style
				styleSet = true
			}

			// Write character
			if cell.Rune == 0 {
				r.writer.WriteRune(' ')
			} else {
				r.writer.WriteRune(cell.Rune)
			}
		}
	}

	// Reset style
	r.writer.Reset()

	// Position cursor
	if r.screen.cursorVisible {
		r.writer.buf.WriteString(CursorTo(r.screen.cursorX, r.screen.cursorY))
		r.writer.ShowCursor()
	}

	// Sync buffers
	for y := range r.screen.current {
		copy(r.screen.previous[y], r.screen.current[y])
	}

	return r.writer.String()
}

// RenderRegion renders only a specific region (useful for partial updates).
func (r *Renderer) RenderRegion(region Region) string {
	r.screen.mu.Lock()
	defer r.screen.mu.Unlock()

	r.writer = NewANSIWriter()

	// Clamp region to screen bounds
	region = region.Intersect(Region{X: 0, Y: 0, Width: r.screen.width, Height: r.screen.height})
	if region.IsEmpty() {
		return ""
	}

	r.writer.HideCursor()

	var lastStyle Style
	styleSet := false
	lastX, lastY := -1, -1

	for y := region.Y; y < region.Y+region.Height; y++ {
		for x := region.X; x < region.X+region.Width; x++ {
			curr := r.screen.current[y][x]
			prev := r.screen.previous[y][x]

			if curr.Width == 0 {
				continue
			}

			if curr.Equal(prev) {
				continue
			}

			if y != lastY || x != lastX+1 {
				r.writer.MoveTo(x, y)
			}

			if !styleSet || !curr.Style.Equal(lastStyle) {
				r.writer.SetStyle(curr.Style)
				lastStyle = curr.Style
				styleSet = true
			}

			if curr.Rune == 0 {
				r.writer.WriteRune(' ')
			} else {
				r.writer.WriteRune(curr.Rune)
			}

			lastX = x + int(curr.Width) - 1
			lastY = y

			// Update previous for this cell
			r.screen.previous[y][x] = curr
		}
	}

	r.writer.Reset()

	if r.screen.cursorVisible {
		r.writer.buf.WriteString(CursorTo(r.screen.cursorX, r.screen.cursorY))
		r.writer.ShowCursor()
	}

	return r.writer.String()
}

// DiffStats returns statistics about what changed between frames.
// Useful for debugging and performance monitoring.
type DiffStats struct {
	TotalCells    int
	ChangedCells  int
	SkippedCells  int // Continuation cells
	StyleChanges  int
	CursorJumps   int
}

// ComputeDiffStats analyzes the current vs previous buffer without rendering.
func (r *Renderer) ComputeDiffStats() DiffStats {
	r.screen.mu.RLock()
	defer r.screen.mu.RUnlock()

	stats := DiffStats{
		TotalCells: r.screen.width * r.screen.height,
	}

	var lastStyle Style
	styleSet := false
	lastX, lastY := -1, -1

	for y := 0; y < r.screen.height; y++ {
		for x := 0; x < r.screen.width; x++ {
			curr := r.screen.current[y][x]
			prev := r.screen.previous[y][x]

			if curr.Width == 0 {
				stats.SkippedCells++
				continue
			}

			if !curr.Equal(prev) {
				stats.ChangedCells++

				if y != lastY || x != lastX+1 {
					stats.CursorJumps++
				}

				if !styleSet || !curr.Style.Equal(lastStyle) {
					stats.StyleChanges++
					lastStyle = curr.Style
					styleSet = true
				}

				lastX = x + int(curr.Width) - 1
				lastY = y
			}
		}
	}

	return stats
}

// Compositor manages the full rendering pipeline.
type Compositor struct {
	screen   *Screen
	renderer *Renderer
	layers   []*Screen // Overlay layers (e.g., popups, menus)
}

// NewCompositor creates a new compositor with the given dimensions.
func NewCompositor(width, height int) *Compositor {
	screen := NewScreen(width, height)
	return &Compositor{
		screen:   screen,
		renderer: NewRenderer(screen),
		layers:   make([]*Screen, 0),
	}
}

// Screen returns the main screen buffer.
func (c *Compositor) Screen() *Screen {
	return c.screen
}

// Resize updates compositor dimensions.
func (c *Compositor) Resize(width, height int) {
	c.screen.Resize(width, height)
	for _, layer := range c.layers {
		layer.Resize(width, height)
	}
}

// AddLayer adds an overlay layer.
func (c *Compositor) AddLayer() *Screen {
	layer := NewScreen(c.screen.width, c.screen.height)
	c.layers = append(c.layers, layer)
	return layer
}

// RemoveLayer removes the topmost layer.
func (c *Compositor) RemoveLayer() {
	if len(c.layers) > 0 {
		c.layers = c.layers[:len(c.layers)-1]
	}
}

// ClearLayers removes all overlay layers.
func (c *Compositor) ClearLayers() {
	c.layers = c.layers[:0]
}

// Compose merges all layers onto the main screen.
// Non-empty cells from higher layers override lower layers.
func (c *Compositor) Compose() {
	// Layers are composited onto the main screen
	// Higher index = higher priority (drawn on top)
	for _, layer := range c.layers {
		c.compositeLayer(layer)
	}
}

// compositeLayer blends a layer onto the main screen.
func (c *Compositor) compositeLayer(layer *Screen) {
	layer.mu.RLock()
	c.screen.mu.Lock()
	defer layer.mu.RUnlock()
	defer c.screen.mu.Unlock()

	for y := 0; y < min(c.screen.height, layer.height); y++ {
		for x := 0; x < min(c.screen.width, layer.width); x++ {
			cell := layer.current[y][x]
			// Only overlay non-empty cells
			if !cell.Empty() {
				c.screen.current[y][x] = cell
			}
		}
	}
}

// Render computes diff and returns ANSI output.
func (c *Compositor) Render() string {
	c.Compose()
	return c.renderer.Render()
}

// RenderFull forces complete redraw.
func (c *Compositor) RenderFull() string {
	c.Compose()
	return c.renderer.RenderFull()
}

// Size returns current dimensions.
func (c *Compositor) Size() (width, height int) {
	return c.screen.Size()
}

// Clear clears all buffers.
func (c *Compositor) Clear() {
	c.screen.Clear()
	for _, layer := range c.layers {
		layer.Clear()
	}
}

// SetCursor sets cursor position and visibility.
func (c *Compositor) SetCursor(x, y int, visible bool) {
	c.screen.SetCursor(x, y, visible)
}

// FrameBuilder provides a fluent API for building frames.
type FrameBuilder struct {
	screen *Screen
}

// NewFrameBuilder creates a builder for the screen.
func NewFrameBuilder(screen *Screen) *FrameBuilder {
	return &FrameBuilder{screen: screen}
}

// Text writes text at position with style.
func (fb *FrameBuilder) Text(x, y int, text string, style Style) *FrameBuilder {
	fb.screen.SetString(x, y, text, style)
	return fb
}

// Fill fills a rectangle.
func (fb *FrameBuilder) Fill(x, y, w, h int, r rune, style Style) *FrameBuilder {
	fb.screen.FillRect(x, y, w, h, r, style)
	return fb
}

// Box draws a box border.
func (fb *FrameBuilder) Box(x, y, w, h int, style Style) *FrameBuilder {
	fb.screen.Box(x, y, w, h, style)
	return fb
}

// HLine draws a horizontal line.
func (fb *FrameBuilder) HLine(x, y, length int, r rune, style Style) *FrameBuilder {
	fb.screen.HLine(x, y, length, r, style)
	return fb
}

// VLine draws a vertical line.
func (fb *FrameBuilder) VLine(x, y, length int, r rune, style Style) *FrameBuilder {
	fb.screen.VLine(x, y, length, r, style)
	return fb
}

// RenderToString is a convenience function for quick rendering.
func RenderToString(width, height int, draw func(*Screen)) string {
	screen := NewScreen(width, height)
	draw(screen)

	var buf strings.Builder
	buf.WriteString(ANSIClearScreen)
	buf.WriteString(ANSICursorHome)

	var lastStyle Style
	styleSet := false

	for y := 0; y < height; y++ {
		if y > 0 {
			buf.WriteString("\r\n")
		}
		for x := 0; x < width; x++ {
			cell := screen.current[y][x]
			if cell.Width == 0 {
				continue
			}
			if !styleSet || !cell.Style.Equal(lastStyle) {
				buf.WriteString(StyleToANSI(cell.Style))
				lastStyle = cell.Style
				styleSet = true
			}
			if cell.Rune == 0 {
				buf.WriteRune(' ')
			} else {
				buf.WriteRune(cell.Rune)
			}
		}
	}

	buf.WriteString(ANSIReset)
	return buf.String()
}
