package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/toast"
)

func TestToastStackDismissOnClick(t *testing.T) {
	stack := NewToastStack()
	var dismissed string
	stack.SetOnDismiss(func(id string) {
		dismissed = id
	})
	stack.SetToasts([]*toast.Toast{
		{ID: "toast-1", Level: toast.ToastInfo, Title: "Hello", Message: "World"},
	})

	stack.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})
	buf := runtime.NewBuffer(40, 10)
	stack.Render(runtime.RenderContext{Buffer: buf})

	if len(stack.toastRects) != 1 {
		t.Fatalf("expected 1 toast rect, got %d", len(stack.toastRects))
	}
	rect := stack.toastRects[0].bounds
	msg := runtime.MouseMsg{
		X:      rect.X + 1,
		Y:      rect.Y,
		Button: runtime.MouseLeft,
		Action: runtime.MouseRelease,
	}
	result := stack.HandleMessage(msg)
	if !result.Handled {
		t.Fatal("expected click to be handled")
	}
	if dismissed != "toast-1" {
		t.Fatalf("expected dismissed toast-1, got %q", dismissed)
	}
}

func TestToastStackUnhandledWithoutToasts(t *testing.T) {
	stack := NewToastStack()
	stack.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})
	result := stack.HandleMessage(runtime.MouseMsg{X: 1, Y: 1, Button: runtime.MouseLeft, Action: runtime.MouseRelease})
	if result.Handled {
		t.Fatal("expected unhandled without toasts")
	}
}
