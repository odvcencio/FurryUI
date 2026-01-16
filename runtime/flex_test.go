package runtime

import "testing"

// testWidget is a simple widget for testing layout.
type testWidget struct {
	preferredSize Size
	bounds        Rect
}

func newTestWidget(w, h int) *testWidget {
	return &testWidget{preferredSize: Size{Width: w, Height: h}}
}

func (t *testWidget) Measure(constraints Constraints) Size {
	return constraints.Constrain(t.preferredSize)
}

func (t *testWidget) Layout(bounds Rect) {
	t.bounds = bounds
}

func (t *testWidget) Render(ctx RenderContext) {}

func (t *testWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func TestVBox_FixedChildren(t *testing.T) {
	w1 := newTestWidget(100, 10)
	w2 := newTestWidget(100, 20)
	w3 := newTestWidget(100, 30)

	vbox := VBox(Fixed(w1), Fixed(w2), Fixed(w3))

	// Measure
	size := vbox.Measure(Loose(100, 100))
	if size.Height != 60 {
		t.Errorf("Measure height = %d, want 60", size.Height)
	}

	// Layout
	vbox.Layout(Rect{0, 0, 100, 100})

	// Check child positions
	if w1.bounds != (Rect{0, 0, 100, 10}) {
		t.Errorf("w1 bounds = %v, want {0,0,100,10}", w1.bounds)
	}
	if w2.bounds != (Rect{0, 10, 100, 20}) {
		t.Errorf("w2 bounds = %v, want {0,10,100,20}", w2.bounds)
	}
	if w3.bounds != (Rect{0, 30, 100, 30}) {
		t.Errorf("w3 bounds = %v, want {0,30,100,30}", w3.bounds)
	}
}

func TestVBox_FlexibleChildren(t *testing.T) {
	fixed := newTestWidget(100, 20)
	flex := newTestWidget(100, 0)

	vbox := VBox(Fixed(fixed), Expanded(flex))
	vbox.Layout(Rect{0, 0, 100, 100})

	// Fixed should keep its size
	if fixed.bounds.Height != 20 {
		t.Errorf("fixed height = %d, want 20", fixed.bounds.Height)
	}

	// Flexible should fill remaining space
	if flex.bounds.Height != 80 {
		t.Errorf("flex height = %d, want 80", flex.bounds.Height)
	}
	if flex.bounds.Y != 20 {
		t.Errorf("flex Y = %d, want 20", flex.bounds.Y)
	}
}

func TestVBox_MultipleFlexible(t *testing.T) {
	fixed := newTestWidget(100, 20)
	flex1 := newTestWidget(100, 0)
	flex2 := newTestWidget(100, 0)

	// flex1 gets 1/3, flex2 gets 2/3
	vbox := VBox(
		Fixed(fixed),
		Flexible(flex1, 1),
		Flexible(flex2, 2),
	)
	vbox.Layout(Rect{0, 0, 100, 80})

	// Remaining: 80 - 20 = 60
	// flex1: 60 * 1/3 = 20
	// flex2: 60 * 2/3 = 40

	if flex1.bounds.Height != 20 {
		t.Errorf("flex1 height = %d, want 20", flex1.bounds.Height)
	}
	if flex2.bounds.Height != 40 {
		t.Errorf("flex2 height = %d, want 40", flex2.bounds.Height)
	}
}

func TestVBox_WithGap(t *testing.T) {
	w1 := newTestWidget(100, 10)
	w2 := newTestWidget(100, 10)
	w3 := newTestWidget(100, 10)

	vbox := VBox(Fixed(w1), Fixed(w2), Fixed(w3)).WithGap(5)

	size := vbox.Measure(Loose(100, 100))
	// 10 + 10 + 10 + 5 + 5 = 40
	if size.Height != 40 {
		t.Errorf("Measure with gap height = %d, want 40", size.Height)
	}

	vbox.Layout(Rect{0, 0, 100, 100})

	if w1.bounds.Y != 0 {
		t.Errorf("w1 Y = %d, want 0", w1.bounds.Y)
	}
	if w2.bounds.Y != 15 { // 10 + 5
		t.Errorf("w2 Y = %d, want 15", w2.bounds.Y)
	}
	if w3.bounds.Y != 30 { // 10 + 5 + 10 + 5
		t.Errorf("w3 Y = %d, want 30", w3.bounds.Y)
	}
}

func TestHBox_FixedChildren(t *testing.T) {
	w1 := newTestWidget(20, 100)
	w2 := newTestWidget(30, 100)
	w3 := newTestWidget(40, 100)

	hbox := HBox(Fixed(w1), Fixed(w2), Fixed(w3))
	hbox.Layout(Rect{0, 0, 100, 100})

	if w1.bounds != (Rect{0, 0, 20, 100}) {
		t.Errorf("w1 bounds = %v, want {0,0,20,100}", w1.bounds)
	}
	if w2.bounds != (Rect{20, 0, 30, 100}) {
		t.Errorf("w2 bounds = %v, want {20,0,30,100}", w2.bounds)
	}
	if w3.bounds != (Rect{50, 0, 40, 100}) {
		t.Errorf("w3 bounds = %v, want {50,0,40,100}", w3.bounds)
	}
}

func TestHBox_FlexibleChildren(t *testing.T) {
	fixed := newTestWidget(20, 100)
	flex := newTestWidget(0, 100)

	hbox := HBox(Fixed(fixed), Expanded(flex))
	hbox.Layout(Rect{0, 0, 100, 100})

	if fixed.bounds.Width != 20 {
		t.Errorf("fixed width = %d, want 20", fixed.bounds.Width)
	}
	if flex.bounds.Width != 80 {
		t.Errorf("flex width = %d, want 80", flex.bounds.Width)
	}
}

func TestSpacer(t *testing.T) {
	left := newTestWidget(20, 50)
	right := newTestWidget(20, 50)

	hbox := HBox(Fixed(left), Space(), Fixed(right))
	hbox.Layout(Rect{0, 0, 100, 50})

	if left.bounds.X != 0 {
		t.Errorf("left X = %d, want 0", left.bounds.X)
	}
	if right.bounds.X != 80 {
		t.Errorf("right X = %d, want 80", right.bounds.X)
	}
}

func TestFixedSpace(t *testing.T) {
	w1 := newTestWidget(20, 100)
	w2 := newTestWidget(20, 100)

	hbox := HBox(Fixed(w1), FixedSpace(10), Fixed(w2))
	hbox.Layout(Rect{0, 0, 100, 100})

	if w1.bounds.X != 0 {
		t.Errorf("w1 X = %d, want 0", w1.bounds.X)
	}
	if w2.bounds.X != 30 {
		t.Errorf("w2 X = %d, want 30", w2.bounds.X)
	}
}

func TestFlex_EmptyContainer(t *testing.T) {
	vbox := VBox()

	size := vbox.Measure(Constraints{MinWidth: 10, MaxWidth: 100, MinHeight: 5, MaxHeight: 50})
	if size.Width != 10 || size.Height != 5 {
		t.Errorf("Empty VBox measure = %v, want {10,5}", size)
	}

	// Should not panic
	vbox.Layout(Rect{0, 0, 100, 100})
	vbox.Render(RenderContext{})
	vbox.HandleMessage(KeyMsg{})
}

func TestFlexChild_Sized(t *testing.T) {
	w := newTestWidget(100, 100)
	vbox := VBox(Sized(w, 30))

	vbox.Layout(Rect{0, 0, 100, 100})

	if w.bounds.Height != 30 {
		t.Errorf("Sized widget height = %d, want 30", w.bounds.Height)
	}
}

func TestFlex_Add(t *testing.T) {
	vbox := VBox()
	w := newTestWidget(100, 20)

	vbox.Add(Fixed(w))

	if len(vbox.Children) != 1 {
		t.Errorf("Add should add a child, got %d children", len(vbox.Children))
	}

	vbox.Layout(Rect{0, 0, 100, 100})
	if w.bounds.Height != 20 {
		t.Errorf("Added widget height = %d, want 20", w.bounds.Height)
	}
}

func TestFlex_HandleMessage(t *testing.T) {
	w1 := &handlingWidget{shouldHandle: false}
	w2 := &handlingWidget{shouldHandle: true}
	w3 := &handlingWidget{shouldHandle: false}

	vbox := VBox(Fixed(w1), Fixed(w2), Fixed(w3))

	result := vbox.HandleMessage(KeyMsg{})

	// w2 handles the message, so result should be handled
	if !result.Handled {
		t.Error("HandleMessage should return handled when a child handles it")
	}
	if !w2.handled {
		t.Error("w2 should have handled the message")
	}
	// w3 should not receive the message
	if w3.received {
		t.Error("w3 should not receive the message after w2 handled it")
	}
}

type handlingWidget struct {
	testWidget
	shouldHandle bool
	handled      bool
	received     bool
}

func (h *handlingWidget) HandleMessage(msg Message) HandleResult {
	h.received = true
	if h.shouldHandle {
		h.handled = true
		return Handled()
	}
	return Unhandled()
}

func TestFlex_HandleMessageNoneHandle(t *testing.T) {
	w1 := &handlingWidget{shouldHandle: false}
	w2 := &handlingWidget{shouldHandle: false}

	vbox := VBox(Fixed(w1), Fixed(w2))

	result := vbox.HandleMessage(KeyMsg{})

	if result.Handled {
		t.Error("HandleMessage should return unhandled when no child handles it")
	}
	// Both should receive the message
	if !w1.received || !w2.received {
		t.Error("All children should receive the message when none handle it")
	}
}

func TestHBox_MeasureWithGap(t *testing.T) {
	w1 := newTestWidget(20, 100)
	w2 := newTestWidget(30, 100)

	hbox := HBox(Fixed(w1), Fixed(w2)).WithGap(10)

	size := hbox.Measure(Loose(200, 100))
	// 20 + 30 + 10 = 60
	if size.Width != 60 {
		t.Errorf("Measure with gap width = %d, want 60", size.Width)
	}
}

func TestVBox_LayoutWithBasisColumn(t *testing.T) {
	w := newTestWidget(100, 100)
	vbox := VBox(Sized(w, 25))

	size := vbox.Measure(Loose(100, 100))
	if size.Height != 25 {
		t.Errorf("Measure with basis height = %d, want 25", size.Height)
	}
}

func TestHBox_LayoutWithBasisRow(t *testing.T) {
	w := newTestWidget(100, 100)
	hbox := HBox(Sized(w, 35))

	size := hbox.Measure(Loose(100, 100))
	if size.Width != 35 {
		t.Errorf("Measure with basis width = %d, want 35", size.Width)
	}
}

func TestVBox_AllFlexibleNegativeAvailable(t *testing.T) {
	// When fixed children exceed available space, flexible children get 0
	fixed := newTestWidget(100, 80)
	flex := newTestWidget(100, 20)

	vbox := VBox(Fixed(fixed), Expanded(flex))
	vbox.Layout(Rect{0, 0, 100, 50}) // Only 50 height, but fixed wants 80

	// flex should get 0 height (available space is negative)
	if flex.bounds.Height != 0 {
		t.Errorf("flex height = %d, want 0 when space is negative", flex.bounds.Height)
	}
}

func TestFlex_Render(t *testing.T) {
	w1 := newTestWidget(100, 20)
	w2 := newTestWidget(100, 20)

	vbox := VBox(Fixed(w1), Fixed(w2))
	vbox.Layout(Rect{0, 0, 100, 40})

	buf := NewBuffer(100, 40)
	ctx := RenderContext{Buffer: buf, Bounds: Rect{0, 0, 100, 40}}

	// Should not panic
	vbox.Render(ctx)
}

func TestFlex_RenderWithMissingBounds(t *testing.T) {
	w1 := newTestWidget(100, 20)
	vbox := VBox(Fixed(w1))
	// Don't call Layout, so childBounds is empty

	buf := NewBuffer(100, 40)
	ctx := RenderContext{Buffer: buf, Bounds: Rect{0, 0, 100, 40}}

	// Should not panic even with missing childBounds
	vbox.Render(ctx)
}

func TestSpacer_Methods(t *testing.T) {
	s := NewSpacer()

	// Measure should return min size
	size := s.Measure(Constraints{MinWidth: 5, MaxWidth: 100, MinHeight: 3, MaxHeight: 50})
	if size.Width != 5 || size.Height != 3 {
		t.Errorf("Spacer Measure = %v, want {5, 3}", size)
	}

	// Layout should not panic
	s.Layout(Rect{0, 0, 10, 10})

	// Render should not panic
	s.Render(RenderContext{})

	// HandleMessage should return unhandled
	result := s.HandleMessage(KeyMsg{})
	if result.Handled {
		t.Error("Spacer HandleMessage should return unhandled")
	}
}

func TestFlex_MeasureHBox(t *testing.T) {
	w1 := newTestWidget(20, 50)
	w2 := newTestWidget(30, 60)

	hbox := HBox(Fixed(w1), Fixed(w2))

	size := hbox.Measure(Loose(100, 100))
	// Width: 20 + 30 = 50, Height: max(50, 60) = 60
	if size.Width != 50 {
		t.Errorf("HBox Measure width = %d, want 50", size.Width)
	}
	if size.Height != 60 {
		t.Errorf("HBox Measure height = %d, want 60", size.Height)
	}
}

func TestFlex_MeasureVBox(t *testing.T) {
	w1 := newTestWidget(50, 20)
	w2 := newTestWidget(60, 30)

	vbox := VBox(Fixed(w1), Fixed(w2))

	size := vbox.Measure(Loose(100, 100))
	// Width: max(50, 60) = 60, Height: 20 + 30 = 50
	if size.Width != 60 {
		t.Errorf("VBox Measure width = %d, want 60", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("VBox Measure height = %d, want 50", size.Height)
	}
}
