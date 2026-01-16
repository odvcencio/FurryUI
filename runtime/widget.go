// Package runtime provides a widget runtime for terminal UIs.
// It implements a constraint-based layout system with focus management
// and a modal stack for overlays.
package runtime

// Constraints define the min/max space available to a widget during measure.
type Constraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

// Tight returns constraints that force an exact size.
func Tight(w, h int) Constraints {
	return Constraints{
		MinWidth:  w,
		MaxWidth:  w,
		MinHeight: h,
		MaxHeight: h,
	}
}

// TightWidth returns constraints with exact width, flexible height.
func TightWidth(w int) Constraints {
	return Constraints{
		MinWidth:  w,
		MaxWidth:  w,
		MinHeight: 0,
		MaxHeight: maxInt,
	}
}

// TightHeight returns constraints with flexible width, exact height.
func TightHeight(h int) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  maxInt,
		MinHeight: h,
		MaxHeight: h,
	}
}

// Loose returns constraints with only max bounds (min = 0).
func Loose(w, h int) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  w,
		MinHeight: 0,
		MaxHeight: h,
	}
}

// Unbounded returns constraints with no limits.
func Unbounded() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  maxInt,
		MinHeight: 0,
		MaxHeight: maxInt,
	}
}

// Constrain clamps a size to fit within these constraints.
func (c Constraints) Constrain(s Size) Size {
	return Size{
		Width:  clamp(s.Width, c.MinWidth, c.MaxWidth),
		Height: clamp(s.Height, c.MinHeight, c.MaxHeight),
	}
}

// IsTight returns true if min equals max for both dimensions.
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// MaxSize returns the maximum size allowed by constraints.
func (c Constraints) MaxSize() Size {
	return Size{Width: c.MaxWidth, Height: c.MaxHeight}
}

// MinSize returns the minimum size required by constraints.
func (c Constraints) MinSize() Size {
	return Size{Width: c.MinWidth, Height: c.MinHeight}
}

// Size is a widget's measured dimensions.
type Size struct {
	Width, Height int
}

// Zero returns true if both dimensions are zero.
func (s Size) Zero() bool {
	return s.Width == 0 && s.Height == 0
}

// Rect is a positioned rectangle.
type Rect struct {
	X, Y, Width, Height int
}

// ZeroRect is the zero value rect.
var ZeroRect = Rect{}

// NewRect creates a rect from position and size.
func NewRect(x, y, w, h int) Rect {
	return Rect{X: x, Y: y, Width: w, Height: h}
}

// RectFromSize creates a rect at origin with the given size.
func RectFromSize(s Size) Rect {
	return Rect{Width: s.Width, Height: s.Height}
}

// Size returns the rect's dimensions as a Size.
func (r Rect) Size() Size {
	return Size{Width: r.Width, Height: r.Height}
}

// Contains returns true if the point is inside the rect.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Intersects returns true if the two rects overlap.
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// Intersection returns the overlapping area of two rects.
func (r Rect) Intersection(other Rect) Rect {
	x := max(r.X, other.X)
	y := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)
	if x2 <= x || y2 <= y {
		return ZeroRect
	}
	return Rect{X: x, Y: y, Width: x2 - x, Height: y2 - y}
}

// Inset returns a rect shrunk by the given amounts.
func (r Rect) Inset(top, right, bottom, left int) Rect {
	return Rect{
		X:      r.X + left,
		Y:      r.Y + top,
		Width:  max(0, r.Width-left-right),
		Height: max(0, r.Height-top-bottom),
	}
}

// Widget is the core interface all UI components implement.
type Widget interface {
	// Measure returns desired size given constraints.
	// This is the first pass of layout.
	Measure(constraints Constraints) Size

	// Layout assigns final position and size.
	// Widget should store this for use in Render.
	Layout(bounds Rect)

	// Render draws the widget to the buffer.
	Render(ctx RenderContext)

	// HandleMessage processes input/events.
	// Returns result indicating if handled and any commands to bubble up.
	HandleMessage(msg Message) HandleResult
}

// BoundsProvider reports the widget's assigned bounds.
type BoundsProvider interface {
	Bounds() Rect
}

// ChildProvider exposes child widgets for container traversal.
type ChildProvider interface {
	ChildWidgets() []Widget
}

// Invalidatable marks widgets that can report whether they need a render pass.
type Invalidatable interface {
	Invalidate()
	NeedsRender() bool
	ClearInvalidation()
}

// Focusable extends Widget for widgets that can receive keyboard focus.
type Focusable interface {
	Widget

	// CanFocus returns true if this widget can currently receive focus.
	CanFocus() bool

	// Focus is called when the widget gains focus.
	Focus()

	// Blur is called when the widget loses focus.
	Blur()

	// IsFocused returns true if this widget currently has focus.
	IsFocused() bool
}

// HandleResult is returned from HandleMessage.
type HandleResult struct {
	Handled  bool      // Was the message consumed?
	Commands []Command // Commands to send to parent/app
}

// Handled returns a result indicating the message was consumed.
func Handled() HandleResult {
	return HandleResult{Handled: true}
}

// Unhandled returns a result indicating the message was not consumed.
func Unhandled() HandleResult {
	return HandleResult{Handled: false}
}

// WithCommand returns a handled result with a single command.
func WithCommand(cmd Command) HandleResult {
	return HandleResult{Handled: true, Commands: []Command{cmd}}
}

// WithCommands returns a handled result with multiple commands.
func WithCommands(cmds ...Command) HandleResult {
	return HandleResult{Handled: true, Commands: cmds}
}

// Helper functions

const maxInt = int(^uint(0) >> 1)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
