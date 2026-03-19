// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/palantir/witchcraft-go-logging/wlog/svclog/svc1log"
)

// Global atomic counters for debugging
var (
	// RefreshableTickerCallCount tracks the number of times NewRefreshableTicker is called
	RefreshableTickerCallCount atomic.Int64
	// RefreshableTickerIterationCount tracks the total number of iterations across all ticker goroutines
	RefreshableTickerIterationCount atomic.Int64
	// RefreshableTickerUpdateCount tracks the number of times updateValidRefreshable is called in ticker goroutines
	RefreshableTickerUpdateCount atomic.Int64
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

func NewRefreshableTickerWithDuration[M any](ctx context.Context, a time.Duration, readerFunc func(context.Context) (M, error), detector ChangeDetector) Validated[M] {
	return NewRefreshableTicker(ctx, time.Tick(a), readerFunc, detector)
}

// NewRefreshableTicker returns a [Validated] refreshable whose current value is read using the provided readerFunc.
// The readerFunc is only called when the [ChangeDetector] indicates the data source has changed.
// The detector's MarkUpdated is called after each successful read.
// The readerFunc is called once initially and then on each tick (subject to the detector) until the context is cancelled.
// If reading fails, the Unvalidated() value will be unchanged. The error is present in v.Validation().
func NewRefreshableTicker[M any](ctx context.Context, updateTicker <-chan time.Time, readerFunc func(context.Context) (M, error), detector ChangeDetector) Validated[M] {
	// Increment global call counter
	callCount := RefreshableTickerCallCount.Add(1)

	// Capture call stack
	callStack := captureCallStack()

	svc1log.FromContext(ctx).Debug("NewRefreshableTicker called",
		svc1log.SafeParam("callCount", callCount),
		svc1log.SafeParam("callStack", callStack))

	v := newValidRefreshable[M]()
	updateValidRefreshable(ctx, v, readerFunc)
	if _, err := v.Validation(); err == nil {
		detector.MarkUpdated()
	}
	go func() {
		localIterationCount := int64(0)
		for {
			select {
			case <-updateTicker:
				localIterationCount++
				globalIterationCount := RefreshableTickerIterationCount.Add(1)

				svc1log.FromContext(ctx).Debug("RefreshableTicker iteration",
					svc1log.SafeParam("localIterationCount", localIterationCount),
					svc1log.SafeParam("globalIterationCount", globalIterationCount),
					svc1log.SafeParam("originalCallStack", callStack),
				)

				if !detector.ShouldUpdate(ctx) {
					continue
				}

				// Increment counter for updateValidRefreshable calls
				updateCount := RefreshableTickerUpdateCount.Add(1)
				svc1log.FromContext(ctx).Debug("Calling updateValidRefreshable",
					svc1log.SafeParam("updateCount", updateCount),
					svc1log.SafeParam("originalCallStack", callStack),
				)

				updateValidRefreshable(ctx, v, readerFunc)
				if _, err := v.Validation(); err == nil {
					detector.MarkUpdated()
				}
			case <-ctx.Done():
				svc1log.FromContext(ctx).Debug("RefreshableTicker goroutine exiting",
					svc1log.SafeParam("localIterations", localIterationCount),
					svc1log.SafeParam("originalCallStack", callStack))
				return
			}
		}
	}()
	return v
}

// captureCallStack captures the current call stack and formats it as "file:line -> file:line -> ..."
// It stops collecting frames once it reaches standard library code (outside vendor/module).
func captureCallStack() string {
	const maxDepth = 20
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pcs) // Skip 3 frames: Callers, captureCallStack, NewRefreshableTicker

	frames := runtime.CallersFrames(pcs[:n])
	var stackParts []string

	for {
		frame, more := frames.Next()

		// Stop collecting if we've reached standard library or runtime code
		// We want to keep vendor/ paths and module paths but exclude stdlib
		if !strings.Contains(frame.File, "/vendor/") &&
			!strings.Contains(frame.File, "apollo-unbundler") &&
			!strings.Contains(frame.File, "github.com/palantir") {
			break
		}

		stackParts = append(stackParts, fmt.Sprintf("%s:%d", frame.File, frame.Line))
		if !more {
			break
		}
	}

	return strings.Join(stackParts, " -> ")
}
