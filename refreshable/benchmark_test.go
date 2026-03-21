// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
)

// BenchmarkParentUpdateAfterDerivedDrop measures the cost of parent.Update()
// after creating and dropping N derived refreshables without calling stop().
// Without runtime.AddCleanup, each dropped derived leaks a subscriber, so
// Update iterates all N callbacks. With cleanup, GC removes them.
func BenchmarkParentUpdateAfterDerivedDrop(b *testing.B) {
	for _, dropped := range []int{0, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("dropped=%d", dropped), func(b *testing.B) {
			parent := refreshable.New(0)
			for i := range dropped {
				derived := refreshable.MapAuto(parent, func(v int) int { return v + i })
				_ = derived.Current()
			}
			runtime.GC()
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			b.ResetTimer()
			for i := range b.N {
				parent.Update(i + 1)
			}
		})
	}
}

// BenchmarkParentSubscriberCountAfterDrop verifies that the parent's subscriber
// list does not grow when derived refreshables are dropped and GC'd.
// This is a correctness-flavored benchmark: without runtime.AddCleanup,
// subscriberCount == dropped; with cleanup, subscriberCount == 0 after GC.
func BenchmarkParentSubscriberCountAfterDrop(b *testing.B) {
	for _, dropped := range []int{100, 1000} {
		b.Run(fmt.Sprintf("dropped=%d", dropped), func(b *testing.B) {
			for range b.N {
				parent := refreshable.New(0)
				var callCount atomic.Int64
				parent.Subscribe(func(int) { callCount.Add(1) })
				callCount.Store(0) // reset after initial Subscribe callback

				for i := range dropped {
					derived := refreshable.MapAuto(parent, func(v int) int { return v + i })
					_ = derived.Current()
				}
				runtime.GC()
				runtime.GC()
				time.Sleep(100 * time.Millisecond)

				callCount.Store(0)
				parent.Update(42)
				if got := callCount.Load(); got != 1 {
					b.Fatalf("expected 1 subscriber callback, got %d (leaked %d)", got, got-1)
				}
			}
		})
	}
}
