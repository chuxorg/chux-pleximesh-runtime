package maestro

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/chuxorg/chux-agent-mesh/runtime/guardian"
	"github.com/chuxorg/chux-agent-mesh/runtime/qa"
)

const (
	intentSchema      = "intent.submission.v0"
	intentMessageType = "intent.submission"

	taskSchema      = "task.evaluate.v0"
	taskMessageType = "task.evaluate"

	outcomeSchema      = "outcome.summary.v0"
	outcomeMessageType = "outcome.summary"

	defaultPosture    = "early_sdlc"
	defaultMaxAttempt = 2

	IntentContextMetadataKey        = "intent_context"
	IntentContextDirective          = "directive"
	IntentContextClarificationReply = "clarification_response"

	executionPlanSchema      = "execution.plan.proposed.v0"
	executionPlanMessageType = "execution.plan.proposed"
)

// Maestro orchestrates message transformations between humans and enforcement agents.
type Maestro struct {
	enforcement EnforcementClient
	outcomes    OutcomePublisher
	runStates   RunStatePublisher
	guidance    GuidancePublisher
	plans       ExecutionPlanPublisher
	maxAttempts int

	mu               sync.Mutex
	awaitingGuidance bool
	runIDGenerator   func() string
	observer         qa.Observer
}

// EnforcementClient emits task.evaluate messages and returns Guardian decisions.
type EnforcementClient interface {
	EvaluateTask(ctx context.Context, msg TaskEvaluate) (guardian.DecisionResult, error)
}

// OutcomePublisher delivers human-readable outcome summaries back to the requester.
type OutcomePublisher interface {
	PublishOutcome(ctx context.Context, summary OutcomeSummary) error
}

// RunStatePublisher emits lifecycle updates that can be consumed by AWACS.
type RunStatePublisher interface {
	PublishRunState(ctx context.Context, update RunStateUpdate) error
}

// GuidancePublisher distributes Maestro guidance events verbatim.
type GuidancePublisher interface {
	PublishGuidance(ctx context.Context, guidance MaestroGuidanceEvent) error
}

// ExecutionPlanPublisher emits dry-run execution plans.
type ExecutionPlanPublisher interface {
	PublishExecutionPlan(ctx context.Context, plan ExecutionPlanEvent) error
}

// New creates a Maestro with sensible defaults.
func New(enforcement EnforcementClient, outcomes OutcomePublisher, runStates RunStatePublisher, guidance GuidancePublisher, plans ExecutionPlanPublisher) *Maestro {
	return &Maestro{
		enforcement: enforcement,
		outcomes:    outcomes,
		runStates:   runStates,
		guidance:    guidance,
		plans:       plans,
		maxAttempts: defaultMaxAttempt,
		runIDGenerator: generateRunID,
		observer:       qa.NopObserver{},
	}
}

// IntentSubmission matches intent.submission.v0.
type IntentSubmission struct {
	Schema      string                 `json:"schema"`
	MessageType string                 `json:"message_type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Payload     IntentPayload          `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IntentPayload wraps human-provided goals plus an artifact to review.
type IntentPayload struct {
	Intent   string   `json:"intent"`
	Artifact Artifact `json:"artifact"`
}

// Artifact represents reviewable material.
type Artifact struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// TaskEvaluate mirrors task.evaluate.v0.
type TaskEvaluate struct {
	Schema      string          `json:"schema"`
	MessageType string          `json:"message_type"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Context     TaskContext     `json:"context"`
	Constraints TaskConstraints `json:"constraints"`
	Artifact    Artifact        `json:"artifact"`
}

// TaskContext provides run metadata for enforcement agents.
type TaskContext struct {
	RunID   string `json:"run_id"`
	Posture string `json:"posture"`
	Intent  string `json:"intent"`
}

// TaskConstraints instruct downstream agents on allowed behaviors.
type TaskConstraints struct {
	PreserveIntent     bool `json:"preserve_intent"`
	AutoCorrectAllowed bool `json:"auto_correct_allowed"`
}

// OutcomeSummary mirrors outcome.summary.v0.
type OutcomeSummary struct {
	Schema      string        `json:"schema"`
	MessageType string        `json:"message_type"`
	From        string        `json:"from"`
	To          string        `json:"to"`
	Status      OutcomeStatus `json:"status"`
	Summary     string        `json:"summary"`
	Learning    []string      `json:"learning"`
}

// OutcomeStatus enumerates possible delivery states.
type OutcomeStatus string

const (
	OutcomeStatusAccepted             OutcomeStatus = "accepted"
	OutcomeStatusCorrectedAndAccepted OutcomeStatus = "corrected_and_accepted"
	OutcomeStatusBlocked              OutcomeStatus = "blocked"
)

// RunState represents Maestro observable phases.
type RunState string

const (
	RunStateReceived   RunState = "received"
	RunStateEvaluating RunState = "evaluating"
	RunStateCorrecting RunState = "correcting"
	RunStateApproved   RunState = "approved"
	RunStateBlocked    RunState = "blocked"
)

// RunStateUpdate carries the run identifier and the current state.
type RunStateUpdate struct {
	RunID   string   `json:"run_id"`
	State   RunState `json:"state"`
	Attempt int      `json:"attempt"`
}

// MaestroGuidanceEvent defines a human-readable prompt Maestro issued.
type MaestroGuidanceEvent struct {
	RunID            string           `json:"run_id"`
	Attempt          int              `json:"attempt"`
	Posture          string           `json:"posture"`
	Explanation      string           `json:"explanation"`
	Options          []GuidanceOption `json:"options"`
	RequiresResponse bool             `json:"requires_response"`
}

// GuidanceOption represents a bounded human option.
type GuidanceOption struct {
	Label     string `json:"label"`
	Rationale string `json:"rationale"`
}

// ExecutionPlanEvent captures a dry-run plan that AWACS can render.
type ExecutionPlanEvent struct {
	Schema      string              `json:"schema"`
	MessageType string              `json:"message_type"`
	RunID       string              `json:"run_id"`
	Posture     string              `json:"posture"`
	Attempt     int                 `json:"attempt"`
	Intent      string              `json:"intent"`
	Steps       []ExecutionPlanStep `json:"steps"`
}

// ExecutionPlanStep enumerates the actors involved in the dry-run.
type ExecutionPlanStep struct {
	StepID           string   `json:"step_id"`
	Description      string   `json:"description"`
	ResponsibleAgent string   `json:"responsible_agent"`
	Inputs           []string `json:"inputs"`
	Outputs          []string `json:"outputs"`
}

// HandleIntent ingests a human submission and coordinates Guardian review.
func (m *Maestro) HandleIntent(ctx context.Context, submission IntentSubmission) error {
	if m == nil {
		return errors.New("maestro is nil")
	}
	if m.enforcement == nil {
		return errors.New("maestro enforcement client is not configured")
	}
	if m.outcomes == nil {
		return errors.New("maestro outcome publisher is not configured")
	}

	m.resolveGuidanceFromIntent(submission.Metadata)
	runID := m.runIDGenerator()
	if runID == "" {
		runID = generateRunID()
	}
	m.observe("maestro.run_id_generated", "random", fmt.Sprintf("run_id=%s", runID))
	posture := m.inferPosture(submission)
	intent := m.inferIntent(submission)
	artifact := Artifact{
		Type:    submission.Payload.Artifact.Type,
		Content: submission.Payload.Artifact.Content,
	}

	m.emitRunState(ctx, runID, RunStateReceived, 0)

	attempt := 0
	for attempt < m.maxAttempts {
		attempt++
		m.emitRunState(ctx, runID, RunStateEvaluating, attempt)
		taskMsg := m.newTaskEvaluate(runID, posture, intent, artifact)
		m.observe("maestro.enforcement.evaluate_task", "external_call", fmt.Sprintf("run_id=%s attempt=%d", runID, attempt))
		decision, err := m.enforcement.EvaluateTask(ctx, taskMsg)
		if err != nil {
			return fmt.Errorf("evaluate task: %w", err)
		}

		if decision.Decision.Outcome == guardian.OutcomeApprove {
			status := OutcomeStatusAccepted
			if attempt > 1 {
				status = OutcomeStatusCorrectedAndAccepted
			}
			summary := fmt.Sprintf("Guardian approved run %s after %d attempt(s).", runID, attempt)
			learning := extractLearning(decision.Guidance.Description)
			m.emitRunState(ctx, runID, RunStateApproved, attempt)
			m.emitExecutionPlan(ctx, runID, posture, intent, artifact, attempt)
			m.observe("maestro.outcome.publish", "event", fmt.Sprintf("run_id=%s status=%s", runID, status))
			return m.outcomes.PublishOutcome(ctx, newOutcomeSummary(status, summary, learning))
		}

		if !decision.Guidance.AutoFixable || attempt == m.maxAttempts {
			summary := fmt.Sprintf("Guardian blocked run %s: %s", runID, decision.Guidance.Description)
			learning := extractLearning(decision.Guidance.Description)
			m.emitRunState(ctx, runID, RunStateBlocked, attempt)
			m.observe("maestro.outcome.publish", "event", fmt.Sprintf("run_id=%s status=%s", runID, OutcomeStatusBlocked))
			return m.outcomes.PublishOutcome(ctx, newOutcomeSummary(OutcomeStatusBlocked, summary, learning))
		}

		m.emitRunState(ctx, runID, RunStateCorrecting, attempt)
		artifact = applyMinimalCorrection(artifact)
	}

	return nil
}

func (m *Maestro) newTaskEvaluate(runID, posture, intent string, artifact Artifact) TaskEvaluate {
	return TaskEvaluate{
		Schema:      taskSchema,
		MessageType: taskMessageType,
		From:        "maestro",
		To:          "enforcement_agent",
		Context: TaskContext{
			RunID:   runID,
			Posture: posture,
			Intent:  intent,
		},
		Constraints: TaskConstraints{
			PreserveIntent:     true,
			AutoCorrectAllowed: true,
		},
		Artifact: artifact,
	}
}

func (m *Maestro) inferPosture(IntentSubmission) string {
	return defaultPosture
}

func (m *Maestro) inferIntent(sub IntentSubmission) string {
	intent := strings.TrimSpace(sub.Payload.Intent)
	if intent != "" {
		return intent
	}
	if t := strings.TrimSpace(sub.Payload.Artifact.Type); t != "" {
		return fmt.Sprintf("refine_%s_artifact", t)
	}
	return "refine_artifact"
}

func newOutcomeSummary(status OutcomeStatus, summary string, learning []string) OutcomeSummary {
	return OutcomeSummary{
		Schema:      outcomeSchema,
		MessageType: outcomeMessageType,
		From:        "maestro",
		To:          "human",
		Status:      status,
		Summary:     summary,
		Learning:    learning,
	}
}

func extractLearning(description string) []string {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return nil
	}
	return []string{desc}
}

func applyMinimalCorrection(artifact Artifact) Artifact {
	content := strings.TrimSpace(artifact.Content)
	lower := strings.ToLower(content)
	if !strings.Contains(lower, "what pleximesh is") {
		content += "\n\nWhat PlexiMesh is: PlexiMesh is a coordination fabric that aligns autonomous agents."
	}
	lower = strings.ToLower(content)
	if !strings.Contains(lower, "what it enables") && !strings.Contains(lower, "what pleximesh enables") {
		content += "\nWhat it enables: The fabric unlocks trustworthy collaboration and runtime governance."
	}

	for utf8.RuneCountInString(content) < 200 {
		content += " PlexiMesh provides consistent checkpoints across engineering lifecycles."
	}

	artifact.Content = content
	return artifact
}

func generateRunID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return "run-" + hex.EncodeToString(buf)
	}
	return "run-fallback"
}

func (m *Maestro) emitRunState(ctx context.Context, runID string, state RunState, attempt int) {
	if m == nil || m.runStates == nil {
		return
	}
	m.observe("maestro.run_state.publish", "event", fmt.Sprintf("run_id=%s state=%s attempt=%d", runID, state, attempt))
	_ = m.runStates.PublishRunState(ctx, RunStateUpdate{
		RunID:   runID,
		State:   state,
		Attempt: attempt,
	})
}

// IssueGuidance emits a maestro.guidance.issued event for AWACS consumption.
func (m *Maestro) IssueGuidance(ctx context.Context, event MaestroGuidanceEvent) error {
	if m == nil || m.guidance == nil {
		return nil
	}
	m.setGuidancePending(event.RequiresResponse)
	m.observe("maestro.guidance.publish", "event", fmt.Sprintf("run_id=%s attempt=%d", event.RunID, event.Attempt))
	return m.guidance.PublishGuidance(ctx, event)
}

func (m *Maestro) emitExecutionPlan(ctx context.Context, runID, posture, intent string, artifact Artifact, attempt int) {
	if m == nil || m.plans == nil {
		return
	}
	if m.isGuidancePending() {
		return
	}
	plan := buildExecutionPlan(runID, posture, intent, artifact, attempt)
	m.observe("maestro.plan.publish", "event", fmt.Sprintf("run_id=%s attempt=%d", runID, attempt))
	_ = m.plans.PublishExecutionPlan(ctx, plan)
}

func buildExecutionPlan(runID, posture, intent string, artifact Artifact, attempt int) ExecutionPlanEvent {
	artifactLabel := fmt.Sprintf("%s artifact", artifact.Type)
	steps := []ExecutionPlanStep{
		{
			StepID:           "step-1",
			ResponsibleAgent: "maestro",
			Description:      fmt.Sprintf("Translate intent \"%s\" into Guardian-aligned constraints for operators.", intent),
			Inputs:           []string{"intent.submission payload", artifactLabel},
			Outputs:          []string{"task.evaluate context"},
		},
		{
			StepID:           "step-2",
			ResponsibleAgent: "enforcement_agent",
			Description:      "Stage compliance boundaries before any execution occurs.",
			Inputs:           []string{"task.evaluate context"},
			Outputs:          []string{"compliance.check message"},
		},
		{
			StepID:           "step-3",
			ResponsibleAgent: "guardian",
			Description:      "Validate the proposed work against runtime guardrails and document approvals.",
			Inputs:           []string{"compliance.check message", fmt.Sprintf("posture: %s", posture)},
			Outputs:          []string{"decision.result", "policy learning summary"},
		},
		{
			StepID:           "step-4",
			ResponsibleAgent: "human_operator",
			Description:      "Execute the approved work using Guardian learning as the guide. No automation occurs here.",
			Inputs:           []string{"decision.result", artifactLabel},
			Outputs:          []string{"delivered work product"},
		},
	}

	return ExecutionPlanEvent{
		Schema:      executionPlanSchema,
		MessageType: executionPlanMessageType,
		RunID:       runID,
		Posture:     posture,
		Attempt:     attempt,
		Intent:      intent,
		Steps:       steps,
	}
}

func (m *Maestro) resolveGuidanceFromIntent(metadata map[string]interface{}) {
	if extractIntentContext(metadata) == IntentContextClarificationReply {
		m.setGuidancePending(false)
	}
}

func (m *Maestro) setGuidancePending(pending bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.awaitingGuidance = pending
}

func (m *Maestro) isGuidancePending() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.awaitingGuidance
}

func extractIntentContext(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[IntentContextMetadataKey]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
