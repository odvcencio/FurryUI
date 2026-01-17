package widgets

import (
	"math"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/toast"
)

const (
	toastMaxWidth = 60
	toastPaddingX = 1
	toastSpacing  = 1
	toastMargin   = 1
	toastMinWidth = 20
	toastSlideMs  = 150
	toastFadeMs   = 200
	toastSlideOff = 1
)

type toastRect struct {
	id     string
	bounds runtime.Rect
	toast  *toast.Toast
}

// ToastStack renders toast notifications.
type ToastStack struct {
	Base
	toasts     []*toast.Toast
	onDismiss  func(id string)
	toastRects []toastRect
	now        time.Time
	animate    bool

	bgStyle      backend.Style
	textStyle    backend.Style
	infoStyle    backend.Style
	successStyle backend.Style
	warnStyle    backend.Style
	errorStyle   backend.Style
}

// NewToastStack creates a new toast stack widget.
func NewToastStack() *ToastStack {
	return &ToastStack{
		bgStyle:      backend.DefaultStyle(),
		textStyle:    backend.DefaultStyle(),
		infoStyle:    backend.DefaultStyle(),
		successStyle: backend.DefaultStyle(),
		warnStyle:    backend.DefaultStyle(),
		errorStyle:   backend.DefaultStyle(),
		animate:      true,
	}
}

// SetToasts updates the toast list.
func (t *ToastStack) SetToasts(toasts []*toast.Toast) {
	t.toasts = toasts
}

// SetOnDismiss registers a handler for dismiss actions.
func (t *ToastStack) SetOnDismiss(fn func(id string)) {
	t.onDismiss = fn
}

// SetNow updates the animation timestamp.
func (t *ToastStack) SetNow(now time.Time) {
	t.now = now
}

// SetAnimationsEnabled toggles toast animations.
func (t *ToastStack) SetAnimationsEnabled(enabled bool) {
	t.animate = enabled
}

// SetStyles configures the toast styles by level.
func (t *ToastStack) SetStyles(bg, text, info, success, warn, err backend.Style) {
	t.bgStyle = bg
	t.textStyle = text
	t.infoStyle = info
	t.successStyle = success
	t.warnStyle = warn
	t.errorStyle = err
}

// Measure fills the available space.
func (t *ToastStack) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight}
}

// Render draws the toast stack.
func (t *ToastStack) Render(ctx runtime.RenderContext) {
	bounds := t.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}

	t.toastRects = t.toastRects[:0]
	if len(t.toasts) == 0 {
		return
	}

	availableWidth := bounds.Width - 2*toastMargin
	if availableWidth <= 0 {
		return
	}
	maxWidth := min(toastMaxWidth, availableWidth)
	if maxWidth < toastMinWidth {
		maxWidth = availableWidth
	}
	if maxWidth <= 0 {
		return
	}

	now := t.now
	if now.IsZero() {
		now = time.Now()
	}
	slideDuration := time.Duration(toastSlideMs) * time.Millisecond
	fadeDuration := time.Duration(toastFadeMs) * time.Millisecond
	y := bounds.Y + bounds.Height - 1 - toastMargin
	for i := len(t.toasts) - 1; i >= 0; i-- {
		toast := t.toasts[i]
		if toast == nil {
			continue
		}
		lines, prefix := t.toastLines(toast, maxWidth-2*toastPaddingX)
		if len(lines) == 0 {
			continue
		}
		width := maxLineLen(lines) + 2*toastPaddingX
		if width < toastMinWidth {
			width = toastMinWidth
		}
		if width > maxWidth {
			width = maxWidth
		}
		height := len(lines)

		yStart := y - height + 1
		age := now.Sub(toast.CreatedAt)
		remaining := toast.Duration - age
		slideOffset := 0
		fade := false
		if t.animate && !toast.CreatedAt.IsZero() {
			if age < slideDuration {
				progress := float64(age) / float64(slideDuration)
				slideOffset = int(math.Round(float64(toastSlideOff) * (1 - progress)))
				if slideOffset < 0 {
					slideOffset = 0
				}
			}
			if remaining > 0 && remaining < fadeDuration {
				fade = true
			}
		}
		yStart += slideOffset
		if yStart < bounds.Y {
			break
		}
		x := bounds.X + bounds.Width - width - toastMargin
		rect := runtime.Rect{X: x, Y: yStart, Width: width, Height: height}
		t.toastRects = append(t.toastRects, toastRect{id: toast.ID, bounds: rect, toast: toast})

		for lineIdx, line := range lines {
			row := runtime.Rect{X: rect.X, Y: rect.Y + lineIdx, Width: rect.Width, Height: 1}
			bgStyle := t.bgStyle
			textStyle := t.textStyle
			infoStyle := t.infoStyle
			successStyle := t.successStyle
			warnStyle := t.warnStyle
			errorStyle := t.errorStyle
			if fade {
				bgStyle = bgStyle.Dim(true)
				textStyle = textStyle.Dim(true)
				infoStyle = infoStyle.Dim(true)
				successStyle = successStyle.Dim(true)
				warnStyle = warnStyle.Dim(true)
				errorStyle = errorStyle.Dim(true)
			}
			ctx.Buffer.Fill(row, ' ', bgStyle)
			if line == "" {
				continue
			}
			startX := rect.X + toastPaddingX
			if lineIdx == 0 && prefix != "" {
				ctx.Buffer.SetString(startX, row.Y, prefix, levelStyle(toast.Level, infoStyle, successStyle, warnStyle, errorStyle))
				ctx.Buffer.SetString(startX+len(prefix), row.Y, line[len(prefix):], textStyle)
			} else {
				ctx.Buffer.SetString(startX, row.Y, line, textStyle)
			}
		}

		y = yStart - toastSpacing
		if y < bounds.Y {
			break
		}
	}
}

// HandleMessage handles dismiss clicks.
func (t *ToastStack) HandleMessage(msg runtime.Message) runtime.HandleResult {
	mouse, ok := msg.(runtime.MouseMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if mouse.Action != runtime.MouseRelease || mouse.Button != runtime.MouseLeft {
		return runtime.Unhandled()
	}
	for _, rect := range t.toastRects {
		if rect.bounds.Contains(mouse.X, mouse.Y) {
			if t.onDismiss != nil {
				t.onDismiss(rect.id)
			}
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// ToastAt returns the toast under the given point.
func (t *ToastStack) ToastAt(x, y int) (*toast.Toast, bool) {
	for _, rect := range t.toastRects {
		if rect.bounds.Contains(x, y) && rect.toast != nil {
			return rect.toast, true
		}
	}
	return nil, false
}

// HasActiveAnimations returns true when any toast is animating.
func (t *ToastStack) HasActiveAnimations(now time.Time) bool {
	if !t.animate || len(t.toasts) == 0 {
		return false
	}
	slideDuration := time.Duration(toastSlideMs) * time.Millisecond
	fadeDuration := time.Duration(toastFadeMs) * time.Millisecond
	for _, toast := range t.toasts {
		if toast == nil || toast.CreatedAt.IsZero() {
			continue
		}
		age := now.Sub(toast.CreatedAt)
		if age < slideDuration {
			return true
		}
		remaining := toast.Duration - age
		if remaining > 0 && remaining < fadeDuration {
			return true
		}
	}
	return false
}

func (t *ToastStack) toastLines(toast *toast.Toast, maxWidth int) ([]string, string) {
	title := strings.TrimSpace(toast.Title)
	if title == "" {
		title = levelLabel(toast.Level)
	}
	prefix := levelIcon(toast.Level) + " "
	contentWidth := maxWidth - len(prefix)
	if contentWidth < 0 {
		contentWidth = 0
	}
	titleLine := prefix + truncateString(title, contentWidth)

	lines := []string{titleLine}

	message := strings.TrimSpace(toast.Message)
	if message != "" {
		if toast.Action != nil && strings.TrimSpace(toast.Action.Label) != "" {
			message = message + " [" + strings.TrimSpace(toast.Action.Label) + "]"
		}
		lines = append(lines, truncateString(message, maxWidth))
	}

	return lines, prefix
}

func levelStyle(level toast.ToastLevel, info, success, warn, err backend.Style) backend.Style {
	switch level {
	case toast.ToastSuccess:
		return success
	case toast.ToastWarning:
		return warn
	case toast.ToastError:
		return err
	default:
		return info
	}
}

func levelLabel(level toast.ToastLevel) string {
	switch level {
	case toast.ToastSuccess:
		return "Success"
	case toast.ToastWarning:
		return "Warning"
	case toast.ToastError:
		return "Error"
	default:
		return "Info"
	}
}

func levelIcon(level toast.ToastLevel) string {
	switch level {
	case toast.ToastSuccess:
		return "+"
	case toast.ToastWarning:
		return "!"
	case toast.ToastError:
		return "x"
	default:
		return "i"
	}
}

func maxLineLen(lines []string) int {
	max := 0
	for _, line := range lines {
		if len(line) > max {
			max = len(line)
		}
	}
	return max
}
