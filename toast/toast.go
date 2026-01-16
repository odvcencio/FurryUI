package toast

import (
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// ToastLevel indicates the severity of a toast notification.
type ToastLevel string

const (
	ToastInfo    ToastLevel = "info"
	ToastSuccess ToastLevel = "success"
	ToastWarning ToastLevel = "warning"
	ToastError   ToastLevel = "error"
)

const (
	DefaultToastDuration = 4 * time.Second
	DefaultMaxToasts     = 5
)

// Toast represents a toast notification.
type Toast struct {
	ID        string
	Level     ToastLevel
	Title     string
	Message   string
	Duration  time.Duration
	CreatedAt time.Time
	Action    *ToastAction
}

// ToastAction represents an optional action for a toast.
type ToastAction struct {
	Label   string
	Command string
}

// ToastManager manages active toast notifications.
type ToastManager struct {
	mu       sync.RWMutex
	toasts   []*Toast
	timers   map[string]*time.Timer
	maxCount int
	onChange func([]*Toast)
}

// NewToastManager creates a new toast manager with default limits.
func NewToastManager() *ToastManager {
	return &ToastManager{
		maxCount: DefaultMaxToasts,
		timers:   make(map[string]*time.Timer),
	}
}

// SetOnChange configures the callback for toast updates.
func (tm *ToastManager) SetOnChange(fn func([]*Toast)) {
	if tm == nil {
		return
	}
	tm.mu.Lock()
	tm.onChange = fn
	snapshot := tm.snapshotLocked()
	tm.mu.Unlock()
	if fn != nil {
		fn(snapshot)
	}
}

// Show creates a new toast and returns its ID.
func (tm *ToastManager) Show(level ToastLevel, title, message string, duration time.Duration) string {
	if tm == nil {
		return ""
	}
	if duration <= 0 {
		duration = DefaultToastDuration
	}
	toast := &Toast{
		ID:        ulid.Make().String(),
		Level:     level,
		Title:     strings.TrimSpace(title),
		Message:   strings.TrimSpace(message),
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	tm.mu.Lock()
	if tm.timers == nil {
		tm.timers = make(map[string]*time.Timer)
	}
	if tm.maxCount <= 0 {
		tm.maxCount = DefaultMaxToasts
	}
	tm.toasts = append(tm.toasts, toast)
	if duration > 0 {
		tm.timers[toast.ID] = time.AfterFunc(duration, func() {
			tm.Dismiss(toast.ID)
		})
	}

	if overflow := len(tm.toasts) - tm.maxCount; overflow > 0 {
		for i := 0; i < overflow; i++ {
			removed := tm.toasts[0]
			tm.toasts = tm.toasts[1:]
			tm.stopTimerLocked(removed.ID)
		}
	}

	snapshot := tm.snapshotLocked()
	cb := tm.onChange
	tm.mu.Unlock()
	if cb != nil {
		cb(snapshot)
	}
	return toast.ID
}

// Info shows an informational toast.
func (tm *ToastManager) Info(title, msg string) {
	tm.Show(ToastInfo, title, msg, DefaultToastDuration)
}

// Success shows a success toast.
func (tm *ToastManager) Success(title, msg string) {
	tm.Show(ToastSuccess, title, msg, DefaultToastDuration)
}

// Warning shows a warning toast.
func (tm *ToastManager) Warning(title, msg string) {
	tm.Show(ToastWarning, title, msg, DefaultToastDuration)
}

// Error shows an error toast.
func (tm *ToastManager) Error(title, msg string) {
	tm.Show(ToastError, title, msg, DefaultToastDuration)
}

// Dismiss removes a toast by ID.
func (tm *ToastManager) Dismiss(id string) {
	if tm == nil || strings.TrimSpace(id) == "" {
		return
	}
	tm.mu.Lock()
	if len(tm.toasts) == 0 {
		tm.mu.Unlock()
		return
	}
	remaining := tm.toasts[:0]
	for _, t := range tm.toasts {
		if t.ID == id {
			tm.stopTimerLocked(id)
			continue
		}
		remaining = append(remaining, t)
	}
	tm.toasts = remaining
	snapshot := tm.snapshotLocked()
	cb := tm.onChange
	tm.mu.Unlock()
	if cb != nil {
		cb(snapshot)
	}
}

func (tm *ToastManager) stopTimerLocked(id string) {
	if tm.timers == nil {
		return
	}
	if timer, ok := tm.timers[id]; ok {
		timer.Stop()
		delete(tm.timers, id)
	}
}

func (tm *ToastManager) snapshotLocked() []*Toast {
	if tm == nil || len(tm.toasts) == 0 {
		return nil
	}
	out := make([]*Toast, len(tm.toasts))
	copy(out, tm.toasts)
	return out
}
