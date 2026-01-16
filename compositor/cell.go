// Package compositor provides a flicker-free terminal rendering system.
// It maintains a virtual screen buffer and outputs only changed cells,
// inspired by modern compositor architecture.
package compositor

// ColorMode defines how a color is represented.
type ColorMode uint8

const (
	// ColorModeNone means no color (inherit default).
	ColorModeNone ColorMode = iota
	// ColorModeDefault uses terminal default color.
	ColorModeDefault
	// ColorMode16 uses basic 16 ANSI colors (0-15).
	ColorMode16
	// ColorMode256 uses extended 256 color palette.
	ColorMode256
	// ColorModeRGB uses 24-bit true color.
	ColorModeRGB
)

// Color represents a terminal color.
type Color struct {
	Mode  ColorMode
	Value uint32 // For 16/256: color index, For RGB: 0xRRGGBB
}

// Pre-defined colors for convenience.
var (
	ColorNone    = Color{Mode: ColorModeNone}
	ColorDefault = Color{Mode: ColorModeDefault}

	// Basic 16 colors
	ColorBlack   = Color{Mode: ColorMode16, Value: 0}
	ColorRed     = Color{Mode: ColorMode16, Value: 1}
	ColorGreen   = Color{Mode: ColorMode16, Value: 2}
	ColorYellow  = Color{Mode: ColorMode16, Value: 3}
	ColorBlue    = Color{Mode: ColorMode16, Value: 4}
	ColorMagenta = Color{Mode: ColorMode16, Value: 5}
	ColorCyan    = Color{Mode: ColorMode16, Value: 6}
	ColorWhite   = Color{Mode: ColorMode16, Value: 7}

	// Bright variants
	ColorBrightBlack   = Color{Mode: ColorMode16, Value: 8}
	ColorBrightRed     = Color{Mode: ColorMode16, Value: 9}
	ColorBrightGreen   = Color{Mode: ColorMode16, Value: 10}
	ColorBrightYellow  = Color{Mode: ColorMode16, Value: 11}
	ColorBrightBlue    = Color{Mode: ColorMode16, Value: 12}
	ColorBrightMagenta = Color{Mode: ColorMode16, Value: 13}
	ColorBrightCyan    = Color{Mode: ColorMode16, Value: 14}
	ColorBrightWhite   = Color{Mode: ColorMode16, Value: 15}
)

// Color256 creates a 256-palette color (0-255).
func Color256(index uint8) Color {
	return Color{Mode: ColorMode256, Value: uint32(index)}
}

// ColorRGB creates a 24-bit true color.
func RGB(r, g, b uint8) Color {
	return Color{Mode: ColorModeRGB, Value: uint32(r)<<16 | uint32(g)<<8 | uint32(b)}
}

// Hex creates a color from hex value (0xRRGGBB).
func Hex(hex uint32) Color {
	return Color{Mode: ColorModeRGB, Value: hex}
}

// Style defines visual attributes for a cell.
type Style struct {
	FG            Color
	BG            Color
	Bold          bool
	Dim           bool
	Italic        bool
	Underline     bool
	Blink         bool
	Reverse       bool
	Strikethrough bool
}

// DefaultStyle returns a style with no attributes.
func DefaultStyle() Style {
	return Style{FG: ColorDefault, BG: ColorDefault}
}

// WithFG returns a copy with foreground color set.
func (s Style) WithFG(c Color) Style {
	s.FG = c
	return s
}

// WithBG returns a copy with background color set.
func (s Style) WithBG(c Color) Style {
	s.BG = c
	return s
}

// WithBold returns a copy with bold set.
func (s Style) WithBold(b bool) Style {
	s.Bold = b
	return s
}

// WithDim returns a copy with dim set.
func (s Style) WithDim(d bool) Style {
	s.Dim = d
	return s
}

// WithItalic returns a copy with italic set.
func (s Style) WithItalic(i bool) Style {
	s.Italic = i
	return s
}

// WithUnderline returns a copy with underline set.
func (s Style) WithUnderline(u bool) Style {
	s.Underline = u
	return s
}

// WithReverse returns a copy with reverse set.
func (s Style) WithReverse(r bool) Style {
	s.Reverse = r
	return s
}

// Equal compares two styles for equality.
func (s Style) Equal(other Style) bool {
	return s.FG == other.FG &&
		s.BG == other.BG &&
		s.Bold == other.Bold &&
		s.Dim == other.Dim &&
		s.Italic == other.Italic &&
		s.Underline == other.Underline &&
		s.Blink == other.Blink &&
		s.Reverse == other.Reverse &&
		s.Strikethrough == other.Strikethrough
}

// Cell represents a single character cell on screen.
type Cell struct {
	Rune  rune
	Width uint8 // Display width (1 for most, 2 for CJK, 0 for continuation)
	Style Style
}

// EmptyCell returns a blank cell with default style.
func EmptyCell() Cell {
	return Cell{Rune: ' ', Width: 1, Style: DefaultStyle()}
}

// Empty returns true if the cell is a space with default style.
func (c Cell) Empty() bool {
	return c.Rune == ' ' && c.Width == 1 && c.Style.Equal(DefaultStyle())
}

// Equal compares two cells for equality.
func (c Cell) Equal(other Cell) bool {
	return c.Rune == other.Rune &&
		c.Width == other.Width &&
		c.Style.Equal(other.Style)
}

// NewCell creates a cell with a rune and style.
func NewCell(r rune, style Style) Cell {
	return Cell{
		Rune:  r,
		Width: 1, // Will be set correctly by screen.Set()
		Style: style,
	}
}
