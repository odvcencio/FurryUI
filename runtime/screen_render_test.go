package runtime

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
)

type renderTestWidget struct {
	bounds Rect
}

func (w *renderTestWidget) Measure(constraints Constraints) Size {
	return Size{Width: 1, Height: 1}
}

func (w *renderTestWidget) Layout(bounds Rect) {
	w.bounds = bounds
}

func (w *renderTestWidget) Render(ctx RenderContext) {
	ctx.Buffer.Set(w.bounds.X, w.bounds.Y, 'x', backend.DefaultStyle())
}

func (w *renderTestWidget) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func TestScreen_RenderPreservesDirtyTracking(t *testing.T) {
	screen := NewScreen(4, 2)
	widget := &renderTestWidget{}
	screen.SetRoot(widget)

	screen.Render()
	buf := screen.Buffer()
	if buf.DirtyCount() == 0 {
		t.Fatal("expected dirty cells after initial render")
	}
	buf.ClearDirty()

	screen.Render()
	if buf.DirtyCount() != 0 {
		t.Fatalf("expected no dirty cells after stable render, got %d", buf.DirtyCount())
	}
}
