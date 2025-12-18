package event

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestEnvelopeMarshalRoundTrip(t *testing.T) {
	reason := "state updated"
	payload := LifecycleStateReportedPayload{
		AgentState: AgentStateWorking,
		Reason:     &reason,
	}

	env := Envelope{
		EventID:       "123e4567-e89b-12d3-a456-426614174000",
		Type:          TypeLifecycleStateReported,
		Domain:        DomainLifecycle,
		SourceAgent:   AgentDescriptor{AgentID: "engineer-runtime", Role: "engineer"},
		Target:        &TargetDescriptor{AgentID: "guardian-core", Scope: "status"},
		Timestamp:     time.Date(2025, time.December, 17, 14, 33, 21, 0, time.UTC),
		CorrelationID: "323e4567-e89b-12d3-a456-426614174000",
		Signature:     "ems-hmac-v1:q3K8fR==",
		Runtime: RuntimeContext{
			RuntimeVersion: "0.1.0",
			InitKitVersion: "2024.09",
		},
	}
	if err := env.BindPayload(payload); err != nil {
		t.Fatalf("BindPayload() error = %v", err)
	}
	if err := env.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("round-trip Validate() error = %v", err)
	}
	if got, want := decoded.EventID, env.EventID; got != want {
		t.Fatalf("EventID mismatch: got %s want %s", got, want)
	}
	if !decoded.Timestamp.Equal(env.Timestamp) {
		t.Fatalf("Timestamp mismatch: got %v want %v", decoded.Timestamp, env.Timestamp)
	}

	var payloadOut LifecycleStateReportedPayload
	if err := decoded.DecodePayload(&payloadOut); err != nil {
		t.Fatalf("DecodePayload() error = %v", err)
	}
	if payloadOut.AgentState != payload.AgentState {
		t.Fatalf("payload agent state mismatch: got %s want %s", payloadOut.AgentState, payload.AgentState)
	}
	if payloadOut.Reason == nil || *payloadOut.Reason != reason {
		t.Fatalf("payload reason mismatch: got %v want %s", payloadOut.Reason, reason)
	}
}

func TestEnvelopeValidateDetectsDomainMismatch(t *testing.T) {
	env := Envelope{
		EventID:     "123",
		Type:        TypeLifecycleStateReported,
		Domain:      DomainIntent,
		SourceAgent: AgentDescriptor{AgentID: "a", Role: "engineer"},
		Timestamp:   time.Unix(0, 0).UTC(),
		Signature:   "ems-hmac-v1:abc",
		Runtime: RuntimeContext{
			RuntimeVersion: "0.1.0",
			InitKitVersion: "2024.09",
		},
		Payload: json.RawMessage(`{}`),
	}
	err := env.Validate()
	if !errors.Is(err, ErrDomainMismatch) {
		t.Fatalf("expected domain mismatch error, got %v", err)
	}
}
