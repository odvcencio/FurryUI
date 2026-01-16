package state

import "sync"

// Subscriptions tracks and clears multiple unsubscribe callbacks.
type Subscriptions struct {
	mu     sync.Mutex
	unsubs []func()
	sched  Scheduler
}

// NewSubscriptions creates a Subscriptions with a default scheduler.
func NewSubscriptions(scheduler Scheduler) *Subscriptions {
	return &Subscriptions{sched: scheduler}
}

// SetScheduler updates the default scheduler.
func (s *Subscriptions) SetScheduler(scheduler Scheduler) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.sched = scheduler
	s.mu.Unlock()
}

// Scheduler returns the default scheduler.
func (s *Subscriptions) Scheduler() Scheduler {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	scheduler := s.sched
	s.mu.Unlock()
	return scheduler
}

// Add registers an unsubscribe callback.
func (s *Subscriptions) Add(unsub func()) {
	if s == nil || unsub == nil {
		return
	}
	s.mu.Lock()
	s.unsubs = append(s.unsubs, unsub)
	s.mu.Unlock()
}

// Subscribe registers a listener and tracks the unsubscribe.
func (s *Subscriptions) Subscribe(sub Subscribable, fn func()) {
	s.SubscribeWithScheduler(sub, nil, fn)
}

// Observe registers a listener using the default scheduler.
func (s *Subscriptions) Observe(sub Subscribable, fn func()) {
	if s == nil {
		return
	}
	scheduler := s.Scheduler()
	s.SubscribeWithScheduler(sub, scheduler, fn)
}

// SubscribeWithScheduler registers a listener using a scheduler and tracks it.
func (s *Subscriptions) SubscribeWithScheduler(sub Subscribable, scheduler Scheduler, fn func()) {
	if s == nil || sub == nil || fn == nil {
		return
	}
	var unsub func()
	if scheduler == nil {
		unsub = sub.Subscribe(fn)
	} else if sched, ok := sub.(interface {
		SubscribeWithScheduler(Scheduler, func()) func()
	}); ok {
		unsub = sched.SubscribeWithScheduler(scheduler, fn)
	} else {
		unsub = sub.Subscribe(fn)
	}
	s.Add(unsub)
}

// Clear unsubscribes all tracked callbacks.
func (s *Subscriptions) Clear() {
	if s == nil {
		return
	}
	s.mu.Lock()
	unsubs := s.unsubs
	s.unsubs = nil
	s.mu.Unlock()
	for _, unsub := range unsubs {
		if unsub != nil {
			unsub()
		}
	}
}
