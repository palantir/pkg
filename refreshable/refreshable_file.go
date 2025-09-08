// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"os"
	"time"
)

const (
	fileRefreshableSyncPeriod = time.Second
)

// NewFileRefreshable creates a NewFileRefreshableWithTicker with time.Tick(time.Second).
func NewFileRefreshable(ctx context.Context, filePath string) Validated[[]byte] {
	return NewFileRefreshableWithTicker(ctx, filePath, time.Tick(fileRefreshableSyncPeriod))
}

// NewFileRefreshableWithTicker returns a new Refreshable whose current value is the bytes of the file at the provided path.
// Calling this function starts a goroutine which re-reads the file on each tick.
// The goroutine will terminate when the provided context is cancelled.
func NewFileRefreshableWithTicker(ctx context.Context, filePath string, updateTicker <-chan time.Time) Validated[[]byte] {
	v := newValidRefreshable[[]byte]()
	go func() {
		for {
			// Read file and update refreshable. If ReadFile fails, the error is present in v.Validation().
			updateValidRefreshable[string, []byte](v, filePath, os.ReadFile)
			// Wait for next tick.
			select {
			case <-updateTicker:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return v
}
