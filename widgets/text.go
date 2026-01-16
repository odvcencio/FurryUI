package widgets

import (
	"strings"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
)

// Text is a simple text display widget.
type Text struct {
	Base
	text  string
	style backend.Style
	lines []string // Cached line splits
}

// NewText creates a new text widget.
func NewText(text string) *Text {
	return &Text{
		text:  text,
		style: backend.DefaultStyle(),
		lines: strings.Split(text, "\n"),
	}
}

// SetText updates the displayed text.
func (t *Text) SetText(text string) {
	t.text = text
	t.lines = strings.Split(text, "\n")
}

// Text returns the current text.
func (t *Text) Text() string {
	return t.text
}

// SetStyle sets the text style.
func (t *Text) SetStyle(style backend.Style) {
	t.style = style
}

// WithStyle sets the style and returns the widget for chaining.
func (t *Text) WithStyle(style backend.Style) *Text {
	t.style = style
	return t
}

// Measure returns the size needed to display the text.
func (t *Text) Measure(constraints runtime.Constraints) runtime.Size {
	// Calculate width: longest line
	maxWidth := 0
	for _, line := range t.lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Height: number of lines
	height := len(t.lines)
	if height == 0 {
		height = 1
	}

	return constraints.Constrain(runtime.Size{
		Width:  maxWidth,
		Height: height,
	})
}

// Render draws the text.
func (t *Text) Render(ctx runtime.RenderContext) {
	bounds := t.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	style := t.style

	for i, line := range t.lines {
		if i >= bounds.Height {
			break
		}
		y := bounds.Y + i
		displayLine := line
		if len(displayLine) > bounds.Width {
			displayLine = displayLine[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, y, displayLine, style)
	}
}

// Label is a single-line text widget often used for headers/labels.
type Label struct {
	Base
	text      string
	style     backend.Style
	alignment Alignment
}

// Alignment specifies text alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// NewLabel creates a new label widget.
func NewLabel(text string) *Label {
	return &Label{
		text:      text,
		style:     backend.DefaultStyle(),
		alignment: AlignLeft,
	}
}

// SetText updates the label text.
func (l *Label) SetText(text string) {
	l.text = text
}

// SetStyle sets the label style.
func (l *Label) SetStyle(style backend.Style) {
	l.style = style
}

// SetAlignment sets text alignment.
func (l *Label) SetAlignment(align Alignment) {
	l.alignment = align
}

// WithStyle sets the style and returns for chaining.
func (l *Label) WithStyle(style backend.Style) *Label {
	l.style = style
	return l
}

// WithAlignment sets alignment and returns for chaining.
func (l *Label) WithAlignment(align Alignment) *Label {
	l.alignment = align
	return l
}

// Measure returns the size needed for the label.
func (l *Label) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{
		Width:  len(l.text),
		Height: 1,
	})
}

// Render draws the label.
func (l *Label) Render(ctx runtime.RenderContext) {
	bounds := l.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	text := l.text
	if len(text) > bounds.Width {
		text = truncateString(text, bounds.Width)
	}

	// Calculate X position based on alignment
	x := bounds.X
	switch l.alignment {
	case AlignCenter:
		x = bounds.X + (bounds.Width-len(text))/2
	case AlignRight:
		x = bounds.X + bounds.Width - len(text)
	}

	ctx.Buffer.SetString(x, bounds.Y, text, l.style)
}
