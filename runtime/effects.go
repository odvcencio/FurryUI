package runtime

import (
	"context"
	"time"
)

// After posts a message after a delay.
func After(delay time.Duration, msg Message) Effect {
	return Effect{
		Run: func(ctx context.Context, post PostFunc) {
			if msg == nil || post == nil || delay <= 0 {
				if msg != nil && post != nil && delay <= 0 {
					post(msg)
				}
				return
			}
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				post(msg)
			}
		},
	}
}

// Every posts messages on a fixed interval.
// Returning nil from fn skips posting.
func Every(interval time.Duration, fn func(time.Time) Message) Effect {
	return Effect{
		Run: func(ctx context.Context, post PostFunc) {
			if interval <= 0 || fn == nil || post == nil {
				return
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case now := <-ticker.C:
					if msg := fn(now); msg != nil {
						post(msg)
					}
				}
			}
		},
	}
}
