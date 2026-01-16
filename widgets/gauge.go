package widgets

import (
	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
)

// GaugeStyle defines the visual appearance of a gauge.
type GaugeStyle struct {
	// Fill characters
	FillChar  rune // Filled portion (default '█')
	EmptyChar rune // Empty portion (default '░')

	// Gradient thresholds and styles (ascending order)
	// Each threshold defines the ratio at which a new color begins
	Thresholds []GaugeThreshold

	// EmptyStyle for unfilled portion
	EmptyStyle backend.Style

	// EdgeStyle for the leading edge (optional glow effect)
	EdgeStyle backend.Style
}

// GaugeThreshold defines a color breakpoint in the gradient.
type GaugeThreshold struct {
	Ratio float64       // Start ratio for this color (0.0-1.0)
	Style backend.Style // Style for this segment
}

// DefaultGaugeStyle returns a green→amber→coral gradient gauge.
func DefaultGaugeStyle(green, amber, coral, edge, empty backend.Style) GaugeStyle {
	return GaugeStyle{
		FillChar:  '█',
		EmptyChar: '░',
		Thresholds: []GaugeThreshold{
			{Ratio: 0.0, Style: green},
			{Ratio: 0.6, Style: amber},
			{Ratio: 0.85, Style: coral},
		},
		EmptyStyle: empty,
		EdgeStyle:  edge,
	}
}

// DrawGauge renders a horizontal gauge bar with gradient fill.
// ratio: 0.0-1.0 fill percentage
// width: total width in characters
// Returns the rendered gauge string and styles for each character.
func DrawGauge(buf *runtime.Buffer, x, y, width int, ratio float64, style GaugeStyle) {
	if buf == nil || width <= 0 {
		return
	}

	// Clamp ratio
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// Calculate fill width
	fill := int(float64(width)*ratio + 0.5)
	if fill > width {
		fill = width
	}

	// Default characters
	fillChar := style.FillChar
	if fillChar == 0 {
		fillChar = '█'
	}
	emptyChar := style.EmptyChar
	if emptyChar == 0 {
		emptyChar = '░'
	}

	// Draw each cell
	for i := 0; i < width; i++ {
		if i < fill {
			// Filled portion - determine color from gradient
			cellRatio := float64(i) / float64(width)
			cellStyle := styleForRatio(cellRatio, style.Thresholds)

			// Apply edge style to leading edge
			if i == fill-1 && style.EdgeStyle != (backend.Style{}) {
				cellStyle = style.EdgeStyle
			}

			buf.Set(x+i, y, fillChar, cellStyle)
		} else {
			// Empty portion
			buf.Set(x+i, y, emptyChar, style.EmptyStyle)
		}
	}
}

// DrawGaugeString renders a gauge and returns it as a string (for inline use).
func DrawGaugeString(width int, ratio float64, style GaugeStyle) string {
	if width <= 0 {
		return ""
	}

	// Clamp ratio
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// Calculate fill width
	fill := int(float64(width)*ratio + 0.5)
	if fill > width {
		fill = width
	}

	// Default characters
	fillChar := style.FillChar
	if fillChar == 0 {
		fillChar = '█'
	}
	emptyChar := style.EmptyChar
	if emptyChar == 0 {
		emptyChar = '░'
	}

	runes := make([]rune, width)
	for i := 0; i < width; i++ {
		if i < fill {
			runes[i] = fillChar
		} else {
			runes[i] = emptyChar
		}
	}
	return string(runes)
}

// styleForRatio returns the style for a given position ratio based on thresholds.
func styleForRatio(ratio float64, thresholds []GaugeThreshold) backend.Style {
	if len(thresholds) == 0 {
		return backend.DefaultStyle()
	}

	// Find the highest threshold that ratio meets or exceeds
	result := thresholds[0].Style
	for _, t := range thresholds {
		if ratio >= t.Ratio {
			result = t.Style
		}
	}
	return result
}

// GaugeSpan represents a styled segment for composite rendering.
type GaugeSpan struct {
	Text  string
	Style backend.Style
}

// DrawGaugeSpans returns gauge as styled spans for scrollback integration.
func DrawGaugeSpans(width int, ratio float64, style GaugeStyle) []GaugeSpan {
	if width <= 0 {
		return nil
	}

	// Clamp ratio
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// Calculate fill width
	fill := int(float64(width)*ratio + 0.5)
	if fill > width {
		fill = width
	}

	// Default characters
	fillChar := style.FillChar
	if fillChar == 0 {
		fillChar = '█'
	}
	emptyChar := style.EmptyChar
	if emptyChar == 0 {
		emptyChar = '░'
	}

	var spans []GaugeSpan

	// Group consecutive cells by style
	if fill > 0 {
		var currentStyle backend.Style
		var currentRunes []rune

		for i := 0; i < fill; i++ {
			cellRatio := float64(i) / float64(width)
			cellStyle := styleForRatio(cellRatio, style.Thresholds)

			// Apply edge style to leading edge
			if i == fill-1 && style.EdgeStyle != (backend.Style{}) {
				cellStyle = style.EdgeStyle
			}

			if len(currentRunes) == 0 {
				currentStyle = cellStyle
				currentRunes = append(currentRunes, fillChar)
			} else if cellStyle == currentStyle {
				currentRunes = append(currentRunes, fillChar)
			} else {
				spans = append(spans, GaugeSpan{Text: string(currentRunes), Style: currentStyle})
				currentStyle = cellStyle
				currentRunes = []rune{fillChar}
			}
		}

		if len(currentRunes) > 0 {
			spans = append(spans, GaugeSpan{Text: string(currentRunes), Style: currentStyle})
		}
	}

	// Empty portion
	if fill < width {
		empty := make([]rune, width-fill)
		for i := range empty {
			empty[i] = emptyChar
		}
		spans = append(spans, GaugeSpan{Text: string(empty), Style: style.EmptyStyle})
	}

	return spans
}
