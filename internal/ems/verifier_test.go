package ems

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

func TestVerifierAcceptsValidSignature(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	km := NewKeyManager(WithTimeSource(func() time.Time { return now }))
	material := bytesWithValue(0x42)
	version, err := km.RegisterKey("agent-1", material[:])
	if err != nil {
		t.Fatalf("register key: %v", err)
	}
	scheduler := NewSliceScheduler(WithSchedulerClock(func() time.Time { return now }))
	verifier := NewVerifier(km, scheduler, WithVerifierClock(func() time.Time { return now }))

	env := sampleEnvelope(now)
	signEnvelope(t, &env, version.Material, scheduler.SliceAt(now))

	if err := verifier.Verify(context.Background(), env); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerifierRejectsInvalidSignature(t *testing.T) {
	now := time.Unix(1_700_000_100, 0).UTC()
	km := NewKeyManager(WithTimeSource(func() time.Time { return now }))
	material := bytesWithValue(0x24)
	version, err := km.RegisterKey("agent-1", material[:])
	if err != nil {
		t.Fatalf("register key: %v", err)
	}
	scheduler := NewSliceScheduler(WithSchedulerClock(func() time.Time { return now }))
	verifier := NewVerifier(km, scheduler, WithVerifierClock(func() time.Time { return now }))

	env := sampleEnvelope(now)
	signEnvelope(t, &env, version.Material, scheduler.SliceAt(now))
	env.Payload = []byte(`{"changed":true}`)

	err = verifier.Verify(context.Background(), env)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifierDetectsReplay(t *testing.T) {
	now := time.Unix(1_700_000_200, 0).UTC()
	km := NewKeyManager(WithTimeSource(func() time.Time { return now }))
	material := bytesWithValue(0x11)
	version, err := km.RegisterKey("agent-1", material[:])
	if err != nil {
		t.Fatalf("register key: %v", err)
	}
	scheduler := NewSliceScheduler(WithSchedulerClock(func() time.Time { return now }))
	verifier := NewVerifier(km, scheduler, WithVerifierClock(func() time.Time { return now }))

	env := sampleEnvelope(now)
	env.EventID = "replay-event"
	signEnvelope(t, &env, version.Material, scheduler.SliceAt(now))

	if err := verifier.Verify(context.Background(), env); err != nil {
		t.Fatalf("first verify failed: %v", err)
	}
	if err := verifier.Verify(context.Background(), env); !errors.Is(err, ErrReplayDetected) {
		t.Fatalf("expected ErrReplayDetected, got %v", err)
	}
}

func signEnvelope(t *testing.T, env *event.Envelope, material [keyMaterialSize]byte, slice SliceIndex) {
	t.Helper()
	env.Signature = ""
	canonical, err := canonicalizeEnvelope(*env)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	ctx := deriveContext(*env, slice)
	ephemeral := hmac.New(sha256.New, material[:])
	ephemeral.Write(ctx)
	ephemeralKey := ephemeral.Sum(nil)

	mac := hmac.New(sha256.New, ephemeralKey)
	mac.Write(canonical)
	sig := mac.Sum(nil)
	env.Signature = signaturePrefix + base64.StdEncoding.EncodeToString(sig)
}

func sampleEnvelope(ts time.Time) event.Envelope {
	return event.Envelope{
		EventID: "evt-123",
		Type:    event.TypeTaskCreated,
		Domain:  event.DomainTask,
		SourceAgent: event.AgentDescriptor{
			AgentID: "agent-1",
			Role:    "engineer",
		},
		Timestamp: ts,
		Payload:   []byte(`{"task":"demo"}`),
		Runtime: event.RuntimeContext{
			RuntimeVersion: "v0",
			InitKitVersion: "initkit",
		},
		CorrelationID: "corr-123",
	}
}

func bytesWithValue(v byte) [keyMaterialSize]byte {
	var material [keyMaterialSize]byte
	for i := range material {
		material[i] = v
	}
	return material
}
