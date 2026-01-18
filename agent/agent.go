// Package agent provides AI-friendly interaction with FluffyUI applications.
// It enables automated testing, AI agents, and scripted interactions by exposing
// a semantic API over the widget tree rather than raw terminal I/O.
package agent

import (
	"encoding/json"
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
	ErrWidgetNotFound = errors.New("widget not found")
	ErrWidgetDisabled = errors.New("widget is disabled")
	ErrNotFocusable   = errors.New("widget is not focusable")
	ErrNotInteractive = errors.New("widget is not interactive")
	ErrTimeout        = errors.New("operation timed out")
	ErrNoApp          = errors.New("no app configured")
)

// Agent provides AI-friendly interaction with a FluffyUI application.
// It wraps a simulation backend and exposes semantic operations over
// the widget tree.
type Agent struct {
	mu       sync.Mutex
	app      *runtime.App
	sim      *sim.Backend
	screen   *runtime.Screen
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
	if top := a.screen.TopLayer(); top != nil && top.Root != nil {
		a.walkWidgets(top.Root, &snap.Widgets)
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
func (a *Agent) walkWidgets(w runtime.Widget, out *[]WidgetInfo) {
	if w == nil {
		return
	}

	info := a.extractWidgetInfo(w)

	// Check for children
	if cp, ok := w.(runtime.ChildProvider); ok {
		children := cp.ChildWidgets()
		for _, child := range children {
			a.walkWidgets(child, &info.Children)
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

	// Infer textbox role/value from focusable text widgets when accessibility is missing.
	if info.Role == "" {
		if f, ok := w.(runtime.Focusable); ok && f.CanFocus() {
			if textWidget, ok := w.(interface{ Text() string }); ok {
				info.Role = accessibility.RoleTextbox
				info.Value = textWidget.Text()
			}
		}
	}

	if info.Value == "" {
		if textWidget, ok := w.(interface{ Text() string }); ok {
			info.Value = textWidget.Text()
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

// FindByType is an alias for FindByRole.
func (a *Agent) FindByType(role accessibility.Role) []WidgetInfo {
	return a.FindByRole(role)
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

// SnapshotJSON returns the current snapshot serialized to JSON.
func (a *Agent) SnapshotJSON() ([]byte, error) {
	snap := a.Snapshot()
	return json.MarshalIndent(snap, "", "  ")
}

// FocusWidget focuses the widget with the given label.
func (a *Agent) FocusWidget(label string) error {
	return a.Focus(label)
}

// Focus moves focus to the widget with the given label.
func (a *Agent) Focus(label string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	return a.focusByID(info.ID)
}

// ActivateWidget activates the widget with the given label.
func (a *Agent) ActivateWidget(label string) error {
	return a.Activate(label)
}

// Activate focuses and activates the widget with the given label.
func (a *Agent) Activate(label string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	if err := a.focusByID(info.ID); err != nil {
		return err
	}
	if err := a.sendKey(terminal.KeyEnter, 0); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// TypeInto focuses the widget and types the given text.
func (a *Agent) TypeInto(label, text string) error {
	return a.Type(label, text)
}

// Type focuses the widget and types the given text.
func (a *Agent) Type(label, text string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	if err := a.focusByID(info.ID); err != nil {
		return err
	}
	if err := a.sendText(text); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// Select focuses the widget and selects the option by label.
func (a *Agent) Select(label, option string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}

	w, acc, err := a.focusWidgetByID(info.ID)
	if err != nil {
		return err
	}
	if acc == nil {
		return ErrNotInteractive
	}

	current := acc.AccessibleLabel()
	if strings.EqualFold(current, option) {
		return nil
	}

	seen := map[string]bool{current: true}
	for i := 0; i < 100; i++ {
		if err := a.sendKey(terminal.KeyDown, 0); err != nil {
			return err
		}
		a.Tick()
		current = acc.AccessibleLabel()
		if strings.EqualFold(current, option) {
			_ = w
			return nil
		}
		if seen[current] {
			break
		}
		seen[current] = true
	}
	return ErrWidgetNotFound
}

// SendKey injects a key into the app.
func (a *Agent) SendKey(key terminal.Key) error {
	if err := a.sendKey(key, 0); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// SendKeyRune injects a key with rune payload.
func (a *Agent) SendKeyRune(key terminal.Key, r rune) error {
	if err := a.sendKey(key, r); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// SendKeyString injects a string as a sequence of key events.
func (a *Agent) SendKeyString(text string) error {
	if err := a.sendText(text); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// WaitForText waits until text appears on screen or timeout occurs.
func (a *Agent) WaitForText(text string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.ContainsText(text) {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForWidget waits until a widget with the given label is present.
func (a *Agent) WaitForWidget(label string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.FindByLabel(label) != nil {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// ListWidgets returns widgets that match the given role.
func (a *Agent) ListWidgets(role accessibility.Role) []WidgetInfo {
	return a.FindByRole(role)
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

func (a *Agent) sendKey(key terminal.Key, r rune) error {
	if a == nil {
		return ErrNoApp
	}
	a.mu.Lock()
	simBackend := a.sim
	app := a.app
	a.mu.Unlock()

	if simBackend != nil {
		simBackend.InjectKey(key, r)
		return nil
	}
	if app != nil {
		app.Post(runtime.KeyMsg{Key: key, Rune: r})
		return nil
	}
	return ErrNoApp
}

func (a *Agent) sendText(text string) error {
	if a == nil {
		return ErrNoApp
	}
	for _, r := range text {
		switch r {
		case '\n':
			if err := a.sendKey(terminal.KeyEnter, 0); err != nil {
				return err
			}
		case '\t':
			if err := a.sendKey(terminal.KeyTab, 0); err != nil {
				return err
			}
		default:
			if err := a.sendKey(terminal.KeyRune, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Agent) focusByID(id string) error {
	_, _, err := a.focusWidgetByID(id)
	return err
}

func (a *Agent) focusWidgetByID(id string) (runtime.Widget, accessibility.Accessible, error) {
	a.mu.Lock()
	screen := a.screen
	a.mu.Unlock()
	if screen == nil {
		return nil, nil, ErrNoApp
	}
	layer := screen.TopLayer()
	if layer == nil || layer.Root == nil {
		return nil, nil, ErrNoApp
	}

	w := findWidgetByID(layer.Root, id)
	if w == nil {
		return nil, nil, ErrWidgetNotFound
	}
	focusable, ok := w.(runtime.Focusable)
	if !ok || !focusable.CanFocus() {
		return w, accessibleFromWidget(w), ErrNotFocusable
	}

	scope := screen.FocusScope()
	if scope == nil {
		return w, accessibleFromWidget(w), ErrNotFocusable
	}

	if !scope.SetFocus(focusable) {
		scope.Reset()
		runtime.RegisterFocusables(scope, layer.Root)
		if !scope.SetFocus(focusable) {
			return w, accessibleFromWidget(w), ErrNotFocusable
		}
	}

	return w, accessibleFromWidget(w), nil
}

func accessibleFromWidget(w runtime.Widget) accessibility.Accessible {
	if w == nil {
		return nil
	}
	if acc, ok := w.(accessibility.Accessible); ok {
		return acc
	}
	return nil
}

func findWidgetByID(w runtime.Widget, id string) runtime.Widget {
	if w == nil {
		return nil
	}
	if widgetID(w) == id {
		return w
	}
	if cp, ok := w.(runtime.ChildProvider); ok {
		for _, child := range cp.ChildWidgets() {
			if found := findWidgetByID(child, id); found != nil {
				return found
			}
		}
	}
	return nil
}
