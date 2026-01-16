package widgets

import (
	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
)

// Panel is a container widget with optional border and background.
type Panel struct {
	Base
	child       runtime.Widget
	style       backend.Style
	borderStyle backend.Style
	hasBorder   bool
	title       string
}

// NewPanel creates a new panel widget.
func NewPanel(child runtime.Widget) *Panel {
	return &Panel{
		child:       child,
		style:       backend.DefaultStyle(),
		borderStyle: backend.DefaultStyle(),
		hasBorder:   false,
	}
}

// SetStyle sets the panel background style.
func (p *Panel) SetStyle(style backend.Style) {
	p.style = style
}

// WithStyle sets the style and returns for chaining.
func (p *Panel) WithStyle(style backend.Style) *Panel {
	p.style = style
	return p
}

// SetBorder enables or disables the border.
func (p *Panel) SetBorder(enabled bool) {
	p.hasBorder = enabled
}

// WithBorder enables border and returns for chaining.
func (p *Panel) WithBorder(style backend.Style) *Panel {
	p.hasBorder = true
	p.borderStyle = style
	return p
}

// SetTitle sets the panel title (shown in border).
func (p *Panel) SetTitle(title string) {
	p.title = title
}

// WithTitle sets title and returns for chaining.
func (p *Panel) WithTitle(title string) *Panel {
	p.title = title
	return p
}

// Measure returns the size needed for the panel.
func (p *Panel) Measure(constraints runtime.Constraints) runtime.Size {
	borderSize := 0
	if p.hasBorder {
		borderSize = 2 // 1 on each side
	}

	if p.child == nil {
		return constraints.Constrain(runtime.Size{
			Width:  borderSize,
			Height: borderSize,
		})
	}

	// Measure child with reduced constraints for border
	childConstraints := runtime.Constraints{
		MinWidth:  max(0, constraints.MinWidth-borderSize),
		MaxWidth:  max(0, constraints.MaxWidth-borderSize),
		MinHeight: max(0, constraints.MinHeight-borderSize),
		MaxHeight: max(0, constraints.MaxHeight-borderSize),
	}

	childSize := p.child.Measure(childConstraints)
	return runtime.Size{
		Width:  childSize.Width + borderSize,
		Height: childSize.Height + borderSize,
	}
}

// Layout positions the panel and its child.
func (p *Panel) Layout(bounds runtime.Rect) {
	p.Base.Layout(bounds)

	if p.child == nil {
		return
	}

	// Calculate child bounds (inside border)
	childBounds := bounds
	if p.hasBorder {
		childBounds = bounds.Inset(1, 1, 1, 1)
	}

	p.child.Layout(childBounds)
}

// Render draws the panel.
func (p *Panel) Render(ctx runtime.RenderContext) {
	bounds := p.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	// Fill background
	ctx.Buffer.Fill(bounds, ' ', p.style)

	// Draw border if enabled
	if p.hasBorder {
		ctx.Buffer.DrawRoundedBox(bounds, p.borderStyle)

		// Draw title in top border
		if p.title != "" {
			title := " " + p.title + " "
			if len(title) > bounds.Width-4 {
				title = title[:bounds.Width-4]
			}
			x := bounds.X + 2
			ctx.Buffer.SetString(x, bounds.Y, title, p.borderStyle)
		}
	}

	// Render child
	if p.child != nil {
		p.child.Render(ctx)
	}
}

// HandleMessage delegates to child.
func (p *Panel) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if p.child != nil {
		return p.child.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the panel's child widget.
func (p *Panel) ChildWidgets() []runtime.Widget {
	if p.child == nil {
		return nil
	}
	return []runtime.Widget{p.child}
}

// Box is a simple container that fills its background.
type Box struct {
	Base
	child runtime.Widget
	style backend.Style
}

// NewBox creates a new box widget.
func NewBox(child runtime.Widget) *Box {
	return &Box{
		child: child,
		style: backend.DefaultStyle(),
	}
}

// SetStyle sets the background style.
func (b *Box) SetStyle(style backend.Style) {
	b.style = style
}

// WithStyle sets style and returns for chaining.
func (b *Box) WithStyle(style backend.Style) *Box {
	b.style = style
	return b
}

// Measure returns the child's size.
func (b *Box) Measure(constraints runtime.Constraints) runtime.Size {
	if b.child == nil {
		return constraints.MinSize()
	}
	return b.child.Measure(constraints)
}

// Layout assigns bounds to the box and child.
func (b *Box) Layout(bounds runtime.Rect) {
	b.Base.Layout(bounds)
	if b.child != nil {
		b.child.Layout(bounds)
	}
}

// Render draws the background and child.
func (b *Box) Render(ctx runtime.RenderContext) {
	// Fill background
	ctx.Buffer.Fill(b.bounds, ' ', b.style)

	// Render child
	if b.child != nil {
		b.child.Render(ctx)
	}
}

// HandleMessage delegates to child.
func (b *Box) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if b.child != nil {
		return b.child.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the box's child widget.
func (b *Box) ChildWidgets() []runtime.Widget {
	if b.child == nil {
		return nil
	}
	return []runtime.Widget{b.child}
}

// max returns the larger of two ints.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
