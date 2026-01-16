package testingagent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chuxorg/chux-agent-mesh/runtime/maestro"
	"github.com/chuxorg/chux-agent-mesh/runtime/transport"
)

const (
	intentEventType              = "intent.submission"
	executionPlanEventType       = "execution.plan.proposed"
	executionGateEvaluatedType   = "execution.gate.evaluated"
	runCompletedEventType        = "run.completed"
	runAbortedEventType          = "run.aborted"
	runStateUpdateEventType      = "run_state.update"
	testingReportGeneratedType   = "testing.report.generated"
	defaultTestingAgentQueueSize = 256
)

// Agent observes runtime events and emits post-run testing reports.
type Agent struct {
	bus            *transport.Bus
	sequence       int
	pendingIntents []pendingIntent
	runs           map[string]*runTrace
	finalizeCh     chan string
}

// Report summarizes a single run's findings.
type Report struct {
	RunID    string          `json:"run_id"`
	Summary  string          `json:"summary"`
	Findings []ReportFinding `json:"findings"`
}

// ReportFinding records an informational, warning, or error observation.
type ReportFinding struct {
	Severity string `json:"severity"` // info | warning | error
	Detail   string `json:"detail"`
}

type pendingIntent struct {
	description string
	order       int
}

type runTrace struct {
	runID             string
	intentObserved    bool
	intentMissing     bool
	intentOrder       int
	intentSummary     string
	planSeen          bool
	planOrder         int
	gateSeen          bool
	gateOrder         int
	runCompletedEvent bool
	runAbortedEvent   bool
	finalState        string
	reported          bool
	expectPlan        bool
}

// New creates a testing agent bound to the shared transport bus.
func New(bus *transport.Bus) *Agent {
	return &Agent{
		bus:        bus,
		runs:       make(map[string]*runTrace),
		finalizeCh: make(chan string, defaultTestingAgentQueueSize),
	}
}

// Start begins observing events and emitting testing reports.
func (a *Agent) Start(ctx context.Context) {
	sub := a.bus.Subscribe(defaultTestingAgentQueueSize)
	go a.consume(ctx, sub)
}

func (a *Agent) consume(ctx context.Context, sub <-chan transport.Event) {
	for {
		select {
		case evt, ok := <-sub:
			if !ok {
				return
			}
			a.handleEvent(ctx, evt)
		case runID := <-a.finalizeCh:
			a.handleFinalizeSignal(ctx, runID)
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) handleEvent(ctx context.Context, evt transport.Event) {
	switch evt.Type {
	case intentEventType:
		a.recordIntentEvent(evt)
	case runStateUpdateEventType:
		a.handleRunStateUpdate(ctx, evt)
	case executionPlanEventType:
		a.handlePlanEvent(evt)
	case executionGateEvaluatedType:
		a.handleGateEvent(evt)
	case runCompletedEventType:
		a.handleCompletionEvent(ctx, evt, true)
	case runAbortedEventType:
		a.handleCompletionEvent(ctx, evt, false)
	case testingReportGeneratedType:
		// ignore self-emitted reports
	default:
		// no-op for other events
	}
}

func (a *Agent) recordIntentEvent(evt transport.Event) {
	a.sequence++
	description := "unknown intent"
	if submission, ok := evt.Payload.(maestro.IntentSubmission); ok {
		text := strings.TrimSpace(submission.Payload.Intent)
		if text == "" {
			text = strings.TrimSpace(submission.Payload.Artifact.Type)
		}
		if text != "" {
			description = text
		}
	}
	a.pendingIntents = append(a.pendingIntents, pendingIntent{
		description: description,
		order:       a.sequence,
	})
}

func (a *Agent) handleRunStateUpdate(ctx context.Context, evt transport.Event) {
	runID := evt.Metadata["run_id"]
	if runID == "" {
		return
	}
	state := strings.ToLower(evt.Metadata["state"])
	trace := a.ensureTrace(runID)
	switch state {
	case string(maestro.RunStateReceived):
		a.assignIntentToRun(trace)
	case string(maestro.RunStateApproved):
		trace.finalState = "approved"
		trace.expectPlan = true
		a.scheduleFinalization(ctx, trace)
	case string(maestro.RunStateBlocked):
		trace.finalState = "blocked"
		a.scheduleFinalization(ctx, trace)
	}
}

func (a *Agent) handlePlanEvent(evt transport.Event) {
	runID := evt.Metadata["run_id"]
	if runID == "" {
		return
	}
	trace := a.ensureTrace(runID)
	if trace.planSeen {
		return
	}
	trace.planSeen = true
	trace.planOrder = a.nextOrder()
}

func (a *Agent) handleGateEvent(evt transport.Event) {
	runID := evt.Metadata["run_id"]
	if runID == "" {
		return
	}
	trace := a.ensureTrace(runID)
	if trace.gateSeen {
		return
	}
	trace.gateSeen = true
	trace.gateOrder = a.nextOrder()
}

func (a *Agent) handleCompletionEvent(ctx context.Context, evt transport.Event, completed bool) {
	runID := evt.Metadata["run_id"]
	if runID == "" {
		return
	}
	trace := a.ensureTrace(runID)
	if completed {
		trace.runCompletedEvent = true
		if trace.finalState == "" {
			trace.finalState = "completed"
		}
	} else {
		trace.runAbortedEvent = true
		if trace.finalState == "" {
			trace.finalState = "aborted"
		}
	}
	a.finalizeRun(ctx, trace)
}

func (a *Agent) assignIntentToRun(trace *runTrace) {
	if trace.intentObserved {
		return
	}
	if len(a.pendingIntents) == 0 {
		trace.intentMissing = true
		return
	}
	intent := a.pendingIntents[0]
	a.pendingIntents = a.pendingIntents[1:]
	trace.intentObserved = true
	trace.intentOrder = intent.order
	trace.intentSummary = intent.description
	if trace.intentOrder > a.sequence {
		a.sequence = trace.intentOrder
	}
}

func (a *Agent) ensureTrace(runID string) *runTrace {
	if trace, ok := a.runs[runID]; ok {
		return trace
	}
	trace := &runTrace{
		runID:         runID,
		intentSummary: "unspecified",
	}
	a.runs[runID] = trace
	return trace
}

func (a *Agent) nextOrder() int {
	a.sequence++
	return a.sequence
}

func (a *Agent) finalizeRun(ctx context.Context, trace *runTrace) {
	if trace.reported {
		return
	}
	findings := trace.evaluateFindings()
	summary := trace.buildSummary(findings)
	report := Report{
		RunID:    trace.runID,
		Summary:  summary,
		Findings: findings,
	}
	meta := map[string]string{"run_id": trace.runID}
	a.bus.Publish(transport.Event{
		Type:     testingReportGeneratedType,
		Payload:  report,
		Metadata: meta,
	})
	log.Printf("testing.agent run_id=%s summary=%s findings=%d", trace.runID, summary, len(findings))
	trace.reported = true
	delete(a.runs, trace.runID)
}

func (t *runTrace) evaluateFindings() []ReportFinding {
	var findings []ReportFinding
	if !t.intentObserved {
		if t.intentMissing {
			findings = append(findings, warning("intent.submission event missing before run_state.received"))
		} else {
			findings = append(findings, warning("intent.submission event not correlated to run"))
		}
	}
	if t.expectPlan && !t.planSeen {
		findings = append(findings, errorFinding("execution.plan.proposed event missing for approved run"))
	}
	if !t.gateSeen {
		findings = append(findings, warning("execution.gate.evaluated event missing for run"))
	}
	if t.finalState == "approved" && !t.runCompletedEvent {
		findings = append(findings, warning("run.completed event missing"))
	}
	if t.finalState == "blocked" && !t.runAbortedEvent {
		findings = append(findings, warning("run.aborted event missing"))
	}
	if t.planSeen && t.intentObserved && t.planOrder < t.intentOrder {
		findings = append(findings, warning("execution plan emitted before intent submission correlation"))
	}
	if t.gateSeen && t.planSeen && t.gateOrder < t.planOrder {
		findings = append(findings, warning("execution gate evaluated before plan proposal"))
	}
	return findings
}

func (t *runTrace) buildSummary(findings []ReportFinding) string {
	status := "completed"
	switch t.finalState {
	case "blocked", "aborted":
		status = "aborted"
	case "":
		status = "finished"
	}
	severity := "no findings"
	hasErrors := false
	hasWarnings := false
	for _, f := range findings {
		switch strings.ToLower(f.Severity) {
		case "error":
			hasErrors = true
		case "warning":
			hasWarnings = true
		}
	}
	if hasErrors {
		severity = "errors detected"
	} else if hasWarnings {
		severity = "warnings detected"
	}
	intent := t.intentSummary
	if intent == "" {
		intent = "unspecified intent"
	}
	return fmt.Sprintf("Run %s (%s) %s with %s", t.runID, intent, status, severity)
}

func warning(detail string) ReportFinding {
	return ReportFinding{Severity: "warning", Detail: detail}
}

func errorFinding(detail string) ReportFinding {
	return ReportFinding{Severity: "error", Detail: detail}
}

const finalizeGrace = 20 * time.Millisecond

func (a *Agent) scheduleFinalization(ctx context.Context, trace *runTrace) {
	runID := trace.runID
	go func() {
		select {
		case <-time.After(finalizeGrace):
			select {
			case a.finalizeCh <- runID:
			case <-ctx.Done():
			}
		case <-ctx.Done():
		}
	}()
}

func (a *Agent) handleFinalizeSignal(ctx context.Context, runID string) {
	trace, ok := a.runs[runID]
	if !ok || trace.reported {
		return
	}
	if trace.runCompletedEvent || trace.runAbortedEvent || trace.finalState != "" {
		a.finalizeRun(ctx, trace)
	}
}
