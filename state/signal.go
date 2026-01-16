// Package state provides minimal reactive primitives for terminal UIs.
package state

import "sync"

// EqualFunc compares two values for equality.
type EqualFunc[T any] func(a, b T) bool

// EqualComparable compares comparable values with ==.
func EqualComparable[T comparable](a, b T) bool {
	return a == b
}

// Subscribable emits change notifications.
type Subscribable interface {
	Subscribe(fn func()) func()
}

type subscriber struct {
	fn        func()
	scheduler Scheduler
}

// Signal holds a value and notifies subscribers on change.
type Signal[T any] struct {
	mu    sync.Mutex
	value T
	subs  map[int]subscriber
	next  int
	equal EqualFunc[T]
}

// NewSignal creates a new signal with an initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{value: initial}
}

// SetEqualFunc configures the equality check used to suppress redundant updates.
func (s *Signal[T]) SetEqualFunc(fn EqualFunc[T]) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.equal = fn
	s.mu.Unlock()
}

// Get returns the current value.
func (s *Signal[T]) Get() T {
	if s == nil {
		var zero T
		return zero
	}
	s.mu.Lock()
	value := s.value
	s.mu.Unlock()
	return value
}

// Set updates the value and notifies subscribers if it changed.
func (s *Signal[T]) Set(value T) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	if s.equal != nil && s.equal(s.value, value) {
		s.mu.Unlock()
		return false
	}
	s.value = value
	subs := s.copySubscribersLocked()
	s.mu.Unlock()

	s.notify(subs)
	return true
}

// Update replaces the value using fn.
// fn runs outside the signal lock; Update is not atomic across goroutines.
func (s *Signal[T]) Update(fn func(T) T) bool {
	if s == nil || fn == nil {
		return false
	}
	current := s.Get()
	next := fn(current)
	return s.Set(next)
}

// Subscribe registers a listener for change notifications.
func (s *Signal[T]) Subscribe(fn func()) func() {
	return s.SubscribeWithScheduler(nil, fn)
}

// SubscribeWithScheduler registers a listener using a scheduler.
// If scheduler is nil, callbacks run synchronously.
func (s *Signal[T]) SubscribeWithScheduler(scheduler Scheduler, fn func()) func() {
	if s == nil || fn == nil {
		return func() {}
	}
	s.mu.Lock()
	if s.subs == nil {
		s.subs = make(map[int]subscriber)
	}
	id := s.next
	s.next++
	s.subs[id] = subscriber{fn: fn, scheduler: scheduler}
	s.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			delete(s.subs, id)
			s.mu.Unlock()
		})
	}
}

func (s *Signal[T]) copySubscribersLocked() []subscriber {
	if len(s.subs) == 0 {
		return nil
	}
	subs := make([]subscriber, 0, len(s.subs))
	for _, sub := range s.subs {
		subs = append(subs, sub)
	}
	return subs
}

func (s *Signal[T]) notify(subs []subscriber) {
	for _, sub := range subs {
		if sub.fn == nil {
			continue
		}
		if sub.scheduler == nil {
			sub.fn()
			continue
		}
		sub.scheduler.Schedule(sub.fn)
	}
}
