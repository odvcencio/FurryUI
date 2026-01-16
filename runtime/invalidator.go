package runtime

import "sync/atomic"

// Invalidator posts an invalidate message with coalescing.
type Invalidator struct {
	post    func(Message) bool
	pending atomic.Bool
}

// NewInvalidator creates an invalidator wired to a post function.
func NewInvalidator(post func(Message) bool) *Invalidator {
	return &Invalidator{post: post}
}

// Invalidate requests a render pass.
func (i *Invalidator) Invalidate() {
	if i == nil || i.post == nil {
		return
	}
	if i.pending.CompareAndSwap(false, true) {
		if !i.post(InvalidateMsg{}) {
			i.pending.Store(false)
		}
	}
}

// Schedule runs fn and requests a render pass.
func (i *Invalidator) Schedule(fn func()) {
	if fn == nil {
		return
	}
	fn()
	i.Invalidate()
}

func (i *Invalidator) resetPending() {
	if i == nil {
		return
	}
	i.pending.Store(false)
}
