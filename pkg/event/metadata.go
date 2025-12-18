package event

import (
	"reflect"
	"sort"
)

// RoutingMetadata describes immutable routing metadata for a canonical event type.
type RoutingMetadata struct {
	Type        Type
	Domain      Domain
	Description string
	PayloadType reflect.Type
}

// NewPayload returns a pointer to the concrete payload type defined for the metadata.
func (m RoutingMetadata) NewPayload() any {
	if m.PayloadType == nil {
		return nil
	}
	return reflect.New(m.PayloadType).Interface()
}

var routingTable = map[Type]RoutingMetadata{}

// MetadataForType returns the routing metadata for the provided event type.
func MetadataForType(t Type) (RoutingMetadata, bool) {
	meta, ok := routingTable[t]
	return meta, ok
}

// KnownTypes returns a deterministic slice of registered event types.
func KnownTypes() []Type {
	types := make([]Type, 0, len(routingTable))
	for t := range routingTable {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })
	return types
}

func registerMetadata(meta RoutingMetadata) {
	if meta.Type == "" {
		panic("event: metadata type is required")
	}
	if meta.Domain == "" {
		panic("event: metadata domain is required")
	}
	if meta.PayloadType == nil {
		panic("event: payload type is required")
	}
	routingTable[meta.Type] = meta
}

func payloadType[T any]() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

// Event type constants.
const (
	TypeLifecycleStateReported       Type = "Lifecycle.StateReported"
	TypeLifecycleTransitionRequested Type = "Lifecycle.TransitionRequested"
	TypeLifecycleTransitionApproved  Type = "Lifecycle.TransitionApproved"
	TypeLifecycleTransitionRejected  Type = "Lifecycle.TransitionRejected"

	TypeIntentCreated   Type = "Intent.Created"
	TypeIntentClarified Type = "Intent.Clarified"

	TypeTaskCreated   Type = "Task.Created"
	TypeTaskAssigned  Type = "Task.Assigned"
	TypeTaskBlocked   Type = "Task.Blocked"
	TypeTaskCompleted Type = "Task.Completed"
	TypeTaskFailed    Type = "Task.Failed"

	TypeVerificationRequested Type = "Verification.Requested"
	TypeVerificationPassed    Type = "Verification.Passed"
	TypeVerificationFailed    Type = "Verification.Failed"

	TypePromptCreated       Type = "Prompt.Created"
	TypePromptSent          Type = "Prompt.Sent"
	TypeLLMResponseReceived Type = "LLM.ResponseReceived"
	TypePromptEvaluated     Type = "Prompt.Evaluated"

	TypeTelemetryMetricEmitted    Type = "Telemetry.MetricEmitted"
	TypeTelemetryDurationMeasured Type = "Telemetry.DurationMeasured"

	TypeStatusReported Type = "Status.Reported"
	TypeHealthReported Type = "Health.Reported"

	TypeViolationLawBreach          Type = "Violation.LawBreach"
	TypeViolationCapabilityExceeded Type = "Violation.CapabilityExceeded"

	TypeDebugTrace Type = "Debug.Trace"
)

func init() {
	registerMetadata(RoutingMetadata{
		Type:        TypeLifecycleStateReported,
		Domain:      DomainLifecycle,
		Description: "Agent lifecycle state change",
		PayloadType: payloadType[LifecycleStateReportedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeLifecycleTransitionRequested,
		Domain:      DomainLifecycle,
		Description: "Lifecycle transition request",
		PayloadType: payloadType[LifecycleTransitionRequestedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeLifecycleTransitionApproved,
		Domain:      DomainLifecycle,
		Description: "Lifecycle transition approved",
		PayloadType: payloadType[LifecycleTransitionApprovedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeLifecycleTransitionRejected,
		Domain:      DomainLifecycle,
		Description: "Lifecycle transition rejected",
		PayloadType: payloadType[LifecycleTransitionRejectedPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeIntentCreated,
		Domain:      DomainIntent,
		Description: "Intent created",
		PayloadType: payloadType[IntentCreatedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeIntentClarified,
		Domain:      DomainIntent,
		Description: "Intent clarified",
		PayloadType: payloadType[IntentClarifiedPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeTaskCreated,
		Domain:      DomainTask,
		Description: "Task created",
		PayloadType: payloadType[TaskCreatedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeTaskAssigned,
		Domain:      DomainTask,
		Description: "Task assigned",
		PayloadType: payloadType[TaskAssignedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeTaskBlocked,
		Domain:      DomainTask,
		Description: "Task blocked",
		PayloadType: payloadType[TaskStatusPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeTaskCompleted,
		Domain:      DomainTask,
		Description: "Task completed",
		PayloadType: payloadType[TaskStatusPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeTaskFailed,
		Domain:      DomainTask,
		Description: "Task failed",
		PayloadType: payloadType[TaskStatusPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeVerificationRequested,
		Domain:      DomainVerification,
		Description: "Verification requested",
		PayloadType: payloadType[VerificationRequestedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeVerificationPassed,
		Domain:      DomainVerification,
		Description: "Verification passed",
		PayloadType: payloadType[VerificationResultPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeVerificationFailed,
		Domain:      DomainVerification,
		Description: "Verification failed",
		PayloadType: payloadType[VerificationResultPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypePromptCreated,
		Domain:      DomainCognitive,
		Description: "Prompt created",
		PayloadType: payloadType[PromptCreatedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypePromptSent,
		Domain:      DomainCognitive,
		Description: "Prompt sent",
		PayloadType: payloadType[PromptSentPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeLLMResponseReceived,
		Domain:      DomainCognitive,
		Description: "LLM response received",
		PayloadType: payloadType[LLMResponseReceivedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypePromptEvaluated,
		Domain:      DomainCognitive,
		Description: "Prompt evaluated",
		PayloadType: payloadType[PromptEvaluatedPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeTelemetryMetricEmitted,
		Domain:      DomainTelemetry,
		Description: "Telemetry metric emitted",
		PayloadType: payloadType[TelemetryMetricEmittedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeTelemetryDurationMeasured,
		Domain:      DomainTelemetry,
		Description: "Telemetry duration measured",
		PayloadType: payloadType[TelemetryDurationMeasuredPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeStatusReported,
		Domain:      DomainStatus,
		Description: "Status reported",
		PayloadType: payloadType[StatusReportedPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeHealthReported,
		Domain:      DomainHealth,
		Description: "Health reported",
		PayloadType: payloadType[HealthReportedPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeViolationLawBreach,
		Domain:      DomainViolation,
		Description: "Violation law breach",
		PayloadType: payloadType[ViolationLawBreachPayload](),
	})
	registerMetadata(RoutingMetadata{
		Type:        TypeViolationCapabilityExceeded,
		Domain:      DomainViolation,
		Description: "Violation capability exceeded",
		PayloadType: payloadType[ViolationCapabilityExceededPayload](),
	})

	registerMetadata(RoutingMetadata{
		Type:        TypeDebugTrace,
		Domain:      DomainDebug,
		Description: "Debug trace",
		PayloadType: payloadType[DebugTracePayload](),
	})
}
