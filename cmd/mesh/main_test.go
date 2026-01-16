package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/runtime/maestro"
	"github.com/chuxorg/chux-agent-mesh/runtime/transport"
)

func TestBusGuidancePublisherPublishesEvent(t *testing.T) {
	bus := transport.NewBus()
	sub := bus.Subscribe(1)
	publisher := &busGuidancePublisher{bus: bus}

	guidance := maestro.MaestroGuidanceEvent{
		RunID:       "run-guidance",
		Attempt:     0,
		Posture:     "early_sdlc",
		Explanation: "Please confirm the target persona for this brochure.",
		Options: []maestro.GuidanceOption{
			{Label: "Proceed", Rationale: "Human approved the current angle."},
		},
		RequiresResponse: true,
	}

	if err := publisher.PublishGuidance(context.Background(), guidance); err != nil {
		t.Fatalf("PublishGuidance returned error: %v", err)
	}

	select {
	case evt := <-sub:
		if evt.Type != "maestro.guidance.issued" {
			t.Fatalf("unexpected event type: %s", evt.Type)
		}
		payload, ok := evt.Payload.(maestro.MaestroGuidanceEvent)
		if !ok {
			t.Fatalf("payload type mismatch: %T", evt.Payload)
		}
		if payload.Explanation != guidance.Explanation {
			t.Fatalf("explanation mismatch: got %s, want %s", payload.Explanation, guidance.Explanation)
		}
		if len(payload.Options) != len(guidance.Options) {
			t.Fatalf("options length mismatch: got %d, want %d", len(payload.Options), len(guidance.Options))
		}
		if payload.Options[0].Label != guidance.Options[0].Label || payload.Options[0].Rationale != guidance.Options[0].Rationale {
			t.Fatalf("option mismatch: %+v", payload.Options[0])
		}
		if payload.RequiresResponse != guidance.RequiresResponse {
			t.Fatalf("requires_response mismatch: %v vs %v", payload.RequiresResponse, guidance.RequiresResponse)
		}

		if evt.Metadata["run_id"] != guidance.RunID {
			t.Fatalf("missing run_id metadata: %+v", evt.Metadata)
		}
		if evt.Metadata["posture"] != guidance.Posture {
			t.Fatalf("missing posture metadata: %+v", evt.Metadata)
		}
		if evt.Metadata["attempt"] != strconv.Itoa(guidance.Attempt) {
			t.Fatalf("missing attempt metadata: %+v", evt.Metadata)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for guidance event")
	}
}

func TestBusExecutionPlanPublisherPublishesEvent(t *testing.T) {
	bus := transport.NewBus()
	sub := bus.Subscribe(1)
	publisher := &busExecutionPlanPublisher{bus: bus}

	plan := maestro.ExecutionPlanEvent{
		RunID:   "run-plan",
		Posture: "early_sdlc",
		Attempt: 1,
		Intent:  "document pleximesh",
		Steps: []maestro.ExecutionPlanStep{
			{
				StepID:           "step-1",
				ResponsibleAgent: "maestro",
				Description:      "Explain the plan",
				Inputs:           []string{"intent"},
				Outputs:          []string{"task.evaluate"},
			},
		},
	}
	if err := publisher.PublishExecutionPlan(context.Background(), plan); err != nil {
		t.Fatalf("PublishExecutionPlan returned error: %v", err)
	}

	select {
	case evt := <-sub:
		if evt.Type != executionPlanEventType {
			t.Fatalf("unexpected event type: %s", evt.Type)
		}
		payload, ok := evt.Payload.(maestro.ExecutionPlanEvent)
		if !ok {
			t.Fatalf("payload type mismatch: %T", evt.Payload)
		}
		if payload.RunID != plan.RunID || len(payload.Steps) != 1 {
			t.Fatalf("plan payload mismatch: %+v", payload)
		}

		if evt.Metadata["run_id"] != plan.RunID {
			t.Fatalf("missing run_id metadata: %+v", evt.Metadata)
		}
		if evt.Metadata["posture"] != plan.Posture {
			t.Fatalf("missing posture metadata: %+v", evt.Metadata)
		}
		if evt.Metadata["attempt"] != strconv.Itoa(plan.Attempt) {
			t.Fatalf("missing attempt metadata: %+v", evt.Metadata)
		}
		if evt.Metadata["intent"] != plan.Intent {
			t.Fatalf("missing intent metadata: %+v", evt.Metadata)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for plan event")
	}
}

func TestNormalizeIntentSetsDirectiveContextByDefault(t *testing.T) {
	submission := maestro.IntentSubmission{
		Payload: maestro.IntentPayload{
			Intent: "document pleximesh principles",
			Artifact: maestro.Artifact{
				Type:    "doc",
				Content: " quick brief ",
			},
		},
	}

	normalizeIntent(&submission)

	contextValue, _ := submission.Metadata[maestro.IntentContextMetadataKey].(string)
	if contextValue != maestro.IntentContextDirective {
		t.Fatalf("expected %s context, got %s", maestro.IntentContextDirective, contextValue)
	}
}

func TestNormalizeIntentDetectsClarificationResponses(t *testing.T) {
	submission := maestro.IntentSubmission{
		Metadata: map[string]interface{}{
			responseToMetadataKey: guidanceEventTypeIssued,
		},
		Payload: maestro.IntentPayload{
			Intent: "clarify target audience",
			Artifact: maestro.Artifact{
				Type:    "doc",
				Content: "Responding to Maestro guidance.",
			},
		},
	}

	normalizeIntent(&submission)

	contextValue, _ := submission.Metadata[maestro.IntentContextMetadataKey].(string)
	if contextValue != maestro.IntentContextClarificationReply {
		t.Fatalf("expected %s context, got %s", maestro.IntentContextClarificationReply, contextValue)
	}
}

func TestBuildIntentMetadataExportsContext(t *testing.T) {
	submission := maestro.IntentSubmission{
		Metadata: map[string]interface{}{
			maestro.IntentContextMetadataKey: maestro.IntentContextClarificationReply,
		},
	}

	meta := buildIntentMetadata(submission)
	if meta[maestro.IntentContextMetadataKey] != maestro.IntentContextClarificationReply {
		t.Fatalf("intent metadata missing context: %+v", meta)
	}
}
