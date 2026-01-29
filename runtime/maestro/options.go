package maestro

import "github.com/chuxorg/chux-agent-mesh/runtime/qa"

// Option configures Maestro behavior for QA and testing.
type Option func(*Maestro)

// WithRunIDGenerator overrides the run ID generator.
func WithRunIDGenerator(fn func() string) Option {
	return func(m *Maestro) {
		if fn != nil {
			m.runIDGenerator = fn
		}
	}
}

// WithObserver attaches a QA observer for boundary tracing.
func WithObserver(observer qa.Observer) Option {
	return func(m *Maestro) {
		if observer != nil {
			m.observer = observer
		}
	}
}

// NewWithOptions constructs a Maestro with optional QA overrides.
func NewWithOptions(enforcement EnforcementClient, outcomes OutcomePublisher, runStates RunStatePublisher, guidance GuidancePublisher, plans ExecutionPlanPublisher, opts ...Option) *Maestro {
	m := New(enforcement, outcomes, runStates, guidance, plans)
	for _, opt := range opts {
		opt(m)
	}
	if m.runIDGenerator == nil {
		m.runIDGenerator = generateRunID
	}
	if m.observer == nil {
		m.observer = qa.NopObserver{}
	}
	return m
}

func (m *Maestro) observe(name, sideEffect, detail string) {
	if m == nil || m.observer == nil {
		return
	}
	m.observer.Record(qa.BoundaryEvent{
		Name:       name,
		SideEffect: sideEffect,
		Detail:     detail,
	})
}
