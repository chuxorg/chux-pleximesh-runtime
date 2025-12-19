package sdk

import (
	"context"
	"fmt"
	"time"
)

// AgentIdentity is immutable metadata describing an agent instance.
type AgentIdentity struct {
	AgentID   string
	AgentType string
}

// Agent is the authoritative interface every PlexiMesh agent must satisfy.
type Agent interface {
	Identity() AgentIdentity
	Capabilities() []Capability
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	HandleEvent(ctx context.Context, evt Event) error
	Shutdown(ctx context.Context) error
}

// BaseConfig wires the dependencies allowed for an agent.
type BaseConfig struct {
	Identity     AgentIdentity
	Capabilities CapabilitySet
	Publisher    EventPublisher
	Subscriber   EventSubscriber
	Acker        AckNackHandler
	Clock        Clock
}

// BaseAgent provides the minimal, embeddable implementation of Agent.
type BaseAgent struct {
	identity     AgentIdentity
	capabilities CapabilitySet
	publisher    EventPublisher
	subscriber   EventSubscriber
	acker        AckNackHandler
	clock        Clock
}

// NewBaseAgent constructs the base after validating identity and capabilities.
func NewBaseAgent(cfg BaseConfig) (*BaseAgent, error) {
	if cfg.Identity.AgentID == "" || cfg.Identity.AgentType == "" {
		return nil, ErrMissingIdentity
	}
	if cfg.Capabilities.Len() == 0 {
		return nil, fmt.Errorf("%w: capabilities required", ErrInvalidCapabilityManifest)
	}
	clock := cfg.Clock
	if clock == nil {
		clock = systemClock{}
	}

	return &BaseAgent{
		identity:     cfg.Identity,
		capabilities: cfg.Capabilities,
		publisher:    cfg.Publisher,
		subscriber:   cfg.Subscriber,
		acker:        cfg.Acker,
		clock:        clock,
	}, nil
}

// Identity returns the immutable identity.
func (a *BaseAgent) Identity() AgentIdentity {
	return a.identity
}

// Capabilities exposes a read-only view of the manifest entries.
func (a *BaseAgent) Capabilities() []Capability {
	return a.capabilities.List()
}

// Init provides a deterministic default no-op lifecycle hook.
func (a *BaseAgent) Init(context.Context) error { return nil }

// Start provides a deterministic default no-op lifecycle hook.
func (a *BaseAgent) Start(context.Context) error { return nil }

// HandleEvent provides a deterministic default no-op lifecycle hook.
func (a *BaseAgent) HandleEvent(context.Context, Event) error { return nil }

// Shutdown provides a deterministic default no-op lifecycle hook.
func (a *BaseAgent) Shutdown(context.Context) error { return nil }

// Publisher returns the configured publisher interface.
func (a *BaseAgent) Publisher() EventPublisher { return a.publisher }

// Subscriber returns the configured subscriber interface.
func (a *BaseAgent) Subscriber() EventSubscriber { return a.subscriber }

// Acker returns the configured ack/nack handler.
func (a *BaseAgent) Acker() AckNackHandler { return a.acker }

// Clock provides deterministic access to time.
func (a *BaseAgent) Clock() Clock { return a.clock }

// Publish emits an outbound event via the SDK publisher.
func (a *BaseAgent) Publish(ctx context.Context, evt Event) error {
	if a.publisher == nil {
		return ErrNoPublisher
	}
	return a.publisher.Publish(ctx, evt.Clone())
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
