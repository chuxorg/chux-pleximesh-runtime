package main

import (
	"context"
	"strconv"

	"github.com/chuxorg/chux-agent-mesh/runtime/maestro"
	"github.com/chuxorg/chux-agent-mesh/runtime/transport"
)

const (
	guidanceEventTypeIssued = "maestro.guidance.issued"
	executionPlanEventType  = "execution.plan.proposed"
	responseToMetadataKey   = "response_to"
)

type busGuidancePublisher struct {
	bus *transport.Bus
}

func (b *busGuidancePublisher) PublishGuidance(ctx context.Context, guidance maestro.MaestroGuidanceEvent) error {
	if b == nil || b.bus == nil {
		return nil
	}
	meta := map[string]string{
		"run_id":  guidance.RunID,
		"posture": guidance.Posture,
		"attempt": strconv.Itoa(guidance.Attempt),
	}
	b.bus.Publish(transport.Event{
		Type:     guidanceEventTypeIssued,
		Payload:  guidance,
		Metadata: meta,
	})
	return nil
}

type busExecutionPlanPublisher struct {
	bus *transport.Bus
}

func (b *busExecutionPlanPublisher) PublishExecutionPlan(ctx context.Context, plan maestro.ExecutionPlanEvent) error {
	if b == nil || b.bus == nil {
		return nil
	}
	meta := map[string]string{
		"run_id":  plan.RunID,
		"posture": plan.Posture,
		"attempt": strconv.Itoa(plan.Attempt),
		"intent":  plan.Intent,
	}
	b.bus.Publish(transport.Event{
		Type:     executionPlanEventType,
		Payload:  plan,
		Metadata: meta,
	})
	return nil
}

func normalizeIntent(submission *maestro.IntentSubmission) {
	if submission == nil {
		return
	}
	if submission.Metadata == nil {
		submission.Metadata = map[string]interface{}{}
	}
	if value, ok := submission.Metadata[responseToMetadataKey]; ok {
		if text, ok := value.(string); ok && text == guidanceEventTypeIssued {
			submission.Metadata[maestro.IntentContextMetadataKey] = maestro.IntentContextClarificationReply
			return
		}
	}
	submission.Metadata[maestro.IntentContextMetadataKey] = maestro.IntentContextDirective
}

func buildIntentMetadata(submission maestro.IntentSubmission) map[string]interface{} {
	meta := map[string]interface{}{}
	if submission.Metadata == nil {
		return meta
	}
	if value, ok := submission.Metadata[maestro.IntentContextMetadataKey]; ok {
		meta[maestro.IntentContextMetadataKey] = value
	}
	return meta
}
