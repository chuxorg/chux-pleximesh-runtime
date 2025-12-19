package messagebus

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

const defaultBufferSize = 16

var (
	// ErrAgentIDRequired is returned when a subscription attempts to register without identifying an agent.
	ErrAgentIDRequired = errors.New("messagebus: agent_id required for subscription")

	errSubscriptionClosed = errors.New("messagebus: subscription closed")
)

// SubscriptionConfig describes a subscriber's interest in specific domains and event types.
type SubscriptionConfig struct {
	AgentID    string
	Domains    []event.Domain
	Types      []event.Type
	BufferSize int
}

// Router coordinates subscriptions and dispatches envelopes to interested subscribers.
type Router struct {
	mu    sync.RWMutex
	subs  []*subscriber
	index map[string]int
	next  uint64
}

// SubscriptionHandle exposes the event stream for a single subscriber.
type SubscriptionHandle struct {
	router *Router
	sub    *subscriber
	once   sync.Once
}

type subscriber struct {
	id      string
	agentID string
	domains map[event.Domain]struct{}
	types   map[event.Type]struct{}
	events  chan event.Envelope
	done    chan struct{}
	wg      sync.WaitGroup
	closed  uint32
}

// NewRouter constructs an empty router ready to accept subscriptions.
func NewRouter() *Router {
	return &Router{
		index: make(map[string]int),
	}
}

// Subscribe registers a subscriber using the provided filter configuration.
func (r *Router) Subscribe(cfg SubscriptionConfig) (*SubscriptionHandle, error) {
	if cfg.AgentID == "" {
		return nil, ErrAgentIDRequired
	}
	buffer := cfg.BufferSize
	if buffer <= 0 {
		buffer = defaultBufferSize
	}
	sub := &subscriber{
		agentID: cfg.AgentID,
		domains: make(map[event.Domain]struct{}, len(cfg.Domains)),
		types:   make(map[event.Type]struct{}, len(cfg.Types)),
		events:  make(chan event.Envelope, buffer),
		done:    make(chan struct{}),
	}
	for _, d := range cfg.Domains {
		sub.domains[d] = struct{}{}
	}
	for _, t := range cfg.Types {
		sub.types[t] = struct{}{}
	}

	id := atomic.AddUint64(&r.next, 1)
	sub.id = fmt.Sprintf("%s-%d", cfg.AgentID, id)

	r.mu.Lock()
	r.subs = append(r.subs, sub)
	r.index[sub.id] = len(r.subs) - 1
	r.mu.Unlock()

	return &SubscriptionHandle{
		router: r,
		sub:    sub,
	}, nil
}

// Events exposes the subscription's event stream.
func (h *SubscriptionHandle) Events() <-chan event.Envelope {
	return h.sub.events
}

// Close unregisters the subscription and releases resources.
func (h *SubscriptionHandle) Close() {
	h.once.Do(func() {
		removed := h.router.remove(h.sub.id)
		if removed != nil {
			atomic.StoreUint32(&removed.closed, 1)
			close(removed.done)
			removed.wg.Wait()
			close(removed.events)
		}
	})
}

// Dispatch delivers the envelope to all matching subscribers, respecting ordered fan-out.
func (r *Router) Dispatch(ctx context.Context, envelope event.Envelope) error {
	subs := r.snapshot()
	for _, sub := range subs {
		if sub == nil {
			continue
		}
		if !sub.matches(envelope) {
			continue
		}
		if err := sub.deliver(ctx, envelope); err != nil {
			if errors.Is(err, errSubscriptionClosed) {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *Router) snapshot() []*subscriber {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.subs) == 0 {
		return nil
	}
	cp := make([]*subscriber, len(r.subs))
	copy(cp, r.subs)
	return cp
}

func (r *Router) remove(id string) *subscriber {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx, ok := r.index[id]
	if !ok {
		return nil
	}
	removed := r.subs[idx]
	copy(r.subs[idx:], r.subs[idx+1:])
	r.subs = r.subs[:len(r.subs)-1]
	delete(r.index, id)
	for i := idx; i < len(r.subs); i++ {
		r.index[r.subs[i].id] = i
	}
	return removed
}

func (s *subscriber) matches(envelope event.Envelope) bool {
	if len(s.types) > 0 {
		if _, ok := s.types[envelope.Type]; !ok {
			return false
		}
	}
	if len(s.domains) > 0 {
		if _, ok := s.domains[envelope.Domain]; !ok {
			return false
		}
	}
	return true
}

func (s *subscriber) deliver(ctx context.Context, envelope event.Envelope) error {
	if atomic.LoadUint32(&s.closed) == 1 {
		return errSubscriptionClosed
	}
	s.wg.Add(1)
	defer s.wg.Done()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return errSubscriptionClosed
	case s.events <- envelope:
		return nil
	}
}
