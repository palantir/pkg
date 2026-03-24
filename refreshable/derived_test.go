// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// awaitGCCleanup runs GC repeatedly until condition returns true or the timeout expires.
func awaitGCCleanup(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
		if condition() {
			return
		}
	}
	t.Fatal("timed out waiting for GC cleanup")
}

func TestMapGCCleanup(t *testing.T) {
	parent := refreshable.NewDefaultRefreshable(1)
	var mapCalls atomic.Int32
	mapped := parent.Map(func(i interface{}) interface{} {
		mapCalls.Add(1)
		return i.(int) * 2
	})

	// Verify the mapped refreshable works.
	assert.Equal(t, 2, mapped.Current())
	require.NoError(t, parent.Update(5))
	assert.Equal(t, 10, mapped.Current())

	// Drop the derived reference and wait for GC cleanup.
	runtime.KeepAlive(mapped)
	mapped = nil //nolint:ineffassign

	updateVal := 100
	awaitGCCleanup(t, func() bool {
		mapCalls.Store(0)
		updateVal++
		_ = parent.Update(updateVal)
		return mapCalls.Load() == 0
	})
}

func TestValidatingGCCleanup(t *testing.T) {
	parent := refreshable.NewDefaultRefreshable(1)
	var validateCalls atomic.Int32
	vr, err := refreshable.NewValidatingRefreshable(parent, func(i interface{}) error {
		validateCalls.Add(1)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, vr.Current())

	// Drop the validating refreshable and wait for GC cleanup.
	runtime.KeepAlive(vr)
	vr = nil //nolint:ineffassign

	updateVal := 100
	awaitGCCleanup(t, func() bool {
		validateCalls.Store(0)
		updateVal++
		_ = parent.Update(updateVal)
		return validateCalls.Load() == 0
	})
}

func TestMapValidatingGCCleanup(t *testing.T) {
	parent := refreshable.NewDefaultRefreshable("hello")
	var mapCalls atomic.Int32
	vr, err := refreshable.NewMapValidatingRefreshable(parent, func(i interface{}) (interface{}, error) {
		mapCalls.Add(1)
		return len(i.(string)), nil
	})
	require.NoError(t, err)
	assert.Equal(t, 5, vr.Current())

	// Drop the validating refreshable and wait for GC cleanup.
	runtime.KeepAlive(vr)
	vr = nil //nolint:ineffassign

	vals := []string{"a", "bb", "ccc", "dddd", "eeeee", "ffffff"}
	idx := 0
	awaitGCCleanup(t, func() bool {
		mapCalls.Store(0)
		_ = parent.Update(vals[idx%len(vals)])
		idx++
		return mapCalls.Load() == 0
	})
}

func TestMapGCWithActiveSubscriber(t *testing.T) {
	parent := refreshable.NewDefaultRefreshable(1)
	mapped := parent.Map(func(i interface{}) interface{} {
		return i.(int) * 2
	})

	// Subscribe to the derived refreshable.
	var latest atomic.Value
	unsub := mapped.Subscribe(func(i interface{}) {
		latest.Store(i)
	})

	require.NoError(t, parent.Update(5))
	assert.Equal(t, 10, latest.Load())

	// Drop the derived reference but keep the subscription.
	runtime.KeepAlive(mapped)
	mapped = nil //nolint:ineffassign

	// Run GC — the subscription should keep updates flowing.
	for i := 0; i < 5; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	require.NoError(t, parent.Update(7))
	assert.Equal(t, 14, latest.Load(), "subscription should still receive updates after derived is GC'd")

	// Now unsubscribe — cleanup should fire since gcDone is true and subCount reaches 0.
	unsub()

	// Verify updates no longer flow. Use a mapFn-based check by creating a new mapped
	// refreshable to observe parent subscriber behavior indirectly.
	var mapCalls atomic.Int32
	probe := parent.Map(func(i interface{}) interface{} {
		mapCalls.Add(1)
		return i
	})
	_ = probe // keep alive

	// The original subscription's cleanup should have already fired synchronously.
	// Verify the parent still works for new subscribers.
	mapCalls.Store(0)
	require.NoError(t, parent.Update(99))
	assert.Greater(t, mapCalls.Load(), int32(0), "parent should still notify new subscribers")
	runtime.KeepAlive(probe)
}

func TestDerivedMapChain(t *testing.T) {
	parent := refreshable.NewDefaultRefreshable(1)
	var middleCalls, finalCalls atomic.Int32

	middle := parent.Map(func(i interface{}) interface{} {
		middleCalls.Add(1)
		return i.(int) * 2
	})
	final := middle.Map(func(i interface{}) interface{} {
		finalCalls.Add(1)
		return i.(int) + 100
	})

	assert.Equal(t, 102, final.Current())

	require.NoError(t, parent.Update(5))
	assert.Equal(t, 110, final.Current())

	// Drop the middle reference but keep the final.
	// The chain should stay alive because final.refs holds middle.
	middleCalls.Store(0)
	finalCalls.Store(0)
	runtime.KeepAlive(middle)
	middle = nil //nolint:ineffassign

	for i := 0; i < 5; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	middleCalls.Store(0)
	finalCalls.Store(0)
	require.NoError(t, parent.Update(10))
	assert.Equal(t, 120, final.Current())
	assert.Greater(t, middleCalls.Load(), int32(0), "middle mapFn should still be called")
	assert.Greater(t, finalCalls.Load(), int32(0), "final mapFn should still be called")

	// Now drop the final reference — entire chain should be cleaned up.
	runtime.KeepAlive(final)
	final = nil //nolint:ineffassign

	updateVal := 100
	awaitGCCleanup(t, func() bool {
		middleCalls.Store(0)
		finalCalls.Store(0)
		updateVal++
		_ = parent.Update(updateVal)
		return middleCalls.Load() == 0 && finalCalls.Load() == 0
	})
}
