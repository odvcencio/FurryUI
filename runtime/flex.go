package runtime

// FlexDirection specifies the main axis of a flex container.
type FlexDirection int

const (
	Column FlexDirection = iota // Vertical (VBox)
	Row                         // Horizontal (HBox)
)

// FlexChild wraps a widget with flex layout properties.
type FlexChild struct {
	Widget Widget
	Grow   float64 // How much to grow (0 = fixed, 1+ = proportional)
	Shrink float64 // How much to shrink (0 = fixed, 1+ = proportional)
	Basis  int     // Base size (-1 = use measured size)
}

// Fixed creates a child that doesn't grow or shrink.
func Fixed(w Widget) FlexChild {
	return FlexChild{Widget: w, Grow: 0, Shrink: 0, Basis: -1}
}

// Flexible creates a child that grows with the given factor.
func Flexible(w Widget, grow float64) FlexChild {
	return FlexChild{Widget: w, Grow: grow, Shrink: 1, Basis: -1}
}

// Expanded creates a child that grows to fill available space (Grow=1).
func Expanded(w Widget) FlexChild {
	return FlexChild{Widget: w, Grow: 1, Shrink: 1, Basis: -1}
}

// Sized creates a child with a fixed basis size.
func Sized(w Widget, basis int) FlexChild {
	return FlexChild{Widget: w, Grow: 0, Shrink: 0, Basis: basis}
}

// Flex is a container that lays out children along an axis.
type Flex struct {
	Direction FlexDirection
	Children  []FlexChild
	Gap       int // Space between children

	// Cached layout
	bounds      Rect
	childBounds []Rect
	measured    Size
}

// VBox creates a vertical flex container.
func VBox(children ...FlexChild) *Flex {
	return &Flex{Direction: Column, Children: children}
}

// HBox creates a horizontal flex container.
func HBox(children ...FlexChild) *Flex {
	return &Flex{Direction: Row, Children: children}
}

// WithGap sets the gap between children.
func (f *Flex) WithGap(gap int) *Flex {
	f.Gap = gap
	return f
}

// Add appends a child to the flex container.
func (f *Flex) Add(child FlexChild) {
	f.Children = append(f.Children, child)
}

// Measure calculates the desired size of the flex container.
func (f *Flex) Measure(constraints Constraints) Size {
	if len(f.Children) == 0 {
		f.measured = constraints.MinSize()
		return f.measured
	}

	// Measure all children with loose constraints on the main axis
	childSizes := make([]Size, len(f.Children))
	totalMain := 0
	maxCross := 0

	for i, child := range f.Children {
		var childConstraints Constraints
		if f.Direction == Column {
			// For column: constrain width, free height
			childConstraints = Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: 0,
				MaxHeight: maxInt,
			}
		} else {
			// For row: free width, constrain height
			childConstraints = Constraints{
				MinWidth:  0,
				MaxWidth:  maxInt,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			}
		}

		if child.Basis >= 0 {
			childSizes[i] = f.sizeWithBasis(child.Basis)
		} else {
			childSizes[i] = child.Widget.Measure(childConstraints)
		}

		if f.Direction == Column {
			totalMain += childSizes[i].Height
			maxCross = max(maxCross, childSizes[i].Width)
		} else {
			totalMain += childSizes[i].Width
			maxCross = max(maxCross, childSizes[i].Height)
		}
	}

	// Add gaps
	if len(f.Children) > 1 {
		totalMain += f.Gap * (len(f.Children) - 1)
	}

	if f.Direction == Column {
		f.measured = constraints.Constrain(Size{Width: maxCross, Height: totalMain})
	} else {
		f.measured = constraints.Constrain(Size{Width: totalMain, Height: maxCross})
	}
	return f.measured
}

// Layout positions all children within the given bounds.
func (f *Flex) Layout(bounds Rect) {
	f.bounds = bounds
	f.childBounds = make([]Rect, len(f.Children))

	if len(f.Children) == 0 {
		return
	}

	// Measure children to get their preferred sizes
	childSizes := make([]Size, len(f.Children))
	totalFixed := 0
	totalGrow := 0.0

	for i, child := range f.Children {
		var childConstraints Constraints
		if f.Direction == Column {
			childConstraints = Loose(bounds.Width, maxInt)
		} else {
			childConstraints = Loose(maxInt, bounds.Height)
		}

		if child.Basis >= 0 {
			childSizes[i] = f.sizeWithBasis(child.Basis)
		} else {
			childSizes[i] = child.Widget.Measure(childConstraints)
		}

		mainSize := f.mainSize(childSizes[i])
		if child.Grow == 0 {
			totalFixed += mainSize
		}
		totalGrow += child.Grow
	}

	// Add gaps to fixed space
	gaps := 0
	if len(f.Children) > 1 {
		gaps = f.Gap * (len(f.Children) - 1)
	}
	totalFixed += gaps

	// Calculate available space for growing
	available := f.mainSize(bounds.Size()) - totalFixed
	if available < 0 {
		available = 0
	}

	// Position children
	offset := 0
	for i, child := range f.Children {
		// Calculate size
		var mainSize int
		if child.Grow > 0 && totalGrow > 0 {
			// Growing child: get proportional share
			share := child.Grow / totalGrow
			mainSize = int(float64(available) * share)
		} else {
			mainSize = f.mainSize(childSizes[i])
		}

		// Create bounds for this child
		var childBounds Rect
		if f.Direction == Column {
			childBounds = Rect{
				X:      bounds.X,
				Y:      bounds.Y + offset,
				Width:  bounds.Width,
				Height: mainSize,
			}
		} else {
			childBounds = Rect{
				X:      bounds.X + offset,
				Y:      bounds.Y,
				Width:  mainSize,
				Height: bounds.Height,
			}
		}

		f.childBounds[i] = childBounds
		child.Widget.Layout(childBounds)

		offset += mainSize + f.Gap
	}
}

// Bounds returns the assigned bounds for the flex container.
func (f *Flex) Bounds() Rect {
	return f.bounds
}

// ChildWidgets returns the flex container's child widgets.
func (f *Flex) ChildWidgets() []Widget {
	if len(f.Children) == 0 {
		return nil
	}
	children := make([]Widget, 0, len(f.Children))
	for _, child := range f.Children {
		if child.Widget != nil {
			children = append(children, child.Widget)
		}
	}
	return children
}

// Render draws all children.
func (f *Flex) Render(ctx RenderContext) {
	for i, child := range f.Children {
		if i < len(f.childBounds) {
			childCtx := ctx.Sub(f.childBounds[i])
			child.Widget.Render(childCtx)
		}
	}
}

// HandleMessage dispatches to children.
// Messages go to all children; first handler wins.
func (f *Flex) HandleMessage(msg Message) HandleResult {
	for _, child := range f.Children {
		result := child.Widget.HandleMessage(msg)
		if result.Handled {
			return result
		}
	}
	return Unhandled()
}

// mainSize returns the size along the main axis.
func (f *Flex) mainSize(s Size) int {
	if f.Direction == Column {
		return s.Height
	}
	return s.Width
}

// crossSize returns the size along the cross axis.
func (f *Flex) crossSize(s Size) int {
	if f.Direction == Column {
		return s.Width
	}
	return s.Height
}

// sizeWithBasis creates a size with the basis on the main axis.
func (f *Flex) sizeWithBasis(basis int) Size {
	if f.Direction == Column {
		return Size{Width: 0, Height: basis}
	}
	return Size{Width: basis, Height: 0}
}

// Spacer is a flexible empty widget for adding space in flex layouts.
type Spacer struct {
	bounds Rect
}

// NewSpacer creates a spacer widget.
func NewSpacer() *Spacer {
	return &Spacer{}
}

func (s *Spacer) Measure(constraints Constraints) Size {
	return constraints.MinSize()
}

func (s *Spacer) Layout(bounds Rect) {
	s.bounds = bounds
}

// Bounds returns the assigned bounds for the spacer.
func (s *Spacer) Bounds() Rect {
	return s.bounds
}

func (s *Spacer) Render(ctx RenderContext) {
	// Spacer is invisible
}

func (s *Spacer) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

// Space creates a flexible spacer that expands to fill available space.
func Space() FlexChild {
	return Expanded(NewSpacer())
}

// FixedSpace creates a fixed-size spacer.
func FixedSpace(size int) FlexChild {
	return Sized(NewSpacer(), size)
}
