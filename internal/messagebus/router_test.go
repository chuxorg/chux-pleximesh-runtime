package messagebus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

func TestRouterFanOutOrdered(t *testing.T) {
	router := NewRouter()

	taskSub, err := router.Subscribe(SubscriptionConfig{
		AgentID: "task-consumer",
		Domains: []event.Domain{event.DomainTask},
	})
	if err != nil {
		t.Fatalf("subscribe task consumer: %v", err)
	}
	intentSub, err := router.Subscribe(SubscriptionConfig{
		AgentID: "intent-consumer",
		Types:   []event.Type{event.TypeIntentCreated},
	})
	if err != nil {
		t.Fatalf("subscribe intent consumer: %v", err)
	}
	t.Cleanup(func() {
		taskSub.Close()
		intentSub.Close()
	})

	var taskEvents []string
	done := make(chan struct{})
	go func() {
		for env := range taskSub.Events() {
			taskEvents = append(taskEvents, env.EventID)
			if len(taskEvents) == 2 {
				close(done)
				return
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	env1 := newEnvelope("evt-1", event.TypeTaskCreated, event.DomainTask)
	env2 := newEnvelope("evt-2", event.TypeIntentCreated, event.DomainIntent)
	env3 := newEnvelope("evt-3", event.TypeTaskCompleted, event.DomainTask)

	if err := router.Dispatch(ctx, env1); err != nil {
		t.Fatalf("dispatch 1: %v", err)
	}
	if err := router.Dispatch(ctx, env2); err != nil {
		t.Fatalf("dispatch 2: %v", err)
	}
	if err := router.Dispatch(ctx, env3); err != nil {
		t.Fatalf("dispatch 3: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for task events")
	}

	if got, want := taskEvents, []string{"evt-1", "evt-3"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("task events out of order, got %v want %v", got, want)
	}
	select {
	case env := <-intentSub.Events():
		if env.EventID != "evt-2" {
			t.Fatalf("intent event mismatch, got %s", env.EventID)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for intent event")
	}
}

func TestRouterBackpressure(t *testing.T) {
	router := NewRouter()

	sub, err := router.Subscribe(SubscriptionConfig{
		AgentID:    "slow-consumer",
		BufferSize: 1,
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Close()

	env1 := newEnvelope("evt-1", event.TypeTaskCreated, event.DomainTask)
	env2 := newEnvelope("evt-2", event.TypeTaskCreated, event.DomainTask)

	if err := router.Dispatch(context.Background(), env1); err != nil {
		t.Fatalf("dispatch first: %v", err)
	}

	blocked := make(chan struct{})
	go func() {
		_ = router.Dispatch(context.Background(), env2)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatalf("dispatch should have blocked due to backpressure")
	case <-time.After(50 * time.Millisecond):
	}

	first := <-sub.Events()
	if first.EventID != "evt-1" {
		t.Fatalf("expected evt-1, got %s", first.EventID)
	}

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatalf("dispatch failed to resume once buffer freed")
	}

	second := <-sub.Events()
	if second.EventID != "evt-2" {
		t.Fatalf("expected evt-2, got %s", second.EventID)
	}
}

func TestRouterConcurrentSubscriptions(t *testing.T) {
	router := NewRouter()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i := 0; i < 10; i++ {
		id := i
		eventType := event.Type(fmt.Sprintf("Custom.%d", id))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				handle, err := router.Subscribe(SubscriptionConfig{
					AgentID: fmt.Sprintf("agent-%d", id),
					Types:   []event.Type{eventType},
				})
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				env := newEnvelope(fmt.Sprintf("evt-%d-%d", id, j), eventType, event.DomainDebug)
				if err := router.Dispatch(context.Background(), env); err != nil {
					select {
					case errCh <- err:
					default:
					}
					handle.Close()
					return
				}
				select {
				case <-handle.Events():
				case <-time.After(time.Second):
					select {
					case errCh <- fmt.Errorf("agent-%d timed out waiting for event", id):
					default:
					}
					handle.Close()
					return
				}
				handle.Close()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("router concurrency error: %v", err)
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for concurrent subscribers")
	}
}

func newEnvelope(id string, typ event.Type, domain event.Domain) event.Envelope {
	return event.Envelope{
		EventID: id,
		Type:    typ,
		Domain:  domain,
		SourceAgent: event.AgentDescriptor{
			AgentID: "source",
			Role:    "tester",
		},
		Timestamp: time.Now().UTC(),
		Payload:   []byte(`{"ok":true}`),
		Signature: "test",
		Runtime: event.RuntimeContext{
			RuntimeVersion: "test",
			InitKitVersion: "test",
		},
	}
}
