package runtime

import (
	"context"
	"testing"
	"time"
)

func TestApp_HandleCommand_SendMsg(t *testing.T) {
	app := NewApp(AppConfig{})
	msg := ResizeMsg{Width: 10, Height: 5}

	if app.handleCommand(SendMsg{Message: msg}) {
		t.Fatalf("expected SendMsg to not force render")
	}

	select {
	case got := <-app.messages:
		if got != msg {
			t.Fatalf("unexpected message: %#v", got)
		}
	default:
		t.Fatal("expected message to be posted")
	}
}

func TestApp_HandleCommand_Effect(t *testing.T) {
	app := NewApp(AppConfig{})
	done := make(chan struct{})

	app.handleCommand(Effect{Run: func(ctx context.Context, post PostFunc) {
		post(ResizeMsg{Width: 1, Height: 2})
		close(done)
	}})

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("effect did not run")
	}

	select {
	case <-app.messages:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected effect to post a message")
	}
}

func TestApp_SpawnPendingEffect(t *testing.T) {
	app := NewApp(AppConfig{})
	ran := make(chan struct{}, 1)

	app.Spawn(Effect{Run: func(ctx context.Context, post PostFunc) {
		ran <- struct{}{}
	}})

	select {
	case <-ran:
		t.Fatal("expected pending effect to wait for start")
	default:
	}

	app.taskCtx = context.Background()
	app.startPendingEffects()

	select {
	case <-ran:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected pending effect to run")
	}
}
