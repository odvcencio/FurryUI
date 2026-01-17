package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// SelectOption represents a selectable option.
type SelectOption struct {
	Label    string
	Value    any
	Disabled bool
}

// Select is a dropdown-like selector (inline).
type Select struct {
	FocusableBase
	accessibility.Base

	options    []SelectOption
	selected   int
	onChange   func(option SelectOption)
	style      backend.Style
	focusStyle backend.Style
}

// NewSelect creates a select widget.
func NewSelect(options ...SelectOption) *Select {
	s := &Select{
		options:    options,
		selected:   0,
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	s.Base.Role = accessibility.RoleTextbox
	s.syncState()
	return s
}

// SetOnChange sets the change handler.
func (s *Select) SetOnChange(fn func(option SelectOption)) {
	if s == nil {
		return
	}
	s.onChange = fn
}

// Selected returns the current selection index.
func (s *Select) Selected() int {
	if s == nil {
		return 0
	}
	return s.selected
}

// SelectedOption returns the current option.
func (s *Select) SelectedOption() (SelectOption, bool) {
	if s == nil || s.selected < 0 || s.selected >= len(s.options) {
		return SelectOption{}, false
	}
	return s.options[s.selected], true
}

// SetSelected updates the selected index.
func (s *Select) SetSelected(index int) {
	if s == nil || index < 0 || index >= len(s.options) {
		return
	}
	if s.options[index].Disabled {
		return
	}
	s.selected = index
	s.syncState()
	if s.onChange != nil {
		s.onChange(s.options[index])
	}
}

// Measure returns the desired size.
func (s *Select) Measure(constraints runtime.Constraints) runtime.Size {
	label := s.currentLabel()
	width := len(label) + 4
	if width < 6 {
		width = 6
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the select.
func (s *Select) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	label := s.currentLabel()
	text := "[" + truncateString(label, bounds.Width-4) + " v]"
	style := s.style
	if s.focused {
		style = s.focusStyle
	}
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
}

// HandleMessage changes selection.
func (s *Select) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || !s.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp, terminal.KeyLeft:
		if s.moveSelection(-1) {
			return runtime.Handled()
		}
	case terminal.KeyDown, terminal.KeyRight:
		if s.moveSelection(1) {
			return runtime.Handled()
		}
	case terminal.KeyHome:
		s.SetSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		s.SetSelected(len(s.options) - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (s *Select) moveSelection(delta int) bool {
	if s == nil || len(s.options) == 0 {
		return false
	}
	index := s.selected
	for i := 0; i < len(s.options); i++ {
		index += delta
		if index < 0 {
			index = len(s.options) - 1
		} else if index >= len(s.options) {
			index = 0
		}
		if !s.options[index].Disabled {
			s.SetSelected(index)
			return true
		}
	}
	return false
}

func (s *Select) currentLabel() string {
	if opt, ok := s.SelectedOption(); ok {
		return opt.Label
	}
	return ""
}

func (s *Select) syncState() {
	if s == nil {
		return
	}
	label := s.currentLabel()
	s.Base.Label = label
}
