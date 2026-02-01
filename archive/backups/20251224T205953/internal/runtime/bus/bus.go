//go:build ignore
// +build ignore

package bus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
	"github.com/chuxorg/chux-agent-mesh/internal/sdk"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

var (
	errHandlerRequired = errors.New("bus: handler is required")
	errUnknownEventID  = errors.New("bus: event not tracked")
)

// Config wires the dependencies required by the runtime bus.
type Config struct {
	Router         *messagebus.Router
	Logger         *slog.Logger
	RuntimeVersion string
	InitKitVersion string
}

// Bus coordinates publishes and fan-out delivery for the runtime.
type Bus struct {
	router         *messagebus.Router
	logger         *slog.Logger
	runtimeVersion string
	initKitVersion string

	inflight sync.Map // event_id -> *deliveryState
}

// New creates the runtime bus.
func New(cfg Config) *Bus {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	if cfg.Router == nil {
		cfg.Router = messagebus.NewRouter()
	}
	return &Bus{
		router:         cfg.Router,
		logger:         logger,
		runtimeVersion: valueOrDefault(cfg.RuntimeVersion, "mesh-dev"),
		initKitVersion: valueOrDefault(cfg.InitKitVersion, "initkit-dev"),
	}
}

// DeliveryReceipt allows callers to await the ack associated with a routed envelope.
type DeliveryReceipt struct {
	eventID string
	state   *deliveryState
}

// EventID returns the tracked event identifier.
func (r *DeliveryReceipt) EventID() string {
	if r == nil {
		return ""
	}
	return r.eventID
}

// WaitAck blocks until the tracked event is acked or the context is cancelled.
func (r *DeliveryReceipt) WaitAck(ctx context.Context) error {
	if r == nil || r.state == nil || r.state.ackCh == nil {
		return nil
	}
	return r.state.waitAck(ctx)
}

// Publish validates the envelope and dispatches it to interested subscribers.
func (b *Bus) Publish(ctx context.Context, env event.Envelope) (*DeliveryReceipt, error) {
	if err := env.Validate(); err != nil {
		return nil, err
	}
	state := b.trackDelivery(env)
	b.logger.Info("bus.publish",
		"event_id", env.EventID,
		"type", env.Type,
		"domain", env.Domain,
		"source_agent", env.SourceAgent.AgentID,
		"target_agent", targetAgentID(env.Target),
		"correlation_id", env.CorrelationID,
	)
	if err := b.router.Dispatch(ctx, env); err != nil {
		if state != nil {
			state.complete(err)
			b.inflight.Delete(env.EventID)
		}
		return nil, err
	}
	return &DeliveryReceipt{eventID: env.EventID, state: state}, nil
}

// PublisherFor returns an sdk-compatible publisher bound to the agent identity.
func (b *Bus) PublisherFor(id sdk.AgentIdentity) sdk.EventPublisher {
	return &publisher{
		bus:     b,
		agent:   id,
		logger:  b.logger.With("agent_id", id.AgentID),
		nowFunc: time.Now,
	}
}

// SubscriberFor returns an sdk EventSubscriber bound to the agent identity.
func (b *Bus) SubscriberFor(id sdk.AgentIdentity) *Subscriber {
	return newSubscriber(b, id)
}

// AckerFor returns an ack handler tied to the runtime bus.
func (b *Bus) AckerFor(id sdk.AgentIdentity) sdk.AckNackHandler {
	return &acker{bus: b, agentID: id.AgentID, logger: b.logger.With("agent_id", id.AgentID)}
}

func (b *Bus) trackDelivery(env event.Envelope) *deliveryState {
	if env.Target == nil || (env.Target.AgentID == "" && env.Target.Scope == "") {
		return nil
	}
	state := newDeliveryState(env)
	b.inflight.Store(env.EventID, state)
	return state
}

// Ack records a positive acknowledgement for the in-flight event.
func (b *Bus) Ack(ctx context.Context, agentID, eventID string) error {
	return b.resolveDelivery(ctx, agentID, eventID, nil)
}

// Nack records a negative acknowledgement for the in-flight event.
func (b *Bus) Nack(ctx context.Context, agentID, eventID string, reason error) error {
	if reason == nil {
		reason = errors.New("nack without reason")
	}
	return b.resolveDelivery(ctx, agentID, eventID, reason)
}

func (b *Bus) resolveDelivery(ctx context.Context, agentID, eventID string, err error) error {
	value, ok := b.inflight.Load(eventID)
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownEventID, eventID)
	}
	state := value.(*deliveryState)
	b.inflight.Delete(eventID)
	state.complete(err)
	result := "ack"
	if err != nil {
		result = "nack"
	}
	b.logger.Info("bus.delivery."+result,
		"event_id", eventID,
		"target_agent", agentID,
		"latency_ms", state.latencyMillis(),
		"error", errorString(err),
		"context_error", ctx.Err(),
	)
	return nil
}

type publisher struct {
	bus     *Bus
	agent   sdk.AgentIdentity
	logger  *slog.Logger
	nowFunc func() time.Time
}

func (p *publisher) Publish(ctx context.Context, evt sdk.Event) error {
	env, err := p.toEnvelope(evt)
	if err != nil {
		return err
	}
	_, err = p.bus.Publish(ctx, env)
	return err
}

func (p *publisher) toEnvelope(evt sdk.Event) (event.Envelope, error) {
	if evt.ID == "" {
		return event.Envelope{}, fmt.Errorf("bus: event id required")
	}
	if evt.Type == "" {
		return event.Envelope{}, fmt.Errorf("bus: event type required")
	}
	meta, ok := event.MetadataForType(event.Type(evt.Type))
	if !ok {
		return event.Envelope{}, fmt.Errorf("bus: %w %s", event.ErrUnknownType, evt.Type)
	}
	timestamp := evt.Timestamp
	if timestamp.IsZero() {
		timestamp = p.nowFunc().UTC()
	}
	env := event.Envelope{
		EventID:   evt.ID,
		Type:      event.Type(evt.Type),
		Domain:    meta.Domain,
		Timestamp: timestamp,
		SourceAgent: event.AgentDescriptor{
			AgentID: p.agent.AgentID,
			Role:    p.agent.AgentType,
		},
		Payload:   evt.Payload,
		Signature: "ems-hmac-v1:dummy-signature",
		Runtime: event.RuntimeContext{
			RuntimeVersion: p.bus.runtimeVersion,
			InitKitVersion: p.bus.initKitVersion,
		},
	}
	if corr := evt.Metadata["correlation_id"]; corr != "" {
		env.CorrelationID = corr
	}
	if target := evt.Metadata["target_agent_id"]; target != "" {
		env.Target = &event.TargetDescriptor{AgentID: target}
	}
	return env, env.Validate()
}

type acker struct {
	bus     *Bus
	agentID string
	logger  *slog.Logger
}

func (a *acker) Ack(ctx context.Context, eventID string) error {
	err := a.bus.Ack(ctx, a.agentID, eventID)
	if err != nil {
		a.logger.Error("ack failed", "event_id", eventID, "error", err)
	}
	return err
}

func (a *acker) Nack(ctx context.Context, eventID string, reason error) error {
	err := a.bus.Nack(ctx, a.agentID, eventID, reason)
	if err != nil {
		a.logger.Error("nack failed", "event_id", eventID, "error", err)
	}
	return err
}

type deliveryState struct {
	envelope  event.Envelope
	delivered time.Time
	ackCh     chan ackResult
	once      sync.Once
}

type ackResult struct {
	err error
}

func newDeliveryState(env event.Envelope) *deliveryState {
	return &deliveryState{
		envelope:  env,
		delivered: time.Now().UTC(),
		ackCh:     make(chan ackResult, 1),
	}
}

func (s *deliveryState) complete(err error) {
	s.once.Do(func() {
		select {
		case s.ackCh <- ackResult{err: err}:
		default:
		}
		close(s.ackCh)
	})
}

func (s *deliveryState) waitAck(ctx context.Context) error {
	select {
	case result, ok := <-s.ackCh:
		if !ok {
			return nil
		}
		return result.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *deliveryState) latencyMillis() float64 {
	if s.delivered.IsZero() {
		return 0
	}
	return float64(time.Since(s.delivered).Milliseconds())
}

// Subscriber multiplexes router subscriptions for a single agent.
type Subscriber struct {
	bus     *Bus
	agentID string
	logger  *slog.Logger

	mu      sync.Mutex
	runners []*subscriptionRunner
	wg      sync.WaitGroup
	seq     atomic.Uint64
}

type subscriptionRunner struct {
	handle *messagebus.SubscriptionHandle
	cancel context.CancelFunc
}

func newSubscriber(bus *Bus, id sdk.AgentIdentity) *Subscriber {
	return &Subscriber{
		bus:     bus,
		agentID: id.AgentID,
		logger:  bus.logger.With("agent_id", id.AgentID),
	}
}

// Subscribe registers the handler for the provided topic (domain name).
func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler sdk.EventHandler) error {
	if handler == nil {
		return errHandlerRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cfg := messagebus.SubscriptionConfig{
		AgentID: fmt.Sprintf("%s-%d", s.agentID, s.seq.Add(1)),
	}
	if topic != "" {
		cfg.Domains = []event.Domain{event.Domain(topic)}
	}
	handle, err := s.bus.router.Subscribe(cfg)
	if err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	runner := &subscriptionRunner{handle: handle, cancel: cancel}

	s.mu.Lock()
	s.runners = append(s.runners, runner)
	s.mu.Unlock()

	s.wg.Add(1)
	go s.consume(runCtx, runner, handler)

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-runCtx.Done():
		}
	}()
	return nil
}

func (s *Subscriber) consume(ctx context.Context, runner *subscriptionRunner, handler sdk.EventHandler) {
	defer func() {
		runner.handle.Close()
		s.wg.Done()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case env, ok := <-runner.handle.Events():
			if !ok {
				return
			}
			if env.SourceAgent.AgentID == s.agentID {
				continue
			}
			if env.Target != nil && env.Target.AgentID != "" && env.Target.AgentID != s.agentID {
				continue
			}
			if err := handler(ctx, envelopeToSDK(env)); err != nil {
				s.logger.Error("handler error", "event_id", env.EventID, "error", err)
			}
		}
	}
}

// Close terminates all active subscriptions.
func (s *Subscriber) Close() {
	s.mu.Lock()
	for _, runner := range s.runners {
		runner.cancel()
		runner.handle.Close()
	}
	s.runners = nil
	s.mu.Unlock()
	s.wg.Wait()
}

func envelopeToSDK(env event.Envelope) sdk.Event {
	metadata := map[string]string{
		"event_domain": string(env.Domain),
		"source_agent": env.SourceAgent.AgentID,
	}
	if env.CorrelationID != "" {
		metadata["correlation_id"] = env.CorrelationID
	}
	if env.Target != nil && env.Target.AgentID != "" {
		metadata["target_agent_id"] = env.Target.AgentID
	}
	return sdk.Event{
		ID:        env.EventID,
		Type:      string(env.Type),
		Payload:   append([]byte(nil), env.Payload...),
		Metadata:  metadata,
		Timestamp: env.Timestamp,
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func targetAgentID(target *event.TargetDescriptor) string {
	if target == nil {
		return ""
	}
	return target.AgentID
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
