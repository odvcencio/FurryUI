package scroll

import (
	"image"
	"testing"

	"github.com/odvcencio/fluffy-ui/runtime"
)

func TestViewportClampOffset(t *testing.T) {
	v := NewViewport(nil)
	v.SetViewSize(runtime.Size{Width: 10, Height: 5})
	v.SetContentSize(runtime.Size{Width: 30, Height: 20})

	v.SetOffset(100, 100)
	if got := v.Offset(); got != (image.Point{X: 20, Y: 15}) {
		t.Fatalf("offset clamp = %+v, want %+v", got, image.Point{X: 20, Y: 15})
	}

	v.SetOffset(-5, -7)
	if got := v.Offset(); got != (image.Point{}) {
		t.Fatalf("offset clamp negative = %+v, want %+v", got, image.Point{})
	}
}

func TestViewportMaxOffsetAndVisibleRect(t *testing.T) {
	v := NewViewport(nil)
	v.SetViewSize(runtime.Size{Width: 10, Height: 5})
	v.SetContentSize(runtime.Size{Width: 8, Height: 4})
	if got := v.MaxOffset(); got != (image.Point{}) {
		t.Fatalf("max offset = %+v, want %+v", got, image.Point{})
	}

	v.SetContentSize(runtime.Size{Width: 30, Height: 20})
	v.SetOffset(4, 3)
	if got := v.VisibleRect(); got != (runtime.Rect{X: 4, Y: 3, Width: 10, Height: 5}) {
		t.Fatalf("visible rect = %+v, want %+v", got, runtime.Rect{X: 4, Y: 3, Width: 10, Height: 5})
	}
}

func TestFixedHeightIndex(t *testing.T) {
	index := FixedHeightIndex{
		Height: 2,
		Count: func() int {
			return 5
		},
	}
	if got := index.TotalHeight(); got != 10 {
		t.Fatalf("total height = %d, want 10", got)
	}
	if got := index.IndexForOffset(0); got != 0 {
		t.Fatalf("index for offset 0 = %d, want 0", got)
	}
	if got := index.IndexForOffset(9); got != 4 {
		t.Fatalf("index for offset 9 = %d, want 4", got)
	}
	if got := index.IndexForOffset(100); got != 4 {
		t.Fatalf("index for offset 100 = %d, want 4", got)
	}
	if got := index.OffsetForIndex(0); got != 0 {
		t.Fatalf("offset for index 0 = %d, want 0", got)
	}
	if got := index.OffsetForIndex(10); got != 8 {
		t.Fatalf("offset for index 10 = %d, want 8", got)
	}
}
