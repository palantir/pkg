// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"os"
	"path/filepath"
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
// It is equivalent to calling NewFileRefreshableWithReaderFunc with os.ReadFile.
func NewFileRefreshableWithTicker(ctx context.Context, filePath string, updateTicker <-chan time.Time) Validated[[]byte] {
	return NewFileRefreshableWithReaderFunc(ctx, filePath, updateTicker, os.ReadFile)
}

// NewFileRefreshableWithReaderFunc returns a [Validated] refreshable whose current value is the bytes read using the provided readerFunc.
// This function is similar to [NewFileRefreshableWithTicker] but allows callers to provide a custom file reading function
// instead of using os.ReadFile directly. This is useful for scenarios where custom file processing is needed
// (e.g., decompression, decryption, or other transformations).
//
// The readerFunc is called once initially and then on each tick until the context is cancelled.
// If reading fails, the Current() value will be unchanged. The error is present in v.Validation().
func NewFileRefreshableWithReaderFunc(ctx context.Context, filePath string, updateTicker <-chan time.Time, readerFunc func(string) ([]byte, error)) Validated[[]byte] {
	v := newValidRefreshable[[]byte]()
	updateValidRefreshable(v, filePath, readerFunc)
	detector := newStatFileChangeDetector(ctx, filePath)
	go func() {
		for {
			select {
			case <-updateTicker:
				if !detector.ShouldUpdate(ctx, filePath) {
					continue
				}
				updateValidRefreshable(v, filePath, readerFunc)
			case <-ctx.Done():
				return
			}
		}
	}()
	return v
}

// FileChangeDetector determines whether a file has changed since the last check.
// Implementations handle internal bookkeeping of previous state.
type FileChangeDetector interface {
	// ShouldUpdate returns true if the file at filePath appears to have changed
	// since the last call, or if the change status cannot be determined.
	ShouldUpdate(ctx context.Context, filePath string) bool
}

type statFileChangeDetector struct {
	lastResolvedPath string
	lastModTime      time.Time
	lastSize         int64
}

func newStatFileChangeDetector(ctx context.Context, filePath string) *statFileChangeDetector {
	d := &statFileChangeDetector{}
	d.ShouldUpdate(ctx, filePath)
	return d
}

func (d *statFileChangeDetector) ShouldUpdate(ctx context.Context, filePath string) bool {
	resolvedPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return true
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return true
	}
	if resolvedPath == d.lastResolvedPath && info.ModTime().Equal(d.lastModTime) && info.Size() == d.lastSize {
		return false
	}
	d.lastResolvedPath = resolvedPath
	d.lastModTime = info.ModTime()
	d.lastSize = info.Size()
	return true
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
