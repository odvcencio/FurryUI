package compositor

import (
	"fmt"
	"strconv"
	"strings"
)

// ANSI escape sequences.
const (
	ANSIEscape       = "\x1b["
	ANSIClearScreen  = "\x1b[2J"
	ANSIClearLine    = "\x1b[2K"
	ANSICursorHome   = "\x1b[H"
	ANSICursorHide   = "\x1b[?25l"
	ANSICursorShow   = "\x1b[?25h"
	ANSIReset        = "\x1b[0m"
	ANSISaveCursor   = "\x1b[s"
	ANSIRestoreCursor = "\x1b[u"
	ANSIAltScreen    = "\x1b[?1049h"
	ANSIMainScreen   = "\x1b[?1049l"
)

// CursorTo returns ANSI sequence to move cursor to (x, y).
// Coordinates are 0-indexed, but ANSI uses 1-indexed.
func CursorTo(x, y int) string {
	return fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)
}

// CursorUp moves cursor up n lines.
func CursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA", n)
}

// CursorDown moves cursor down n lines.
func CursorDown(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dB", n)
}

// CursorForward moves cursor right n columns.
func CursorForward(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dC", n)
}

// CursorBack moves cursor left n columns.
func CursorBack(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dD", n)
}

// StyleToANSI converts a Style to ANSI escape sequence.
func StyleToANSI(s Style) string {
	var parts []string

	// Start with reset
	parts = append(parts, "0")

	// Attributes
	if s.Bold {
		parts = append(parts, "1")
	}
	if s.Dim {
		parts = append(parts, "2")
	}
	if s.Italic {
		parts = append(parts, "3")
	}
	if s.Underline {
		parts = append(parts, "4")
	}
	if s.Blink {
		parts = append(parts, "5")
	}
	if s.Reverse {
		parts = append(parts, "7")
	}
	if s.Strikethrough {
		parts = append(parts, "9")
	}

	// Foreground color
	parts = append(parts, colorToANSI(s.FG, true)...)

	// Background color
	parts = append(parts, colorToANSI(s.BG, false)...)

	return ANSIEscape + strings.Join(parts, ";") + "m"
}

// colorToANSI converts a Color to ANSI SGR parameters.
func colorToANSI(c Color, fg bool) []string {
	switch c.Mode {
	case ColorModeNone, ColorModeDefault:
		// Use default color (39 for FG, 49 for BG)
		if fg {
			return []string{"39"}
		}
		return []string{"49"}

	case ColorMode16:
		// Basic 16 colors: 30-37 for FG (normal), 90-97 for FG (bright)
		// 40-47 for BG (normal), 100-107 for BG (bright)
		idx := c.Value
		if fg {
			if idx < 8 {
				return []string{strconv.Itoa(30 + int(idx))}
			}
			return []string{strconv.Itoa(90 + int(idx) - 8)}
		}
		if idx < 8 {
			return []string{strconv.Itoa(40 + int(idx))}
		}
		return []string{strconv.Itoa(100 + int(idx) - 8)}

	case ColorMode256:
		// 256-color: 38;5;n for FG, 48;5;n for BG
		if fg {
			return []string{"38", "5", strconv.Itoa(int(c.Value))}
		}
		return []string{"48", "5", strconv.Itoa(int(c.Value))}

	case ColorModeRGB:
		// True color: 38;2;r;g;b for FG, 48;2;r;g;b for BG
		r := (c.Value >> 16) & 0xFF
		g := (c.Value >> 8) & 0xFF
		b := c.Value & 0xFF
		if fg {
			return []string{"38", "2", strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))}
		}
		return []string{"48", "2", strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))}
	}

	return nil
}

// StyleDelta returns ANSI codes to change from 'from' style to 'to' style.
// This is more efficient than always resetting and setting new style.
func StyleDelta(from, to Style) string {
	if from.Equal(to) {
		return ""
	}

	// For simplicity, we always do a full reset and set.
	// A more optimized version could compute minimal changes.
	return StyleToANSI(to)
}

// ANSIWriter helps build ANSI output efficiently.
type ANSIWriter struct {
	buf       strings.Builder
	lastStyle Style
	styleSet  bool
	lastX     int
	lastY     int
	posSet    bool
}

// NewANSIWriter creates a new ANSI writer.
func NewANSIWriter() *ANSIWriter {
	return &ANSIWriter{
		lastX: -1,
		lastY: -1,
	}
}

// MoveTo positions cursor, optimizing for sequential writes.
func (w *ANSIWriter) MoveTo(x, y int) {
	if w.posSet && w.lastY == y && w.lastX == x {
		// Cursor is already at the right position after last write
		return
	}

	if w.posSet && w.lastY == y {
		// Same line, use relative movement
		delta := x - w.lastX
		if delta > 0 && delta < 5 {
			w.buf.WriteString(CursorForward(delta))
			w.lastX = x
			return
		}
	}

	// Full position
	w.buf.WriteString(CursorTo(x, y))
	w.lastX = x
	w.lastY = y
	w.posSet = true
}

// SetStyle changes the current style.
func (w *ANSIWriter) SetStyle(s Style) {
	if w.styleSet && w.lastStyle.Equal(s) {
		return
	}
	w.buf.WriteString(StyleToANSI(s))
	w.lastStyle = s
	w.styleSet = true
}

// WriteRune writes a single rune.
func (w *ANSIWriter) WriteRune(r rune) {
	w.buf.WriteRune(r)
	w.lastX++ // Advance cursor position
}

// WriteString writes a string.
func (w *ANSIWriter) WriteString(s string) {
	w.buf.WriteString(s)
	w.lastX += len([]rune(s))
}

// Reset adds a style reset.
func (w *ANSIWriter) Reset() {
	w.buf.WriteString(ANSIReset)
	w.styleSet = false
}

// ShowCursor adds cursor show sequence.
func (w *ANSIWriter) ShowCursor() {
	w.buf.WriteString(ANSICursorShow)
}

// HideCursor adds cursor hide sequence.
func (w *ANSIWriter) HideCursor() {
	w.buf.WriteString(ANSICursorHide)
}

// String returns the accumulated output.
func (w *ANSIWriter) String() string {
	return w.buf.String()
}

// Len returns current buffer length.
func (w *ANSIWriter) Len() int {
	return w.buf.Len()
}

// Grow pre-allocates buffer capacity.
func (w *ANSIWriter) Grow(n int) {
	w.buf.Grow(n)
}
