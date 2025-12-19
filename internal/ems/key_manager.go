package ems

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const keyMaterialSize = 32

var (
	// ErrKeyExists signals that a key already exists for the requested agent.
	ErrKeyExists = errors.New("ems: key already registered")
	// ErrUnknownAgent indicates that the agent has not been provisioned yet.
	ErrUnknownAgent = errors.New("ems: unknown agent")
	// ErrKeyRevoked is raised when a revoked key is requested.
	ErrKeyRevoked = errors.New("ems: key revoked")
	// ErrInvalidMaterial indicates that the root key did not meet size requirements.
	ErrInvalidMaterial = fmt.Errorf("ems: root key must be %d bytes", keyMaterialSize)
)

// RootKeyVersion contains metadata about a root key generation.
type RootKeyVersion struct {
	AgentID    string
	Generation uint32
	Material   [keyMaterialSize]byte
	RotatedAt  time.Time
	Revoked    bool
}

type rotationHistory struct {
	version RootKeyVersion
}

type keyRecord struct {
	current   RootKeyVersion
	history   []rotationHistory
	revokedAt *time.Time
}

// KeyManager stores root keys for agents and handles rotation history.
type KeyManager struct {
	mu      sync.RWMutex
	nowFunc func() time.Time
	records map[string]*keyRecord
}

// KeyManagerOption configures a KeyManager instance.
type KeyManagerOption func(*KeyManager)

// WithTimeSource overrides the clock used for rotation timestamps.
func WithTimeSource(now func() time.Time) KeyManagerOption {
	return func(km *KeyManager) {
		km.nowFunc = now
	}
}

// NewKeyManager constructs a KeyManager with sane defaults.
func NewKeyManager(opts ...KeyManagerOption) *KeyManager {
	km := &KeyManager{
		nowFunc: time.Now,
		records: make(map[string]*keyRecord),
	}
	for _, opt := range opts {
		opt(km)
	}
	return km
}

// RegisterKey provisions the first root key for an agent.
func (km *KeyManager) RegisterKey(agentID string, material []byte) (RootKeyVersion, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if _, ok := km.records[agentID]; ok {
		return RootKeyVersion{}, ErrKeyExists
	}
	keyMaterial, err := normalizeMaterial(material)
	if err != nil {
		return RootKeyVersion{}, err
	}
	version := RootKeyVersion{
		AgentID:    agentID,
		Generation: 1,
		Material:   keyMaterial,
		RotatedAt:  km.nowFunc(),
	}
	km.records[agentID] = &keyRecord{current: version}
	return version, nil
}

// RotateKey replaces the current key for the agent and records the previous generation.
func (km *KeyManager) RotateKey(agentID string, material []byte) (RootKeyVersion, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	rec, ok := km.records[agentID]
	if !ok {
		return RootKeyVersion{}, ErrUnknownAgent
	}
	if rec.current.Revoked {
		return RootKeyVersion{}, ErrKeyRevoked
	}
	keyMaterial, err := normalizeMaterial(material)
	if err != nil {
		return RootKeyVersion{}, err
	}
	rec.history = append(rec.history, rotationHistory{version: rec.current})
	rec.current = RootKeyVersion{
		AgentID:    agentID,
		Generation: rec.history[len(rec.history)-1].version.Generation + 1,
		Material:   keyMaterial,
		RotatedAt:  km.nowFunc(),
	}
	return rec.current, nil
}

// ActiveKey returns the active, non-revoked key for the agent.
func (km *KeyManager) ActiveKey(agentID string) (RootKeyVersion, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	rec, ok := km.records[agentID]
	if !ok {
		return RootKeyVersion{}, ErrUnknownAgent
	}
	if rec.current.Revoked {
		return RootKeyVersion{}, ErrKeyRevoked
	}
	return rec.current, nil
}

// KeyHistory returns a copy of the rotation history for the agent.
func (km *KeyManager) KeyHistory(agentID string) ([]RootKeyVersion, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	rec, ok := km.records[agentID]
	if !ok {
		return nil, ErrUnknownAgent
	}
	history := make([]RootKeyVersion, len(rec.history))
	for i, entry := range rec.history {
		history[i] = entry.version
	}
	return history, nil
}

// RevokeKey marks the active key as revoked and prevents further use.
func (km *KeyManager) RevokeKey(agentID string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	rec, ok := km.records[agentID]
	if !ok {
		return ErrUnknownAgent
	}
	if rec.current.Revoked {
		return nil
	}
	rec.current.Revoked = true
	ts := km.nowFunc()
	rec.revokedAt = &ts
	return nil
}

func normalizeMaterial(material []byte) ([keyMaterialSize]byte, error) {
	var key [keyMaterialSize]byte
	if len(material) != keyMaterialSize {
		return key, ErrInvalidMaterial
	}
	copy(key[:], material)
	return key, nil
}
