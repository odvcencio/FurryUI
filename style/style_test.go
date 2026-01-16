package style

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/compositor"
)

func TestToBackend_DefaultColors(t *testing.T) {
	cs := compositor.Style{
		FG: compositor.ColorNone,
		BG: compositor.ColorNone,
	}

	style := ToBackend(cs)
	fg, bg, _ := style.Decompose()

	if fg != backend.ColorDefault {
		t.Fatalf("FG = %v, want default", fg)
	}
	if bg != backend.ColorDefault {
		t.Fatalf("BG = %v, want default", bg)
	}
}

func TestToBackend_ColorsAndAttrs(t *testing.T) {
	cs := compositor.DefaultStyle().
		WithFG(compositor.RGB(10, 20, 30)).
		WithBG(compositor.Color256(7)).
		WithBold(true).
		WithItalic(true).
		WithUnderline(true).
		WithDim(true)
	cs.Blink = true
	cs.Reverse = true
	cs.Strikethrough = true

	style := ToBackend(cs)
	fg, bg, attrs := style.Decompose()

	if !fg.IsRGB() {
		t.Fatalf("FG expected RGB, got %v", fg)
	}
	r, g, b := fg.RGB()
	if r != 10 || g != 20 || b != 30 {
		t.Fatalf("FG RGB = %d,%d,%d, want 10,20,30", r, g, b)
	}
	if bg != backend.Color(7) {
		t.Fatalf("BG = %v, want 7", bg)
	}

	expectedAttrs := backend.AttrBold | backend.AttrItalic | backend.AttrUnderline |
		backend.AttrDim | backend.AttrBlink | backend.AttrReverse | backend.AttrStrikeThrough
	if attrs&expectedAttrs != expectedAttrs {
		t.Fatalf("attrs = %v, want %v", attrs, expectedAttrs)
	}
}
