package widgets

import (
	"fmt"
	"image"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// ScrollView provides a scrollable container.
type ScrollView struct {
	FocusableBase
	content    runtime.Widget
	virtual    scroll.VirtualContent
	viewport   *scroll.Viewport
	behavior   scroll.ScrollBehavior
	style      backend.Style
	services   runtime.Services
	vScrollbar scroll.Scrollbar
	hScrollbar scroll.Scrollbar
	childBuf   *runtime.Buffer
}

// NewScrollView creates a scroll view for content.
func NewScrollView(content runtime.Widget) *ScrollView {
	vp := scroll.NewViewport(content)
	view := &ScrollView{
		content:  content,
		virtual:  asVirtual(content),
		viewport: vp,
		behavior: scroll.ScrollBehavior{Vertical: scroll.ScrollAuto, Horizontal: scroll.ScrollAuto, MouseWheel: 3, PageSize: 1},
		style:    backend.DefaultStyle(),
		vScrollbar: scroll.Scrollbar{
			Orientation:  scroll.Vertical,
			Track:        backend.DefaultStyle(),
			Thumb:        backend.DefaultStyle().Reverse(true),
			MinThumbSize: 1,
			Chars:        scroll.DefaultScrollbarChars(),
		},
		hScrollbar: scroll.Scrollbar{
			Orientation:  scroll.Horizontal,
			Track:        backend.DefaultStyle(),
			Thumb:        backend.DefaultStyle().Reverse(true),
			MinThumbSize: 1,
			Chars:        scroll.DefaultScrollbarChars(),
		},
	}
	view.setViewportCallbacks()
	return view
}

// SetContent updates the scroll content.
func (s *ScrollView) SetContent(content runtime.Widget) {
	if s == nil {
		return
	}
	s.content = content
	s.virtual = asVirtual(content)
	if s.viewport != nil {
		s.viewport.SetContent(content)
	}
	s.setViewportCallbacks()
}

// SetBehavior updates scroll behavior.
func (s *ScrollView) SetBehavior(behavior scroll.ScrollBehavior) {
	if s == nil {
		return
	}
	s.behavior = behavior
}

// Bind attaches app services.
func (s *ScrollView) Bind(services runtime.Services) {
	s.services = services
	s.setViewportCallbacks()
}

// Unbind releases app services.
func (s *ScrollView) Unbind() {
	s.services = runtime.Services{}
}

// Measure returns the desired size.
func (s *ScrollView) Measure(constraints runtime.Constraints) runtime.Size {
	if s == nil {
		return constraints.MinSize()
	}
	if s.virtual != nil {
		contentSize := s.virtualContentSize(constraints)
		if s.viewport != nil {
			s.viewport.SetContentSize(contentSize)
		}
		return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight})
	}
	if s.content == nil {
		return constraints.MinSize()
	}
	maxInt := int(^uint(0) >> 1)
	contentSize := s.content.Measure(runtime.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: 0,
		MaxHeight: maxInt,
	})
	if s.viewport != nil {
		s.viewport.SetContentSize(contentSize)
	}
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight})
}

// Layout positions the content.
func (s *ScrollView) Layout(bounds runtime.Rect) {
	s.FocusableBase.Layout(bounds)
	if s.viewport == nil {
		return
	}
	s.viewport.SetViewSize(bounds.Size())
	if s.virtual != nil {
		contentSize := s.viewport.ContentSize()
		if contentSize.Width <= 0 || contentSize.Height <= 0 {
			contentSize = s.virtualContentSize(runtime.Constraints{
				MinWidth:  bounds.Width,
				MaxWidth:  bounds.Width,
				MinHeight: 0,
				MaxHeight: int(^uint(0) >> 1),
			})
			s.viewport.SetContentSize(contentSize)
		}
		return
	}
	if s.content == nil {
		return
	}
	contentSize := s.viewport.ContentSize()
	if contentSize.Width <= 0 {
		contentSize.Width = bounds.Width
	}
	if contentSize.Height <= 0 {
		contentSize.Height = bounds.Height
	}
	s.content.Layout(runtime.Rect{X: 0, Y: 0, Width: contentSize.Width, Height: contentSize.Height})
}

// Render draws the visible portion of content.
func (s *ScrollView) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.Fill(bounds, ' ', s.style)
	if s.viewport == nil {
		return
	}
	if s.virtual != nil {
		s.renderVirtual(ctx)
		s.drawScrollbars(ctx)
		return
	}
	if s.content == nil {
		return
	}
	contentSize := s.viewport.ContentSize()
	if contentSize.Width <= 0 || contentSize.Height <= 0 {
		return
	}
	if s.childBuf == nil {
		s.childBuf = runtime.NewBuffer(contentSize.Width, contentSize.Height)
	} else {
		s.childBuf.Resize(contentSize.Width, contentSize.Height)
	}
	s.childBuf.Clear()
	childCtx := runtime.RenderContext{
		Buffer: s.childBuf,
		Bounds: runtime.Rect{Width: contentSize.Width, Height: contentSize.Height},
	}
	s.content.Render(childCtx)

	offset := s.viewport.Offset()
	for y := 0; y < bounds.Height; y++ {
		srcY := y + offset.Y
		if srcY < 0 || srcY >= contentSize.Height {
			continue
		}
		for x := 0; x < bounds.Width; x++ {
			srcX := x + offset.X
			if srcX < 0 || srcX >= contentSize.Width {
				continue
			}
			cell := s.childBuf.Get(srcX, srcY)
			ctx.Buffer.Set(bounds.X+x, bounds.Y+y, cell.Rune, cell.Style)
		}
	}
	s.drawScrollbars(ctx)
}

// HandleMessage handles scrolling input.
func (s *ScrollView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || s.viewport == nil {
		return runtime.Unhandled()
	}
	if s.content != nil {
		if result := s.content.HandleMessage(msg); result.Handled {
			return result
		}
	}
	switch ev := msg.(type) {
	case runtime.KeyMsg:
		if !s.focused {
			return runtime.Unhandled()
		}
		switch ev.Key {
		case terminal.KeyUp:
			s.ScrollBy(0, -1)
			return runtime.Handled()
		case terminal.KeyDown:
			s.ScrollBy(0, 1)
			return runtime.Handled()
		case terminal.KeyPageUp:
			s.PageBy(-1)
			return runtime.Handled()
		case terminal.KeyPageDown:
			s.PageBy(1)
			return runtime.Handled()
		case terminal.KeyHome:
			s.ScrollToStart()
			return runtime.Handled()
		case terminal.KeyEnd:
			s.ScrollToEnd()
			return runtime.Handled()
		}
	case runtime.MouseMsg:
		if ev.Button == runtime.MouseWheelUp {
			s.ScrollBy(0, -s.behavior.MouseWheel)
			return runtime.Handled()
		}
		if ev.Button == runtime.MouseWheelDown {
			s.ScrollBy(0, s.behavior.MouseWheel)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the content widget.
func (s *ScrollView) ChildWidgets() []runtime.Widget {
	if s == nil || s.content == nil {
		return nil
	}
	return []runtime.Widget{s.content}
}

// ScrollBy scrolls the view by delta.
func (s *ScrollView) ScrollBy(dx, dy int) {
	if s == nil || s.viewport == nil {
		return
	}
	if s.virtual != nil {
		s.virtualScrollBy(dy)
		return
	}
	s.viewport.ScrollBy(dx, dy)
}

// ScrollTo scrolls to the specified offset.
func (s *ScrollView) ScrollTo(x, y int) {
	if s == nil || s.viewport == nil {
		return
	}
	if s.virtual != nil {
		index := s.virtualIndexForOffset(y)
		s.viewport.ScrollTo(0, s.virtualOffsetForIndex(index))
		return
	}
	s.viewport.ScrollTo(x, y)
}

// PageBy scrolls by page count.
func (s *ScrollView) PageBy(pages int) {
	if s == nil {
		return
	}
	delta := s.pageSize() * pages
	s.ScrollBy(0, delta)
}

// ScrollToStart scrolls to the top-left.
func (s *ScrollView) ScrollToStart() {
	s.ScrollTo(0, 0)
}

// ScrollToEnd scrolls to the bottom-right.
func (s *ScrollView) ScrollToEnd() {
	if s == nil || s.viewport == nil {
		return
	}
	max := s.viewport.MaxOffset()
	s.ScrollTo(max.X, max.Y)
}

func (s *ScrollView) pageSize() int {
	if s == nil {
		return 1
	}
	if s.behavior.PageSize > 0 {
		return int(float64(s.bounds.Height) * s.behavior.PageSize)
	}
	if s.bounds.Height > 0 {
		return s.bounds.Height
	}
	return 1
}

func (s *ScrollView) setViewportCallbacks() {
	if s == nil || s.viewport == nil {
		return
	}
	s.viewport.SetOnChange(func(offset image.Point, content runtime.Size, view runtime.Size) {
		s.invalidate()
		s.announceScroll(offset, content, view)
	})
}

func (s *ScrollView) invalidate() {
	if s == nil {
		return
	}
	s.Invalidate()
	s.services.Invalidate()
}

func (s *ScrollView) announceScroll(offset image.Point, content runtime.Size, view runtime.Size) {
	announcer := s.services.Announcer()
	if announcer == nil {
		return
	}
	if content.Height <= 0 {
		return
	}
	line := offset.Y + 1
	if line < 1 {
		line = 1
	}
	if line > content.Height {
		line = content.Height
	}
	message := fmt.Sprintf("Line %d of %d", line, content.Height)
	announcer.Announce(message, 0)
}

func (s *ScrollView) drawScrollbars(ctx runtime.RenderContext) {
	if s == nil || s.viewport == nil {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	content := s.viewport.ContentSize()
	view := s.viewport.ViewSize()
	offset := s.viewport.Offset()
	if s.behavior.Vertical != scroll.ScrollNever {
		shouldDraw := s.behavior.Vertical == scroll.ScrollAlways || content.Height > view.Height
		if shouldDraw {
			s.drawVerticalScrollbar(ctx, bounds, content, view, offset)
		}
	}
	if s.behavior.Horizontal != scroll.ScrollNever {
		shouldDraw := s.behavior.Horizontal == scroll.ScrollAlways || content.Width > view.Width
		if shouldDraw {
			s.drawHorizontalScrollbar(ctx, bounds, content, view, offset)
		}
	}
}

func (s *ScrollView) drawVerticalScrollbar(ctx runtime.RenderContext, bounds runtime.Rect, content runtime.Size, view runtime.Size, offset image.Point) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	x := bounds.X + bounds.Width - 1
	if x < bounds.X {
		return
	}
	trackChar := s.vScrollbar.Chars.Track
	thumbChar := s.vScrollbar.Chars.Thumb
	if trackChar == 0 {
		trackChar = '|'
	}
	if thumbChar == 0 {
		thumbChar = '#'
	}
	for y := bounds.Y; y < bounds.Y+bounds.Height; y++ {
		ctx.Buffer.Set(x, y, trackChar, s.vScrollbar.Track)
	}
	if content.Height <= 0 || view.Height <= 0 {
		return
	}
	maxOffset := content.Height - view.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	thumbSize := int(float64(view.Height) / float64(content.Height) * float64(view.Height))
	if thumbSize < s.vScrollbar.MinThumbSize {
		thumbSize = s.vScrollbar.MinThumbSize
	}
	if thumbSize > view.Height {
		thumbSize = view.Height
	}
	thumbStart := 0
	if maxOffset > 0 {
		thumbStart = int(float64(offset.Y) / float64(maxOffset) * float64(view.Height-thumbSize))
	}
	for i := 0; i < thumbSize; i++ {
		y := bounds.Y + thumbStart + i
		if y >= bounds.Y && y < bounds.Y+bounds.Height {
			ctx.Buffer.Set(x, y, thumbChar, s.vScrollbar.Thumb)
		}
	}
}

func (s *ScrollView) drawHorizontalScrollbar(ctx runtime.RenderContext, bounds runtime.Rect, content runtime.Size, view runtime.Size, offset image.Point) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	y := bounds.Y + bounds.Height - 1
	if y < bounds.Y {
		return
	}
	trackChar := s.hScrollbar.Chars.Track
	thumbChar := s.hScrollbar.Chars.Thumb
	if trackChar == 0 {
		trackChar = '-'
	}
	if thumbChar == 0 {
		thumbChar = '#'
	}
	for x := bounds.X; x < bounds.X+bounds.Width; x++ {
		ctx.Buffer.Set(x, y, trackChar, s.hScrollbar.Track)
	}
	if content.Width <= 0 || view.Width <= 0 {
		return
	}
	maxOffset := content.Width - view.Width
	if maxOffset < 0 {
		maxOffset = 0
	}
	thumbSize := int(float64(view.Width) / float64(content.Width) * float64(view.Width))
	if thumbSize < s.hScrollbar.MinThumbSize {
		thumbSize = s.hScrollbar.MinThumbSize
	}
	if thumbSize > view.Width {
		thumbSize = view.Width
	}
	thumbStart := 0
	if maxOffset > 0 {
		thumbStart = int(float64(offset.X) / float64(maxOffset) * float64(view.Width-thumbSize))
	}
	for i := 0; i < thumbSize; i++ {
		x := bounds.X + thumbStart + i
		if x >= bounds.X && x < bounds.X+bounds.Width {
			ctx.Buffer.Set(x, y, thumbChar, s.hScrollbar.Thumb)
		}
	}
}

func (s *ScrollView) renderVirtual(ctx runtime.RenderContext) {
	if s == nil || s.viewport == nil || s.virtual == nil {
		return
	}
	bounds := s.bounds
	contentSize := s.viewport.ContentSize()
	if contentSize.Height <= 0 {
		contentSize = s.virtualContentSize(runtime.Constraints{
			MinWidth:  bounds.Width,
			MaxWidth:  bounds.Width,
			MinHeight: 0,
			MaxHeight: int(^uint(0) >> 1),
		})
		s.viewport.SetContentSize(contentSize)
	}
	offset := s.viewport.Offset()
	start := s.virtualIndexForOffset(offset.Y)
	y := s.virtualOffsetForIndex(start)
	count := s.virtual.ItemCount()
	for i := start; i < count; i++ {
		itemHeight := s.virtual.ItemHeight(i)
		if itemHeight <= 0 {
			continue
		}
		itemBounds := runtime.Rect{
			X:      bounds.X,
			Y:      bounds.Y + (y - offset.Y),
			Width:  bounds.Width,
			Height: itemHeight,
		}
		if itemBounds.Y >= bounds.Y+bounds.Height {
			break
		}
		if itemBounds.Y+itemBounds.Height <= bounds.Y {
			y += itemHeight
			continue
		}
		s.virtual.RenderItem(i, ctx.Sub(itemBounds))
		y += itemHeight
	}
}

func (s *ScrollView) virtualContentSize(constraints runtime.Constraints) runtime.Size {
	if s == nil || s.virtual == nil {
		return runtime.Size{}
	}
	if sizer, ok := s.virtual.(scroll.VirtualSizer); ok {
		total := sizer.TotalHeight()
		if total < 0 {
			total = 0
		}
		width := constraints.MaxWidth
		if width <= 0 {
			width = constraints.MinWidth
		}
		return runtime.Size{Width: width, Height: total}
	}
	count := s.virtual.ItemCount()
	totalHeight := 0
	for i := 0; i < count; i++ {
		height := s.virtual.ItemHeight(i)
		if height > 0 {
			totalHeight += height
		}
	}
	width := constraints.MaxWidth
	if width <= 0 {
		width = constraints.MinWidth
	}
	return runtime.Size{Width: width, Height: totalHeight}
}

func (s *ScrollView) virtualIndexForOffset(offset int) int {
	if s == nil || s.virtual == nil {
		return 0
	}
	if indexer, ok := s.virtual.(scroll.VirtualIndexer); ok {
		return indexer.IndexForOffset(offset)
	}
	if offset <= 0 {
		return 0
	}
	total := 0
	count := s.virtual.ItemCount()
	for i := 0; i < count; i++ {
		height := s.virtual.ItemHeight(i)
		if height <= 0 {
			continue
		}
		if total+height > offset {
			return i
		}
		total += height
	}
	if count > 0 {
		return count - 1
	}
	return 0
}

func (s *ScrollView) virtualOffsetForIndex(index int) int {
	if s == nil || s.virtual == nil || index <= 0 {
		return 0
	}
	if indexer, ok := s.virtual.(scroll.VirtualIndexer); ok {
		return indexer.OffsetForIndex(index)
	}
	total := 0
	count := s.virtual.ItemCount()
	if index >= count {
		index = count - 1
	}
	for i := 0; i < index; i++ {
		height := s.virtual.ItemHeight(i)
		if height > 0 {
			total += height
		}
	}
	return total
}

func (s *ScrollView) virtualScrollBy(delta int) {
	if s == nil || s.virtual == nil || s.viewport == nil || delta == 0 {
		return
	}
	offset := s.viewport.Offset()
	index := s.virtualIndexForOffset(offset.Y)
	index += delta
	if index < 0 {
		index = 0
	}
	maxIndex := s.virtual.ItemCount() - 1
	if maxIndex < 0 {
		maxIndex = 0
	}
	if index > maxIndex {
		index = maxIndex
	}
	s.viewport.ScrollTo(0, s.virtualOffsetForIndex(index))
}

func asVirtual(content runtime.Widget) scroll.VirtualContent {
	if content == nil {
		return nil
	}
	virtual, ok := content.(scroll.VirtualContent)
	if !ok {
		return nil
	}
	return virtual
}

var _ scroll.Controller = (*ScrollView)(nil)
