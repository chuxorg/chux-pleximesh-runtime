package ems

import (
	"testing"
	"time"
)

func TestSliceSchedulerWindow(t *testing.T) {
	base := time.Unix(120, 0).UTC()
	s := NewSliceScheduler(WithSchedulerClock(func() time.Time {
		return base
	}))

	if got := s.CurrentSlice(); got != SliceIndex(4) {
		t.Fatalf("unexpected current slice: got %d want 4", got)
	}

	window := s.AcceptableSlices(base)
	expected := []SliceIndex{3, 4, 5}
	if len(window) != len(expected) {
		t.Fatalf("unexpected window length: got %d want %d", len(window), len(expected))
	}
	for i, slice := range expected {
		if window[i] != slice {
			t.Fatalf("window[%d] mismatch: got %d want %d", i, window[i], slice)
		}
	}

	if !s.WithinWindow(SliceIndex(5), base.Add(5*time.Second)) {
		t.Fatalf("expected slice 5 to be within tolerance")
	}
	if s.WithinWindow(SliceIndex(6), base.Add(5*time.Second)) {
		t.Fatalf("did not expect slice 6 to be accepted")
	}
}
