package widgets

import (
	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/state"
)

// SignalLabel is a tiny label bound to a signal.
// It demonstrates managing subscriptions in Mount/Unmount with a state.Scheduler.
type SignalLabel struct {
	Base
	source    state.Readable[string]
	scheduler state.Scheduler
	subs      state.Subscriptions
	text      string
	style     backend.Style
	alignment Alignment
	mounted   bool
}

// NewSignalLabel creates a new signal-backed label.
func NewSignalLabel(source state.Readable[string], scheduler state.Scheduler) *SignalLabel {
	label := &SignalLabel{
		source:    source,
		scheduler: scheduler,
		style:     backend.DefaultStyle(),
		alignment: AlignLeft,
	}
	label.subs.SetScheduler(scheduler)
	if source != nil {
		label.text = source.Get()
	}
	return label
}

// Text returns the current label text.
func (s *SignalLabel) Text() string {
	return s.text
}

// SetStyle sets the label style.
func (s *SignalLabel) SetStyle(style backend.Style) {
	s.style = style
}

// SetAlignment sets text alignment.
func (s *SignalLabel) SetAlignment(align Alignment) {
	s.alignment = align
}

// Measure returns the size needed for the label.
func (s *SignalLabel) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{
		Width:  len(s.text),
		Height: 1,
	})
}

// Render draws the label.
func (s *SignalLabel) Render(ctx runtime.RenderContext) {
	bounds := s.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	text := s.text
	if len(text) > bounds.Width {
		text = truncateString(text, bounds.Width)
	}

	x := bounds.X
	switch s.alignment {
	case AlignCenter:
		x = bounds.X + (bounds.Width-len(text))/2
	case AlignRight:
		x = bounds.X + bounds.Width - len(text)
	}

	ctx.Buffer.SetString(x, bounds.Y, text, s.style)
}

// Mount subscribes to signal changes.
func (s *SignalLabel) Mount() {
	s.mounted = true
	s.subscribe()
}

// Unmount unsubscribes from signal changes.
func (s *SignalLabel) Unmount() {
	s.mounted = false
	s.subs.Clear()
}

func (s *SignalLabel) subscribe() {
	s.subs.Clear()
	if s.source == nil {
		s.text = ""
		return
	}
	s.text = s.source.Get()
	s.subs.Observe(s.source, s.onSignal)
}

func (s *SignalLabel) onSignal() {
	if !s.mounted || s.source == nil {
		return
	}
	s.text = s.source.Get()
}
