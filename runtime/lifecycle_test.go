package runtime

import "testing"

type lifecycleWidget struct {
	children  []Widget
	mounted   int
	unmounted int
}

func (w *lifecycleWidget) Measure(constraints Constraints) Size {
	return Size{}
}

func (w *lifecycleWidget) Layout(bounds Rect) {}

func (w *lifecycleWidget) Render(ctx RenderContext) {}

func (w *lifecycleWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func (w *lifecycleWidget) ChildWidgets() []Widget {
	return w.children
}

func (w *lifecycleWidget) Mount() {
	w.mounted++
}

func (w *lifecycleWidget) Unmount() {
	w.unmounted++
}

func TestScreen_LifecycleRoot(t *testing.T) {
	child := &lifecycleWidget{}
	root := &lifecycleWidget{children: []Widget{child}}
	screen := NewScreen(10, 5)

	screen.SetRoot(root)
	if root.mounted != 1 || child.mounted != 1 {
		t.Fatalf("expected mounted calls root=1 child=1, got root=%d child=%d", root.mounted, child.mounted)
	}

	screen.SetRoot(nil)
	if root.unmounted != 1 || child.unmounted != 1 {
		t.Fatalf("expected unmounted calls root=1 child=1, got root=%d child=%d", root.unmounted, child.unmounted)
	}
}

func TestScreen_LifecycleLayer(t *testing.T) {
	root := &lifecycleWidget{}
	overlay := &lifecycleWidget{}
	screen := NewScreen(10, 5)

	screen.SetRoot(root)
	screen.PushLayer(overlay, true)
	if overlay.mounted != 1 {
		t.Fatalf("expected overlay mounted once, got %d", overlay.mounted)
	}

	if !screen.PopLayer() {
		t.Fatalf("expected PopLayer to succeed")
	}
	if overlay.unmounted != 1 {
		t.Fatalf("expected overlay unmounted once, got %d", overlay.unmounted)
	}
	if root.unmounted != 0 {
		t.Fatalf("expected root to remain mounted, got %d", root.unmounted)
	}
}
