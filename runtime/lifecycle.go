package runtime

// Lifecycle is implemented by widgets that need mount/unmount hooks.
type Lifecycle interface {
	Mount()
	Unmount()
}

// MountTree calls Mount on widgets that implement Lifecycle.
func MountTree(root Widget) {
	mountWidget(root)
}

// UnmountTree calls Unmount on widgets that implement Lifecycle.
func UnmountTree(root Widget) {
	unmountWidget(root)
}

func mountWidget(w Widget) {
	if w == nil {
		return
	}
	if m, ok := w.(Lifecycle); ok {
		m.Mount()
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			mountWidget(child)
		}
	}
}

func unmountWidget(w Widget) {
	if w == nil {
		return
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			unmountWidget(child)
		}
	}
	if m, ok := w.(Lifecycle); ok {
		m.Unmount()
	}
}
