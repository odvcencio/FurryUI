package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// TextArea is a multi-line text input widget.
type TextArea struct {
	FocusableBase
	accessibility.Base

	text       []rune
	cursor     int
	scrollY    int
	style      backend.Style
	focusStyle backend.Style
	onChange   func(text string)
	services   runtime.Services
}

// NewTextArea creates a new text area.
func NewTextArea() *TextArea {
	ta := &TextArea{
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	ta.Base.Role = accessibility.RoleTextbox
	return ta
}

// Bind attaches app services.
func (t *TextArea) Bind(services runtime.Services) {
	t.services = services
}

// Unbind releases app services.
func (t *TextArea) Unbind() {
	t.services = runtime.Services{}
}

// SetText sets the text and moves the cursor to the end.
func (t *TextArea) SetText(text string) {
	if t == nil {
		return
	}
	t.text = []rune(text)
	t.cursor = len(t.text)
	t.syncValue()
}

// Text returns the current text.
func (t *TextArea) Text() string {
	if t == nil {
		return ""
	}
	return string(t.text)
}

// OnChange registers a callback for text changes.
func (t *TextArea) OnChange(fn func(text string)) {
	if t == nil {
		return
	}
	t.onChange = fn
}

// Measure returns the desired size.
func (t *TextArea) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight})
}

// Render draws the text area.
func (t *TextArea) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	bounds := t.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := t.style
	if t.focused {
		style = t.focusStyle
	}
	ctx.Buffer.Fill(bounds, ' ', style)

	lineStarts, lineLengths := t.lineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	t.scrollY = min(max(t.scrollY, 0), max(0, len(lineStarts)-1))
	if line < t.scrollY {
		t.scrollY = line
	} else if line >= t.scrollY+bounds.Height {
		t.scrollY = line - bounds.Height + 1
	}
	scrollX := 0
	if col >= bounds.Width {
		scrollX = col - bounds.Width + 1
	}

	for row := 0; row < bounds.Height; row++ {
		lineIndex := t.scrollY + row
		if lineIndex >= len(lineStarts) {
			break
		}
		lineText := t.lineText(lineIndex, lineStarts, lineLengths)
		if scrollX < len(lineText) {
			lineText = lineText[scrollX:]
		} else {
			lineText = ""
		}
		if len(lineText) > bounds.Width {
			lineText = lineText[:bounds.Width]
		}
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, lineText, style)
	}

	if t.focused {
		cursorRow := line - t.scrollY
		cursorCol := col - scrollX
		if cursorRow >= 0 && cursorRow < bounds.Height && cursorCol >= 0 && cursorCol < bounds.Width {
			cursorX := bounds.X + cursorCol
			cursorY := bounds.Y + cursorRow
			ch := ' '
			lineText := t.lineText(line, lineStarts, lineLengths)
			if col < len(lineText) {
				ch = rune(lineText[col])
			}
			ctx.Buffer.Set(cursorX, cursorY, ch, style.Reverse(true))
		}
	}
}

// HandleMessage processes keyboard input.
func (t *TextArea) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlC:
		if t.copyToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlX:
		if t.cutToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlV:
		if t.pasteFromClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		t.insertRune('\n')
		return runtime.Handled()
	case terminal.KeyBackspace:
		if t.cursor > 0 {
			t.deleteRune(t.cursor - 1)
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if t.cursor < len(t.text) {
			t.deleteRune(t.cursor)
		}
		return runtime.Handled()
	case terminal.KeyLeft:
		if t.cursor > 0 {
			t.cursor--
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if t.cursor < len(t.text) {
			t.cursor++
		}
		return runtime.Handled()
	case terminal.KeyUp:
		t.moveVertical(-1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.moveVertical(1)
		return runtime.Handled()
	case terminal.KeyHome:
		t.moveLineBoundary(true)
		return runtime.Handled()
	case terminal.KeyEnd:
		t.moveLineBoundary(false)
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune != 0 {
			t.insertRune(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *TextArea) insertRune(r rune) {
	t.text = append(t.text[:t.cursor], append([]rune{r}, t.text[t.cursor:]...)...)
	t.cursor++
	t.syncValue()
}

func (t *TextArea) insertText(text string) {
	if text == "" {
		return
	}
	runes := []rune(text)
	t.text = append(t.text[:t.cursor], append(runes, t.text[t.cursor:]...)...)
	t.cursor += len(runes)
	t.syncValue()
}

func (t *TextArea) deleteRune(index int) {
	if index < 0 || index >= len(t.text) {
		return
	}
	t.text = append(t.text[:index], t.text[index+1:]...)
	if t.cursor > index {
		t.cursor--
	}
	t.syncValue()
}

func (t *TextArea) moveVertical(delta int) {
	lineStarts, lineLengths := t.lineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	target := line + delta
	if target < 0 || target >= len(lineStarts) {
		return
	}
	targetLen := lineLengths[target]
	if col > targetLen {
		col = targetLen
	}
	t.cursor = lineStarts[target] + col
}

func (t *TextArea) moveLineBoundary(start bool) {
	lineStarts, lineLengths := t.lineMeta()
	line, _ := t.cursorLineCol(lineStarts, lineLengths)
	if line < 0 || line >= len(lineStarts) {
		return
	}
	if start {
		t.cursor = lineStarts[line]
		return
	}
	t.cursor = lineStarts[line] + lineLengths[line]
}

func (t *TextArea) lineMeta() ([]int, []int) {
	if t == nil {
		return []int{0}, []int{0}
	}
	starts := []int{0}
	var lengths []int
	for i, r := range t.text {
		if r == '\n' {
			lengths = append(lengths, i-starts[len(starts)-1])
			starts = append(starts, i+1)
		}
	}
	lastStart := starts[len(starts)-1]
	lengths = append(lengths, len(t.text)-lastStart)
	return starts, lengths
}

func (t *TextArea) lineText(line int, starts []int, lengths []int) string {
	if line < 0 || line >= len(starts) {
		return ""
	}
	start := starts[line]
	end := start + lengths[line]
	if start > len(t.text) || end > len(t.text) || start > end {
		return ""
	}
	return string(t.text[start:end])
}

func (t *TextArea) cursorLineCol(starts []int, lengths []int) (int, int) {
	if len(starts) == 0 {
		return 0, 0
	}
	for i, start := range starts {
		end := start + lengths[i]
		if t.cursor <= end {
			return i, t.cursor - start
		}
	}
	last := len(starts) - 1
	return last, lengths[last]
}

func (t *TextArea) syncValue() {
	t.Base.Label = t.Text()
	if t.onChange != nil {
		t.onChange(t.Text())
	}
}

// ClipboardCopy returns the current text.
func (t *TextArea) ClipboardCopy() (string, bool) {
	if t == nil {
		return "", false
	}
	return t.Text(), true
}

// ClipboardCut returns the current text and clears it.
func (t *TextArea) ClipboardCut() (string, bool) {
	if t == nil {
		return "", false
	}
	text := t.Text()
	t.text = nil
	t.cursor = 0
	t.scrollY = 0
	t.syncValue()
	return text, true
}

// ClipboardPaste inserts text at the cursor.
func (t *TextArea) ClipboardPaste(text string) bool {
	if t == nil || text == "" {
		return false
	}
	t.insertText(text)
	return true
}

func (t *TextArea) copyToClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := t.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (t *TextArea) cutToClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := t.ClipboardCut()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (t *TextArea) pasteFromClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, err := cb.Read()
	if err != nil || text == "" {
		return false
	}
	return t.ClipboardPaste(text)
}

var _ clipboard.Target = (*TextArea)(nil)
