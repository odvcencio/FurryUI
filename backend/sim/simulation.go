// Package sim provides a simulation backend for testing.
package sim

import (
	"strings"
	"sync"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/backend/tcell"
	"github.com/odvcencio/furry-ui/terminal"
	tcellv2 "github.com/gdamore/tcell/v2"
)

// Backend is a testable backend using tcell's simulation screen.
type Backend struct {
	*tcell.Backend
	screen tcellv2.SimulationScreen
	mu     sync.Mutex
}

// New creates a new simulation backend with the given dimensions.
func New(width, height int) *Backend {
	screen := tcellv2.NewSimulationScreen("")
	screen.SetSize(width, height)

	return &Backend{
		Backend: tcell.NewWithScreen(screen),
		screen:  screen,
	}
}

// Resize changes the simulation screen size.
func (s *Backend) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.screen.SetSize(width, height)
}

// InjectKey injects a key event into the simulation.
func (s *Backend) InjectKey(key terminal.Key, r rune) {
	s.PostEvent(terminal.KeyEvent{Key: key, Rune: r})
}

// InjectKeyRune injects a regular character keypress.
func (s *Backend) InjectKeyRune(r rune) {
	s.InjectKey(terminal.KeyRune, r)
}

// InjectKeyString injects a string as a sequence of key events.
func (s *Backend) InjectKeyString(str string) {
	for _, r := range str {
		s.InjectKeyRune(r)
	}
}

// InjectResize injects a resize event.
func (s *Backend) InjectResize(width, height int) {
	s.mu.Lock()
	s.screen.SetSize(width, height)
	s.mu.Unlock()
	s.PostEvent(terminal.ResizeEvent{Width: width, Height: height})
}

// Capture captures the current screen content as a string.
func (s *Backend) Capture() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, h := s.screen.Size()
	var lines []string

	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			mainc, comb, _, _ := s.screen.GetContent(x, y)
			if mainc == 0 {
				mainc = ' '
			}
			line.WriteRune(mainc)
			for _, c := range comb {
				line.WriteRune(c)
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// CaptureCell returns the content and style of a single cell.
func (s *Backend) CaptureCell(x, y int) (mainc rune, comb []rune, style backend.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, c, tcStyle, _ := s.screen.GetContent(x, y)
	return m, c, convertTcellStyle(tcStyle)
}

// CaptureRegion captures a rectangular region of the screen.
func (s *Backend) CaptureRegion(x, y, w, h int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lines []string
	for row := y; row < y+h; row++ {
		var line strings.Builder
		for col := x; col < x+w; col++ {
			mainc, _, _, _ := s.screen.GetContent(col, row)
			if mainc == 0 {
				mainc = ' '
			}
			line.WriteRune(mainc)
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

// FindText searches for text on the screen and returns its position.
func (s *Backend) FindText(text string) (x, y int) {
	capture := s.Capture()
	lines := strings.Split(capture, "\n")

	for row, line := range lines {
		if col := strings.Index(line, text); col >= 0 {
			return col, row
		}
	}
	return -1, -1
}

// ContainsText returns true if the text appears anywhere on screen.
func (s *Backend) ContainsText(text string) bool {
	x, y := s.FindText(text)
	return x >= 0 && y >= 0
}

// convertTcellStyle converts tcellv2.Style to backend.Style.
func convertTcellStyle(ts tcellv2.Style) backend.Style {
	fg, bg, attrs := ts.Decompose()
	style := backend.DefaultStyle().
		Foreground(convertTcellColor(fg)).
		Background(convertTcellColor(bg))

	if attrs&tcellv2.AttrBold != 0 {
		style = style.Bold(true)
	}
	if attrs&tcellv2.AttrItalic != 0 {
		style = style.Italic(true)
	}
	if attrs&tcellv2.AttrUnderline != 0 {
		style = style.Underline(true)
	}
	if attrs&tcellv2.AttrDim != 0 {
		style = style.Dim(true)
	}
	if attrs&tcellv2.AttrBlink != 0 {
		style = style.Blink(true)
	}
	if attrs&tcellv2.AttrReverse != 0 {
		style = style.Reverse(true)
	}
	if attrs&tcellv2.AttrStrikeThrough != 0 {
		style = style.StrikeThrough(true)
	}

	return style
}

// convertTcellColor converts tcellv2.Color to backend.Color.
func convertTcellColor(tc tcellv2.Color) backend.Color {
	if tc == tcellv2.ColorDefault {
		return backend.ColorDefault
	}
	if tc&tcellv2.ColorIsRGB != 0 {
		r, g, b := tc.RGB()
		return backend.ColorRGB(uint8(r), uint8(g), uint8(b))
	}
	return backend.Color(tc & 0xFF)
}

// Ensure Backend implements backend.Backend
var _ backend.Backend = (*Backend)(nil)
