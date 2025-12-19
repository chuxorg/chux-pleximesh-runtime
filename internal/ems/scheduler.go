package ems

import "time"

// SliceIndex uniquely identifies a 30-second slice.
type SliceIndex int64

// SliceScheduler calculates current and acceptable EMS slices.
type SliceScheduler struct {
	sliceDuration time.Duration
	tolerance     int64
	nowFunc       func() time.Time
}

// SchedulerOption configures a SliceScheduler.
type SchedulerOption func(*SliceScheduler)

// WithSliceDuration overrides the slice granularity.
func WithSliceDuration(d time.Duration) SchedulerOption {
	return func(s *SliceScheduler) {
		if d > 0 {
			s.sliceDuration = d
		}
	}
}

// WithSliceTolerance sets the ± tolerance in slices.
func WithSliceTolerance(v int64) SchedulerOption {
	return func(s *SliceScheduler) {
		if v >= 0 {
			s.tolerance = v
		}
	}
}

// WithSchedulerClock overrides the time source.
func WithSchedulerClock(now func() time.Time) SchedulerOption {
	return func(s *SliceScheduler) {
		s.nowFunc = now
	}
}

// NewSliceScheduler creates a scheduler that defaults to 30-second slices with ±1 tolerance.
func NewSliceScheduler(opts ...SchedulerOption) *SliceScheduler {
	s := &SliceScheduler{
		sliceDuration: 30 * time.Second,
		tolerance:     1,
		nowFunc:       time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CurrentSlice returns the slice for the configured clock.
func (s *SliceScheduler) CurrentSlice() SliceIndex {
	return s.SliceAt(s.nowFunc())
}

// SliceAt returns the slice index for the provided instant.
func (s *SliceScheduler) SliceAt(ts time.Time) SliceIndex {
	if s.sliceDuration <= 0 {
		return SliceIndex(0)
	}
	return SliceIndex(ts.UnixNano() / s.sliceDuration.Nanoseconds())
}

// AcceptableSlices returns the slice window ± tolerance for the provided instant.
func (s *SliceScheduler) AcceptableSlices(ts time.Time) []SliceIndex {
	current := s.SliceAt(ts)
	if s.tolerance == 0 {
		return []SliceIndex{current}
	}
	window := make([]SliceIndex, 0, 1+2*int(s.tolerance))
	for offset := -s.tolerance; offset <= s.tolerance; offset++ {
		window = append(window, current+SliceIndex(offset))
	}
	return window
}

// WithinWindow reports whether candidate lies within the acceptable window for the instant.
func (s *SliceScheduler) WithinWindow(candidate SliceIndex, ts time.Time) bool {
	current := s.SliceAt(ts)
	diff := int64(candidate - current)
	if diff < 0 {
		diff = -diff
	}
	return diff <= s.tolerance
}
