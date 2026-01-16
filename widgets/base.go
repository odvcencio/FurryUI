// Package widgets provides reusable widgets for terminal UIs.
package widgets

import (
	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
)

// Base provides common functionality for widgets.
// Embed this in widget structs to get default implementations.
type Base struct {
	bounds      runtime.Rect
	focused     bool
	needsRender bool
}

// Layout stores the assigned bounds.
func (b *Base) Layout(bounds runtime.Rect) {
	if b.bounds != bounds {
		b.bounds = bounds
		b.needsRender = true
	}
}

// Bounds returns the widget's assigned bounds.
func (b *Base) Bounds() runtime.Rect {
	return b.bounds
}

// HandleMessage returns Unhandled by default.
func (b *Base) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// CanFocus returns false by default.
func (b *Base) CanFocus() bool {
	return false
}

// Focus marks the widget as focused.
func (b *Base) Focus() {
	b.focused = true
}

// Blur marks the widget as unfocused.
func (b *Base) Blur() {
	b.focused = false
}

// IsFocused returns whether the widget is focused.
func (b *Base) IsFocused() bool {
	return b.focused
}

// Invalidate marks the widget as needing a render pass.
func (b *Base) Invalidate() {
	b.needsRender = true
}

// NeedsRender reports whether the widget needs to re-render.
func (b *Base) NeedsRender() bool {
	return b.needsRender
}

// ClearInvalidation clears the render-needed flag.
func (b *Base) ClearInvalidation() {
	b.needsRender = false
}

// FocusableBase extends Base for focusable widgets.
type FocusableBase struct {
	Base
}

// CanFocus returns true for focusable widgets.
func (f *FocusableBase) CanFocus() bool {
	return true
}

// drawText is a helper to draw text with word wrapping.
func drawText(buf *runtime.Buffer, bounds runtime.Rect, text string, style backend.Style) {
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.Width
	maxY := bounds.Y + bounds.Height

	for _, r := range text {
		if r == '\n' {
			x = bounds.X
			y++
			if y >= maxY {
				break
			}
			continue
		}

		if x >= maxX {
			x = bounds.X
			y++
			if y >= maxY {
				break
			}
		}

		buf.Set(x, y, r, style)
		x++
	}
}

// fillRect fills a rectangle with a character.
func fillRect(buf *runtime.Buffer, bounds runtime.Rect, ch rune, style backend.Style) {
	buf.Fill(bounds, ch, style)
}

// truncateString truncates a string to fit within maxWidth.
// Adds "..." if truncated.
func truncateString(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

// padRight pads a string with spaces to reach the given width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	result := make([]byte, width)
	copy(result, s)
	for i := len(s); i < width; i++ {
		result[i] = ' '
	}
	return string(result)
}

// centerString centers a string within the given width.
func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	result := make([]byte, width)
	for i := range result {
		result[i] = ' '
	}
	copy(result[padding:], s)
	return string(result)
}
