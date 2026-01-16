package transport

import (
	"testing"
	"time"
)

func TestBusPublishesToMultipleSubscribers(t *testing.T) {
	bus := NewBus()

	subOne := bus.Subscribe(4)
	defer bus.Unsubscribe(subOne)

	subTwo := bus.Subscribe(4)
	defer bus.Unsubscribe(subTwo)

	event := Event{
		Type:    "test.event",
		Payload: map[string]string{"message": "hello"},
	}

	bus.Publish(event)

	select {
	case received := <-subOne:
		if received.Type != event.Type {
			t.Fatalf("subscriber one received wrong type: %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber one did not receive event")
	}

	select {
	case received := <-subTwo:
		if received.Type != event.Type {
			t.Fatalf("subscriber two received wrong type: %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber two did not receive event")
	}
}

func TestBusDropWhenSubscribersSlow(t *testing.T) {
	bus := NewBus()
	sub := bus.Subscribe(1)
	defer bus.Unsubscribe(sub)

	// Publish multiple events without draining to saturate the buffer.
	bus.Publish(Event{Type: "first"})
	bus.Publish(Event{Type: "second"})
	bus.Publish(Event{Type: "third"})

	select {
	case evt := <-sub:
		if evt.Type != "first" {
			t.Fatalf("expected to receive first event, got %s", evt.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive buffered event")
	}

	// No further events should be queued because the subscriber was slow and drops apply.
	select {
	case evt := <-sub:
		t.Fatalf("expected no additional events, got %s", evt.Type)
	default:
	}
}
