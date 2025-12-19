package qaharness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

const (
	defaultLoadEvents       = 10000
	defaultLoadSubscribers  = 3
	defaultScenarioTimeout  = 30 * time.Second
	scenarioInvalidEnvelope = "invalid_envelope_rejected"
	scenarioLoadFanOut      = "high_volume_fanout"
)

// DeliverFunc models the ingress pipeline exercised by the harness.
type DeliverFunc func(context.Context, event.Envelope) error

// Option configures Harness behavior.
type Option func(*Harness)

// WithClock overrides the timestamp source used in reports.
func WithClock(clock func() time.Time) Option {
	return func(h *Harness) {
		if clock != nil {
			h.clock = clock
		}
	}
}

// WithLoadEvents adjusts the target event count for the load scenario.
func WithLoadEvents(count int) Option {
	return func(h *Harness) {
		if count > 0 {
			h.loadEvents = count
		}
	}
}

// WithLoadSubscribers adjusts the total subscribers participating in the load scenario.
func WithLoadSubscribers(count int) Option {
	return func(h *Harness) {
		if count > 0 {
			h.loadSubscribers = count
		}
	}
}

// WithDeliverFunc overrides the ingress pipeline used for dispatch attempts.
func WithDeliverFunc(fn DeliverFunc) Option {
	return func(h *Harness) {
		if fn != nil {
			h.deliver = fn
		}
	}
}

// Harness executes QA scenarios against a router and records structured findings.
type Harness struct {
	router          *messagebus.Router
	clock           func() time.Time
	loadEvents      int
	loadSubscribers int
	deliver         DeliverFunc
}

// New constructs a QA harness rooted in the provided router.
func New(router *messagebus.Router, opts ...Option) *Harness {
	if router == nil {
		panic("qaharness: router is required")
	}
	h := &Harness{
		router:          router,
		clock:           time.Now,
		loadEvents:      defaultLoadEvents,
		loadSubscribers: defaultLoadSubscribers,
	}
	h.deliver = func(ctx context.Context, env event.Envelope) error {
		if err := env.Validate(); err != nil {
			return err
		}
		return router.Dispatch(ctx, env)
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Report summarizes all executed QA scenarios.
type Report struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Passed      bool             `json:"passed"`
	Scenarios   []ScenarioResult `json:"scenarios"`
}

// ScenarioResult records the outcome of a single QA scenario.
type ScenarioResult struct {
	Name          string             `json:"name"`
	Passed        bool               `json:"passed"`
	DurationMS    float64            `json:"duration_ms"`
	Metrics       map[string]float64 `json:"metrics,omitempty"`
	Details       map[string]string  `json:"details,omitempty"`
	FailureReason string             `json:"failure,omitempty"`
}

// Run executes all MB QA scenarios and returns a structured report.
func (h *Harness) Run(ctx context.Context) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	report := Report{
		GeneratedAt: h.clock().UTC(),
		Scenarios:   make([]ScenarioResult, 0, 2),
	}
	invalid := h.runInvalidEnvelopeScenario(ctx)
	report.Scenarios = append(report.Scenarios, invalid)
	load := h.runLoadFanOutScenario(ctx)
	report.Scenarios = append(report.Scenarios, load)
	report.Passed = invalid.Passed && load.Passed
	if !report.Passed {
		return report, errors.New("qaharness: one or more scenarios failed")
	}
	return report, nil
}

func (h *Harness) runInvalidEnvelopeScenario(ctx context.Context) ScenarioResult {
	start := time.Now()
	result := ScenarioResult{
		Name:    scenarioInvalidEnvelope,
		Metrics: map[string]float64{"dispatch_attempts": 1},
	}
	defer func() {
		result.DurationMS = elapsedMillis(start)
	}()
	handle, err := h.router.Subscribe(messagebus.SubscriptionConfig{
		AgentID: fmt.Sprintf("qa-invalid-%d", start.UnixNano()),
		Domains: []event.Domain{event.DomainTask},
	})
	if err != nil {
		result.FailureReason = fmt.Sprintf("subscription failed: %v", err)
		return result
	}
	defer handle.Close()

	invalid := h.validEnvelope("qa-invalid-event")
	invalid.Domain = ""

	if err := h.deliver(ctx, invalid); err == nil {
		result.FailureReason = "expected envelope validation to fail"
		return result
	} else {
		result.Details = map[string]string{"validation_error": err.Error()}
	}

	select {
	case <-handle.Events():
		result.FailureReason = "subscriber received invalid envelope"
	case <-time.After(25 * time.Millisecond):
		result.Passed = true
	}
	return result
}

func (h *Harness) runLoadFanOutScenario(parent context.Context) ScenarioResult {
	start := time.Now()
	result := ScenarioResult{
		Name: scenarioLoadFanOut,
		Metrics: map[string]float64{
			"events_dispatched": float64(h.loadEvents),
			"subscribers":       float64(h.loadSubscribers),
		},
	}
	defer func() {
		result.DurationMS = elapsedMillis(start)
	}()

	ctx, cancel := context.WithTimeout(parent, defaultScenarioTimeout)
	defer cancel()

	handles := make([]*messagebus.SubscriptionHandle, 0, h.loadSubscribers)
	for i := 0; i < h.loadSubscribers; i++ {
		handle, err := h.router.Subscribe(messagebus.SubscriptionConfig{
			AgentID:    fmt.Sprintf("qa-load-%d", i),
			Domains:    []event.Domain{event.DomainTask},
			BufferSize: 4096,
		})
		if err != nil {
			result.FailureReason = fmt.Sprintf("subscriber %d: %v", i, err)
			return result
		}
		handles = append(handles, handle)
	}
	defer func() {
		for _, handle := range handles {
			handle.Close()
		}
	}()

	counts := make([]int64, len(handles))
	var wg sync.WaitGroup
	var once sync.Once
	var scenarioErr error
	fail := func(err error) {
		if err == nil {
			return
		}
		once.Do(func() {
			scenarioErr = err
			cancel()
		})
	}

	for idx, handle := range handles {
		wg.Add(1)
		go func(i int, hdl *messagebus.SubscriptionHandle) {
			defer wg.Done()
			for received := 0; received < h.loadEvents; received++ {
				select {
				case <-ctx.Done():
					fail(ctx.Err())
					return
				case env, ok := <-hdl.Events():
					if !ok {
						fail(errors.New("subscription closed unexpectedly"))
						return
					}
					if env.EventID == "" {
						fail(errors.New("empty event delivered"))
						return
					}
					counts[i]++
				}
			}
		}(idx, handle)
	}

	dispatchStart := time.Now()
	for i := 0; i < h.loadEvents; i++ {
		env := h.validEnvelope(fmt.Sprintf("qa-load-%d", i))
		if err := h.deliver(ctx, env); err != nil {
			fail(fmt.Errorf("dispatch %d: %w", i, err))
			break
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		fail(fmt.Errorf("fan-out scenario timeout: %w", ctx.Err()))
	}

	duration := time.Since(dispatchStart)
	rate := 0.0
	if duration > 0 {
		rate = float64(h.loadEvents) / duration.Seconds()
	}
	result.Metrics["events_per_second"] = rate
	for i, count := range counts {
		result.Metrics[fmt.Sprintf("subscriber_%d_received", i)] = float64(count)
	}

	if scenarioErr != nil {
		result.FailureReason = scenarioErr.Error()
		return result
	}
	for idx, count := range counts {
		if count != int64(h.loadEvents) {
			result.FailureReason = fmt.Sprintf("subscriber %d received %d events", idx, count)
			return result
		}
	}
	result.Passed = true
	if result.Details == nil {
		result.Details = make(map[string]string)
	}
	result.Details["throughput"] = fmt.Sprintf("%.2f events/sec", rate)
	return result
}

func elapsedMillis(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}

func (h *Harness) validEnvelope(id string) event.Envelope {
	payload := event.TaskCreatedPayload{
		TaskID:       fmt.Sprintf("task-%s", id),
		IntentID:     "intent-qa",
		Summary:      "QA harness load validation",
		AssignedRole: "qa",
		Constraints:  []string{"none"},
	}
	body, _ := json.Marshal(payload)
	return event.Envelope{
		EventID: id,
		Type:    event.TypeTaskCreated,
		Domain:  event.DomainTask,
		SourceAgent: event.AgentDescriptor{
			AgentID: "qa-publisher",
			Role:    "harness",
		},
		Timestamp: time.Now().UTC(),
		Payload:   body,
		Signature: "ems-hmac-v1:dGVzdC1zaWc=",
		Runtime: event.RuntimeContext{
			RuntimeVersion: "qa-harness",
			InitKitVersion: "qa-harness",
		},
	}
}
