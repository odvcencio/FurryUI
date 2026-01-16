package runtime

// Bindable widgets receive app services when mounted into a screen.
type Bindable interface {
	Bind(services Services)
}

// Unbindable widgets release app services when removed.
type Unbindable interface {
	Unbind()
}

// BindTree calls Bind on widgets that implement Bindable.
func BindTree(root Widget, services Services) {
	if services.isZero() {
		return
	}
	bindWidget(root, services)
}

// UnbindTree calls Unbind on widgets that implement Unbindable.
func UnbindTree(root Widget) {
	unbindWidget(root)
}

func bindWidget(w Widget, services Services) {
	if w == nil {
		return
	}
	if b, ok := w.(Bindable); ok {
		b.Bind(services)
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			bindWidget(child, services)
		}
	}
}

func unbindWidget(w Widget) {
	if w == nil {
		return
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			unbindWidget(child)
		}
	}
	if u, ok := w.(Unbindable); ok {
		u.Unbind()
	}
}
