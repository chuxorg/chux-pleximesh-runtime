package event

import "testing"

var canonicalTypes = []Type{
	TypeLifecycleStateReported,
	TypeLifecycleTransitionRequested,
	TypeLifecycleTransitionApproved,
	TypeLifecycleTransitionRejected,
	TypeIntentCreated,
	TypeIntentClarified,
	TypeTaskCreated,
	TypeTaskAssigned,
	TypeTaskBlocked,
	TypeTaskCompleted,
	TypeTaskFailed,
	TypeVerificationRequested,
	TypeVerificationPassed,
	TypeVerificationFailed,
	TypePromptCreated,
	TypePromptSent,
	TypeLLMResponseReceived,
	TypePromptEvaluated,
	TypeTelemetryMetricEmitted,
	TypeTelemetryDurationMeasured,
	TypeStatusReported,
	TypeHealthReported,
	TypeViolationLawBreach,
	TypeViolationCapabilityExceeded,
	TypeDebugTrace,
}

func TestRoutingMetadataCoverage(t *testing.T) {
	if got, want := len(routingTable), len(canonicalTypes); got != want {
		t.Fatalf("routing table length mismatch: got %d want %d", got, want)
	}
	for _, typ := range canonicalTypes {
		meta, ok := MetadataForType(typ)
		if !ok {
			t.Fatalf("missing metadata for %s", typ)
		}
		if meta.Domain == "" {
			t.Fatalf("metadata domain missing for %s", typ)
		}
		if meta.PayloadType == nil {
			t.Fatalf("payload type missing for %s", typ)
		}
	}
}

func TestRoutingMetadataNewPayload(t *testing.T) {
	meta, ok := MetadataForType(TypeTaskCreated)
	if !ok {
		t.Fatalf("metadata missing for %s", TypeTaskCreated)
	}
	payload, ok := meta.NewPayload().(*TaskCreatedPayload)
	if !ok || payload == nil {
		t.Fatalf("expected *TaskCreatedPayload, got %#v", payload)
	}
}
