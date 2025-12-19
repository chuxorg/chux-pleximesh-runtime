package ems

import (
	"testing"
	"time"
)

func TestKeyManagerRotationAndRevocation(t *testing.T) {
	base := time.Unix(1_700_000_000, 0).UTC()
	ticks := []time.Time{
		base,
		base.Add(30 * time.Second),
		base.Add(90 * time.Second),
	}
	var idx int
	km := NewKeyManager(WithTimeSource(func() time.Time {
		defer func() { idx++ }()
		return ticks[idx%len(ticks)]
	}))

	registerMaterial := bytesOf(0x01)
	version, err := km.RegisterKey("agent-1", registerMaterial)
	if err != nil {
		t.Fatalf("RegisterKey() error = %v", err)
	}
	if version.Generation != 1 {
		t.Fatalf("unexpected generation: got %d want 1", version.Generation)
	}
	if !version.RotatedAt.Equal(base) {
		t.Fatalf("unexpected rotatedAt timestamp: got %v want %v", version.RotatedAt, base)
	}
	active, err := km.ActiveKey("agent-1")
	if err != nil {
		t.Fatalf("ActiveKey() error = %v", err)
	}
	if active.Generation != 1 {
		t.Fatalf("active generation mismatch: got %d want 1", active.Generation)
	}

	rotatedMaterial := bytesOf(0x02)
	rotated, err := km.RotateKey("agent-1", rotatedMaterial)
	if err != nil {
		t.Fatalf("RotateKey() error = %v", err)
	}
	if rotated.Generation != 2 {
		t.Fatalf("rotation generation mismatch: got %d want 2", rotated.Generation)
	}
	if !rotated.RotatedAt.Equal(ticks[1]) {
		t.Fatalf("rotation timestamp mismatch: got %v want %v", rotated.RotatedAt, ticks[1])
	}
	history, err := km.KeyHistory("agent-1")
	if err != nil {
		t.Fatalf("KeyHistory() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].Generation != 1 {
		t.Fatalf("history generation mismatch: got %d want 1", history[0].Generation)
	}

	if err := km.RevokeKey("agent-1"); err != nil {
		t.Fatalf("RevokeKey() error = %v", err)
	}
	if _, err := km.ActiveKey("agent-1"); err != ErrKeyRevoked {
		t.Fatalf("expected ErrKeyRevoked, got %v", err)
	}
	if _, err := km.RotateKey("agent-1", registerMaterial); err != ErrKeyRevoked {
		t.Fatalf("expected ErrKeyRevoked on rotate, got %v", err)
	}
}

func TestKeyManagerRejectsInvalidMaterial(t *testing.T) {
	km := NewKeyManager()
	if _, err := km.RegisterKey("agent-1", []byte{0x00}); err != ErrInvalidMaterial {
		t.Fatalf("expected ErrInvalidMaterial, got %v", err)
	}
	if err := km.RevokeKey("ghost"); err != ErrUnknownAgent {
		t.Fatalf("expected ErrUnknownAgent revoke, got %v", err)
	}
}

func bytesOf(b byte) []byte {
	material := make([]byte, keyMaterialSize)
	for i := range material {
		material[i] = b
	}
	return material
}
