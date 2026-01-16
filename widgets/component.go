package widgets

import (
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/state"
)

// Component is a base widget with bound services and subscriptions.
type Component struct {
	Base
	Services runtime.Services
	Subs     state.Subscriptions
}

// Bind attaches app services to the component.
func (c *Component) Bind(services runtime.Services) {
	c.Services = services
	c.Subs.SetScheduler(services.Scheduler())
}

// Unbind releases app services and subscriptions.
func (c *Component) Unbind() {
	c.Subs.Clear()
	c.Services = runtime.Services{}
}

// Invalidate requests a render pass.
func (c *Component) Invalidate() {
	c.Services.Invalidate()
}

// Observe registers a subscription using the default scheduler.
func (c *Component) Observe(sub state.Subscribable, fn func()) {
	c.Subs.Observe(sub, fn)
}
