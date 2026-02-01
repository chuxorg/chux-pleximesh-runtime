package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type EventEnvelope struct {
	ID            string          `json:"id"`
	TS            time.Time       `json:"ts"`
	Source        string          `json:"source"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlationId,omitempty"`
}

type EmitRequest struct {
	Source        string          `json:"source"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlationId,omitempty"`
}

type EventBus struct {
	mu     sync.RWMutex
	subs   map[int]chan EventEnvelope
	nextID int
}

func NewEventBus() *EventBus {
	return &EventBus{subs: make(map[int]chan EventEnvelope)}
}

func (b *EventBus) Subscribe(buffer int) (int, <-chan EventEnvelope) {
	if buffer <= 0 {
		buffer = 1
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.nextID
	b.nextID++
	ch := make(chan EventEnvelope, buffer)
	b.subs[id] = ch
	return id, ch
}

func (b *EventBus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch, ok := b.subs[id]
	if !ok {
		return
	}
	delete(b.subs, id)
	close(ch)
}

func (b *EventBus) Publish(evt EventEnvelope) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- evt:
		default:
			// best-effort; drop when subscriber is slow
		}
	}
}

type RingBuffer struct {
	mu     sync.RWMutex
	size   int
	events []EventEnvelope
	next   int
	count  int
}

func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 1
	}
	return &RingBuffer{
		size:   size,
		events: make([]EventEnvelope, size),
	}
}

func (r *RingBuffer) Add(evt EventEnvelope) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[r.next] = evt
	r.next = (r.next + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *RingBuffer) Snapshot(limit int) []EventEnvelope {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.count == 0 {
		return nil
	}
	if limit <= 0 || limit > r.count {
		limit = r.count
	}
	oldest := 0
	if r.count == r.size {
		oldest = r.next
	}
	start := oldest
	if limit < r.count {
		start = (oldest + (r.count - limit)) % r.size
	}
	out := make([]EventEnvelope, 0, limit)
	for i := 0; i < limit; i++ {
		idx := (start + i) % r.size
		out = append(out, r.events[idx])
	}
	return out
}

var eventCounter uint64

func newEventID() string {
	count := atomic.AddUint64(&eventCounter, 1)
	return fmt.Sprintf("evt-%d-%d", time.Now().UnixNano(), count)
}

func main() {
	port := getenvInt("RUNTIME_PORT", 8787)
	busBuffer := getenvInt("BUS_SUBSCRIBER_BUFFER", 64)
	ringSize := getenvInt("EVENT_BUFFER_SIZE", 200)
	snapshotLimit := getenvInt("SNAPSHOT_LIMIT", ringSize)

	bus := NewEventBus()
	buffer := NewRingBuffer(ringSize)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
			"ts": time.Now().UTC(),
		})
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "streaming_unsupported")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		id, ch := bus.Subscribe(busBuffer)
		defer bus.Unsubscribe(id)

		flusher.Flush()
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(evt)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	mux.HandleFunc("/emit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		var req EmitRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		if req.Source == "" || req.Kind == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "missing_required_fields")
			return
		}
		if len(req.Payload) == 0 {
			req.Payload = json.RawMessage("null")
		}
		env := EventEnvelope{
			ID:            newEventID(),
			TS:            time.Now().UTC(),
			Source:        req.Source,
			Kind:          req.Kind,
			Name:          req.Name,
			Payload:       req.Payload,
			CorrelationID: req.CorrelationID,
		}
		buffer.Add(env)
		bus.Publish(env)
		log.Printf("event_published id=%s source=%s kind=%s name=%s correlation_id=%s", env.ID, env.Source, env.Kind, env.Name, env.CorrelationID)
		writeJSON(w, http.StatusAccepted, env)
	})

	mux.HandleFunc("/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		limit := snapshotLimit
		if raw := r.URL.Query().Get("n"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil {
				limit = parsed
			}
		}
		writeJSON(w, http.StatusOK, buffer.Snapshot(limit))
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("runtime_daemon_starting port=%d buffer_size=%d", port, ringSize)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("runtime_daemon_error err=%v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, reason string) {
	writeJSON(w, status, map[string]string{"error": reason})
}

func getenvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
