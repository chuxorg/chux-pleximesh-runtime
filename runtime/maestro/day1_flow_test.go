package maestro

import (
	"context"
	"strings"
	"testing"

	"github.com/chuxorg/chux-agent-mesh/runtime/guardian"
)

func TestDayOneEngineerFlowWithGuardianStub(t *testing.T) {
	stubGuardian := guardian.NewStub()
	enforcer := &stubEnforcement{
		guardian: stubGuardian,
	}
	publisher := &recordingOutcomePublisher{}
	runStates := &recordingRunStatePublisher{}
	guidance := &recordingGuidancePublisher{}
	plans := &recordingExecutionPlanPublisher{}

	m := New(enforcer, publisher, runStates, guidance, plans)
	shortDraft := "Quick overview without required structure."
	submission := IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		From:        "human",
		To:          "maestro",
		Payload: IntentPayload{
			Intent: "day1-engineer-draft",
			Artifact: Artifact{
				Type:    "doc",
				Content: shortDraft,
			},
		},
	}

	if err := m.HandleIntent(context.Background(), submission); err != nil {
		t.Fatalf("HandleIntent failed: %v", err)
	}

	if len(enforcer.tasks) != 2 {
		t.Fatalf("expected 2 task.evaluate attempts, got %d", len(enforcer.tasks))
	}

	first := enforcer.tasks[0]
	second := enforcer.tasks[1]

	assertTaskMatchesSchema(t, first)
	assertTaskMatchesSchema(t, second)

	if first.Artifact.Content != shortDraft {
		t.Fatalf("first attempt should send original draft, got: %q", first.Artifact.Content)
	}

	if second.Artifact.Content == shortDraft {
		t.Fatalf("second attempt should contain Maestro corrections")
	}

	if !containsWhatVsEnables(second.Artifact.Content) {
		t.Fatalf("corrected artifact missing PlexiMesh explanations: %s", second.Artifact.Content)
	}

	if utf8Len(second.Artifact.Content) < 200 {
		t.Fatalf("corrected artifact still shorter than 200 chars: %d", utf8Len(second.Artifact.Content))
	}

	if publisher.summary.Schema != outcomeSchema || publisher.summary.MessageType != outcomeMessageType {
		t.Fatalf("outcome message schema mismatch: %+v", publisher.summary)
	}

	if publisher.summary.Status != OutcomeStatusCorrectedAndAccepted {
		t.Fatalf("expected corrected_and_accepted status, got %s", publisher.summary.Status)
	}

	if publisher.summary.From != "maestro" || publisher.summary.To != "human" {
		t.Fatalf("outcome routing invalid: %+v", publisher.summary)
	}

	if publisher.summary.Summary == "" {
		t.Fatalf("outcome summary should not be empty")
	}

	expectedStates := []RunState{
		RunStateReceived,
		RunStateEvaluating,
		RunStateCorrecting,
		RunStateEvaluating,
		RunStateApproved,
	}

	if len(runStates.updates) != len(expectedStates) {
		t.Fatalf("expected %d run state updates, got %d", len(expectedStates), len(runStates.updates))
	}

	for i, state := range expectedStates {
		if runStates.updates[i].State != state {
			t.Fatalf("expected run state %s at position %d, got %s", state, i, runStates.updates[i].State)
		}
	}

	if len(plans.events) != 1 {
		t.Fatalf("expected a single execution plan, got %d", len(plans.events))
	}
}

type stubEnforcement struct {
	guardian guardian.Guardian
	tasks    []TaskEvaluate
}

func (s *stubEnforcement) EvaluateTask(ctx context.Context, msg TaskEvaluate) (guardian.DecisionResult, error) {
	s.tasks = append(s.tasks, msg)
	check := guardian.ComplianceCheck{
		Schema:      "compliance.check.v0",
		MessageType: "compliance.check",
		From:        "enforcement_agent",
		To:          "guardian",
		Artifact: guardian.ComplianceArtifact{
			Type:    msg.Artifact.Type,
			Content: msg.Artifact.Content,
		},
		Expectation: guardian.ComplianceExpectation{
			Posture: msg.Context.Posture,
		},
	}
	return s.guardian.EvaluateCompliance(check)
}

type recordingOutcomePublisher struct {
	summary OutcomeSummary
}

func (r *recordingOutcomePublisher) PublishOutcome(ctx context.Context, summary OutcomeSummary) error {
	r.summary = summary
	return nil
}

type recordingRunStatePublisher struct {
	updates []RunStateUpdate
}

func (r *recordingRunStatePublisher) PublishRunState(ctx context.Context, update RunStateUpdate) error {
	r.updates = append(r.updates, update)
	return nil
}

type recordingGuidancePublisher struct {
	events []MaestroGuidanceEvent
}

func (r *recordingGuidancePublisher) PublishGuidance(ctx context.Context, guidance MaestroGuidanceEvent) error {
	r.events = append(r.events, guidance)
	return nil
}

type recordingExecutionPlanPublisher struct {
	events []ExecutionPlanEvent
}

func (r *recordingExecutionPlanPublisher) PublishExecutionPlan(ctx context.Context, plan ExecutionPlanEvent) error {
	r.events = append(r.events, plan)
	return nil
}

func assertTaskMatchesSchema(t *testing.T, task TaskEvaluate) {
	t.Helper()
	if task.Schema != taskSchema {
		t.Fatalf("task schema mismatch: %s", task.Schema)
	}
	if task.MessageType != taskMessageType {
		t.Fatalf("task message type mismatch: %s", task.MessageType)
	}
	if task.From != "maestro" || task.To != "enforcement_agent" {
		t.Fatalf("task routing invalid: from=%s to=%s", task.From, task.To)
	}
}

func containsWhatVsEnables(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "what pleximesh is") && (strings.Contains(lower, "what it enables") || strings.Contains(lower, "what pleximesh enables"))
}
