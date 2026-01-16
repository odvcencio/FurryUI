package runtime

import "testing"

func TestShouldFlushQueue(t *testing.T) {
	cases := []struct {
		policy QueueFlushPolicy
		msg    Message
		want   bool
	}{
		{FlushManual, ResizeMsg{Width: 1, Height: 1}, false},
		{FlushManual, QueueFlushMsg{}, true},
		{FlushOnTick, TickMsg{}, true},
		{FlushOnTick, ResizeMsg{Width: 1, Height: 1}, false},
		{FlushOnTick, QueueFlushMsg{}, true},
		{FlushOnMessage, TickMsg{}, false},
		{FlushOnMessage, ResizeMsg{Width: 1, Height: 1}, true},
		{FlushOnMessageAndTick, TickMsg{}, true},
		{FlushOnMessageAndTick, ResizeMsg{Width: 1, Height: 1}, true},
	}

	for i, tc := range cases {
		if got := shouldFlushQueue(tc.policy, tc.msg); got != tc.want {
			t.Fatalf("case %d policy=%d msg=%T got %v want %v", i, tc.policy, tc.msg, got, tc.want)
		}
	}
}
