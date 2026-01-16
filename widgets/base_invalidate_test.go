package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/runtime"
)

func TestBase_InvalidateLifecycle(t *testing.T) {
	var b Base
	if b.NeedsRender() {
		t.Fatal("expected new base to not need render")
	}

	b.Invalidate()
	if !b.NeedsRender() {
		t.Fatal("expected base to need render after invalidate")
	}

	b.ClearInvalidation()
	if b.NeedsRender() {
		t.Fatal("expected base to clear invalidation")
	}
}

func TestBase_LayoutMarksRender(t *testing.T) {
	var b Base
	bounds := runtime.Rect{X: 1, Y: 2, Width: 3, Height: 4}

	b.Layout(bounds)
	if !b.NeedsRender() {
		t.Fatal("expected layout to mark render")
	}

	b.ClearInvalidation()
	b.Layout(bounds)
	if b.NeedsRender() {
		t.Fatal("expected layout with same bounds to not mark render")
	}

	b.Layout(runtime.Rect{X: 2, Y: 2, Width: 3, Height: 4})
	if !b.NeedsRender() {
		t.Fatal("expected layout with new bounds to mark render")
	}
}
