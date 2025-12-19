package ems

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chuxorg/chux-agent-mesh/pkg/event"
)

const (
	signaturePrefix  = "ems-hmac-v1:"
	defaultReplayTTL = 2 * time.Minute
	contextSeparator = "|"
)

var (
	// ErrMissingSignature indicates that the envelope lacked an EMS signature.
	ErrMissingSignature = errors.New("ems: missing signature")
	// ErrInvalidSignature indicates that signature verification failed.
	ErrInvalidSignature = errors.New("ems: invalid signature")
	// ErrReplayDetected signals that the event_id has already been processed.
	ErrReplayDetected = errors.New("ems: replay detected")
)

// Verifier validates EMS signatures on ingress.
type Verifier struct {
	keys      *KeyManager
	scheduler *SliceScheduler
	now       func() time.Time
	replay    *replayTracker
}

// VerifierOption configures verifier behavior.
type VerifierOption func(*Verifier)

// WithVerifierClock overrides the time source used for slice calculation.
func WithVerifierClock(now func() time.Time) VerifierOption {
	return func(v *Verifier) {
		if now != nil {
			v.now = now
		}
	}
}

// WithReplayWindow sets the replay detection window duration.
func WithReplayWindow(d time.Duration) VerifierOption {
	return func(v *Verifier) {
		if d > 0 {
			v.replay = newReplayTracker(d)
		}
	}
}

// NewVerifier constructs an EMS verifier backed by the provided key manager and scheduler.
func NewVerifier(keys *KeyManager, scheduler *SliceScheduler, opts ...VerifierOption) *Verifier {
	if keys == nil {
		panic("ems: key manager is required")
	}
	if scheduler == nil {
		scheduler = NewSliceScheduler()
	}
	v := &Verifier{
		keys:      keys,
		scheduler: scheduler,
		now:       time.Now,
		replay:    newReplayTracker(defaultReplayTTL),
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Verify checks the EMS signature and replay status for the provided envelope.
func (v *Verifier) Verify(ctx context.Context, envelope event.Envelope) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if envelope.EventID == "" {
		return errors.New("ems: event_id required")
	}
	if envelope.SourceAgent.AgentID == "" {
		return errors.New("ems: source agent id required")
	}
	if envelope.Signature == "" {
		return ErrMissingSignature
	}
	signatureBytes, err := parseSignature(envelope.Signature)
	if err != nil {
		return err
	}
	key, err := v.keys.ActiveKey(envelope.SourceAgent.AgentID)
	if err != nil {
		return fmt.Errorf("ems: active key: %w", err)
	}
	canonical, err := canonicalizeEnvelope(envelope)
	if err != nil {
		return fmt.Errorf("ems: canonicalize envelope: %w", err)
	}
	now := v.now()
	slices := v.scheduler.AcceptableSlices(now)
	if len(slices) == 0 {
		slices = []SliceIndex{v.scheduler.SliceAt(now)}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if !v.validateAcrossSlices(ctx, canonical, envelope, key.Material, slices, signatureBytes) {
		return ErrInvalidSignature
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !v.replay.mark(envelope.EventID, now) {
		return ErrReplayDetected
	}
	return nil
}

func (v *Verifier) validateAcrossSlices(ctx context.Context, canonical []byte, envelope event.Envelope, material [keyMaterialSize]byte, slices []SliceIndex, signature []byte) bool {
	for _, slice := range slices {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		if verifySliceSignature(canonical, envelope, material, slice, signature) {
			return true
		}
	}
	return false
}

func verifySliceSignature(canonical []byte, envelope event.Envelope, material [keyMaterialSize]byte, slice SliceIndex, signature []byte) bool {
	ctxValue := deriveContext(envelope, slice)
	ephemeral := hmac.New(sha256.New, material[:])
	ephemeral.Write(ctxValue)
	ephemeralKey := ephemeral.Sum(nil)

	mac := hmac.New(sha256.New, ephemeralKey)
	mac.Write(canonical)
	expected := mac.Sum(nil)
	return hmac.Equal(expected, signature)
}

func parseSignature(signature string) ([]byte, error) {
	if !strings.HasPrefix(signature, signaturePrefix) {
		return nil, ErrInvalidSignature
	}
	raw := signature[len(signaturePrefix):]
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, ErrInvalidSignature
	}
	return decoded, nil
}

func canonicalizeEnvelope(envelope event.Envelope) ([]byte, error) {
	canonical := envelope
	canonical.Signature = ""
	canonical.Timestamp = canonical.Timestamp.UTC()
	if len(envelope.Payload) > 0 {
		canonical.Payload = append([]byte(nil), envelope.Payload...)
	}
	return json.Marshal(canonical)
}

func deriveContext(envelope event.Envelope, slice SliceIndex) []byte {
	builder := strings.Builder{}
	builder.WriteString(envelope.SourceAgent.AgentID)
	builder.WriteString(contextSeparator)
	builder.WriteString(string(envelope.Domain))
	builder.WriteString(contextSeparator)
	builder.WriteString(string(envelope.Type))
	builder.WriteString(contextSeparator)
	builder.WriteString(envelope.CorrelationID)
	builder.WriteString(contextSeparator)
	builder.WriteString(strconv.FormatInt(int64(slice), 10))
	return []byte(builder.String())
}

type replayTracker struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
}

func newReplayTracker(ttl time.Duration) *replayTracker {
	if ttl <= 0 {
		ttl = defaultReplayTTL
	}
	return &replayTracker{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
}

func (r *replayTracker) mark(id string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gc(now)
	if _, ok := r.seen[id]; ok {
		return false
	}
	r.seen[id] = now
	return true
}

func (r *replayTracker) gc(now time.Time) {
	cutoff := now.Add(-r.ttl)
	for id, ts := range r.seen {
		if ts.Before(cutoff) {
			delete(r.seen, id)
		}
	}
}
