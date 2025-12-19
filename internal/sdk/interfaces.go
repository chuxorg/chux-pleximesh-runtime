package sdk

import "context"

// EventPublisher is the minimal contract an agent can use to emit events.
type EventPublisher interface {
	Publish(ctx context.Context, evt Event) error
}

// EventHandler is invoked for inbound events delivered via a Subscriber.
type EventHandler func(ctx context.Context, evt Event) error

// EventSubscriber is a minimal interface for registering inbound event handlers.
type EventSubscriber interface {
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
}

// AckNackHandler allows an agent to ack or nack inbound events deterministically.
type AckNackHandler interface {
	Ack(ctx context.Context, eventID string) error
	Nack(ctx context.Context, eventID string, reason error) error
}
