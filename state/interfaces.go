package state

// Readable exposes read-only reactive state.
type Readable[T any] interface {
	Get() T
	Subscribe(fn func()) func()
	SubscribeWithScheduler(scheduler Scheduler, fn func()) func()
}

// Writable exposes read/write reactive state.
type Writable[T any] interface {
	Readable[T]
	Set(value T) bool
	Update(fn func(T) T) bool
}
