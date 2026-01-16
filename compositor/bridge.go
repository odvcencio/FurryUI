// Package compositor bridge provides integration with Bubbletea's rendering model.
// Since Bubbletea expects View() to return a string and handles terminal updates
// internally, we provide adapters that:
// 1. Track what was previously rendered
// 2. Compute minimal updates for high-churn areas
// 3. Return optimized ANSI output that bubbletea can pass through
//
// For backend.Backend integration, see pkg/ui/render which provides BackendWriter.
package compositor

import (
	"io"
	"strings"
	"sync"
)

// StreamRenderer provides flicker-free rendering for a streaming text area.
// It tracks the previous content and outputs only the delta.
type StreamRenderer struct {
	mu           sync.Mutex
	width        int
	height       int
	lastContent  string
	lastLines    []string
	screen       *Screen
	renderer     *Renderer
	initialized  bool
}

// NewStreamRenderer creates a renderer for streaming content.
func NewStreamRenderer(width, height int) *StreamRenderer {
	screen := NewScreen(width, height)
	return &StreamRenderer{
		width:    width,
		height:   height,
		screen:   screen,
		renderer: NewRenderer(screen),
	}
}

// Resize updates the dimensions.
func (sr *StreamRenderer) Resize(width, height int) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if width == sr.width && height == sr.height {
		return
	}

	sr.width = width
	sr.height = height
	sr.screen.Resize(width, height)
	sr.lastContent = ""
	sr.lastLines = nil
	sr.initialized = false
}

// Update renders new content, returning optimized ANSI output.
// If the content hasn't changed, returns empty string.
func (sr *StreamRenderer) Update(content string, style Style) string {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// No change
	if content == sr.lastContent && sr.initialized {
		return ""
	}

	// Split into lines
	lines := strings.Split(content, "\n")

	// Limit to visible height
	if len(lines) > sr.height {
		lines = lines[len(lines)-sr.height:]
	}

	// Clear screen and write new content
	sr.screen.Clear()
	for y, line := range lines {
		if y >= sr.height {
			break
		}
		sr.screen.SetString(0, y, line, style)
	}

	sr.lastContent = content
	sr.lastLines = lines

	// First render or after resize
	if !sr.initialized {
		sr.initialized = true
		return sr.renderer.RenderFull()
	}

	return sr.renderer.Render()
}

// View returns the current content as a plain string (for fallback).
func (sr *StreamRenderer) View() string {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if len(sr.lastLines) == 0 {
		return ""
	}
	return strings.Join(sr.lastLines, "\n")
}

// DirectWriter provides direct terminal output for compositor rendering.
// Use this when you want to bypass bubbletea's renderer entirely.
type DirectWriter struct {
	out        io.Writer
	compositor *Compositor
	mu         sync.Mutex
}

// NewDirectWriter creates a writer that outputs directly to a terminal.
func NewDirectWriter(out io.Writer, width, height int) *DirectWriter {
	return &DirectWriter{
		out:        out,
		compositor: NewCompositor(width, height),
	}
}

// Screen returns the main screen for drawing.
func (dw *DirectWriter) Screen() *Screen {
	return dw.compositor.Screen()
}

// Resize updates dimensions.
func (dw *DirectWriter) Resize(width, height int) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.compositor.Resize(width, height)
}

// Flush renders and writes to output.
func (dw *DirectWriter) Flush() error {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	output := dw.compositor.Render()
	_, err := io.WriteString(dw.out, output)
	return err
}

// FlushFull forces a complete redraw.
func (dw *DirectWriter) FlushFull() error {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	output := dw.compositor.RenderFull()
	_, err := io.WriteString(dw.out, output)
	return err
}

// Clear clears all buffers.
func (dw *DirectWriter) Clear() {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.compositor.Clear()
}

// TextBox renders text in a box with word wrapping.
// Returns the rendered output as a string.
func TextBox(text string, x, y, width, height int, style Style, borderStyle Style) string {
	screen := NewScreen(width, height)

	// Draw border if we have room
	if width >= 2 && height >= 2 {
		screen.Box(0, 0, width, height, borderStyle)
	}

	// Render text inside the box
	innerX := 1
	innerY := 1
	innerWidth := width - 2
	innerHeight := height - 2

	if innerWidth <= 0 || innerHeight <= 0 {
		return ""
	}

	// Word wrap and render
	lines := wrapText(text, innerWidth)
	for i, line := range lines {
		if i >= innerHeight {
			break
		}
		screen.SetString(innerX, innerY+i, line, style)
	}

	return RenderToString(width, height, func(s *Screen) {
		for row := 0; row < height; row++ {
			for col := 0; col < width; col++ {
				cell := screen.Get(col, row)
				s.Set(col, row, cell.Rune, cell.Style)
			}
		}
	})
}

// wrapText wraps text to a given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return nil
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var line strings.Builder
		lineLen := 0

		for i, word := range words {
			wordLen := len([]rune(word))

			if lineLen > 0 && lineLen+1+wordLen > width {
				lines = append(lines, line.String())
				line.Reset()
				lineLen = 0
			}

			if lineLen > 0 {
				line.WriteRune(' ')
				lineLen++
			}

			// Handle words longer than width
			if wordLen > width && lineLen == 0 {
				runes := []rune(word)
				for len(runes) > 0 {
					chunk := min(len(runes), width)
					line.WriteString(string(runes[:chunk]))
					lines = append(lines, line.String())
					line.Reset()
					runes = runes[chunk:]
				}
				lineLen = 0
			} else {
				line.WriteString(word)
				lineLen += wordLen
			}

			// Last word
			if i == len(words)-1 && lineLen > 0 {
				lines = append(lines, line.String())
			}
		}
	}

	return lines
}

// ProgressBar renders a progress bar.
func ProgressBar(current, total, width int, filled, empty rune, filledStyle, emptyStyle Style) string {
	if width <= 0 || total <= 0 {
		return ""
	}

	screen := NewScreen(width, 1)
	progress := float64(current) / float64(total)
	filledCount := int(progress * float64(width))

	for x := 0; x < width; x++ {
		if x < filledCount {
			screen.Set(x, 0, filled, filledStyle)
		} else {
			screen.Set(x, 0, empty, emptyStyle)
		}
	}

	return RenderToString(width, 1, func(s *Screen) {
		for x := 0; x < width; x++ {
			cell := screen.Get(x, 0)
			s.Set(x, 0, cell.Rune, cell.Style)
		}
	})
}

// Spinner renders a spinner frame.
var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// SpinnerFrame returns the current spinner character.
func SpinnerFrame(tick int) rune {
	return spinnerFrames[tick%len(spinnerFrames)]
}

// Table renders a simple table.
type TableRenderer struct {
	screen    *Screen
	columns   []int // Column widths
	row       int
	rowHeight int
}

// NewTableRenderer creates a table renderer.
func NewTableRenderer(width, height int, columns []int) *TableRenderer {
	return &TableRenderer{
		screen:    NewScreen(width, height),
		columns:   columns,
		rowHeight: 1,
	}
}

// SetRowHeight sets height per row.
func (tr *TableRenderer) SetRowHeight(h int) {
	tr.rowHeight = h
}

// AddRow adds a row of cells.
func (tr *TableRenderer) AddRow(cells []string, style Style) {
	x := 0
	for i, cell := range cells {
		if i >= len(tr.columns) {
			break
		}
		tr.screen.SetString(x, tr.row, cell, style)
		x += tr.columns[i]
	}
	tr.row += tr.rowHeight
}

// AddSeparator adds a horizontal line.
func (tr *TableRenderer) AddSeparator(style Style) {
	w, _ := tr.screen.Size()
	tr.screen.HLine(0, tr.row, w, '─', style)
	tr.row++
}

// Render returns the table as a string.
func (tr *TableRenderer) Render() string {
	w, h := tr.screen.Size()
	return RenderToString(w, h, func(s *Screen) {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				cell := tr.screen.Get(x, y)
				s.Set(x, y, cell.Rune, cell.Style)
			}
		}
	})
}
