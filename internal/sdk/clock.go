package sdk

import (
	"sync"
	"time"
)

// Clock exposes the minimal time dependency agents are allowed to observe.
type Clock interface {
	Now() time.Time
}

// DeterministicClock is a controllable clock used by the harness for repeatable tests.
type DeterministicClock struct {
	mu  sync.Mutex
	now time.Time
}

// NewDeterministicClock constructs a clock fixed at start.
func NewDeterministicClock(start time.Time) *DeterministicClock {
	return &DeterministicClock{now: start}
}

// Now returns the deterministic instant.
func (c *DeterministicClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forwards by d.
func (c *DeterministicClock) Advance(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return c.now
}
