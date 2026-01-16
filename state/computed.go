package state

import "sync"

// Computed derives its value from other signals.
type Computed[T any] struct {
	signal    *Signal[T]
	compute   func() T
	mu        sync.Mutex
	unsubs    []func()
	scheduler Scheduler
}

// NewComputed creates a derived value from dependencies.
func NewComputed[T any](compute func() T, deps ...Subscribable) *Computed[T] {
	return NewComputedWithScheduler(nil, compute, deps...)
}

// NewComputedWithScheduler creates a derived value and schedules recomputes.
func NewComputedWithScheduler[T any](scheduler Scheduler, compute func() T, deps ...Subscribable) *Computed[T] {
	if compute == nil {
		compute = func() T {
			var zero T
			return zero
		}
	}
	c := &Computed[T]{
		signal:    NewSignal(compute()),
		compute:   compute,
		scheduler: scheduler,
	}
	for _, dep := range deps {
		if dep == nil {
			continue
		}
		unsub := dep.Subscribe(c.enqueueRecompute)
		if unsub != nil {
			c.unsubs = append(c.unsubs, unsub)
		}
	}
	return c
}

// SetEqualFunc configures the equality check used to suppress redundant updates.
func (c *Computed[T]) SetEqualFunc(fn EqualFunc[T]) {
	if c == nil {
		return
	}
	c.signal.SetEqualFunc(fn)
}

// Get returns the current computed value.
func (c *Computed[T]) Get() T {
	if c == nil {
		var zero T
		return zero
	}
	return c.signal.Get()
}

// Subscribe registers a listener for change notifications.
func (c *Computed[T]) Subscribe(fn func()) func() {
	if c == nil {
		return func() {}
	}
	return c.signal.Subscribe(fn)
}

// SubscribeWithScheduler registers a listener using a scheduler.
// If scheduler is nil, callbacks run synchronously.
func (c *Computed[T]) SubscribeWithScheduler(scheduler Scheduler, fn func()) func() {
	if c == nil {
		return func() {}
	}
	return c.signal.SubscribeWithScheduler(scheduler, fn)
}

// Stop unsubscribes from dependency updates.
func (c *Computed[T]) Stop() {
	if c == nil {
		return
	}
	c.mu.Lock()
	unsubs := c.unsubs
	c.unsubs = nil
	c.mu.Unlock()
	for _, unsub := range unsubs {
		if unsub != nil {
			unsub()
		}
	}
}

func (c *Computed[T]) recompute() {
	if c == nil {
		return
	}
	c.signal.Set(c.compute())
}

func (c *Computed[T]) enqueueRecompute() {
	if c == nil {
		return
	}
	if c.scheduler == nil {
		c.recompute()
		return
	}
	c.scheduler.Schedule(c.recompute)
}
