package sdk

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestHarnessLifecycleAndOutboundCapture(t *testing.T) {
	caps, err := NewCapabilitySet([]Capability{{Name: "emit", Version: "v1"}})
	if err != nil {
		t.Fatalf("caps: %v", err)
	}

	factory := func(cfg BaseConfig) (Agent, error) {
		base, err := NewBaseAgent(cfg)
		if err != nil {
			return nil, err
		}
		return &dummyAgent{BaseAgent: base}, nil
	}

	harness, err := NewHarness(factory, HarnessOptions{
		Identity:     AgentIdentity{AgentID: "agent-1", AgentType: "tester"},
		Capabilities: caps,
		StartTime:    time.Unix(10, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("harness: %v", err)
	}

	if err := harness.Start(context.Background()); !errors.Is(err, ErrLifecycleViolation) {
		t.Fatalf("start before init should violate lifecycle: %v", err)
	}

	if err := harness.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := harness.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	evt := Event{ID: "in-1", Type: "test", Payload: []byte("payload")}
	if err := harness.InjectEvent(context.Background(), evt); err != nil {
		t.Fatalf("inject: %v", err)
	}

	if err := harness.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	agent := harness.agent.(*dummyAgent)
	wantOrder := []string{"init", "start", "handle", "shutdown"}
	if len(agent.calls) != len(wantOrder) {
		t.Fatalf("got %v calls, want %v", agent.calls, wantOrder)
	}
	for i, stage := range wantOrder {
		if agent.calls[i] != stage {
			t.Fatalf("stage %d mismatch: got %s want %s", i, agent.calls[i], stage)
		}
	}

	outbound := harness.OutboundEvents()
	if len(outbound) != 1 {
		t.Fatalf("expected one outbound event, got %d", len(outbound))
	}
	if outbound[0].ID != "out-in-1" {
		t.Fatalf("unexpected outbound id %q", outbound[0].ID)
	}

	if got := harness.AckedEvents(); len(got) != 1 || got[0] != "in-1" {
		t.Fatalf("acked mismatch: %v", got)
	}

	if got := harness.NackedEvents(); len(got) != 0 {
		t.Fatalf("unexpected nacks: %v", got)
	}

	if got := harness.SubscriptionTopics(); len(got) != 1 || got[0] != "telemetry" {
		t.Fatalf("subscription topics mismatch: %v", got)
	}

	transitions := harness.LifecycleTransitions()
	if len(transitions) != 5 { // includes failed start attempt
		t.Fatalf("unexpected transition count %d", len(transitions))
	}
	if transitions[0].Stage != StageStart || !errors.Is(transitions[0].Err, ErrLifecycleViolation) {
		t.Fatalf("first transition should record lifecycle violation: %+v", transitions[0])
	}
}

func TestManifestLoaderValidation(t *testing.T) {
	manifest := `{"capabilities":[{"name":"emit","version":"v1"}]}`
	set, err := LoadCapabilityManifest(strings.NewReader(manifest))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if set.Len() != 1 || !set.Has("emit") {
		t.Fatalf("manifest not parsed: %+v", set)
	}

	bad := `{"capabilities":[{"name":"","version":""}]}`
	if _, err := LoadCapabilityManifest(strings.NewReader(bad)); !errors.Is(err, ErrInvalidCapabilityManifest) {
		t.Fatalf("expected invalid manifest error, got %v", err)
	}
}

func TestDeterministicErrors(t *testing.T) {
	caps, err := NewCapabilitySet([]Capability{{Name: "emit", Version: "v1"}})
	if err != nil {
		t.Fatalf("caps: %v", err)
	}

	factory := func(cfg BaseConfig) (Agent, error) {
		return NewBaseAgent(cfg)
	}

	harness, err := NewHarness(factory, HarnessOptions{
		Identity:     AgentIdentity{AgentID: "agent-err", AgentType: "tester"},
		Capabilities: caps,
	})
	if err != nil {
		t.Fatalf("harness: %v", err)
	}

	if err := harness.InjectEvent(context.Background(), Event{}); !errors.Is(err, ErrLifecycleViolation) {
		t.Fatalf("inject before start should error: %v", err)
	}

	if err := harness.Shutdown(context.Background()); !errors.Is(err, ErrLifecycleViolation) {
		t.Fatalf("shutdown before init should error: %v", err)
	}
}

func TestBaseAgentPublishWithoutPublisher(t *testing.T) {
	caps, _ := NewCapabilitySet([]Capability{{Name: "emit", Version: "v1"}})
	base, err := NewBaseAgent(BaseConfig{
		Identity:     AgentIdentity{AgentID: "agent", AgentType: "tester"},
		Capabilities: caps,
	})
	if err != nil {
		t.Fatalf("base: %v", err)
	}
	if err := base.Publish(context.Background(), Event{ID: "1"}); !errors.Is(err, ErrNoPublisher) {
		t.Fatalf("expected ErrNoPublisher, got %v", err)
	}
}

type dummyAgent struct {
	*BaseAgent
	calls []string
}

func (d *dummyAgent) Init(ctx context.Context) error {
	d.calls = append(d.calls, "init")
	if sub := d.Subscriber(); sub != nil {
		_ = sub.Subscribe(ctx, "telemetry", func(context.Context, Event) error { return nil })
	}
	return nil
}

func (d *dummyAgent) Start(context.Context) error {
	d.calls = append(d.calls, "start")
	return nil
}

func (d *dummyAgent) HandleEvent(ctx context.Context, evt Event) error {
	d.calls = append(d.calls, "handle")
	if ack := d.Acker(); ack != nil {
		_ = ack.Ack(ctx, evt.ID)
	}
	evt.ID = "out-" + evt.ID
	return d.Publish(ctx, evt)
}

func (d *dummyAgent) Shutdown(context.Context) error {
	d.calls = append(d.calls, "shutdown")
	return nil
}
