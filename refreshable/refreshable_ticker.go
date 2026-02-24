package refreshable

import (
	"context"
	"time"
)

// ChangeDetector determines whether an underlying data source has changed since the last successful read.
// Implementations handle internal bookkeeping of previous state.
type ChangeDetector interface {
	// ShouldUpdate returns true if the data source appears to have changed
	// since the last call to MarkUpdated, or if the change status cannot be determined.
	ShouldUpdate(ctx context.Context) bool
	// MarkUpdated commits the pending state from the last ShouldUpdate call,
	// so that subsequent ShouldUpdate calls compare against it.
	MarkUpdated()
}

type alwaysCheckChangeDetector struct{}

func NewAlwaysCheckChangeDetector() ChangeDetector {
	return &alwaysCheckChangeDetector{}
}

func (alwaysCheckChangeDetector) ShouldUpdate(context.Context) bool { return true }
func (alwaysCheckChangeDetector) MarkUpdated()                      {}

func NewRefreshableTickerWithDuration[M any](ctx context.Context, a time.Duration, readerFunc func() (M, error), detector ChangeDetector) Validated[M] {
	return NewRefreshableTicker(ctx, time.Tick(a), readerFunc, detector)
}

// NewRefreshableTicker returns a [Validated] refreshable whose current value is read using the provided readerFunc.
// The readerFunc is only called when the [ChangeDetector] indicates the data source has changed.
// The detector's MarkUpdated is called after each successful read.
// The readerFunc is called once initially and then on each tick (subject to the detector) until the context is cancelled.
// If reading fails, the Current() value will be unchanged. The error is present in v.Validation().
func NewRefreshableTicker[M any](ctx context.Context, updateTicker <-chan time.Time, readerFunc func() (M, error), detector ChangeDetector) Validated[M] {
	v := newValidRefreshable[M]()
	updateValidRefreshable(v, readerFunc)
	if _, err := v.Validation(); err == nil {
		detector.MarkUpdated()
	}
	go func() {
		for {
			select {
			case <-updateTicker:
				if !detector.ShouldUpdate(ctx) {
					continue
				}
				updateValidRefreshable(v, readerFunc)
				if _, err := v.Validation(); err == nil {
					detector.MarkUpdated()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return v
}
