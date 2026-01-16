package backend

// Color represents a terminal color.
// Values 0-255 are palette colors, values >= 256 are true colors.
type Color int32

// Color constants
const (
	ColorDefault Color = -1
	ColorBlack   Color = 0
	ColorRed     Color = 1
	ColorGreen   Color = 2
	ColorYellow  Color = 3
	ColorBlue    Color = 4
	ColorMagenta Color = 5
	ColorCyan    Color = 6
	ColorWhite   Color = 7

	// Bright variants
	ColorBrightBlack   Color = 8
	ColorBrightRed     Color = 9
	ColorBrightGreen   Color = 10
	ColorBrightYellow  Color = 11
	ColorBrightBlue    Color = 12
	ColorBrightMagenta Color = 13
	ColorBrightCyan    Color = 14
	ColorBrightWhite   Color = 15
)

// ColorRGB creates a true color from RGB components.
func ColorRGB(r, g, b uint8) Color {
	return Color(int32(r)<<16 | int32(g)<<8 | int32(b) | 0x01000000)
}

// IsRGB returns true if this is a true color (not palette).
func (c Color) IsRGB() bool {
	return c&0x01000000 != 0
}

// RGB returns the red, green, blue components of an RGB color.
// Returns 0, 0, 0 for non-RGB colors.
func (c Color) RGB() (r, g, b uint8) {
	if !c.IsRGB() {
		return 0, 0, 0
	}
	return uint8((c >> 16) & 0xFF), uint8((c >> 8) & 0xFF), uint8(c & 0xFF)
}

// AttrMask represents text attributes.
type AttrMask uint32

// Attribute flags
const (
	AttrBold AttrMask = 1 << iota
	AttrBlink
	AttrReverse
	AttrUnderline
	AttrDim
	AttrItalic
	AttrStrikeThrough
)

// Style combines foreground, background colors and attributes.
type Style struct {
	fg    Color
	bg    Color
	attrs AttrMask
}

// DefaultStyle returns the default style (default colors, no attributes).
func DefaultStyle() Style {
	return Style{fg: ColorDefault, bg: ColorDefault}
}

// Foreground sets the foreground color.
func (s Style) Foreground(c Color) Style {
	s.fg = c
	return s
}

// Background sets the background color.
func (s Style) Background(c Color) Style {
	s.bg = c
	return s
}

// Bold enables or disables bold.
func (s Style) Bold(on bool) Style {
	if on {
		s.attrs |= AttrBold
	} else {
		s.attrs &^= AttrBold
	}
	return s
}

// Italic enables or disables italic.
func (s Style) Italic(on bool) Style {
	if on {
		s.attrs |= AttrItalic
	} else {
		s.attrs &^= AttrItalic
	}
	return s
}

// Dim enables or disables dim.
func (s Style) Dim(on bool) Style {
	if on {
		s.attrs |= AttrDim
	} else {
		s.attrs &^= AttrDim
	}
	return s
}

// Underline enables or disables underline.
func (s Style) Underline(on bool) Style {
	if on {
		s.attrs |= AttrUnderline
	} else {
		s.attrs &^= AttrUnderline
	}
	return s
}

// Reverse enables or disables reverse video.
func (s Style) Reverse(on bool) Style {
	if on {
		s.attrs |= AttrReverse
	} else {
		s.attrs &^= AttrReverse
	}
	return s
}

// Blink enables or disables blink.
func (s Style) Blink(on bool) Style {
	if on {
		s.attrs |= AttrBlink
	} else {
		s.attrs &^= AttrBlink
	}
	return s
}

// StrikeThrough enables or disables strikethrough.
func (s Style) StrikeThrough(on bool) Style {
	if on {
		s.attrs |= AttrStrikeThrough
	} else {
		s.attrs &^= AttrStrikeThrough
	}
	return s
}

// Attributes returns all attributes.
func (s Style) Attributes() AttrMask {
	return s.attrs
}

// FG returns the foreground color.
func (s Style) FG() Color {
	return s.fg
}

// BG returns the background color.
func (s Style) BG() Color {
	return s.bg
}

// Decompose returns the foreground, background, and attributes.
func (s Style) Decompose() (fg, bg Color, attrs AttrMask) {
	return s.fg, s.bg, s.attrs
}
