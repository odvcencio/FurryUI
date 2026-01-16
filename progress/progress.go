package progress

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// ProgressType describes how progress should be presented.
type ProgressType string

const (
	ProgressIndeterminate ProgressType = "indeterminate"
	ProgressDeterminate   ProgressType = "determinate"
	ProgressSteps         ProgressType = "steps"
)

// Progress captures the current state of a long-running task.
type Progress struct {
	ID          string
	Type        ProgressType
	Label       string
	Current     int
	Total       int
	Percent     float64
	Cancellable bool
	StartedAt   time.Time
}

// ProgressManager tracks active progress entries.
type ProgressManager struct {
	mu       sync.RWMutex
	active   map[string]*Progress
	onChange func([]Progress)
}

// NewProgressManager creates a new manager.
func NewProgressManager() *ProgressManager {
	return &ProgressManager{
		active: make(map[string]*Progress),
	}
}

// SetOnChange configures the callback for progress updates.
func (pm *ProgressManager) SetOnChange(fn func([]Progress)) {
	if pm == nil {
		return
	}
	pm.mu.Lock()
	pm.onChange = fn
	snapshot := pm.snapshotLocked()
	pm.mu.Unlock()
	if fn != nil {
		fn(snapshot)
	}
}

// Start begins tracking a progress entry.
func (pm *ProgressManager) Start(id, label string, ptype ProgressType, total int) {
	if pm == nil || strings.TrimSpace(id) == "" {
		return
	}
	if ptype == "" {
		ptype = ProgressIndeterminate
	}
	if total < 0 {
		total = 0
	}
	now := time.Now()
	pm.mu.Lock()
	if pm.active == nil {
		pm.active = make(map[string]*Progress)
	}
	pm.active[id] = &Progress{
		ID:        id,
		Type:      ptype,
		Label:     strings.TrimSpace(label),
		Current:   0,
		Total:     total,
		Percent:   computePercent(0, total),
		StartedAt: now,
	}
	snapshot := pm.snapshotLocked()
	cb := pm.onChange
	pm.mu.Unlock()
	if cb != nil {
		cb(snapshot)
	}
}

// Update updates the current value for a progress entry.
func (pm *ProgressManager) Update(id string, current int) {
	if pm == nil || strings.TrimSpace(id) == "" {
		return
	}
	pm.mu.Lock()
	if pm.active == nil {
		pm.mu.Unlock()
		return
	}
	entry, ok := pm.active[id]
	if !ok {
		pm.mu.Unlock()
		return
	}
	if current < 0 {
		current = 0
	}
	entry.Current = current
	entry.Percent = computePercent(entry.Current, entry.Total)
	snapshot := pm.snapshotLocked()
	cb := pm.onChange
	pm.mu.Unlock()
	if cb != nil {
		cb(snapshot)
	}
}

// Done removes a progress entry.
func (pm *ProgressManager) Done(id string) {
	if pm == nil || strings.TrimSpace(id) == "" {
		return
	}
	pm.mu.Lock()
	if pm.active == nil {
		pm.mu.Unlock()
		return
	}
	delete(pm.active, id)
	snapshot := pm.snapshotLocked()
	cb := pm.onChange
	pm.mu.Unlock()
	if cb != nil {
		cb(snapshot)
	}
}

func computePercent(current, total int) float64 {
	if total <= 0 {
		return 0
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	return float64(current) / float64(total)
}

func (pm *ProgressManager) snapshotLocked() []Progress {
	if pm == nil || len(pm.active) == 0 {
		return nil
	}
	out := make([]Progress, 0, len(pm.active))
	for _, entry := range pm.active {
		out = append(out, *entry)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.Before(out[j].StartedAt)
	})
	return out
}
