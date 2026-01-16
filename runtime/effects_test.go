package runtime

import (
	"context"
	"testing"
	"time"
)

func TestAfter_Immediate(t *testing.T) {
	calls := 0
	effect := After(0, ResizeMsg{Width: 1, Height: 1})
	effect.Run(context.Background(), func(Message) bool {
		calls++
		return true
	})
	if calls != 1 {
		t.Fatalf("expected immediate post, got %d", calls)
	}
}

func TestEvery_Invalid(t *testing.T) {
	calls := 0
	effect := Every(0, func(time.Time) Message { return ResizeMsg{Width: 1, Height: 1} })
	effect.Run(context.Background(), func(Message) bool {
		calls++
		return true
	})
	if calls != 0 {
		t.Fatalf("expected no posts for invalid interval, got %d", calls)
	}

	calls = 0
	effect = Every(10*time.Millisecond, nil)
	effect.Run(context.Background(), func(Message) bool {
		calls++
		return true
	})
	if calls != 0 {
		t.Fatalf("expected no posts for nil callback, got %d", calls)
	}
}
