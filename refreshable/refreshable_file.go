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

// NewFileRefreshable creates a Validated refreshable that reads from a file every second.
// It is equivalent to calling NewFileRefreshableWithTicker with time.Tick(time.Second).
func NewFileRefreshable(ctx context.Context, filePath string) Validated[[]byte] {
	return NewFileRefreshableWithTicker(ctx, filePath, time.Tick(fileRefreshableSyncPeriod))
}

// NewFileRefreshableWithTicker returns a Validated refreshable whose current value is the bytes of the file at the provided path.
// This function reads the file once then starts a goroutine which re-reads the file on each tick until the provided context is cancelled.
// If reading the file fails, the Current() value will be unchanged. The error is present in v.Validation().
func NewFileRefreshableWithTicker(ctx context.Context, filePath string, updateTicker <-chan time.Time) Validated[[]byte] {
	v := newValidRefreshable[[]byte]()
	updateValidRefreshable[string, []byte](v, filePath, os.ReadFile)
	go func() {
		for {
			select {
			case <-updateTicker:
				// Read file and update refreshable. If ReadFile fails, the error is present in v.Validation().
				updateValidRefreshable[string, []byte](v, filePath, os.ReadFile)
			case <-ctx.Done():
				return
			}
		}
	}()
	return v
}

// NewMultiFileRefreshable creates a Validated Refreshable that tracks the contents of multiple files.
// The input is a Refreshable of a set of file paths (map keys). The output is a Validated Refreshable
// of a map from file path to file contents. When files are added to or removed from the input set,
// the corresponding file watchers are created or destroyed. Each file is read periodically
// using NewFileRefreshable.
//
// Current() returns a map containing only successfully read files.
// Validation() returns the map and a joined error of all file read failures.
func NewMultiFileRefreshable(ctx context.Context, paths Refreshable[map[string]struct{}]) Validated[map[string][]byte] {
	return MapValues(ctx, paths, func(ctx context.Context, path string, _ struct{}) Validated[[]byte] {
		return NewFileRefreshable(ctx, path)
	})
}
