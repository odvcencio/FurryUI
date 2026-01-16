package widgets

import (
	"strings"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

// Input is a text input widget with cursor support.
type Input struct {
	FocusableBase

	text       strings.Builder
	cursorPos  int
	style      backend.Style
	focusStyle backend.Style
	placeholder string

	// Callbacks
	onSubmit func(text string)
	onChange func(text string)
}

// NewInput creates a new input widget.
func NewInput() *Input {
	return &Input{
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Bold(true),
	}
}

// SetPlaceholder sets the placeholder text shown when empty.
func (i *Input) SetPlaceholder(text string) {
	i.placeholder = text
}

// SetStyle sets the normal style.
func (i *Input) SetStyle(style backend.Style) {
	i.style = style
}

// SetFocusStyle sets the focused style.
func (i *Input) SetFocusStyle(style backend.Style) {
	i.focusStyle = style
}

// OnSubmit sets the callback for when Enter is pressed.
func (i *Input) OnSubmit(fn func(text string)) {
	i.onSubmit = fn
}

// OnChange sets the callback for when text changes.
func (i *Input) OnChange(fn func(text string)) {
	i.onChange = fn
}

// Text returns the current input text.
func (i *Input) Text() string {
	return i.text.String()
}

// SetText sets the input text and moves cursor to end.
func (i *Input) SetText(text string) {
	i.text.Reset()
	i.text.WriteString(text)
	i.cursorPos = i.text.Len()
}

// Clear clears the input text.
func (i *Input) Clear() {
	i.text.Reset()
	i.cursorPos = 0
}

// CursorPos returns the current cursor position.
func (i *Input) CursorPos() int {
	return i.cursorPos
}

// Measure returns the size needed for the input.
func (i *Input) Measure(constraints runtime.Constraints) runtime.Size {
	// Input is typically 1 line tall, fills available width
	return runtime.Size{
		Width:  constraints.MaxWidth,
		Height: 1,
	}
}

// Render draws the input field.
func (i *Input) Render(ctx runtime.RenderContext) {
	bounds := i.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	style := i.style
	if i.focused {
		style = i.focusStyle
	}

	// Clear the input area
	ctx.Buffer.Fill(bounds, ' ', style)

	text := i.text.String()

	// Show placeholder if empty and not focused
	if text == "" && !i.focused && i.placeholder != "" {
		placeholderStyle := style.Dim(true)
		display := i.placeholder
		if len(display) > bounds.Width {
			display = display[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y, display, placeholderStyle)
		return
	}

	// Calculate visible portion of text
	// Scroll so cursor is always visible
	visibleStart := 0
	if i.cursorPos >= bounds.Width {
		visibleStart = i.cursorPos - bounds.Width + 1
	}

	visibleEnd := visibleStart + bounds.Width
	if visibleEnd > len(text) {
		visibleEnd = len(text)
	}

	visible := ""
	if visibleStart < len(text) {
		visible = text[visibleStart:visibleEnd]
	}

	// Draw text
	ctx.Buffer.SetString(bounds.X, bounds.Y, visible, style)

	// Draw cursor if focused (by inverting the cell)
	if i.focused {
		cursorX := bounds.X + i.cursorPos - visibleStart
		if cursorX >= bounds.X && cursorX < bounds.X+bounds.Width {
			var cursorChar rune = ' '
			if i.cursorPos < len(text) {
				cursorChar = rune(text[i.cursorPos])
			}
			cursorStyle := style.Reverse(true)
			ctx.Buffer.Set(cursorX, bounds.Y, cursorChar, cursorStyle)
		}
	}
}

// HandleMessage processes keyboard input.
func (i *Input) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if !i.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEnter:
		if i.onSubmit != nil {
			text := i.text.String()
			i.onSubmit(text)
		}
		return runtime.WithCommand(runtime.Submit{Text: i.text.String()})

	case terminal.KeyBackspace:
		if i.cursorPos > 0 {
			text := i.text.String()
			i.text.Reset()
			i.text.WriteString(text[:i.cursorPos-1])
			i.text.WriteString(text[i.cursorPos:])
			i.cursorPos--
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyDelete:
		text := i.text.String()
		if i.cursorPos < len(text) {
			i.text.Reset()
			i.text.WriteString(text[:i.cursorPos])
			i.text.WriteString(text[i.cursorPos+1:])
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if key.Ctrl {
			// Word left
			i.cursorPos = i.wordBoundaryLeft()
		} else if i.cursorPos > 0 {
			i.cursorPos--
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if key.Ctrl {
			// Word right
			i.cursorPos = i.wordBoundaryRight()
		} else if i.cursorPos < i.text.Len() {
			i.cursorPos++
		}
		return runtime.Handled()

	case terminal.KeyHome:
		i.cursorPos = 0
		return runtime.Handled()

	case terminal.KeyEnd:
		i.cursorPos = i.text.Len()
		return runtime.Handled()

	case terminal.KeyRune:
		// Insert character
		text := i.text.String()
		i.text.Reset()
		i.text.WriteString(text[:i.cursorPos])
		i.text.WriteRune(key.Rune)
		i.text.WriteString(text[i.cursorPos:])
		i.cursorPos++
		i.notifyChange()
		return runtime.Handled()

	case terminal.KeyTab:
		// Tab might be focus navigation
		if key.Shift {
			return runtime.WithCommand(runtime.FocusPrev{})
		}
		return runtime.WithCommand(runtime.FocusNext{})

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

func (i *Input) notifyChange() {
	if i.onChange != nil {
		i.onChange(i.text.String())
	}
}

func (i *Input) wordBoundaryLeft() int {
	text := i.text.String()
	pos := i.cursorPos - 1

	// Skip whitespace
	for pos > 0 && text[pos] == ' ' {
		pos--
	}
	// Skip word characters
	for pos > 0 && text[pos-1] != ' ' {
		pos--
	}
	return pos
}

func (i *Input) wordBoundaryRight() int {
	text := i.text.String()
	pos := i.cursorPos

	// Skip word characters
	for pos < len(text) && text[pos] != ' ' {
		pos++
	}
	// Skip whitespace
	for pos < len(text) && text[pos] == ' ' {
		pos++
	}
	return pos
}

// MultilineInput is a text input that supports multiple lines.
type MultilineInput struct {
	FocusableBase

	lines     []string
	cursorX   int
	cursorY   int
	scrollY   int // First visible line
	style     backend.Style
	focusStyle backend.Style

	onSubmit func(text string)
	onChange func(text string)
}

// NewMultilineInput creates a new multiline input widget.
func NewMultilineInput() *MultilineInput {
	return &MultilineInput{
		lines:      []string{""},
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle(),
	}
}

// Text returns the full text content.
func (m *MultilineInput) Text() string {
	return strings.Join(m.lines, "\n")
}

// SetText sets the content.
func (m *MultilineInput) SetText(text string) {
	m.lines = strings.Split(text, "\n")
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	m.cursorY = len(m.lines) - 1
	m.cursorX = len(m.lines[m.cursorY])
}

// Clear clears all content.
func (m *MultilineInput) Clear() {
	m.lines = []string{""}
	m.cursorX = 0
	m.cursorY = 0
	m.scrollY = 0
}

// OnSubmit sets the callback (Ctrl+Enter to submit).
func (m *MultilineInput) OnSubmit(fn func(text string)) {
	m.onSubmit = fn
}

// Measure returns the preferred size.
func (m *MultilineInput) Measure(constraints runtime.Constraints) runtime.Size {
	// Prefer to be at least 3 lines tall, up to content or max
	height := len(m.lines)
	if height < 3 {
		height = 3
	}
	return constraints.Constrain(runtime.Size{
		Width:  constraints.MaxWidth,
		Height: height,
	})
}

// Render draws the multiline input.
func (m *MultilineInput) Render(ctx runtime.RenderContext) {
	bounds := m.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	style := m.style
	if m.focused {
		style = m.focusStyle
	}

	// Clear area
	ctx.Buffer.Fill(bounds, ' ', style)

	// Draw visible lines
	for i := 0; i < bounds.Height; i++ {
		lineIdx := m.scrollY + i
		if lineIdx >= len(m.lines) {
			break
		}

		line := m.lines[lineIdx]
		if len(line) > bounds.Width {
			line = line[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+i, line, style)
	}

	// Draw cursor
	if m.focused {
		cursorScreenY := m.cursorY - m.scrollY
		if cursorScreenY >= 0 && cursorScreenY < bounds.Height {
			cursorX := bounds.X + m.cursorX
			if cursorX >= bounds.X && cursorX < bounds.X+bounds.Width {
				var ch rune = ' '
				if m.cursorY < len(m.lines) && m.cursorX < len(m.lines[m.cursorY]) {
					ch = rune(m.lines[m.cursorY][m.cursorX])
				}
				ctx.Buffer.Set(cursorX, bounds.Y+cursorScreenY, ch, style.Reverse(true))
			}
		}
	}
}

// HandleMessage processes input for multiline editing.
func (m *MultilineInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if !m.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEnter:
		if key.Ctrl && m.onSubmit != nil {
			m.onSubmit(m.Text())
			return runtime.WithCommand(runtime.Submit{Text: m.Text()})
		}
		// Insert newline
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX]
		newLine := line[m.cursorX:]
		m.lines = append(m.lines[:m.cursorY+1], append([]string{newLine}, m.lines[m.cursorY+1:]...)...)
		m.cursorY++
		m.cursorX = 0
		m.ensureCursorVisible()
		return runtime.Handled()

	case terminal.KeyBackspace:
		if m.cursorX > 0 {
			line := m.lines[m.cursorY]
			m.lines[m.cursorY] = line[:m.cursorX-1] + line[m.cursorX:]
			m.cursorX--
		} else if m.cursorY > 0 {
			// Join with previous line
			prevLine := m.lines[m.cursorY-1]
			m.cursorX = len(prevLine)
			m.lines[m.cursorY-1] = prevLine + m.lines[m.cursorY]
			m.lines = append(m.lines[:m.cursorY], m.lines[m.cursorY+1:]...)
			m.cursorY--
		}
		return runtime.Handled()

	case terminal.KeyUp:
		if m.cursorY > 0 {
			m.cursorY--
			if m.cursorX > len(m.lines[m.cursorY]) {
				m.cursorX = len(m.lines[m.cursorY])
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			if m.cursorX > len(m.lines[m.cursorY]) {
				m.cursorX = len(m.lines[m.cursorY])
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if m.cursorX > 0 {
			m.cursorX--
		} else if m.cursorY > 0 {
			m.cursorY--
			m.cursorX = len(m.lines[m.cursorY])
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if m.cursorX < len(m.lines[m.cursorY]) {
			m.cursorX++
		} else if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.cursorX = 0
		}
		return runtime.Handled()

	case terminal.KeyRune:
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX] + string(key.Rune) + line[m.cursorX:]
		m.cursorX++
		return runtime.Handled()

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

func (m *MultilineInput) ensureCursorVisible() {
	if m.cursorY < m.scrollY {
		m.scrollY = m.cursorY
	} else if m.cursorY >= m.scrollY+m.bounds.Height {
		m.scrollY = m.cursorY - m.bounds.Height + 1
	}
}
