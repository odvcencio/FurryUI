// Package terminal provides terminal event types used throughout the UI.
package terminal

// Event represents a terminal input event.
type Event interface {
	eventMarker()
}

// KeyEvent represents a key press.
type KeyEvent struct {
	Key   Key
	Rune  rune
	Alt   bool
	Ctrl  bool
	Shift bool
}

func (KeyEvent) eventMarker() {}

// ResizeEvent indicates terminal size changed.
type ResizeEvent struct {
	Width  int
	Height int
}

func (ResizeEvent) eventMarker() {}

// MouseEvent represents a mouse input event.
type MouseEvent struct {
	X, Y   int
	Button MouseButton
	Action MouseAction
	Alt    bool
	Ctrl   bool
	Shift  bool
}

func (MouseEvent) eventMarker() {}

// PasteEvent represents bracketed paste content.
type PasteEvent struct {
	Text string
}

func (PasteEvent) eventMarker() {}

// MouseButton identifies which mouse button was involved.
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
)

// MouseAction identifies what happened with the mouse.
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMove
)

// Key represents special keys.
type Key int

const (
	KeyNone Key = iota
	KeyRune     // Regular character
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEscape
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlF
	KeyCtrlP
	KeyCtrlZ
)
