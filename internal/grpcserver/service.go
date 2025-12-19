package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/ems"
	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
	"github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type busService struct {
	meshpb.UnimplementedEventBusServer
	router   *messagebus.Router
	verifier *ems.Verifier
}

func (s *busService) Publish(stream meshpb.EventBus_PublishServer) error {
	ctx := stream.Context()
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		pbEnv := req.GetEnvelope()
		eventID := ""
		if pbEnv != nil {
			eventID = pbEnv.GetEventId()
		}
		if err != nil {
			return err
		}
		env, err := protoToEnvelope(pbEnv)
		if err != nil {
			if sendErr := stream.Send(reject(eventID, fmt.Sprintf("invalid envelope: %v", err))); sendErr != nil {
				return sendErr
			}
			continue
		}
		if err := env.Validate(); err != nil {
			if sendErr := stream.Send(reject(env.EventID, err.Error())); sendErr != nil {
				return sendErr
			}
			continue
		}
		if err := s.verifier.Verify(ctx, env); err != nil {
			if sendErr := stream.Send(reject(env.EventID, err.Error())); sendErr != nil {
				return sendErr
			}
			continue
		}
		if err := s.router.Dispatch(ctx, env); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			if sendErr := stream.Send(reject(env.EventID, err.Error())); sendErr != nil {
				return sendErr
			}
			continue
		}
		if err := stream.Send(ack(env.EventID)); err != nil {
			return err
		}
	}
}

func (s *busService) Subscribe(req *meshpb.SubscriptionRequest, stream meshpb.EventBus_SubscribeServer) error {
	if req.GetAgentId() == "" {
		return status.Error(codes.InvalidArgument, "agent_id is required")
	}
	config := messagebus.SubscriptionConfig{
		AgentID: req.GetAgentId(),
	}
	if len(req.GetDomains()) > 0 {
		config.Domains = make([]event.Domain, len(req.GetDomains()))
		for i, domain := range req.GetDomains() {
			config.Domains[i] = event.Domain(domain)
		}
	}
	if len(req.GetEventTypes()) > 0 {
		config.Types = make([]event.Type, len(req.GetEventTypes()))
		for i, typ := range req.GetEventTypes() {
			config.Types[i] = event.Type(typ)
		}
	}
	handle, err := s.router.Subscribe(config)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	defer handle.Close()

	for {
		select {
		case env, ok := <-handle.Events():
			if !ok {
				return nil
			}
			if err := stream.Send(envelopeToProto(env)); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func protoToEnvelope(pb *meshpb.EventEnvelope) (event.Envelope, error) {
	if pb == nil {
		return event.Envelope{}, errors.New("missing envelope")
	}
	ts := pb.GetTimestamp()
	if ts == "" {
		return event.Envelope{}, errors.New("timestamp is required")
	}
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return event.Envelope{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	env := event.Envelope{
		EventID:       pb.GetEventId(),
		Type:          event.Type(pb.GetEventType()),
		Domain:        event.Domain(pb.GetEventDomain()),
		Timestamp:     parsed,
		CorrelationID: pb.GetCorrelationId(),
		Payload:       append([]byte(nil), pb.GetPayload()...),
		Signature:     pb.GetSignature(),
	}
	if src := pb.GetSourceAgent(); src != nil {
		env.SourceAgent = event.AgentDescriptor{
			AgentID: src.GetAgentId(),
			Role:    src.GetRole(),
		}
	}
	if tgt := pb.GetTarget(); tgt != nil && (tgt.GetAgentId() != "" || tgt.GetScope() != "") {
		env.Target = &event.TargetDescriptor{
			AgentID: tgt.GetAgentId(),
			Scope:   tgt.GetScope(),
		}
	}
	if rt := pb.GetRuntime(); rt != nil {
		env.Runtime = event.RuntimeContext{
			RuntimeVersion: rt.GetRuntimeVersion(),
			InitKitVersion: rt.GetInitkitVersion(),
		}
	}
	return env, nil
}

func envelopeToProto(env event.Envelope) *meshpb.EventEnvelope {
	pb := &meshpb.EventEnvelope{
		EventId:       env.EventID,
		EventType:     string(env.Type),
		EventDomain:   string(env.Domain),
		Timestamp:     env.Timestamp.UTC().Format(time.RFC3339Nano),
		CorrelationId: env.CorrelationID,
		Payload:       append([]byte(nil), env.Payload...),
		Signature:     env.Signature,
		Runtime: &meshpb.RuntimeContext{
			RuntimeVersion: env.Runtime.RuntimeVersion,
			InitkitVersion: env.Runtime.InitKitVersion,
		},
		SourceAgent: &meshpb.AgentDescriptor{
			AgentId: env.SourceAgent.AgentID,
			Role:    env.SourceAgent.Role,
		},
	}
	if env.Target != nil {
		pb.Target = &meshpb.TargetDescriptor{
			AgentId: env.Target.AgentID,
			Scope:   env.Target.Scope,
		}
	}
	return pb
}

func ack(eventID string) *meshpb.BusIngressResponse {
	return &meshpb.BusIngressResponse{
		EventId: eventID,
		Status:  meshpb.AckStatus_ACK_STATUS_ACCEPTED,
		Message: "accepted",
	}
}

func reject(eventID, message string) *meshpb.BusIngressResponse {
	return &meshpb.BusIngressResponse{
		EventId: eventID,
		Status:  meshpb.AckStatus_ACK_STATUS_REJECTED,
		Message: message,
	}
}
