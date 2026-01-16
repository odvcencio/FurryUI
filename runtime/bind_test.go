package runtime

import "testing"

type bindTestWidget struct {
	children []Widget
	bound    int
	unbound  int
}

func (b *bindTestWidget) Measure(c Constraints) Size { return Size{} }
func (b *bindTestWidget) Layout(bounds Rect)         {}
func (b *bindTestWidget) Render(ctx RenderContext)   {}
func (b *bindTestWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}
func (b *bindTestWidget) ChildWidgets() []Widget { return b.children }
func (b *bindTestWidget) Bind(services Services) { b.bound++ }
func (b *bindTestWidget) Unbind()                { b.unbound++ }

func TestScreen_BindRoot(t *testing.T) {
	child := &bindTestWidget{}
	root := &bindTestWidget{children: []Widget{child}}
	screen := NewScreen(10, 5)
	app := NewApp(AppConfig{})
	screen.SetServices(app.Services())

	screen.SetRoot(root)
	if root.bound != 1 || child.bound != 1 {
		t.Fatalf("expected bind calls root=1 child=1, got root=%d child=%d", root.bound, child.bound)
	}

	screen.SetRoot(nil)
	if root.unbound != 1 || child.unbound != 1 {
		t.Fatalf("expected unbind calls root=1 child=1, got root=%d child=%d", root.unbound, child.unbound)
	}
}

func TestScreen_BindLayer(t *testing.T) {
	root := &bindTestWidget{}
	overlay := &bindTestWidget{}
	screen := NewScreen(10, 5)
	app := NewApp(AppConfig{})
	screen.SetServices(app.Services())

	screen.SetRoot(root)
	screen.PushLayer(overlay, true)
	if overlay.bound != 1 {
		t.Fatalf("expected overlay bound once, got %d", overlay.bound)
	}

	if !screen.PopLayer() {
		t.Fatalf("expected PopLayer to succeed")
	}
	if overlay.unbound != 1 {
		t.Fatalf("expected overlay unbound once, got %d", overlay.unbound)
	}
	if root.unbound != 0 {
		t.Fatalf("expected root to remain bound, got %d", root.unbound)
	}
}
