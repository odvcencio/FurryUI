package runtime

import (
	"time"

	"github.com/odvcencio/furry-ui/state"
)

// Services exposes app-level scheduling and messaging helpers.
type Services struct {
	app *App
}

// Services returns a service handle for the app.
func (a *App) Services() Services {
	return Services{app: a}
}

func (s Services) isZero() bool {
	return s.app == nil
}

// Scheduler returns the app state scheduler.
func (s Services) Scheduler() state.Scheduler {
	if s.app == nil {
		return nil
	}
	return s.app.StateScheduler()
}

// InvalidateScheduler returns the app invalidation scheduler.
func (s Services) InvalidateScheduler() state.Scheduler {
	if s.app == nil {
		return nil
	}
	return s.app.InvalidateScheduler()
}

// Invalidate requests a render pass.
func (s Services) Invalidate() {
	if s.app == nil {
		return
	}
	s.app.Invalidate()
}

// Post sends a message into the app loop.
func (s Services) Post(msg Message) bool {
	if s.app == nil {
		return false
	}
	return s.app.tryPost(msg)
}

// Spawn starts an effect using the app task context.
func (s Services) Spawn(effect Effect) {
	if s.app == nil {
		return
	}
	s.app.Spawn(effect)
}

// After schedules a delayed message.
func (s Services) After(delay time.Duration, msg Message) {
	if s.app == nil {
		return
	}
	s.app.After(delay, msg)
}

// Every schedules a recurring message.
func (s Services) Every(interval time.Duration, fn func(time.Time) Message) {
	if s.app == nil {
		return
	}
	s.app.Every(interval, fn)
}
