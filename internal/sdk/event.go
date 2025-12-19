package sdk

import "time"

// Event is the canonical representation exchanged between agents and the harness.
type Event struct {
	ID        string
	Type      string
	Payload   []byte
	Metadata  map[string]string
	Timestamp time.Time
}

// Clone returns a deep copy of the event to prevent callers from mutating shared state.
func (e Event) Clone() Event {
	clone := Event{
		ID:        e.ID,
		Type:      e.Type,
		Timestamp: e.Timestamp,
	}
	if len(e.Payload) > 0 {
		clone.Payload = append([]byte(nil), e.Payload...)
	}
	if len(e.Metadata) > 0 {
		clone.Metadata = make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}
