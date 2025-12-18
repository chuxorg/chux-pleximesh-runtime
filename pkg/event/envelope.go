package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrUnknownType signals that no routing metadata exists for the requested event type.
	ErrUnknownType = errors.New("event: unknown type")
	// ErrDomainMismatch is raised when the envelope domain conflicts with the registered metadata.
	ErrDomainMismatch = errors.New("event: domain mismatch")
)

// Domain represents the canonical routing domain for an event.
type Domain string

// Canonical event domains.
const (
	DomainLifecycle    Domain = "lifecycle"
	DomainIntent       Domain = "intent"
	DomainTask         Domain = "task"
	DomainVerification Domain = "verification"
	DomainCognitive    Domain = "cognitive"
	DomainTelemetry    Domain = "telemetry"
	DomainStatus       Domain = "status"
	DomainHealth       Domain = "health"
	DomainViolation    Domain = "violation"
	DomainDebug        Domain = "debug"
)

// Type uniquely identifies a canonical event class.
type Type string

// Envelope captures the canonical event envelope shared by all event classes.
type Envelope struct {
	EventID       string            `json:"event_id"`
	Type          Type              `json:"event_type"`
	Domain        Domain            `json:"event_domain"`
	SourceAgent   AgentDescriptor   `json:"source_agent"`
	Target        *TargetDescriptor `json:"target,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Payload       json.RawMessage   `json:"payload"`
	Signature     string            `json:"signature"`
	Runtime       RuntimeContext    `json:"runtime"`
}

// AgentDescriptor identifies the emitting agent.
type AgentDescriptor struct {
	AgentID string `json:"agent_id"`
	Role    string `json:"role"`
}

// TargetDescriptor optionally scopes the intended receiver or audience.
type TargetDescriptor struct {
	AgentID string `json:"agent_id,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

// RuntimeContext records versions that produced the event.
type RuntimeContext struct {
	RuntimeVersion string `json:"runtime_version"`
	InitKitVersion string `json:"initkit_version"`
}

// Validate asserts that the envelope satisfies the canonical contract and the routing metadata.
func (e Envelope) Validate() error {
	if e.EventID == "" {
		return errors.New("event: event_id is required")
	}
	if e.Type == "" {
		return errors.New("event: event_type is required")
	}
	meta, ok := MetadataForType(e.Type)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownType, e.Type)
	}
	if e.Domain == "" {
		return errors.New("event: event_domain is required")
	}
	if e.Domain != meta.Domain {
		return fmt.Errorf("%w: %s is registered under %s but envelope used %s", ErrDomainMismatch, e.Type, meta.Domain, e.Domain)
	}
	if e.SourceAgent.AgentID == "" || e.SourceAgent.Role == "" {
		return errors.New("event: source_agent requires agent_id and role")
	}
	if e.Timestamp.IsZero() {
		return errors.New("event: timestamp is required")
	}
	if len(e.Payload) == 0 {
		return errors.New("event: payload must not be empty")
	}
	if e.Signature == "" {
		return errors.New("event: signature must not be empty")
	}
	if e.Runtime.RuntimeVersion == "" || e.Runtime.InitKitVersion == "" {
		return errors.New("event: runtime versions must be provided")
	}
	return nil
}

// BindPayload marshals the provided payload into the envelope payload field.
func (e *Envelope) BindPayload(payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("event: marshal payload: %w", err)
	}
	e.Payload = body
	return nil
}

// DecodePayload unmarshals the envelope payload into v.
func (e Envelope) DecodePayload(v any) error {
	if len(e.Payload) == 0 {
		return errors.New("event: payload is empty")
	}
	if err := json.Unmarshal(e.Payload, v); err != nil {
		return fmt.Errorf("event: decode payload: %w", err)
	}
	return nil
}
