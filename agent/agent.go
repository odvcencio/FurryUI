// Package agent provides AI-friendly interaction with FluffyUI applications.
// It enables automated testing, AI agents, and scripted interactions by exposing
// a semantic API over the widget tree rather than raw terminal I/O.
package agent

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Common errors returned by Agent methods.
var (
	ErrWidgetNotFound  = errors.New("widget not found")
	ErrWidgetDisabled  = errors.New("widget is disabled")
	ErrNotFocusable    = errors.New("widget is not focusable")
	ErrNotInteractive  = errors.New("widget is not interactive")
	ErrTimeout         = errors.New("operation timed out")
	ErrNoApp           = errors.New("no app configured")
)

// Agent provides AI-friendly interaction with a FluffyUI application.
// It wraps a simulation backend and exposes semantic operations over
// the widget tree.
type Agent struct {
	mu      sync.Mutex
	app     *runtime.App
	sim     *sim.Backend
	screen  *runtime.Screen
	tickRate time.Duration
}

// Config configures an Agent.
type Config struct {
	// App is the FluffyUI application to control.
	App *runtime.App

	// Sim is the simulation backend. If nil, one will be created.
	Sim *sim.Backend

	// Width and Height set the terminal dimensions (default 80x24).
	Width, Height int

	// TickRate is how long to wait between operations for UI to settle.
	// Default is 50ms.
	TickRate time.Duration
}

// New creates a new Agent with the given configuration.
func New(cfg Config) *Agent {
	width, height := cfg.Width, cfg.Height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	s := cfg.Sim
	if s == nil {
		s = sim.New(width, height)
	}

	tickRate := cfg.TickRate
	if tickRate <= 0 {
		tickRate = 50 * time.Millisecond
	}

	return &Agent{
		app:      cfg.App,
		sim:      s,
		tickRate: tickRate,
	}
}

// Backend returns the underlying simulation backend.
func (a *Agent) Backend() *sim.Backend {
	if a == nil {
		return nil
	}
	return a.sim
}

// SetScreen sets the screen reference for widget tree access.
// This is typically called after App.Run() starts.
func (a *Agent) SetScreen(screen *runtime.Screen) {
	if a == nil {
		return
	}
	a.mu.Lock()
	a.screen = screen
	a.mu.Unlock()
}

// Screen returns the current screen.
func (a *Agent) Screen() *runtime.Screen {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.screen
}

// Tick waits for the UI to process pending events.
func (a *Agent) Tick() {
	if a == nil {
		return
	}
	time.Sleep(a.tickRate)
}

// Snapshot returns a structured representation of the current UI state.
func (a *Agent) Snapshot() Snapshot {
	if a == nil {
		return Snapshot{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	snap := Snapshot{
		Timestamp: time.Now(),
	}

	if a.sim != nil {
		snap.Text = a.sim.Capture()
		snap.Width, _ = a.sim.Size()
		_, snap.Height = a.sim.Size()
	}

	if a.screen == nil {
		return snap
	}

	snap.Width, snap.Height = a.screen.Size()
	snap.LayerCount = a.screen.LayerCount()

	// Walk widget tree
	if root := a.screen.Root(); root != nil {
		a.walkWidgets(root, &snap.Widgets, nil)
	}

	// Find focused widget
	if scope := a.screen.FocusScope(); scope != nil {
		if focused := scope.Current(); focused != nil {
			snap.FocusedID = widgetID(focused)
			for i := range snap.Widgets {
				if snap.Widgets[i].ID == snap.FocusedID {
					snap.Focused = &snap.Widgets[i]
					break
				}
			}
		}
	}

	return snap
}

// walkWidgets recursively collects widget info from the tree.
func (a *Agent) walkWidgets(w runtime.Widget, out *[]WidgetInfo, parent *WidgetInfo) {
	if w == nil {
		return
	}

	info := a.extractWidgetInfo(w)

	// Check for children
	if cp, ok := w.(runtime.ChildProvider); ok {
		children := cp.ChildWidgets()
		for _, child := range children {
			a.walkWidgets(child, &info.Children, &info)
		}
	}

	*out = append(*out, info)
}

// extractWidgetInfo builds WidgetInfo from a widget.
func (a *Agent) extractWidgetInfo(w runtime.Widget) WidgetInfo {
	info := WidgetInfo{
		ID: widgetID(w),
	}

	// Get bounds
	if bp, ok := w.(runtime.BoundsProvider); ok {
		info.Bounds = bp.Bounds()
	}

	// Get accessibility info
	if acc, ok := w.(accessibility.Accessible); ok {
		info.Role = acc.AccessibleRole()
		info.Label = acc.AccessibleLabel()
		info.Description = acc.AccessibleDescription()
		info.State = acc.AccessibleState()
		if val := acc.AccessibleValue(); val != nil {
			info.Value = val.Text
			info.ValueInfo = val
		}
	}

	// Check focusable
	if f, ok := w.(runtime.Focusable); ok {
		info.Focusable = f.CanFocus()
		info.Focused = f.IsFocused()
	}

	// Determine available actions based on role
	info.Actions = actionsForRole(info.Role, info.State)

	return info
}

// widgetID generates a unique identifier for a widget.
func widgetID(w runtime.Widget) string {
	if w == nil {
		return ""
	}
	// Use pointer address as unique ID
	return fmt.Sprintf("%p", w)
}

// actionsForRole returns available actions based on widget role and state.
func actionsForRole(role accessibility.Role, state accessibility.StateSet) []string {
	if state.Disabled {
		return nil
	}

	switch role {
	case accessibility.RoleButton:
		return []string{"activate", "focus"}
	case accessibility.RoleCheckbox, accessibility.RoleRadio:
		return []string{"toggle", "focus"}
	case accessibility.RoleTextbox:
		return []string{"type", "clear", "focus"}
	case accessibility.RoleList, accessibility.RoleTree:
		return []string{"select", "focus", "scroll"}
	case accessibility.RoleMenuItem:
		return []string{"activate"}
	case accessibility.RoleTab:
		return []string{"activate", "focus"}
	default:
		return []string{"focus"}
	}
}

// FindByLabel finds the first widget with a matching label (case-insensitive substring).
func (a *Agent) FindByLabel(label string) *WidgetInfo {
	snap := a.Snapshot()
	return findByLabelIn(snap.Widgets, label)
}

func findByLabelIn(widgets []WidgetInfo, label string) *WidgetInfo {
	label = strings.ToLower(label)
	for i := range widgets {
		w := &widgets[i]
		if strings.Contains(strings.ToLower(w.Label), label) {
			return w
		}
		if found := findByLabelIn(w.Children, label); found != nil {
			return found
		}
	}
	return nil
}

// FindByRole finds all widgets with the given role.
func (a *Agent) FindByRole(role accessibility.Role) []WidgetInfo {
	snap := a.Snapshot()
	var results []WidgetInfo
	findByRoleIn(snap.Widgets, role, &results)
	return results
}

func findByRoleIn(widgets []WidgetInfo, role accessibility.Role, out *[]WidgetInfo) {
	for _, w := range widgets {
		if w.Role == role {
			*out = append(*out, w)
		}
		findByRoleIn(w.Children, role, out)
	}
}

// FindByID finds a widget by its ID.
func (a *Agent) FindByID(id string) *WidgetInfo {
	snap := a.Snapshot()
	return findByIDIn(snap.Widgets, id)
}

func findByIDIn(widgets []WidgetInfo, id string) *WidgetInfo {
	for i := range widgets {
		w := &widgets[i]
		if w.ID == id {
			return w
		}
		if found := findByIDIn(w.Children, id); found != nil {
			return found
		}
	}
	return nil
}

// GetFocused returns the currently focused widget.
func (a *Agent) GetFocused() *WidgetInfo {
	snap := a.Snapshot()
	return snap.Focused
}

// IsFocused checks if a widget with the given label is focused.
func (a *Agent) IsFocused(label string) bool {
	w := a.FindByLabel(label)
	return w != nil && w.Focused
}

// IsEnabled checks if a widget with the given label is enabled.
func (a *Agent) IsEnabled(label string) bool {
	w := a.FindByLabel(label)
	return w != nil && !w.State.Disabled
}

// IsChecked checks if a checkbox/radio with the given label is checked.
func (a *Agent) IsChecked(label string) bool {
	w := a.FindByLabel(label)
	if w == nil || w.State.Checked == nil {
		return false
	}
	return *w.State.Checked
}

// GetValue returns the value of an input widget.
func (a *Agent) GetValue(label string) (string, error) {
	w := a.FindByLabel(label)
	if w == nil {
		return "", ErrWidgetNotFound
	}
	return w.Value, nil
}

// ContainsText checks if the given text appears on screen.
func (a *Agent) ContainsText(text string) bool {
	if a == nil || a.sim == nil {
		return false
	}
	return a.sim.ContainsText(text)
}

// FindText returns the position of text on screen, or (-1, -1) if not found.
func (a *Agent) FindText(text string) (x, y int) {
	if a == nil || a.sim == nil {
		return -1, -1
	}
	return a.sim.FindText(text)
}

// CaptureText returns the raw text content of the screen.
func (a *Agent) CaptureText() string {
	if a == nil || a.sim == nil {
		return ""
	}
	return a.sim.Capture()
}
