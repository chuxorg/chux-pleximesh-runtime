//go:build ignore
// +build ignore

package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/runtime/bus"
	"github.com/chuxorg/chux-agent-mesh/internal/runtime/dummy"
	"github.com/chuxorg/chux-agent-mesh/internal/runtime/librarian"
	"github.com/chuxorg/chux-agent-mesh/internal/sdk"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

// Config wires the runtime dependencies.
type Config struct {
	Logger  *slog.Logger
	DataDir string
}

// Runtime drives the happy-path execution slice for the PlexiMesh runtime.
type Runtime struct {
	logger    *slog.Logger
	bus       *bus.Bus
	librarian *librarian.Writer
}

// New wires the runtime with sane defaults.
func New(cfg Config) *Runtime {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	dataDir := cfg.DataDir
	if dataDir == "" {
		dataDir = "data"
	}
	_ = os.MkdirAll(dataDir, 0o755)

	runtimeBus := bus.New(bus.Config{
		Logger:         logger.With("component", "bus"),
		RuntimeVersion: "mesh-dev",
		InitKitVersion: "initkit-dev",
	})
	librarianWriter := librarian.NewWriter(filepath.Join(dataDir, "librarian"), logger.With("component", "librarian"))

	return &Runtime{
		logger:    logger,
		bus:       runtimeBus,
		librarian: librarianWriter,
	}
}

// Run executes the DummyAgent slice and captures librarian evidence.
func (r *Runtime) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	runIDs := newRunIdentifiers()
	caps, err := sdk.NewCapabilitySet([]sdk.Capability{
		{Name: "dummy.bus", Version: "v1"},
	})
	if err != nil {
		return err
	}

	labels := []string{"alpha", "bravo"}
	records := make([]*agentRecord, 0, len(labels))
	for _, label := range labels {
		identity := sdk.AgentIdentity{
			AgentID:   fmt.Sprintf("dummy-%s-%s", label, runIDs.timestamp),
			AgentType: fmt.Sprintf("dummy-%s", label),
		}
		execCtx := newAgentExecution(runIDs, label)
		headers := librarian.ExecutionHeaders{
			ExecutionID:     execCtx.ExecutionID,
			RootExecutionID: execCtx.RootExecutionID,
			CorrelationID:   execCtx.CorrelationID,
			AgentID:         identity.AgentID,
		}

		subscriber := r.bus.SubscriberFor(identity)
		base := sdk.BaseConfig{
			Identity:     identity,
			Capabilities: caps,
			Publisher:    r.bus.PublisherFor(identity),
			Subscriber:   subscriber,
			Acker:        r.bus.AckerFor(identity),
		}
		agent, err := dummy.New(dummy.Config{
			BaseConfig: base,
			Logger:     r.logger.With("component", "dummy_agent", "agent_label", label),
			ExecCtx:    execCtx,
		})
		if err != nil {
			subscriber.Close()
			return err
		}

		promptID, promptErr := r.librarian.SubmitPrompt(ctx, headers, librarian.PromptPayload{
			Title:        fmt.Sprintf("DummyAgent-%s runtime prove-out", label),
			Instructions: "Observe message bus fan-out under multiple subscribers.",
			Inputs: map[string]string{
				"execution_id": execCtx.ExecutionID,
				"agent_label":  label,
				"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
			},
		})
		if promptErr != nil {
			r.logger.Warn("runtime.prompt_submission_failed", "agent_id", identity.AgentID, "error", promptErr)
			promptID = fmt.Sprintf("prompt-%s", execCtx.ExecutionID)
		}

		record := &agentRecord{
			identity:   identity,
			execCtx:    execCtx,
			headers:    headers,
			agent:      agent,
			subscriber: subscriber,
			promptID:   promptID,
		}
		records = append(records, record)
	}
	defer r.teardownAgents(ctx, records)

	for _, record := range records {
		if err := record.agent.Init(ctx); err != nil {
			return err
		}
	}
	for _, record := range records {
		if err := record.agent.Start(ctx); err != nil {
			return err
		}
	}

	receipt, dispatchedID, err := r.sendTask(ctx, runIDs)
	if err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if receipt != nil {
		if err := receipt.WaitAck(waitCtx); err != nil {
			r.logger.Warn("runtime.ack_wait_failed", "event_id", receipt.EventID(), "error", err)
		} else {
			r.logger.Info("runtime.ack_received", "event_id", receipt.EventID())
		}
	}

	for _, record := range records {
		r.persistResult(ctx, record, dispatchedID)
	}

	return nil
}

func (r *Runtime) sendTask(ctx context.Context, ids runIdentifiers) (*bus.DeliveryReceipt, string, error) {
	payload := event.TaskCreatedPayload{
		TaskID:       fmt.Sprintf("task-%s", ids.timestamp),
		IntentID:     "intent-dummy",
		Summary:      "Exercise runtime message bus",
		AssignedRole: "dummy",
		Constraints:  []string{"log lifecycle", "ack inbound"},
	}
	env := event.Envelope{
		EventID: fmt.Sprintf("evt-runtime-%d", time.Now().UnixNano()),
		Type:    event.TypeTaskCreated,
		Domain:  event.DomainTask,
		SourceAgent: event.AgentDescriptor{
			AgentID: "runtime-supervisor",
			Role:    "runtime",
		},
		Target: &event.TargetDescriptor{
			Scope: "fanout",
		},
		Timestamp:     time.Now().UTC(),
		CorrelationID: ids.correlationID,
		Signature:     "ems-hmac-v1:runtime-signature",
		Runtime: event.RuntimeContext{
			RuntimeVersion: "mesh-dev",
			InitKitVersion: "initkit-dev",
		},
	}
	if err := env.BindPayload(payload); err != nil {
		return nil, "", err
	}
	r.logger.Info("runtime.dispatch_task",
		"event_id", env.EventID,
		"fanout_scope", env.Target.Scope,
		"correlation_id", ids.correlationID,
	)
	receipt, err := r.bus.Publish(ctx, env)
	return receipt, env.EventID, err
}

type runIdentifiers struct {
	timestamp     string
	rootID        string
	correlationID string
}

func newRunIdentifiers() runIdentifiers {
	now := time.Now().UTC()
	ts := now.Format("20060102T150405")
	return runIdentifiers{
		timestamp:     ts,
		rootID:        fmt.Sprintf("root-%s", ts),
		correlationID: fmt.Sprintf("corr-%d", now.UnixNano()),
	}
}

func newAgentExecution(ids runIdentifiers, label string) dummy.ExecutionContext {
	return dummy.ExecutionContext{
		ExecutionID:     fmt.Sprintf("exec-%s-%s", ids.timestamp, label),
		RootExecutionID: ids.rootID,
		CorrelationID:   ids.correlationID,
	}
}

type agentRecord struct {
	identity   sdk.AgentIdentity
	execCtx    dummy.ExecutionContext
	headers    librarian.ExecutionHeaders
	agent      *dummy.Agent
	subscriber *bus.Subscriber
	promptID   string
}

func (r *Runtime) teardownAgents(ctx context.Context, records []*agentRecord) {
	for _, record := range records {
		if record == nil {
			continue
		}
		if record.subscriber != nil {
			record.subscriber.Close()
		}
		r.shutdownAgent(ctx, record.agent)
	}
}

func (r *Runtime) shutdownAgent(ctx context.Context, agent *dummy.Agent) {
	if agent == nil {
		return
	}
	if err := agent.Shutdown(ctx); err != nil {
		r.logger.Warn("runtime.agent_shutdown_error", "error", err)
	}
}

func (r *Runtime) persistResult(ctx context.Context, record *agentRecord, dispatchedEventID string) {
	summary, logs := record.agent.LastReport()
	if summary == "" {
		summary = fmt.Sprintf("agent %s completed run", record.identity.AgentID)
	}
	details := map[string]string{
		"agent_id":         record.identity.AgentID,
		"execution_id":     record.execCtx.ExecutionID,
		"dispatched_event": dispatchedEventID,
	}
	if err := r.librarian.SubmitResult(ctx, record.headers, record.promptID, librarian.ResultPayload{
		Outcome: summary,
		Details: details,
		Logs:    logs,
	}); err != nil {
		r.logger.Warn("runtime.result_submission_failed", "agent_id", record.identity.AgentID, "error", err)
	}
}
