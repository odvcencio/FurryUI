// Package tcell provides a Backend implementation using tcell.
package tcell

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Backend implements backend.Backend using tcell.
type Backend struct {
	screen tcell.Screen

	// Bracketed paste state
	inPaste     bool
	pasteBuffer strings.Builder

	styleCache    map[backend.Style]tcell.Style
	styleCacheCap int
}

// New creates a new tcell backend.
func New() (*Backend, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	return &Backend{screen: screen}, nil
}

// NewWithScreen creates a backend with an existing tcell screen (for testing).
func NewWithScreen(screen tcell.Screen) *Backend {
	return &Backend{screen: screen}
}

// Init initializes the backend.
func (b *Backend) Init() error {
	if err := b.screen.Init(); err != nil {
		return err
	}
	b.screen.EnableMouse()
	b.screen.EnablePaste()
	return nil
}

// Fini cleans up the backend.
func (b *Backend) Fini() {
	b.screen.Fini()
}

// Size returns the terminal dimensions.
func (b *Backend) Size() (width, height int) {
	return b.screen.Size()
}

// SetContent sets a cell at position (x, y).
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {
	b.screen.SetContent(x, y, mainc, comb, b.cachedStyle(style))
}

// SetRow updates a row using the runtime row-writer fast path.
func (b *Backend) SetRow(y int, startX int, cells []backend.Cell) {
	if startX < 0 || len(cells) == 0 {
		return
	}
	x := startX
	for _, cell := range cells {
		b.screen.SetContent(x, y, cell.Rune, nil, b.cachedStyle(cell.Style))
		x++
	}
}

// SetRect updates a rectangle using row-major cells.
func (b *Backend) SetRect(x, y, width, height int, cells []backend.Cell) {
	if width <= 0 || height <= 0 {
		return
	}
	expected := width * height
	if len(cells) < expected {
		return
	}
	for row := 0; row < height; row++ {
		rowStart := row * width
		b.SetRow(y+row, x, cells[rowStart:rowStart+width])
	}
}

// Show synchronizes the buffer to the terminal.
func (b *Backend) Show() {
	b.screen.Show()
}

// Clear clears the screen.
func (b *Backend) Clear() {
	b.screen.Clear()
}

// HideCursor hides the cursor.
func (b *Backend) HideCursor() {
	b.screen.HideCursor()
}

// ShowCursor shows the cursor.
func (b *Backend) ShowCursor() {
	// tcell shows cursor when we set its position
}

// SetCursorPos sets the cursor position.
func (b *Backend) SetCursorPos(x, y int) {
	b.screen.ShowCursor(x, y)
}

// PollEvent blocks until an event is available.
func (b *Backend) PollEvent() terminal.Event {
	for {
		ev := b.screen.PollEvent()
		if ev == nil {
			return nil
		}

		// Handle bracketed paste state machine
		switch e := ev.(type) {
		case *tcell.EventPaste:
			if e.Start() {
				// Begin paste mode, buffer subsequent key events
				b.inPaste = true
				b.pasteBuffer.Reset()
				continue
			}
			if e.End() {
				// End paste mode, emit PasteEvent with accumulated content
				b.inPaste = false
				text := b.pasteBuffer.String()
				b.pasteBuffer.Reset()
				if text != "" {
					return terminal.PasteEvent{Text: text}
				}
				continue
			}

		case *tcell.EventKey:
			if b.inPaste {
				// Accumulate runes during paste
				if e.Key() == tcell.KeyRune {
					b.pasteBuffer.WriteRune(e.Rune())
				} else if e.Key() == tcell.KeyEnter {
					b.pasteBuffer.WriteRune('\n')
				} else if e.Key() == tcell.KeyTab {
					b.pasteBuffer.WriteRune('\t')
				}
				continue
			}
		}

		// Normal event handling
		return convertEvent(ev)
	}
}

// PostEvent injects an event into the queue.
func (b *Backend) PostEvent(ev terminal.Event) error {
	tev := reverseConvertEvent(ev)
	if tev != nil {
		return b.screen.PostEvent(tev)
	}
	return nil
}

// Beep emits an audible bell.
func (b *Backend) Beep() {
	b.screen.Beep()
}

// Sync forces a full redraw.
func (b *Backend) Sync() {
	b.screen.Sync()
}

const defaultStyleCacheCap = 256

func (b *Backend) cachedStyle(s backend.Style) tcell.Style {
	if b.styleCache == nil {
		b.styleCacheCap = defaultStyleCacheCap
		b.styleCache = make(map[backend.Style]tcell.Style, b.styleCacheCap)
	}
	if cached, ok := b.styleCache[s]; ok {
		return cached
	}
	if b.styleCacheCap <= 0 {
		b.styleCacheCap = defaultStyleCacheCap
	}
	if len(b.styleCache) >= b.styleCacheCap {
		b.styleCache = make(map[backend.Style]tcell.Style, b.styleCacheCap)
	}
	style := convertStyle(s)
	b.styleCache[s] = style
	return style
}

// convertStyle converts backend.Style to tcell.Style.
func convertStyle(s backend.Style) tcell.Style {
	fg, bg, attrs := s.Decompose()
	style := tcell.StyleDefault.
		Foreground(convertColor(fg)).
		Background(convertColor(bg))

	if attrs&backend.AttrBold != 0 {
		style = style.Bold(true)
	}
	if attrs&backend.AttrItalic != 0 {
		style = style.Italic(true)
	}
	if attrs&backend.AttrUnderline != 0 {
		style = style.Underline(true)
	}
	if attrs&backend.AttrDim != 0 {
		style = style.Dim(true)
	}
	if attrs&backend.AttrBlink != 0 {
		style = style.Blink(true)
	}
	if attrs&backend.AttrReverse != 0 {
		style = style.Reverse(true)
	}
	if attrs&backend.AttrStrikeThrough != 0 {
		style = style.StrikeThrough(true)
	}

	return style
}

// convertColor converts backend.Color to tcell.Color.
func convertColor(c backend.Color) tcell.Color {
	if c == backend.ColorDefault {
		return tcell.ColorDefault
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		return tcell.NewRGBColor(int32(r), int32(g), int32(b))
	}
	return tcell.PaletteColor(int(c))
}

// convertEvent converts a tcell event to terminal.Event.
func convertEvent(ev tcell.Event) terminal.Event {
	switch e := ev.(type) {
	case *tcell.EventKey:
		return terminal.KeyEvent{
			Key:   convertKey(e.Key()),
			Rune:  e.Rune(),
			Alt:   e.Modifiers()&tcell.ModAlt != 0,
			Ctrl:  e.Modifiers()&tcell.ModCtrl != 0,
			Shift: e.Modifiers()&tcell.ModShift != 0,
		}
	case *tcell.EventResize:
		w, h := e.Size()
		return terminal.ResizeEvent{Width: w, Height: h}
	case *tcell.EventMouse:
		x, y := e.Position()
		mods := e.Modifiers()
		return terminal.MouseEvent{
			X:      x,
			Y:      y,
			Button: convertMouseButton(e.Buttons()),
			Action: convertMouseAction(e.Buttons()),
			Alt:    mods&tcell.ModAlt != 0,
			Ctrl:   mods&tcell.ModCtrl != 0,
			Shift:  mods&tcell.ModShift != 0,
		}
	default:
		return nil
	}
}

// convertKey converts tcell.Key to terminal.Key.
func convertKey(k tcell.Key) terminal.Key {
	switch k {
	case tcell.KeyRune:
		return terminal.KeyRune
	case tcell.KeyUp:
		return terminal.KeyUp
	case tcell.KeyDown:
		return terminal.KeyDown
	case tcell.KeyRight:
		return terminal.KeyRight
	case tcell.KeyLeft:
		return terminal.KeyLeft
	case tcell.KeyPgUp:
		return terminal.KeyPageUp
	case tcell.KeyPgDn:
		return terminal.KeyPageDown
	case tcell.KeyHome:
		return terminal.KeyHome
	case tcell.KeyEnd:
		return terminal.KeyEnd
	case tcell.KeyInsert:
		return terminal.KeyInsert
	case tcell.KeyDelete:
		return terminal.KeyDelete
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return terminal.KeyBackspace
	case tcell.KeyTab:
		return terminal.KeyTab
	case tcell.KeyEnter:
		return terminal.KeyEnter
	case tcell.KeyEscape:
		return terminal.KeyEscape
	case tcell.KeyCtrlB:
		return terminal.KeyCtrlB
	case tcell.KeyCtrlC:
		return terminal.KeyCtrlC
	case tcell.KeyCtrlD:
		return terminal.KeyCtrlD
	case tcell.KeyCtrlF:
		return terminal.KeyCtrlF
	case tcell.KeyCtrlP:
		return terminal.KeyCtrlP
	case tcell.KeyCtrlV:
		return terminal.KeyCtrlV
	case tcell.KeyCtrlX:
		return terminal.KeyCtrlX
	case tcell.KeyCtrlZ:
		return terminal.KeyCtrlZ
	case tcell.KeyF1:
		return terminal.KeyF1
	case tcell.KeyF2:
		return terminal.KeyF2
	case tcell.KeyF3:
		return terminal.KeyF3
	case tcell.KeyF4:
		return terminal.KeyF4
	case tcell.KeyF5:
		return terminal.KeyF5
	case tcell.KeyF6:
		return terminal.KeyF6
	case tcell.KeyF7:
		return terminal.KeyF7
	case tcell.KeyF8:
		return terminal.KeyF8
	case tcell.KeyF9:
		return terminal.KeyF9
	case tcell.KeyF10:
		return terminal.KeyF10
	case tcell.KeyF11:
		return terminal.KeyF11
	case tcell.KeyF12:
		return terminal.KeyF12
	default:
		return terminal.KeyNone
	}
}

// convertMouseButton converts tcell button mask to terminal.MouseButton.
func convertMouseButton(buttons tcell.ButtonMask) terminal.MouseButton {
	switch {
	case buttons&tcell.WheelUp != 0:
		return terminal.MouseWheelUp
	case buttons&tcell.WheelDown != 0:
		return terminal.MouseWheelDown
	case buttons&tcell.Button1 != 0:
		return terminal.MouseLeft
	case buttons&tcell.Button2 != 0:
		return terminal.MouseMiddle
	case buttons&tcell.Button3 != 0:
		return terminal.MouseRight
	default:
		return terminal.MouseNone
	}
}

// convertMouseAction determines the mouse action from button state.
func convertMouseAction(buttons tcell.ButtonMask) terminal.MouseAction {
	if buttons == tcell.ButtonNone {
		return terminal.MouseRelease
	}
	if buttons&(tcell.WheelUp|tcell.WheelDown) != 0 {
		return terminal.MousePress // Wheel events are instantaneous
	}
	return terminal.MousePress
}

// reverseConvertEvent converts terminal.Event to tcell.Event for PostEvent.
func reverseConvertEvent(ev terminal.Event) tcell.Event {
	switch e := ev.(type) {
	case terminal.ResizeEvent:
		return tcell.NewEventResize(e.Width, e.Height)
	default:
		return nil
	}
}

// Ensure Backend implements backend.Backend
var _ backend.Backend = (*Backend)(nil)
