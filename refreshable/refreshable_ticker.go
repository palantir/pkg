package refreshable

import (
	"context"
	"time"
)

func NewRefreshableTickerWithDuration[M any](ctx context.Context, a time.Duration, readerFunc func() (M, error)) Validated[M] {
	return NewRefreshableTicker(ctx, time.Tick(a), readerFunc)
}

// NewFileRefreshableWithReaderFunc returns a [Validated] refreshable whose current value is the bytes read using the provided readerFunc.
// This function is similar to [NewFileRefreshableWithTicker] but allows callers to provide a custom file reading function
// instead of using os.ReadFile directly. This is useful for scenarios where custom file processing is needed
// (e.g., decompression, decryption, or other transformations).
//
// The readerFunc is called once initially and then on each tick until the context is cancelled.
// If reading fails, the Current() value will be unchanged. The error is present in v.Validation().
func NewRefreshableTicker[M any](ctx context.Context, updateTicker <-chan time.Time, readerFunc func() (M, error)) Validated[M] {
	v := newValidRefreshable[M]()
	updateValidRefreshable(v, readerFunc)
	go func() {
		for {
			select {
			case <-updateTicker:
				// Read file and update refreshable. If readerFunc fails, the error is present in v.Validation().
				updateValidRefreshable(v, readerFunc)
			case <-ctx.Done():
				return
			}
		}
	}()
	return v
}
