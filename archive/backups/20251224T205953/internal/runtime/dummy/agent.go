//go:build ignore
// +build ignore

package dummy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/sdk"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

// ExecutionContext captures the IDs and correlation headers shared across the run.
type ExecutionContext struct {
	ExecutionID     string
	RootExecutionID string
	CorrelationID   string
}

// Agent exercises the runtime message bus and librarian wiring.
type Agent struct {
	*sdk.BaseAgent

	logger  *slog.Logger
	execCtx ExecutionContext

	mu        sync.Mutex
	lastLog   string
	eventLog  []string
	eventSeed atomic.Uint64
}

// Config wires the dependencies required by the dummy agent.
type Config struct {
	BaseConfig sdk.BaseConfig
	Logger     *slog.Logger
	ExecCtx    ExecutionContext
}

// New constructs the dummy runtime agent.
func New(cfg Config) (*Agent, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	base, err := sdk.NewBaseAgent(cfg.BaseConfig)
	if err != nil {
		return nil, err
	}
	return &Agent{
		BaseAgent: base,
		logger:    cfg.Logger.With("agent_id", cfg.BaseConfig.Identity.AgentID),
		execCtx:   cfg.ExecCtx,
	}, nil
}

// Init subscribes to the task domain and reports readiness.
func (a *Agent) Init(ctx context.Context) error {
	a.logger.Info("dummy_agent.init",
		"execution_id", a.execCtx.ExecutionID,
		"root_execution_id", a.execCtx.RootExecutionID,
		"correlation_id", a.execCtx.CorrelationID,
	)
	sub := a.Subscriber()
	if sub == nil {
		return fmt.Errorf("dummy agent subscriber not configured")
	}
	if err := sub.Subscribe(ctx, string(event.DomainTask), a.handleSubscribedEvent); err != nil {
		return err
	}
	return nil
}

// Start emits a lifecycle event proving the runtime publisher path.
func (a *Agent) Start(ctx context.Context) error {
	a.logger.Info("dummy_agent.start")
	payload := event.LifecycleStateReportedPayload{AgentState: event.AgentStateWorking}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	evt := sdk.Event{
		ID:        a.nextEventID("lifecycle"),
		Type:      string(event.TypeLifecycleStateReported),
		Payload:   body,
		Metadata:  a.baseMetadata(),
		Timestamp: a.Clock().Now(),
	}
	if err := a.Publish(ctx, evt); err != nil {
		return err
	}
	a.recordLog(fmt.Sprintf("published lifecycle %s", evt.ID))
	return nil
}

func (a *Agent) handleSubscribedEvent(ctx context.Context, evt sdk.Event) error {
	return a.HandleEvent(ctx, evt)
}

// HandleEvent records the inbound task, ACKs it, and emits a completion event.
func (a *Agent) HandleEvent(ctx context.Context, evt sdk.Event) error {
	if evt.Type != string(event.TypeTaskCreated) {
		a.logger.Info("dummy_agent.ignore_event",
			"inbound_event_id", evt.ID,
			"type", evt.Type,
			"correlation_id", evt.Metadata["correlation_id"],
		)
		return nil
	}
	a.logger.Info("dummy_agent.handle",
		"inbound_event_id", evt.ID,
		"type", evt.Type,
		"correlation_id", evt.Metadata["correlation_id"],
	)
	var task event.TaskCreatedPayload
	if err := json.Unmarshal(evt.Payload, &task); err != nil {
		a.logger.Warn("dummy_agent.payload_decode_failed", "event_id", evt.ID, "error", err)
	}
	if ack := a.Acker(); ack != nil {
		if err := ack.Ack(ctx, evt.ID); err != nil {
			a.logger.Error("dummy_agent.ack_failed", "event_id", evt.ID, "error", err)
		}
	}

	payload := event.TaskStatusPayload{
		TaskID:    task.TaskID,
		Status:    event.TaskStatusCompleted,
		Reason:    "dummy agent satisfied task",
		Artifacts: []string{"artifact://dummy/evidence.txt"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	response := sdk.Event{
		ID:        a.nextEventID("task-completed"),
		Type:      string(event.TypeTaskCompleted),
		Payload:   body,
		Metadata:  a.baseMetadata(),
		Timestamp: a.Clock().Now(),
	}
	if err := a.Publish(ctx, response); err != nil {
		return err
	}
	a.recordLog(fmt.Sprintf("handled %s -> %s", evt.ID, response.ID))
	return nil
}

// Shutdown records the lifecycle closure in logs.
func (a *Agent) Shutdown(context.Context) error {
	a.logger.Info("dummy_agent.shutdown")
	return nil
}

// LastReport exposes the agent's last execution summary and log lines.
func (a *Agent) LastReport() (string, []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	logCopy := make([]string, len(a.eventLog))
	copy(logCopy, a.eventLog)
	return a.lastLog, logCopy
}

func (a *Agent) recordLog(entry string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastLog = entry
	a.eventLog = append(a.eventLog, fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339Nano), entry))
}

func (a *Agent) baseMetadata() map[string]string {
	return map[string]string{
		"correlation_id":  a.execCtx.CorrelationID,
		"target_agent_id": "",
	}
}

func (a *Agent) nextEventID(prefix string) string {
	counter := a.eventSeed.Add(1)
	return fmt.Sprintf("%s-%d", prefix, counter)
}
