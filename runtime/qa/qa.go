package qa

import "time"

// BoundaryEvent records a runtime boundary observation for QA inspection.
type BoundaryEvent struct {
	Name       string
	SideEffect string
	Detail     string
}

// Observer receives boundary observations.
type Observer interface {
	Record(BoundaryEvent)
}

// NopObserver is a no-op observer used by default.
type NopObserver struct{}

// Record implements Observer as a no-op.
func (NopObserver) Record(BoundaryEvent) {}

// Clock provides time primitives for deterministic tests.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

// SystemClock is the default wall-clock implementation.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time { return time.Now() }

// After waits for the duration to elapse.
func (SystemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// ResolveObserver returns a non-nil observer.
func ResolveObserver(obs Observer) Observer {
	if obs == nil {
		return NopObserver{}
	}
	return obs
}

// ResolveClock returns a non-nil clock.
func ResolveClock(clock Clock) Clock {
	if clock == nil {
		return SystemClock{}
	}
	return clock
}
