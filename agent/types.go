package agent

import (
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Snapshot captures a structured view of the current UI state.
type Snapshot struct {
	Timestamp  time.Time    `json:"timestamp"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	LayerCount int          `json:"layer_count,omitempty"`
	Text       string       `json:"text,omitempty"`
	Widgets    []WidgetInfo `json:"widgets,omitempty"`
	FocusedID  string       `json:"focused_id,omitempty"`
	Focused    *WidgetInfo  `json:"focused,omitempty"`
}

// WidgetInfo describes a widget in the UI tree.
type WidgetInfo struct {
	ID          string                   `json:"id"`
	Role        accessibility.Role       `json:"type"`
	Label       string                   `json:"label,omitempty"`
	Description string                   `json:"description,omitempty"`
	Value       string                   `json:"value,omitempty"`
	ValueInfo   *accessibility.ValueInfo `json:"value_info,omitempty"`
	State       accessibility.StateSet   `json:"state,omitempty"`
	Bounds      runtime.Rect             `json:"bounds"`
	Children    []WidgetInfo             `json:"children,omitempty"`
	Actions     []string                 `json:"actions,omitempty"`
	Focusable   bool                     `json:"focusable,omitempty"`
	Focused     bool                     `json:"focused,omitempty"`
}
