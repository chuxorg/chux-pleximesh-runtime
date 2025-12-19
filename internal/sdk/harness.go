package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentFactory constructs an agent wired to the provided base config.
type AgentFactory func(cfg BaseConfig) (Agent, error)

// HarnessOptions control how the harness instantiates the agent.
type HarnessOptions struct {
	Identity     AgentIdentity
	Capabilities CapabilitySet
	StartTime    time.Time
}

// LifecycleStage enumerates the lifecycle transitions captured by the harness.
type LifecycleStage string

const (
	StageInit     LifecycleStage = "init"
	StageStart    LifecycleStage = "start"
	StageHandle   LifecycleStage = "handle"
	StageShutdown LifecycleStage = "shutdown"
)

// LifecycleTransition captures the ordered lifecycle events for verification.
type LifecycleTransition struct {
	Stage     LifecycleStage
	Err       error
	Timestamp time.Time
}

// Harness is a deterministic execution environment for agents.
type Harness struct {
	agent       Agent
	clock       *DeterministicClock
	publisher   *recordingPublisher
	subscriber  *recordingSubscriber
	acker       *recordingAck
	transitions []LifecycleTransition
	state       lifecycleState
}

type lifecycleState int

const (
	stateNew lifecycleState = iota
	stateInitialized
	stateStarted
	stateShutdown
)

// NewHarness constructs the deterministic harness.
func NewHarness(factory AgentFactory, opts HarnessOptions) (*Harness, error) {
	if factory == nil {
		return nil, fmt.Errorf("factory must be provided")
	}
	if opts.Identity.AgentID == "" || opts.Identity.AgentType == "" {
		return nil, ErrMissingIdentity
	}
	if opts.Capabilities.Len() == 0 {
		return nil, fmt.Errorf("%w: capabilities required", ErrInvalidCapabilityManifest)
	}
	start := opts.StartTime
	if start.IsZero() {
		start = time.Unix(0, 0).UTC()
	}

	clock := NewDeterministicClock(start)
	publisher := &recordingPublisher{}
	subscriber := &recordingSubscriber{}
	acker := &recordingAck{}

	cfg := BaseConfig{
		Identity:     opts.Identity,
		Capabilities: opts.Capabilities,
		Publisher:    publisher,
		Subscriber:   subscriber,
		Acker:        acker,
		Clock:        clock,
	}

	agent, err := factory(cfg)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, fmt.Errorf("factory returned nil agent")
	}

	return &Harness{
		agent:      agent,
		clock:      clock,
		publisher:  publisher,
		subscriber: subscriber,
		acker:      acker,
	}, nil
}

// Init executes the agent init hook.
func (h *Harness) Init(ctx context.Context) error {
	if h.state != stateNew {
		h.recordTransition(StageInit, ErrLifecycleViolation)
		return ErrLifecycleViolation
	}
	err := h.agent.Init(ctx)
	h.recordTransition(StageInit, err)
	if err == nil {
		h.state = stateInitialized
	}
	return err
}

// Start executes the agent start hook.
func (h *Harness) Start(ctx context.Context) error {
	if h.state != stateInitialized {
		h.recordTransition(StageStart, ErrLifecycleViolation)
		return ErrLifecycleViolation
	}
	err := h.agent.Start(ctx)
	h.recordTransition(StageStart, err)
	if err == nil {
		h.state = stateStarted
	}
	return err
}

// InjectEvent delivers evt to the agent's HandleEvent hook.
func (h *Harness) InjectEvent(ctx context.Context, evt Event) error {
	if h.state != stateStarted {
		h.recordTransition(StageHandle, ErrLifecycleViolation)
		return ErrLifecycleViolation
	}
	err := h.agent.HandleEvent(ctx, evt.Clone())
	h.recordTransition(StageHandle, err)
	return err
}

// Shutdown executes the agent shutdown hook.
func (h *Harness) Shutdown(ctx context.Context) error {
	if h.state == stateShutdown || h.state == stateNew {
		h.recordTransition(StageShutdown, ErrLifecycleViolation)
		return ErrLifecycleViolation
	}
	err := h.agent.Shutdown(ctx)
	h.recordTransition(StageShutdown, err)
	h.state = stateShutdown
	return err
}

// OutboundEvents returns a copy of captured outbound events.
func (h *Harness) OutboundEvents() []Event {
	return h.publisher.Events()
}

// LifecycleTransitions returns a copy of recorded transitions.
func (h *Harness) LifecycleTransitions() []LifecycleTransition {
	out := make([]LifecycleTransition, len(h.transitions))
	copy(out, h.transitions)
	return out
}

// AckedEvents exposes the identifiers acked by the agent.
func (h *Harness) AckedEvents() []string {
	return h.acker.Acked()
}

// NackedEvents exposes the identifiers nacked by the agent.
func (h *Harness) NackedEvents() []string {
	return h.acker.Nacked()
}

// SubscriptionTopics exposes the topics the agent attempted to subscribe to.
func (h *Harness) SubscriptionTopics() []string {
	return h.subscriber.Subscriptions()
}

// Clock exposes the deterministic clock to the test harness.
func (h *Harness) Clock() *DeterministicClock {
	return h.clock
}

func (h *Harness) recordTransition(stage LifecycleStage, err error) {
	h.transitions = append(h.transitions, LifecycleTransition{
		Stage:     stage,
		Err:       err,
		Timestamp: h.clock.Now(),
	})
}

type recordingPublisher struct {
	mu     sync.Mutex
	events []Event
}

func (p *recordingPublisher) Publish(_ context.Context, evt Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, evt.Clone())
	return nil
}

func (p *recordingPublisher) Events() []Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.events) == 0 {
		return nil
	}
	out := make([]Event, len(p.events))
	copy(out, p.events)
	return out
}

type recordingSubscriber struct {
	mu            sync.Mutex
	subscriptions []subscription
}

type subscription struct {
	topic   string
	handler EventHandler
}

func (s *recordingSubscriber) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions = append(s.subscriptions, subscription{topic: topic, handler: handler})
	return ctx.Err()
}

func (s *recordingSubscriber) Subscriptions() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.subscriptions))
	for i, sub := range s.subscriptions {
		out[i] = sub.topic
	}
	return out
}

type recordingAck struct {
	mu     sync.Mutex
	acked  []string
	nacked []string
}

func (a *recordingAck) Ack(_ context.Context, eventID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.acked = append(a.acked, eventID)
	return nil
}

func (a *recordingAck) Nack(_ context.Context, eventID string, _ error) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nacked = append(a.nacked, eventID)
	return nil
}

func (a *recordingAck) Acked() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.acked))
	copy(out, a.acked)
	return out
}

func (a *recordingAck) Nacked() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.nacked))
	copy(out, a.nacked)
	return out
}
