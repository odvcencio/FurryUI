package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// RadioGroup manages a set of radio buttons.
type RadioGroup struct {
	selected *state.Signal[int]
	options  []*Radio
	onChange func(index int)
}

// NewRadioGroup creates an empty group.
func NewRadioGroup() *RadioGroup {
	return &RadioGroup{selected: state.NewSignal(-1)}
}

// Selected returns the selected index.
func (g *RadioGroup) Selected() int {
	if g == nil || g.selected == nil {
		return 0
	}
	return g.selected.Get()
}

// SetSelected updates the selected index.
func (g *RadioGroup) SetSelected(index int) {
	if g == nil || g.selected == nil {
		return
	}
	g.selected.Set(index)
	if g.onChange != nil {
		g.onChange(index)
	}
}

// OnChange registers a selection callback.
func (g *RadioGroup) OnChange(fn func(index int)) {
	if g == nil {
		return
	}
	g.onChange = fn
}

// Radio is a single radio option.
type Radio struct {
	FocusableBase
	accessibility.Base

	label      *state.Signal[string]
	group      *RadioGroup
	index      int
	disabled   bool
	style      backend.Style
	focusStyle backend.Style
}

// NewRadio creates a radio option and registers it with the group.
func NewRadio(label string, group *RadioGroup) *Radio {
	r := &Radio{
		label:      state.NewSignal(label),
		group:      group,
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	r.Base.Role = accessibility.RoleRadio
	r.Base.Label = label
	if group != nil {
		r.index = len(group.options)
		group.options = append(group.options, r)
	}
	r.syncState()
	return r
}

// SetDisabled updates disabled state.
func (r *Radio) SetDisabled(disabled bool) {
	if r == nil {
		return
	}
	r.disabled = disabled
	r.Base.State.Disabled = disabled
}

// Measure returns the desired size.
func (r *Radio) Measure(constraints runtime.Constraints) runtime.Size {
	label := ""
	if r.label != nil {
		label = r.label.Get()
	}
	width := 4 + len(label)
	if width < 4 {
		width = 4
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the radio.
func (r *Radio) Render(ctx runtime.RenderContext) {
	if r == nil {
		return
	}
	bounds := r.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	selected := r.isSelected()
	marker := "( )"
	if selected {
		marker = "(*)"
	}
	label := ""
	if r.label != nil {
		label = r.label.Get()
	}
	text := marker + " " + truncateString(label, bounds.Width-4)
	style := r.style
	if r.focused {
		style = r.focusStyle
	}
	if r.disabled {
		style = style.Dim(true)
	}
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
}

// HandleMessage handles selection.
func (r *Radio) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if r == nil || !r.focused || r.disabled {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		if r.group != nil {
			r.group.SetSelected(r.index)
			r.syncState()
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (r *Radio) isSelected() bool {
	if r == nil || r.group == nil {
		return false
	}
	return r.group.Selected() == r.index
}

func (r *Radio) syncState() {
	if r == nil {
		return
	}
	selected := r.isSelected()
	r.Base.State.Selected = selected
}
