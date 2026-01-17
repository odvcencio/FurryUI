// Package scroll provides viewport and scrollbar primitives.
package scroll

import (
	"image"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// ScrollPolicy configures when scrollbars appear.
type ScrollPolicy int

const (
	ScrollAuto ScrollPolicy = iota
	ScrollAlways
	ScrollNever
)

// Orientation describes scrollbar orientation.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// ScrollBehavior controls scroll policies and interactions.
type ScrollBehavior struct {
	Horizontal   ScrollPolicy
	Vertical     ScrollPolicy
	MouseWheel   int
	PageSize     float64
	SmoothScroll bool
}

// Controller provides scroll control for widgets.
type Controller interface {
	ScrollBy(dx, dy int)
	ScrollTo(x, y int)
	PageBy(pages int)
	ScrollToStart()
	ScrollToEnd()
}

// Viewport tracks the visible region of scrollable content.
type Viewport struct {
	content     runtime.Widget
	offset      image.Point
	contentSize runtime.Size
	viewSize    runtime.Size
	onChange    func(offset image.Point, content runtime.Size, view runtime.Size)
}

// NewViewport creates a viewport.
func NewViewport(content runtime.Widget) *Viewport {
	return &Viewport{content: content}
}

// SetContent assigns the content widget.
func (v *Viewport) SetContent(content runtime.Widget) {
	if v == nil {
		return
	}
	v.content = content
}

// Content returns the content widget.
func (v *Viewport) Content() runtime.Widget {
	if v == nil {
		return nil
	}
	return v.content
}

// SetContentSize updates the content size and clamps the offset.
func (v *Viewport) SetContentSize(size runtime.Size) {
	if v == nil {
		return
	}
	v.contentSize = size
	v.SetOffset(v.offset.X, v.offset.Y)
}

// ContentSize returns the content size.
func (v *Viewport) ContentSize() runtime.Size {
	if v == nil {
		return runtime.Size{}
	}
	return v.contentSize
}

// SetViewSize updates the view size and clamps the offset.
func (v *Viewport) SetViewSize(size runtime.Size) {
	if v == nil {
		return
	}
	v.viewSize = size
	v.SetOffset(v.offset.X, v.offset.Y)
}

// ViewSize returns the view size.
func (v *Viewport) ViewSize() runtime.Size {
	if v == nil {
		return runtime.Size{}
	}
	return v.viewSize
}

// Offset returns the current offset.
func (v *Viewport) Offset() image.Point {
	if v == nil {
		return image.Point{}
	}
	return v.offset
}

// SetOnChange sets a callback for offset updates.
func (v *Viewport) SetOnChange(fn func(offset image.Point, content runtime.Size, view runtime.Size)) {
	if v == nil {
		return
	}
	v.onChange = fn
}

// SetOffset sets the scroll offset.
func (v *Viewport) SetOffset(x, y int) {
	if v == nil {
		return
	}
	next := clampOffset(image.Point{X: x, Y: y}, v.contentSize, v.viewSize)
	if next == v.offset {
		return
	}
	v.offset = next
	if v.onChange != nil {
		v.onChange(v.offset, v.contentSize, v.viewSize)
	}
}

// ScrollBy adjusts the offset.
func (v *Viewport) ScrollBy(dx, dy int) {
	if v == nil {
		return
	}
	v.SetOffset(v.offset.X+dx, v.offset.Y+dy)
}

// ScrollTo scrolls to absolute coordinates.
func (v *Viewport) ScrollTo(x, y int) {
	if v == nil {
		return
	}
	v.SetOffset(x, y)
}

// MaxOffset returns the maximum scrollable offset.
func (v *Viewport) MaxOffset() image.Point {
	if v == nil {
		return image.Point{}
	}
	maxX := v.contentSize.Width - v.viewSize.Width
	maxY := v.contentSize.Height - v.viewSize.Height
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	return image.Point{X: maxX, Y: maxY}
}

// VisibleRect returns the visible rectangle within content.
func (v *Viewport) VisibleRect() runtime.Rect {
	if v == nil {
		return runtime.Rect{}
	}
	return runtime.Rect{
		X:      v.offset.X,
		Y:      v.offset.Y,
		Width:  v.viewSize.Width,
		Height: v.viewSize.Height,
	}
}

func clampOffset(offset image.Point, content runtime.Size, view runtime.Size) image.Point {
	maxX := content.Width - view.Width
	maxY := content.Height - view.Height
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	if offset.X < 0 {
		offset.X = 0
	}
	if offset.Y < 0 {
		offset.Y = 0
	}
	if offset.X > maxX {
		offset.X = maxX
	}
	if offset.Y > maxY {
		offset.Y = maxY
	}
	return offset
}

// VirtualContent provides virtualized content rendering.
type VirtualContent interface {
	ItemCount() int
	ItemHeight(index int) int
	RenderItem(index int, ctx runtime.RenderContext)
	ItemAt(index int) any
}

// VirtualSizer optionally provides total height for virtual content.
type VirtualSizer interface {
	TotalHeight() int
}

// VirtualIndexer optionally provides fast offset/index mapping.
type VirtualIndexer interface {
	IndexForOffset(offset int) int
	OffsetForIndex(index int) int
}

// Scrollbar configures scrollbar rendering.
type Scrollbar struct {
	Orientation  Orientation
	Track        backend.Style
	Thumb        backend.Style
	MinThumbSize int
	Chars        ScrollbarChars
}

// ScrollbarChars defines characters used to render the scrollbar.
type ScrollbarChars struct {
	Track     rune
	Thumb     rune
	ArrowUp   rune
	ArrowDown rune
}

// DefaultScrollbarChars returns ASCII defaults.
func DefaultScrollbarChars() ScrollbarChars {
	return ScrollbarChars{
		Track:     '|',
		Thumb:     '#',
		ArrowUp:   '^',
		ArrowDown: 'v',
	}
}
