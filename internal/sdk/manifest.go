package sdk

import (
	"encoding/json"
	"fmt"
	"io"
)

// Capability describes a discrete permission or feature exposed by an agent.
type Capability struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// CapabilityManifest declares the capabilities granted to an agent.
type CapabilityManifest struct {
	Capabilities []Capability `json:"capabilities"`
}

// CapabilitySet is an immutable view over the manifest entries.
type CapabilitySet struct {
	capabilities []Capability
}

// LoadCapabilityManifest parses and validates a manifest from r.
func LoadCapabilityManifest(r io.Reader) (CapabilitySet, error) {
	var manifest CapabilityManifest
	dec := json.NewDecoder(r)
	if err := dec.Decode(&manifest); err != nil {
		return CapabilitySet{}, fmt.Errorf("%w: %v", ErrInvalidCapabilityManifest, err)
	}
	return manifest.toSet()
}

// NewCapabilitySet constructs a set after validating the provided capabilities.
func NewCapabilitySet(capabilities []Capability) (CapabilitySet, error) {
	manifest := CapabilityManifest{Capabilities: capabilities}
	return manifest.toSet()
}

func (m CapabilityManifest) toSet() (CapabilitySet, error) {
	if len(m.Capabilities) == 0 {
		return CapabilitySet{}, fmt.Errorf("%w: no capabilities declared", ErrInvalidCapabilityManifest)
	}

	seen := make(map[string]struct{})
	out := make([]Capability, len(m.Capabilities))
	for i, cap := range m.Capabilities {
		if cap.Name == "" {
			return CapabilitySet{}, fmt.Errorf("%w: capability %d missing name", ErrInvalidCapabilityManifest, i)
		}
		if cap.Version == "" {
			return CapabilitySet{}, fmt.Errorf("%w: capability %s missing version", ErrInvalidCapabilityManifest, cap.Name)
		}
		key := cap.Name + ":" + cap.Version
		if _, ok := seen[key]; ok {
			return CapabilitySet{}, fmt.Errorf("%w: duplicate capability %s %s", ErrInvalidCapabilityManifest, cap.Name, cap.Version)
		}
		seen[key] = struct{}{}
		out[i] = cap
	}

	return CapabilitySet{capabilities: out}, nil
}

// List returns a defensive copy of the capabilities.
func (c CapabilitySet) List() []Capability {
	if len(c.capabilities) == 0 {
		return nil
	}
	dup := make([]Capability, len(c.capabilities))
	copy(dup, c.capabilities)
	return dup
}

// Has reports whether the given capability name exists, regardless of version.
func (c CapabilitySet) Has(name string) bool {
	for _, cap := range c.capabilities {
		if cap.Name == name {
			return true
		}
	}
	return false
}

// Len returns the number of declared capabilities.
func (c CapabilitySet) Len() int {
	return len(c.capabilities)
}
