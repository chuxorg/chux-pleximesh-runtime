package testingagent

import (
	"context"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/runtime/maestro"
	"github.com/chuxorg/chux-agent-mesh/runtime/transport"
)

func TestAgentEmitsReportWithAllEventsPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := transport.NewBus()
	agent := New(bus)
	agent.Start(ctx)

	reportCh := bus.Subscribe(32)

	runID := "run-agent-all"
	publishIntent(bus, "document pleximesh")
	publishRunState(bus, runID, maestro.RunStateReceived, 0)
	publishRunState(bus, runID, maestro.RunStateEvaluating, 1)
	publishPlan(bus, runID)
	publishGate(bus, runID)
	publishRunState(bus, runID, maestro.RunStateApproved, 1)
	publishRunCompleted(bus, runID)

	report := waitForReport(t, reportCh)
	if report.RunID != runID {
		t.Fatalf("unexpected run id in report: %s", report.RunID)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings, got %d: %+v", len(report.Findings), report.Findings)
	}
}

func TestAgentWarnsWhenEventsMissingForBlockedRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := transport.NewBus()
	agent := New(bus)
	agent.Start(ctx)

	reportCh := bus.Subscribe(32)

	runID := "run-agent-blocked"
	publishIntent(bus, "rough draft")
	publishRunState(bus, runID, maestro.RunStateReceived, 0)
	publishRunState(bus, runID, maestro.RunStateEvaluating, 1)
	publishRunState(bus, runID, maestro.RunStateBlocked, 1)

	report := waitForReport(t, reportCh)
	if report.RunID != runID {
		t.Fatalf("unexpected run id in report: %s", report.RunID)
	}
	if len(report.Findings) == 0 {
		t.Fatalf("expected findings for missing events")
	}
	foundGate := false
	foundAbort := false
	for _, f := range report.Findings {
		if f.Detail == "execution.gate.evaluated event missing for run" {
			foundGate = true
		}
		if f.Detail == "run.aborted event missing" {
			foundAbort = true
		}
	}
	if !foundGate || !foundAbort {
		t.Fatalf("expected gate and abort findings, got %+v", report.Findings)
	}
}

func publishIntent(bus *transport.Bus, intent string) {
	bus.Publish(transport.Event{
		Type: intentEventType,
		Payload: maestro.IntentSubmission{
			Schema: intentSchema,
			Payload: maestro.IntentPayload{
				Intent: intent,
				Artifact: maestro.Artifact{
					Type:    "doc",
					Content: "placeholder",
				},
			},
		},
	})
}

const intentSchema = "intent.submission.v0"

func publishRunState(bus *transport.Bus, runID string, state maestro.RunState, attempt int) {
	meta := map[string]string{
		"run_id": runID,
		"state":  string(state),
	}
	bus.Publish(transport.Event{
		Type:     runStateUpdateEventType,
		Payload:  maestro.RunStateUpdate{RunID: runID, State: state, Attempt: attempt},
		Metadata: meta,
	})
}

func publishPlan(bus *transport.Bus, runID string) {
	step := maestro.ExecutionPlanStep{
		StepID:           "step-1",
		Description:      "Explain action",
		ResponsibleAgent: "maestro",
		Inputs:           []string{"intent"},
		Outputs:          []string{"task.evaluate"},
	}
	meta := map[string]string{
		"run_id":  runID,
		"posture": "early_sdlc",
		"intent":  "document pleximesh",
	}
	bus.Publish(transport.Event{
		Type:     executionPlanEventType,
		Payload:  maestro.ExecutionPlanEvent{RunID: runID, Posture: "early_sdlc", Attempt: 1, Intent: "document pleximesh", Steps: []maestro.ExecutionPlanStep{step}},
		Metadata: meta,
	})
}

func publishGate(bus *transport.Bus, runID string) {
	meta := map[string]string{"run_id": runID}
	bus.Publish(transport.Event{
		Type:     executionGateEvaluatedType,
		Payload:  map[string]string{"status": "ok"},
		Metadata: meta,
	})
}

func publishRunCompleted(bus *transport.Bus, runID string) {
	meta := map[string]string{"run_id": runID}
	bus.Publish(transport.Event{
		Type:     runCompletedEventType,
		Payload:  map[string]string{"result": "success"},
		Metadata: meta,
	})
}

func waitForReport(t *testing.T, ch <-chan transport.Event) Report {
	t.Helper()
	timeout := time.After(time.Second)
	for {
		select {
		case evt := <-ch:
			if evt.Type != testingReportGeneratedType {
				continue
			}
			report, ok := evt.Payload.(Report)
			if !ok {
				t.Fatalf("unexpected payload type: %T", evt.Payload)
			}
			return report
		case <-timeout:
			t.Fatalf("timed out waiting for testing report")
		}
	}
}
