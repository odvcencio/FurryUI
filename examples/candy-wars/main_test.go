package main

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func renderViewToString(view *GameView, width, height int) string {
	buf := runtime.NewBuffer(width, height)
	ctx := runtime.RenderContext{
		Buffer: buf,
		Bounds: runtime.Rect{X: 0, Y: 0, Width: width, Height: height},
	}
	view.Render(ctx)

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

func TestTradeDialogRendersQuantityInput(t *testing.T) {
	game := NewGame()
	view := NewGameView(game)
	view.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})

	view.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'b'})
	view.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '9'})
	view.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '7'})

	output := renderViewToString(view, 80, 24)
	if !strings.Contains(output, "Qty: 97") {
		t.Fatalf("expected typed quantity to render in trade dialog\n\nOutput:\n%s", output)
	}
}
