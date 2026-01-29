package testingagent

import (
	"time"

	"github.com/chuxorg/chux-agent-mesh/runtime/qa"
	"github.com/chuxorg/chux-agent-mesh/runtime/transport"
)

const defaultFinalizeGrace = 20 * time.Millisecond

// Options configures the testing agent for QA control.
type Options struct {
	Clock         qa.Clock
	Observer      qa.Observer
	FinalizeGrace time.Duration
}

// NewWithOptions creates a testing agent with optional QA hooks.
func NewWithOptions(bus *transport.Bus, opts Options) *Agent {
	clock := qa.ResolveClock(opts.Clock)
	observer := qa.ResolveObserver(opts.Observer)
	grace := opts.FinalizeGrace
	if grace <= 0 {
		grace = defaultFinalizeGrace
	}
	return &Agent{
		bus:           bus,
		runs:          make(map[string]*runTrace),
		finalizeCh:    make(chan string, defaultTestingAgentQueueSize),
		clock:         clock,
		observer:      observer,
		finalizeGrace: grace,
	}
}
