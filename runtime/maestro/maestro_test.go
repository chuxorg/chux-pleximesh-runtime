package maestro

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chuxorg/chux-agent-mesh/runtime/guardian"
)

func TestHandleIntentApprovedFirstAttempt(t *testing.T) {
	enforcer := &fakeEnforcer{
		decisions: []guardian.DecisionResult{
			mockDecision(guardian.OutcomeApprove, false, "Ready to publish."),
		},
	}
	publisher := &fakePublisher{}

	m := New(enforcer, publisher, &noopRunStatePublisher{}, &noopGuidancePublisher{}, &noopExecutionPlanPublisher{})
	err := m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Payload: IntentPayload{
			Intent: "document pleximesh principles",
			Artifact: Artifact{
				Type:    "doc",
				Content: compliantContent(),
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error: %v", err)
	}

	if len(enforcer.tasks) != 1 {
		t.Fatalf("expected 1 task evaluate, got %d", len(enforcer.tasks))
	}

	if publisher.summary.Status != OutcomeStatusAccepted {
		t.Fatalf("expected accepted status, got %s", publisher.summary.Status)
	}

	if publisher.summary.Summary == "" || !strings.Contains(publisher.summary.Summary, "Guardian approved run") {
		t.Fatalf("unexpected summary: %s", publisher.summary.Summary)
	}
}

func TestHandleIntentAutoFixAndApprove(t *testing.T) {
	enforcer := &fakeEnforcer{
		decisions: []guardian.DecisionResult{
			mockDecision(guardian.OutcomeReject, true, "Needs clarity."),
			mockDecision(guardian.OutcomeApprove, false, "Looks solid."),
		},
	}
	publisher := &fakePublisher{}

	m := New(enforcer, publisher, &noopRunStatePublisher{}, &noopGuidancePublisher{}, &noopExecutionPlanPublisher{})
	err := m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Payload: IntentPayload{
			Intent: "explain pleximesh",
			Artifact: Artifact{
				Type:    "doc",
				Content: "brief blurb missing structure",
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error: %v", err)
	}

	if len(enforcer.tasks) != 2 {
		t.Fatalf("expected 2 task evaluate attempts, got %d", len(enforcer.tasks))
	}

	second := enforcer.tasks[1]
	lower := strings.ToLower(second.Artifact.Content)
	if !strings.Contains(lower, "what pleximesh is") || !strings.Contains(lower, "what it enables") {
		t.Fatalf("auto-correction missing required phrasing: %s", second.Artifact.Content)
	}

	if utf8Len(second.Artifact.Content) < 200 {
		t.Fatalf("auto-correction did not meet length requirement: %d", utf8Len(second.Artifact.Content))
	}

	if publisher.summary.Status != OutcomeStatusCorrectedAndAccepted {
		t.Fatalf("expected corrected_and_accepted status, got %s", publisher.summary.Status)
	}
}

func TestHandleIntentBlockedWhenNotAutoFixable(t *testing.T) {
	enforcer := &fakeEnforcer{
		decisions: []guardian.DecisionResult{
			mockDecision(guardian.OutcomeReject, false, "Policy violation."),
		},
	}
	publisher := &fakePublisher{}

	m := New(enforcer, publisher, &noopRunStatePublisher{}, &noopGuidancePublisher{}, &noopExecutionPlanPublisher{})
	err := m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Payload: IntentPayload{
			Intent: "",
			Artifact: Artifact{
				Type:    "doc",
				Content: "Content with [POLICY_VIOLATION]",
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error: %v", err)
	}

	if len(enforcer.tasks) != 1 {
		t.Fatalf("expected single attempt, got %d", len(enforcer.tasks))
	}
	if publisher.summary.Status != OutcomeStatusBlocked {
		t.Fatalf("expected blocked status, got %s", publisher.summary.Status)
	}
	if !strings.Contains(publisher.summary.Summary, "blocked") {
		t.Fatalf("summary should explain blockage: %s", publisher.summary.Summary)
	}
}

func TestHandleIntentEmitsExecutionPlanAfterApproval(t *testing.T) {
	enforcer := &fakeEnforcer{
		decisions: []guardian.DecisionResult{
			mockDecision(guardian.OutcomeApprove, false, "Ship it."),
		},
	}
	publisher := &fakePublisher{}
	plans := &recordingExecutionPlanPublisher{}

	m := New(enforcer, publisher, &noopRunStatePublisher{}, &noopGuidancePublisher{}, plans)
	err := m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Payload: IntentPayload{
			Intent: "draft pleximesh overview",
			Artifact: Artifact{
				Type:    "doc",
				Content: compliantContent(),
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error: %v", err)
	}

	if len(plans.events) != 1 {
		t.Fatalf("expected execution plan to be emitted, got %d", len(plans.events))
	}

	plan := plans.events[0]
	if plan.Schema != executionPlanSchema || plan.MessageType != executionPlanMessageType {
		t.Fatalf("plan schema mismatch: %+v", plan)
	}
	if plan.RunID == "" || len(plan.Steps) == 0 {
		t.Fatalf("plan missing details: %+v", plan)
	}
	for _, step := range plan.Steps {
		if step.StepID == "" || step.ResponsibleAgent == "" || len(step.Inputs) == 0 || len(step.Outputs) == 0 {
			t.Fatalf("plan step missing fields: %+v", step)
		}
	}
}

func TestExecutionPlanWaitsForGuidanceResolution(t *testing.T) {
	enforcer := &fakeEnforcer{
		decisions: []guardian.DecisionResult{
			mockDecision(guardian.OutcomeApprove, false, "Cleared."),
			mockDecision(guardian.OutcomeApprove, false, "Cleared again."),
		},
	}
	publisher := &fakePublisher{}
	plans := &recordingExecutionPlanPublisher{}

	m := New(enforcer, publisher, &noopRunStatePublisher{}, &noopGuidancePublisher{}, plans)
	guidanceEvent := MaestroGuidanceEvent{
		RunID:            "pending-run",
		Attempt:          1,
		Posture:          "early_sdlc",
		Explanation:      "Need clarification.",
		RequiresResponse: true,
	}
	if err := m.IssueGuidance(context.Background(), guidanceEvent); err != nil {
		t.Fatalf("IssueGuidance returned error: %v", err)
	}

	err := m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Payload: IntentPayload{
			Intent: "new directive",
			Artifact: Artifact{
				Type:    "doc",
				Content: compliantContent(),
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error: %v", err)
	}

	if len(plans.events) != 0 {
		t.Fatalf("plan should not emit while guidance pending")
	}

	err = m.HandleIntent(context.Background(), IntentSubmission{
		Schema:      intentSchema,
		MessageType: intentMessageType,
		Metadata: map[string]interface{}{
			IntentContextMetadataKey: IntentContextClarificationReply,
		},
		Payload: IntentPayload{
			Intent: "follow-up directive",
			Artifact: Artifact{
				Type:    "doc",
				Content: compliantContent(),
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleIntent returned error on clarification: %v", err)
	}

	if len(plans.events) != 1 {
		t.Fatalf("expected plan emission after clarification, got %d", len(plans.events))
	}
}

type fakeEnforcer struct {
	decisions []guardian.DecisionResult
	tasks     []TaskEvaluate
}

func (f *fakeEnforcer) EvaluateTask(ctx context.Context, msg TaskEvaluate) (guardian.DecisionResult, error) {
	f.tasks = append(f.tasks, msg)
	if len(f.decisions) == 0 {
		return guardian.DecisionResult{}, fmt.Errorf("no decision configured")
	}
	decision := f.decisions[0]
	f.decisions = f.decisions[1:]
	return decision, nil
}

type fakePublisher struct {
	summary OutcomeSummary
}

func (p *fakePublisher) PublishOutcome(ctx context.Context, summary OutcomeSummary) error {
	p.summary = summary
	return nil
}

type noopRunStatePublisher struct{}

func (n *noopRunStatePublisher) PublishRunState(ctx context.Context, update RunStateUpdate) error {
	return nil
}

type noopGuidancePublisher struct{}

func (n *noopGuidancePublisher) PublishGuidance(ctx context.Context, guidance MaestroGuidanceEvent) error {
	return nil
}

type noopExecutionPlanPublisher struct{}

func (n *noopExecutionPlanPublisher) PublishExecutionPlan(ctx context.Context, plan ExecutionPlanEvent) error {
	return nil
}

func mockDecision(outcome guardian.Outcome, autoFix bool, desc string) guardian.DecisionResult {
	return guardian.DecisionResult{
		Schema:      "decision.result.v0",
		MessageType: "decision.result",
		From:        "guardian",
		To:          "maestro",
		Decision: guardian.DecisionPayload{
			Outcome: outcome,
		},
		Guidance: guardian.GuidancePayload{
			Category:    guardian.GuidanceCategoryCorrectness,
			Description: desc,
			AutoFixable: autoFix,
		},
	}
}

func compliantContent() string {
	text := "This note explains what PlexiMesh is and walks through what it enables across the runtime. "
	text += "What PlexiMesh is: PlexiMesh is a collaboration fabric that sequences approvals. "
	text += "What it enables: The same workflow unlocks confident iteration across engineering and QA. "
	for utf8Len(text) < 220 {
		text += "Its governance anchors early SDL stages with clarity. "
	}
	return text
}

func utf8Len(s string) int {
	return len([]rune(s))
}
