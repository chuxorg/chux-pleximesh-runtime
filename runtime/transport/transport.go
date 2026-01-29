package transport

import (
	"fmt"
	"sync"

	"github.com/chuxorg/chux-agent-mesh/runtime/qa"
)

// Event represents a single runtime observation (message or run state update).
type Event struct {
	Type     string            `json:"type"`
	Payload  interface{}       `json:"payload"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Bus fan-outs runtime events to any number of subscribers for observability.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[int]chan Event
	nextID      int
	observer    qa.Observer
}

// NewBus constructs an in-memory event bus.
func NewBus() *Bus {
	return NewBusWithObserver(nil)
}

// NewBusWithObserver constructs a bus with an optional QA observer.
func NewBusWithObserver(observer qa.Observer) *Bus {
	return &Bus{
		subscribers: make(map[int]chan Event),
		observer:    qa.ResolveObserver(observer),
	}
}

// Subscribe registers a new subscriber with the given buffer size.
func (b *Bus) Subscribe(buffer int) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if buffer <= 0 {
		buffer = 1
	}

	ch := make(chan Event, buffer)
	id := b.nextID
	b.nextID++
	b.subscribers[id] = ch

	b.observe("bus.subscribe", "channel", fmt.Sprintf("subscriber_id=%d buffer=%d", id, buffer))
	return ch
}

// Unsubscribe removes a subscriber channel.
func (b *Bus) Unsubscribe(ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, sub := range b.subscribers {
		if sub == ch {
			delete(b.subscribers, id)
			close(sub)
			b.observe("bus.unsubscribe", "channel", fmt.Sprintf("subscriber_id=%d", id))
			return
		}
	}
}

// Publish delivers the event to all subscribers without delivery guarantees.
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.observe("bus.publish", "channel", fmt.Sprintf("type=%s subscribers=%d", event.Type, len(b.subscribers)))
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// drop when subscriber is slow; observability is best-effort
		}
	}
}

func (b *Bus) observe(name, sideEffect, detail string) {
	if b == nil || b.observer == nil {
		return
	}
	b.observer.Record(qa.BoundaryEvent{
		Name:       name,
		SideEffect: sideEffect,
		Detail:     detail,
	})
}
