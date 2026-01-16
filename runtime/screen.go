package runtime

import "github.com/odvcencio/furry-ui/backend"

// Layer represents a layer in the modal stack.
// Each layer has its own widget tree and focus scope.
type Layer struct {
	Root       Widget
	FocusScope *FocusScope
	Modal      bool // If true, blocks input to layers below
}

// Screen manages the widget tree, modal stack, and rendering.
type Screen struct {
	width, height int
	layers        []*Layer
	buffer        *Buffer
	hitGrid       *HitGrid
	hitGridModal  bool
	services      Services
}

// NewScreen creates a new screen with the given dimensions.
func NewScreen(w, h int) *Screen {
	return &Screen{
		width:   w,
		height:  h,
		buffer:  NewBuffer(w, h),
		hitGrid: NewHitGrid(w, h),
	}
}

// SetServices configures app services for bindable widgets.
func (s *Screen) SetServices(services Services) {
	s.services = services
}

// Size returns the screen dimensions.
func (s *Screen) Size() (w, h int) {
	return s.width, s.height
}

// Resize changes the screen dimensions.
func (s *Screen) Resize(w, h int) {
	s.width = w
	s.height = h
	s.buffer.Resize(w, h)
	if s.hitGrid != nil {
		s.hitGrid.Resize(w, h)
	}

	// Re-layout all layers
	bounds := Rect{0, 0, w, h}
	for _, layer := range s.layers {
		if layer.Root != nil {
			layer.Root.Layout(bounds)
		}
	}
}

// Buffer returns the screen's render buffer.
func (s *Screen) Buffer() *Buffer {
	return s.buffer
}

// SetRoot sets the root widget of the base layer.
// Creates the base layer if it doesn't exist.
func (s *Screen) SetRoot(root Widget) {
	var oldRoot Widget
	if len(s.layers) == 0 {
		s.layers = append(s.layers, &Layer{
			Root:       root,
			FocusScope: NewFocusScope(),
			Modal:      false,
		})
	} else {
		oldRoot = s.layers[0].Root
		s.layers[0].Root = root
	}

	if oldRoot != nil {
		UnmountTree(oldRoot)
		UnbindTree(oldRoot)
	}

	// Layout the root widget
	if root != nil {
		BindTree(root, s.services)
		root.Layout(Rect{0, 0, s.width, s.height})
		MountTree(root)
	}
}

// Root returns the base layer's root widget.
func (s *Screen) Root() Widget {
	if len(s.layers) == 0 {
		return nil
	}
	return s.layers[0].Root
}

// PushLayer adds a new layer on top of the stack.
// If modal is true, input won't pass to layers below.
func (s *Screen) PushLayer(root Widget, modal bool) {
	layer := &Layer{
		Root:       root,
		FocusScope: NewFocusScope(),
		Modal:      modal,
	}
	s.layers = append(s.layers, layer)

	// Layout the new layer
	if root != nil {
		BindTree(root, s.services)
		root.Layout(Rect{0, 0, s.width, s.height})
		MountTree(root)
	}
}

// PopLayer removes the top layer from the stack.
// Returns false if only the base layer remains (can't pop it).
func (s *Screen) PopLayer() bool {
	if len(s.layers) <= 1 {
		return false
	}

	// Clear focus on the layer being removed
	top := s.layers[len(s.layers)-1]
	top.FocusScope.ClearFocus()
	if top.Root != nil {
		UnmountTree(top.Root)
		UnbindTree(top.Root)
	}

	s.layers = s.layers[:len(s.layers)-1]
	return true
}

// TopLayer returns the topmost layer.
func (s *Screen) TopLayer() *Layer {
	if len(s.layers) == 0 {
		return nil
	}
	return s.layers[len(s.layers)-1]
}

// LayerCount returns the number of layers.
func (s *Screen) LayerCount() int {
	return len(s.layers)
}

// FocusScope returns the focus scope of the top layer.
func (s *Screen) FocusScope() *FocusScope {
	if top := s.TopLayer(); top != nil {
		return top.FocusScope
	}
	return nil
}

// Render draws all layers to the buffer.
func (s *Screen) Render() {
	ctx := RenderContext{
		Buffer:  s.buffer,
		Focused: false,
		Bounds:  Rect{0, 0, s.width, s.height},
	}

	// Render layers from bottom to top
	for i, layer := range s.layers {
		if layer.Root == nil {
			continue
		}

		// Determine if this layer contains focus
		isTopLayer := i == len(s.layers)-1
		ctx.Focused = isTopLayer

		layer.Root.Render(ctx)
	}

	s.buildHitGrid()
}

// HandleMessage dispatches a message to the appropriate layer.
// Messages go to the top layer. If not handled and not modal,
// they bubble down to lower layers.
func (s *Screen) HandleMessage(msg Message) HandleResult {
	if mouse, ok := msg.(MouseMsg); ok && s.hitGrid != nil {
		if target := s.hitGrid.WidgetAt(mouse.X, mouse.Y); target != nil {
			result := target.HandleMessage(msg)
			for _, cmd := range result.Commands {
				s.handleCommand(cmd)
			}
			if result.Handled || s.hitGridModal {
				return result
			}
		} else if s.hitGridModal {
			return Unhandled()
		}
	}

	// Process from top to bottom
	for i := len(s.layers) - 1; i >= 0; i-- {
		layer := s.layers[i]
		if layer.Root == nil {
			continue
		}

		result := layer.Root.HandleMessage(msg)

		// Process any commands
		for _, cmd := range result.Commands {
			s.handleCommand(cmd)
		}

		if result.Handled {
			return result
		}

		// If modal, don't pass to lower layers
		if layer.Modal {
			break
		}
	}

	return Unhandled()
}

// handleCommand processes a command from a widget.
func (s *Screen) handleCommand(cmd Command) {
	switch c := cmd.(type) {
	case FocusNext:
		if scope := s.FocusScope(); scope != nil {
			scope.FocusNext()
		}
	case FocusPrev:
		if scope := s.FocusScope(); scope != nil {
			scope.FocusPrev()
		}
	case PopOverlay:
		s.PopLayer()
	case PushOverlay:
		s.PushLayer(c.Widget, c.Modal)
	}
	// Other commands bubble up to App
}

func (s *Screen) buildHitGrid() {
	if s.hitGrid == nil {
		s.hitGrid = NewHitGrid(s.width, s.height)
	} else {
		s.hitGrid.Resize(s.width, s.height)
		s.hitGrid.Clear()
	}
	s.hitGridModal = false
	if len(s.layers) == 0 {
		return
	}

	start := 0
	if top := s.layers[len(s.layers)-1]; top != nil && top.Modal {
		start = len(s.layers) - 1
		s.hitGridModal = true
	}
	for i := start; i < len(s.layers); i++ {
		layer := s.layers[i]
		if layer == nil || layer.Root == nil {
			continue
		}
		s.addHitWidgets(layer.Root)
	}
}

func (s *Screen) addHitWidgets(widget Widget) {
	if widget == nil {
		return
	}
	if container, ok := widget.(ChildProvider); ok {
		children := container.ChildWidgets()
		if len(children) > 0 {
			for _, child := range children {
				s.addHitWidgets(child)
			}
			return
		}
	}
	boundsProvider, ok := widget.(BoundsProvider)
	if !ok {
		return
	}
	bounds := boundsProvider.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	s.hitGrid.Add(widget, bounds)
}

// RenderContext provides context to widgets during rendering.
type RenderContext struct {
	Buffer  *Buffer
	Focused bool // Is the containing layer focused?
	Bounds  Rect // Widget's allocated bounds
}

// Sub creates a new context for a child widget with adjusted bounds.
func (ctx RenderContext) Sub(bounds Rect) RenderContext {
	return RenderContext{
		Buffer:  ctx.Buffer,
		Focused: ctx.Focused,
		Bounds:  bounds,
	}
}

// Clear fills the context bounds with spaces using the provided style.
func (ctx RenderContext) Clear(style backend.Style) {
	if ctx.Buffer == nil {
		return
	}
	ctx.Buffer.Fill(ctx.Bounds, ' ', style)
}

// SubBuffer returns a buffer view clipped to the context bounds.
func (ctx RenderContext) SubBuffer() *SubBuffer {
	return ctx.Buffer.Sub(ctx.Bounds)
}
