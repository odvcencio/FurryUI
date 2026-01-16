package style

import (
	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/compositor"
)

// ToBackend converts a compositor.Style to backend.Style.
func ToBackend(cs compositor.Style) backend.Style {
	style := backend.DefaultStyle()

	if cs.FG.Mode == compositor.ColorModeRGB {
		r := uint8((cs.FG.Value >> 16) & 0xFF)
		g := uint8((cs.FG.Value >> 8) & 0xFF)
		b := uint8(cs.FG.Value & 0xFF)
		style = style.Foreground(backend.ColorRGB(r, g, b))
	} else if cs.FG.Mode != compositor.ColorModeDefault && cs.FG.Mode != compositor.ColorModeNone {
		style = style.Foreground(backend.Color(cs.FG.Value & 0xFF))
	}

	if cs.BG.Mode == compositor.ColorModeRGB {
		r := uint8((cs.BG.Value >> 16) & 0xFF)
		g := uint8((cs.BG.Value >> 8) & 0xFF)
		b := uint8(cs.BG.Value & 0xFF)
		style = style.Background(backend.ColorRGB(r, g, b))
	} else if cs.BG.Mode != compositor.ColorModeDefault && cs.BG.Mode != compositor.ColorModeNone {
		style = style.Background(backend.Color(cs.BG.Value & 0xFF))
	}

	if cs.Bold {
		style = style.Bold(true)
	}
	if cs.Italic {
		style = style.Italic(true)
	}
	if cs.Underline {
		style = style.Underline(true)
	}
	if cs.Dim {
		style = style.Dim(true)
	}
	if cs.Blink {
		style = style.Blink(true)
	}
	if cs.Reverse {
		style = style.Reverse(true)
	}
	if cs.Strikethrough {
		style = style.StrikeThrough(true)
	}

	return style
}
