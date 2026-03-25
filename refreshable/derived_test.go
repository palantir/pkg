// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GC cleanup: dropping a derived refreshable with no active subscribers
// removes the upstream subscription.
// ---------------------------------------------------------------------------

func TestMapGCCleanup(t *testing.T) {
	parent := refreshable.New(1)

	var mapCalls atomic.Int64
	derived := refreshable.MapAuto(parent, func(v int) int {
		mapCalls.Add(1)
		return v * 2
	})
	require.Equal(t, int64(1), mapCalls.Load())
	require.Equal(t, 2, derived.Current())

	parent.Update(2)
	require.Equal(t, int64(2), mapCalls.Load())
	require.Equal(t, 4, derived.Current())

	derived = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100) // unique value to avoid DeepEqual skip
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "subscription should be cleaned up after GC")
}

func TestMapGCCleanup_ExplicitUnsub(t *testing.T) {
	parent := refreshable.New(1)

	var mapCalls atomic.Int64
	derived, stop := refreshable.Map(parent, func(v int) int {
		mapCalls.Add(1)
		return v * 2
	})
	require.Equal(t, 2, derived.Current())

	stop()
	before := mapCalls.Load()
	parent.Update(99)
	require.Equal(t, before, mapCalls.Load(), "should not call map after explicit unsub")
}

func TestCachedGCCleanup(t *testing.T) {
	parent := refreshable.New(1)

	var subscribeCalls atomic.Int64
	wrapper := refreshable.View(parent, func(v int) int {
		subscribeCalls.Add(1)
		return v
	})

	cached := refreshable.CachedAuto(wrapper)
	require.Equal(t, 1, cached.Current())

	cached = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := subscribeCalls.Load()
		parent.Update(int(before) + 100)
		return subscribeCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "cached subscription should be cleaned up after GC")
}

func TestMergeGCCleanup(t *testing.T) {
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)

	var mergeCalls atomic.Int64
	merged := refreshable.MergeAuto(r1, r2, func(a, b int) int {
		mergeCalls.Add(1)
		return a + b
	})
	require.Equal(t, 3, merged.Current())

	merged = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mergeCalls.Load()
		r1.Update(int(before) + 100)
		return mergeCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "merge subscription should be cleaned up after GC")
}

func TestCollectGCCleanup(t *testing.T) {
	r1 := refreshable.New(10)
	r2 := refreshable.New(20)

	collected := refreshable.CollectAuto(r1, r2)
	require.Equal(t, []int{10, 20}, collected.Current())

	collected = nil
	runtime.GC()
	runtime.GC()

	var newSubCalls atomic.Int64
	r1.Subscribe(func(int) { newSubCalls.Add(1) })

	assert.Eventually(t, func() bool {
		before := newSubCalls.Load()
		r1.Update(int(before) + 200)
		return newSubCalls.Load()-before == 1
	}, time.Second, 10*time.Millisecond)
}

// TestMergeGCCleanup_DerivedInputs verifies that Merge cleans up when its
// inputs are derived refreshables (not plain Updatables). Without push-based
// value propagation, the doUpdate closure would capture the input derived
// wrappers, preventing their runtime.AddCleanup from firing.
func TestMergeGCCleanup_DerivedInputs(t *testing.T) {
	parent1 := refreshable.New(1)
	parent2 := refreshable.New(2)

	d1 := refreshable.MapAuto(parent1, func(v int) int { return v * 10 })
	d2 := refreshable.MapAuto(parent2, func(v int) int { return v * 100 })

	var mergeCalls atomic.Int64
	merged := refreshable.MergeAuto(d1, d2, func(a, b int) int {
		mergeCalls.Add(1)
		return a + b
	})
	require.Equal(t, 210, merged.Current())

	d1 = nil     //nolint:ineffassign
	d2 = nil     //nolint:ineffassign
	merged = nil //nolint:ineffassign

	forceGCAndCleanup()

	assert.Eventually(t, func() bool {
		before := mergeCalls.Load()
		parent1.Update(int(before) + 100)
		return mergeCalls.Load() == before
	}, 5*time.Second, 10*time.Millisecond, "merge with derived inputs should clean up after GC")
}

// TestCollectGCCleanup_DerivedInputs verifies that Collect cleans up when its
// inputs are derived refreshables. Without push-based value propagation, the
// doUpdate closure would capture the input derived wrappers via the
// refreshables slice, preventing their runtime.AddCleanup from firing.
func TestCollectGCCleanup_DerivedInputs(t *testing.T) {
	parent1 := refreshable.New(10)
	parent2 := refreshable.New(20)

	d1 := refreshable.MapAuto(parent1, func(v int) int { return v * 2 })
	d2 := refreshable.MapAuto(parent2, func(v int) int { return v * 3 })

	var subCalls atomic.Int64
	parent1.Subscribe(func(int) { subCalls.Add(1) })

	collected := refreshable.CollectAuto(d1, d2)
	require.Equal(t, []int{20, 60}, collected.Current())

	d1 = nil        //nolint:ineffassign
	d2 = nil        //nolint:ineffassign
	collected = nil //nolint:ineffassign

	forceGCAndCleanup()

	assert.Eventually(t, func() bool {
		before := subCalls.Load()
		parent1.Update(int(before) + 200)
		return subCalls.Load()-before == 1
	}, 5*time.Second, 10*time.Millisecond, "collect with derived inputs should clean up after GC")
}

func TestValidateGCCleanup(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(10)

	var validateCalls atomic.Int64
	validated, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, v int) error {
		validateCalls.Add(1)
		if v < 0 {
			return errors.New("negative")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 10, validated.Unvalidated())

	validated = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := validateCalls.Load()
		parent.Update(int(before) + 100)
		return validateCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "validate subscription should be cleaned up after GC")
}

func TestMergeValidatedGCCleanup(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, err := refreshable.ValidateAuto(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	vr2, err := refreshable.ValidateAuto(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var mergeCalls atomic.Int64
	merged := refreshable.MergeValidatedAuto(vr1, vr2, func(a, b int) int {
		mergeCalls.Add(1)
		return a + b
	})
	require.Equal(t, 3, merged.Unvalidated())

	merged = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mergeCalls.Load()
		r1.Update(int(before) + 100)
		return mergeCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "merge validated subscription should be cleaned up after GC")
}

func TestMapFromValidatedGCCleanup(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(10)
	vr, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var mapCalls atomic.Int64
	mapped := refreshable.MapFromValidatedAuto(vr, func(v int) int {
		mapCalls.Add(1)
		return v * 2
	})
	require.Equal(t, 20, mapped.Current())

	mapped = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "MapFromValidated subscription should be cleaned up after GC")
}

// TestMapWithErrorGCCleanup verifies cleanup for MapWithError, which creates
// an internal intermediate via validatedFromRefreshable.
func TestMapWithErrorGCCleanup(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(10)

	var mapCalls atomic.Int64
	mapped, err := refreshable.MapWithErrorAuto(ctx, parent, func(_ context.Context, v int) (string, error) {
		mapCalls.Add(1)
		if v < 0 {
			return "", errors.New("negative")
		}
		return fmt.Sprintf("val=%d", v), nil
	})
	require.NoError(t, err)
	require.Equal(t, "val=10", mapped.Unvalidated())

	mapped = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "MapWithError subscription should be cleaned up after GC")
}

// TestMergeValidatedAndRefreshableGCCleanup verifies cleanup for
// MergeValidatedAndRefreshable, which wraps a plain Refreshable in a
// Validated internally.
func TestMergeValidatedAndRefreshableGCCleanup(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, err := refreshable.ValidateAuto(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var mergeCalls atomic.Int64
	merged := refreshable.MergeValidatedAndRefreshableAuto(ctx, vr1, r2, func(a, b int) int {
		mergeCalls.Add(1)
		return a + b
	})
	require.Equal(t, 3, merged.Unvalidated())

	merged = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mergeCalls.Load()
		r1.Update(int(before) + 100)
		return mergeCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "MergeValidatedAndRefreshable subscription should be cleaned up after GC")
}

// TestCollectValidatedGCCleanup_NoSubscribers verifies that dropping a
// CollectValidated with no subscribers cleans up immediately.
func TestCollectValidatedGCCleanup_NoSubscribers(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, err := refreshable.ValidateAuto(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	vr2, err := refreshable.ValidateAuto(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var subCalls atomic.Int64
	collected := refreshable.CollectValidatedAuto(vr1, vr2)
	require.Equal(t, []int{1, 2}, collected.Unvalidated())

	r1.Subscribe(func(int) { subCalls.Add(1) })

	collected = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := subCalls.Load()
		r1.Update(int(before) + 100)
		return subCalls.Load()-before == 1
	}, time.Second, 10*time.Millisecond, "CollectValidated subscription should be cleaned up after GC when no subscribers")
}

func TestMapValidatedGCCleanup(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(10)
	vr, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var mapCalls atomic.Int64
	mapped, err := refreshable.MapValidatedAuto(ctx, vr, func(_ context.Context, v int) (int, error) {
		mapCalls.Add(1)
		return v * 2, nil
	})
	require.NoError(t, err)
	require.Equal(t, 20, mapped.Unvalidated())

	mapped = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "MapValidated subscription should be cleaned up after GC")
}

// ---------------------------------------------------------------------------
// Direct subscribe: fire-and-forget on a non-derived refreshable is
// unaffected by GC cleanup.
// ---------------------------------------------------------------------------

func TestDirectSubscribeNotAffected(t *testing.T) {
	parent := refreshable.New(1)

	var calls atomic.Int64
	parent.Subscribe(func(int) { calls.Add(1) })

	runtime.GC()
	runtime.GC()

	before := calls.Load()
	parent.Update(2)
	require.Equal(t, before+1, calls.Load(), "direct subscribe should survive GC")
	parent.Update(3)
	require.Equal(t, before+2, calls.Load(), "direct subscribe should survive GC")
}

// ---------------------------------------------------------------------------
// Chain keep-alive: holding the leaf of a derived chain keeps the entire
// upstream chain alive. Dropping the leaf cascades cleanup.
// ---------------------------------------------------------------------------

// TestDerivedRefreshableHoldsSubscriptionAlive verifies that holding the
// derived keeps its upstream subscription active without the UnsubscribeFunc.
func TestDerivedRefreshableHoldsSubscriptionAlive(t *testing.T) {
	parent := refreshable.New(1)
	derived := refreshable.MapAuto(parent, func(v int) int { return v * 10 })

	runtime.GC()
	runtime.GC()

	parent.Update(5)
	require.Equal(t, 50, derived.Current())
	parent.Update(10)
	require.Equal(t, 100, derived.Current())
}

// TestChainedDerivedKeepsUpstreamAlive verifies that a chain of derived
// refreshables keeps the entire upstream chain alive via refs.
func TestChainedDerivedKeepsUpstreamAlive(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(1)

	var validateCalls atomic.Int64
	intermediate, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, v int) error {
		validateCalls.Add(1)
		if v < 0 {
			return errors.New("negative")
		}
		return nil
	})
	require.NoError(t, err)

	var mapCalls atomic.Int64
	leaf := refreshable.MapFromValidatedAuto(intermediate, func(v int) int {
		mapCalls.Add(1)
		return v * 3
	})
	require.Equal(t, 3, leaf.Current())

	intermediate = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	parent.Update(5)
	require.Equal(t, 15, leaf.Current())
	parent.Update(10)
	require.Equal(t, 30, leaf.Current())

	parent.Update(-1)
	require.Equal(t, 30, leaf.Current(), "should retain last valid mapped value")
	parent.Update(7)
	require.Equal(t, 21, leaf.Current())
}

// TestMapValidatedChainGCCleanup verifies that dropping the leaf AND
// intermediate cascades cleanup to the root.
func TestMapValidatedChainGCCleanup(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(1)

	var validateCalls atomic.Int64
	intermediate, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, v int) error {
		validateCalls.Add(1)
		return nil
	})
	require.NoError(t, err)

	leaf := refreshable.MapFromValidatedAuto(intermediate, func(v int) int { return v })
	require.Equal(t, 1, leaf.Current())

	leaf = nil
	intermediate = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := validateCalls.Load()
		parent.Update(int(before) + 100)
		return validateCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "chained subscription should be cleaned up after GC")
}

// TestLongMapChain verifies a 5-level Map chain stays alive when only the
// leaf is held, and cascades cleanup when the leaf is dropped.
func TestLongMapChain(t *testing.T) {
	parent := refreshable.New(1)
	var callCounts [5]atomic.Int64

	var current refreshable.Refreshable[int] = parent
	for i := range 5 {
		idx := i
		next := refreshable.MapAuto(current, func(v int) int {
			callCounts[idx].Add(1)
			return v + 1
		})
		current = next
	}
	leaf := current
	require.Equal(t, 6, leaf.Current()) // 1 + 5

	current = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	parent.Update(10)
	require.Equal(t, 15, leaf.Current()) // 10 + 5

	leaf = nil

	assert.Eventually(t, func() bool {
		runtime.GC()
		before := callCounts[0].Load()
		parent.Update(int(before) + 100)
		return callCounts[0].Load() == before
	}, 5*time.Second, 10*time.Millisecond, "long chain should be cleaned up after dropping leaf")
}

// TestMapWithErrorFanOutMapOnResult tests: MapWithError → fan-out → Map on one result → drop.
func TestMapWithErrorFanOutMapOnResult(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(42)

	var mapCalls atomic.Int64
	validParams, err := refreshable.MapWithErrorAuto(ctx, parent, func(_ context.Context, v int) (int, error) {
		mapCalls.Add(1)
		return v * 2, nil
	})
	require.NoError(t, err)

	// Fan out
	r1 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 1 })
	r2 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 2 })

	// Map on one fan-out result (subscribes to r1)
	derived := refreshable.MapAuto(r1, func(v int) int { return v * 10 })
	_ = derived.Current()

	_ = r2
	r1 = nil          //nolint:ineffassign
	r2 = nil          //nolint:ineffassign
	validParams = nil //nolint:ineffassign
	derived = nil     //nolint:ineffassign

	forceGCAndCleanup()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, 5*time.Second, 10*time.Millisecond, "fan-out with Map should clean up after all are dropped")
}

// TestMapWithErrorFanOutMergeValidatedAndRefreshable tests:
// MapWithError → fan-out → MergeValidatedAndRefreshable → drop.
func TestMapWithErrorFanOutMergeValidatedAndRefreshable(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(42)

	var mapCalls atomic.Int64
	validParams, err := refreshable.MapWithErrorAuto(ctx, parent, func(_ context.Context, v int) (int, error) {
		mapCalls.Add(1)
		return v * 2, nil
	})
	require.NoError(t, err)

	// Fan out
	r1 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 1 })

	// MergeValidatedAndRefreshable (subscribes to validParams and r1)
	merged := refreshable.MergeValidatedAndRefreshableAuto(ctx, validParams, r1, func(a, b int) int {
		return a + b
	})
	_ = merged.Unvalidated()

	r1 = nil          //nolint:ineffassign
	validParams = nil //nolint:ineffassign
	merged = nil      //nolint:ineffassign

	forceGCAndCleanup()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, 5*time.Second, 10*time.Millisecond, "fan-out with MergeValidatedAndRefreshable should clean up after all are dropped")
}

// TestMapWithErrorFanOutWithSubscribers mimics the pattern from
// NewClientFromRefreshableConfig: MapWithError → many MapFromValidated fan-out,
// plus additional Map/Subscribe calls on some of the fan-out results.
// This is a more complex cleanup chain than the simple fan-out test.
func TestMapWithErrorFanOutWithSubscribers(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(42)

	var mapCalls atomic.Int64
	validParams, err := refreshable.MapWithErrorAuto(ctx, parent, func(_ context.Context, v int) (int, error) {
		mapCalls.Add(1)
		return v * 2, nil
	})
	require.NoError(t, err)
	require.Equal(t, 84, validParams.Unvalidated())

	// Fan out: create multiple MapFromValidated results
	const fanOutCount = 10
	fanOut := make([]refreshable.Refreshable[int], fanOutCount)
	for i := range fanOutCount {
		idx := i
		fanOut[i] = refreshable.MapFromValidatedAuto(validParams, func(v int) int {
			return v + idx
		})
	}

	// Subscribe to one of the fan-out results (like Map(b.TransportParams, ...))
	derived := refreshable.MapAuto(fanOut[2], func(v int) int { return v * 10 })
	_ = derived.Current()

	// MergeValidatedAndRefreshable on validParams and the derived (like getRefreshableTLSConfig)
	merged := refreshable.MergeValidatedAndRefreshableAuto(ctx, validParams, derived, func(a, b int) int {
		return a + b
	})
	_ = merged.Unvalidated()

	// Also try subscribing to one of the fan-out results with Map then MergeValidatedAndRefreshable
	derived2 := refreshable.MapAuto(fanOut[3], func(v int) int { return v * 100 })
	merged2 := refreshable.MergeValidatedAndRefreshableAuto(ctx, validParams, derived2, func(a, b int) int {
		return a + b
	})
	_ = merged2.Unvalidated()

	// Drop everything
	for i := range fanOut {
		fanOut[i] = nil
	}
	validParams = nil //nolint:ineffassign
	derived = nil     //nolint:ineffassign
	merged = nil      //nolint:ineffassign
	derived2 = nil    //nolint:ineffassign
	merged2 = nil     //nolint:ineffassign

	forceGCAndCleanup()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, 5*time.Second, 10*time.Millisecond, "fan-out with subscribers should clean up after all are dropped")
}

// TestCollectValidatedMapValidatedPipeline mirrors a real-world TLS cert
// pipeline: CollectValidated -> MapValidated.
func TestCollectValidatedMapValidatedPipeline(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New([]byte("cert1"))
	r2 := refreshable.New([]byte("cert2"))
	vr1, err := refreshable.ValidateAuto(ctx, r1, func(_ context.Context, _ []byte) error { return nil })
	require.NoError(t, err)
	vr2, err := refreshable.ValidateAuto(ctx, r2, func(_ context.Context, _ []byte) error { return nil })
	require.NoError(t, err)

	collected := refreshable.CollectValidatedAuto(vr1, vr2)

	var flattenCalls atomic.Int64
	leaf, err := refreshable.MapValidatedAuto(ctx, collected, func(_ context.Context, certs [][]byte) (string, error) {
		flattenCalls.Add(1)
		var combined strings.Builder
		for _, c := range certs {
			combined.WriteString(string(c) + ";")
		}
		return combined.String(), nil
	})
	require.NoError(t, err)
	require.Equal(t, "cert1;cert2;", leaf.Unvalidated())

	collected = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	r1.Update([]byte("new-cert1"))
	require.Contains(t, leaf.Unvalidated(), "new-cert1")

	leaf = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := flattenCalls.Load()
		r1.Update(fmt.Appendf(nil, "val%d", before+100))
		return flattenCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "CollectValidated->MapValidated pipeline should clean up after GC")
}

// TestMapFromValidatedFanOut verifies that fanning a single Validated out
// into many MapFromValidated refreshables keeps all alive while held and
// cleans up when all are dropped.
func TestMapFromValidatedFanOut(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(42)
	validated, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	const fanOutCount = 10
	var mapCalls [fanOutCount]atomic.Int64
	refreshables := make([]refreshable.Refreshable[int], fanOutCount)
	for i := range fanOutCount {
		idx := i
		r := refreshable.MapFromValidatedAuto(validated, func(v int) int {
			mapCalls[idx].Add(1)
			return v + idx
		})
		refreshables[i] = r
	}

	for i, r := range refreshables {
		require.Equal(t, 42+i, r.Current())
	}

	parent.Update(100)
	for i, r := range refreshables {
		require.Equal(t, 100+i, r.Current())
	}

	for i := range refreshables {
		refreshables[i] = nil
	}
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls[0].Load()
		parent.Update(int(before) + 200)
		return mapCalls[0].Load() == before
	}, time.Second, 10*time.Millisecond, "fan-out subscriptions should clean up after all are dropped")
}

// TestDeepMapValidatedChain verifies a 4-level Validate -> MapValidated chain
// stays alive when only the leaf is held, and cascades cleanup when dropped.
func TestDeepMapValidatedChain(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(1)

	var callCounts [4]atomic.Int64
	v, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, i int) error {
		callCounts[0].Add(1)
		return nil
	})
	require.NoError(t, err)
	var current refreshable.Validated[int] = v

	for i := 1; i < 4; i++ {
		idx := i
		next, err := refreshable.MapValidatedAuto(ctx, current, func(_ context.Context, v int) (int, error) {
			callCounts[idx].Add(1)
			return v + 1, nil
		})
		require.NoError(t, err)
		current = next
	}
	leaf := current
	require.Equal(t, 4, leaf.Unvalidated()) // 1 + 3 increments

	current = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	v = nil       //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	parent.Update(10)
	require.Equal(t, 13, leaf.Unvalidated()) // 10 + 3

	leaf = nil

	assert.Eventually(t, func() bool {
		runtime.GC()
		before := callCounts[0].Load()
		parent.Update(int(before) + 100)
		return callCounts[0].Load() == before
	}, 5*time.Second, 10*time.Millisecond, "deep MapValidated chain should clean up after dropping leaf")
}

// ---------------------------------------------------------------------------
// Deferred cleanup: subscribing on a derived refreshable then dropping the
// derived reference defers cleanup until all subscribers unsubscribe.
// ---------------------------------------------------------------------------

// TestSubscribeOnDerivedThenDrop verifies that subscribers on a derived
// refreshable continue receiving updates after the derived is dropped.
func TestSubscribeOnDerivedThenDrop(t *testing.T) {
	parent := refreshable.New(1)
	derived := refreshable.MapAuto(parent, func(v int) int { return v * 10 })

	var subCalls atomic.Int64
	derived.Subscribe(func(int) { subCalls.Add(1) })
	require.Equal(t, int64(1), subCalls.Load())

	parent.Update(2)
	require.Equal(t, int64(2), subCalls.Load())

	derived = nil
	runtime.GC()
	runtime.GC()

	before := subCalls.Load()
	parent.Update(999)
	require.Equal(t, before+1, subCalls.Load(), "subscriber should still fire after derived is dropped")
}

// TestSubscribeOnDerivedThenDrop_UnsubCleansUp verifies that unsubscribing
// after the derived is dropped triggers the deferred upstream cleanup.
func TestSubscribeOnDerivedThenDrop_UnsubCleansUp(t *testing.T) {
	parent := refreshable.New(1)

	var mapCalls atomic.Int64
	derived := refreshable.MapAuto(parent, func(v int) int {
		mapCalls.Add(1)
		return v * 10
	})

	var subCalls atomic.Int64
	unsub := derived.Subscribe(func(int) { subCalls.Add(1) })
	require.Equal(t, int64(1), subCalls.Load())

	derived = nil
	runtime.GC()
	runtime.GC()

	before := subCalls.Load()
	parent.Update(999)
	require.Equal(t, before+1, subCalls.Load(), "subscriber should still fire after derived is dropped")

	unsub()

	assert.Eventually(t, func() bool {
		runtime.GC()
		mapBefore := mapCalls.Load()
		parent.Update(int(mapBefore) + 1000)
		return mapCalls.Load() == mapBefore
	}, time.Second, 10*time.Millisecond, "map function should stop after unsubscribe triggers deferred cleanup")
}

// TestSubscribeValidatedOnDerivedThenDrop verifies deferred cleanup for
// Validated refreshables.
func TestSubscribeValidatedOnDerivedThenDrop(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(1)
	validated, err := refreshable.ValidateAuto(ctx, parent, func(_ context.Context, v int) error { return nil })
	require.NoError(t, err)

	var subCalls atomic.Int64
	validated.SubscribeValidated(func(refreshable.Validated[int]) { subCalls.Add(1) })
	require.Equal(t, int64(1), subCalls.Load())

	parent.Update(2)
	require.Equal(t, int64(2), subCalls.Load())

	validated = nil
	runtime.GC()
	runtime.GC()

	before := subCalls.Load()
	parent.Update(999)
	require.Equal(t, before+1, subCalls.Load(), "validated subscriber should still fire after derived is dropped")
}

// TestCollectValidatedGCCleanup verifies that dropping a CollectValidated
// with an active subscriber defers cleanup.
func TestCollectValidatedGCCleanup(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, err := refreshable.ValidateAuto(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	vr2, err := refreshable.ValidateAuto(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)

	var subCalls atomic.Int64
	collected := refreshable.CollectValidatedAuto(vr1, vr2)
	collected.SubscribeValidated(func(refreshable.Validated[[]int]) { subCalls.Add(1) })
	require.Equal(t, []int{1, 2}, collected.Unvalidated())

	collected = nil
	runtime.GC()
	runtime.GC()

	before := subCalls.Load()
	r1.Update(int(before) + 100)
	require.Greater(t, subCalls.Load(), before, "subscriber should still fire after collected is dropped")
}

// TestLongMapChainSubscribeThenDrop subscribes on the leaf of a chain, drops
// all derived references, verifies the subscription keeps firing, then
// unsubscribes and verifies the entire chain cascades cleanup.
func TestLongMapChainSubscribeThenDrop(t *testing.T) {
	parent := refreshable.New(1)
	var callCounts [3]atomic.Int64

	var current refreshable.Refreshable[int] = parent
	for i := range 3 {
		idx := i
		next := refreshable.MapAuto(current, func(v int) int {
			callCounts[idx].Add(1)
			return v + 1
		})
		current = next
	}
	leaf := current
	require.Equal(t, 4, leaf.Current()) // 1 + 3

	var subCalls atomic.Int64
	unsub := leaf.Subscribe(func(int) { subCalls.Add(1) })
	require.Equal(t, int64(1), subCalls.Load())

	leaf = nil
	current = nil //nolint:ineffassign // intentional: clear stack reference to test GC behavior
	runtime.GC()
	runtime.GC()

	before := subCalls.Load()
	parent.Update(10)
	require.Equal(t, before+1, subCalls.Load(), "subscriber should still fire after dropping chain")

	unsub()

	assert.Eventually(t, func() bool {
		runtime.GC()
		before := callCounts[0].Load()
		parent.Update(int(before) + 100)
		return callCounts[0].Load() == before
	}, 5*time.Second, 10*time.Millisecond, "entire chain should clean up after unsubscribing from dropped leaf")
}

// TestMultipleSubscribersOnDerived verifies that cleanup waits for all
// subscribers to unsubscribe, not just the first.
func TestMultipleSubscribersOnDerived(t *testing.T) {
	parent := refreshable.New(1)
	var mapCalls atomic.Int64
	derived := refreshable.MapAuto(parent, func(v int) int {
		mapCalls.Add(1)
		return v * 10
	})

	var sub1Calls, sub2Calls atomic.Int64
	unsub1 := derived.Subscribe(func(int) { sub1Calls.Add(1) })
	unsub2 := derived.Subscribe(func(int) { sub2Calls.Add(1) })
	require.Equal(t, int64(1), sub1Calls.Load())
	require.Equal(t, int64(1), sub2Calls.Load())

	derived = nil
	runtime.GC()
	runtime.GC()

	before1 := sub1Calls.Load()
	before2 := sub2Calls.Load()
	parent.Update(5)
	require.Equal(t, before1+1, sub1Calls.Load(), "sub1 should still fire")
	require.Equal(t, before2+1, sub2Calls.Load(), "sub2 should still fire")

	// First unsub: cleanup should NOT fire yet (subCount=1).
	unsub1()
	mapBefore := mapCalls.Load()
	parent.Update(6)
	require.Greater(t, mapCalls.Load(), mapBefore, "map should still be called with one subscriber remaining")

	// Second unsub: cleanup fires (subCount=0).
	unsub2()

	assert.Eventually(t, func() bool {
		runtime.GC()
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "cleanup should fire after all subscribers unsubscribe")
}

// ---------------------------------------------------------------------------
// Idempotent unsub: explicit UnsubscribeFunc + later GC cleanup must not
// panic (double-unsub safety).
// ---------------------------------------------------------------------------

func TestExplicitUnsubThenGC(t *testing.T) {
	parent := refreshable.New(1)
	derived, stop := refreshable.Map(parent, func(v int) int { return v * 2 })
	require.Equal(t, 2, derived.Current())

	stop()
	before := derived.Current()
	parent.Update(99)
	require.Equal(t, before, derived.Current(), "no updates after explicit unsub")

	derived = nil
	runtime.GC()
	runtime.GC()
	// No panic means success.
	parent.Update(50)
	require.Equal(t, 50, parent.Current())
}

func TestExplicitUnsubThenGC_Validated(t *testing.T) {
	ctx := context.Background()
	parent := refreshable.New(10)
	validated, stop, err := refreshable.Validate(ctx, parent, func(_ context.Context, v int) error {
		if v < 0 {
			return errors.New("negative")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 10, validated.Unvalidated())

	stop()
	validated = nil
	runtime.GC()
	runtime.GC()
	parent.Update(20)
	require.Equal(t, 20, parent.Current())
}

// TestMapContextGCBeforeCancel verifies that GC cleanup before context
// cancellation does not panic when the goroutine later calls stop().
func TestMapContextGCBeforeCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parent := refreshable.New(1)

	var mapCalls atomic.Int64
	mapped := refreshable.MapContext(ctx, parent, func(v int) int {
		mapCalls.Add(1)
		return v * 2
	})
	require.Equal(t, 2, mapped.Current())

	mapped = nil
	runtime.GC()
	runtime.GC()

	assert.Eventually(t, func() bool {
		before := mapCalls.Load()
		parent.Update(int(before) + 100)
		return mapCalls.Load() == before
	}, time.Second, 10*time.Millisecond, "MapContext subscription should be cleaned up after GC")

	cancel()
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

	parent.Update(999)
	require.Equal(t, 999, parent.Current())
}

// ---------------------------------------------------------------------------
// Concurrency stress test.
// ---------------------------------------------------------------------------

func TestConcurrentCreateDropDerived(t *testing.T) {
	parent := refreshable.New(0)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Go(func() {
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				parent.Update(i)
			}
		}
	})

	for range 20 {
		wg.Go(func() {
			for range 100 {
				derived := refreshable.MapAuto(parent, func(v int) int { return v * 2 })
				_ = derived.Current()
				runtime.Gosched()
			}
		})
	}

	wg.Go(func() {
		for range 20 {
			runtime.GC()
			time.Sleep(time.Millisecond)
		}
	})

	done := make(chan struct{})
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
		close(done)
	}()
	<-done
	wg.Wait()

	parent.Update(42)
	require.Equal(t, 42, parent.Current())
}

// ---------------------------------------------------------------------------
// Subscriber count helpers for GC cleanup verification via reflection.
// ---------------------------------------------------------------------------

// updatableSubscriberCount uses reflection to read the length of the internal
// subscribers slice on a defaultRefreshable. This provides a direct measure of
// subscription accumulation without triggering subscriber callbacks.
// The updatable parameter is passed as any to support different generic
// instantiations of Updatable.
func updatableSubscriberCount(t *testing.T, updatable any) int {
	t.Helper()
	v := reflect.ValueOf(updatable)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	subs := v.FieldByName("subscribers")
	require.True(t, subs.IsValid(), "could not find subscribers field via reflection")
	return subs.Len()
}

// forceGCAndCleanup runs multiple GC cycles with pauses to allow
// runtime.AddCleanup callbacks to execute. The cleanup chain has multiple
// levels, each requiring a separate GC cycle followed by async cleanup
// completion before the next level becomes unreachable.
func forceGCAndCleanup() {
	for range 20 {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// Simple cleanup: MapWithError + MapFromValidated fan-out, drop all, verify
// config subscriber count returns to zero.
// ---------------------------------------------------------------------------

// TestMapWithErrorMapFromValidatedFanOutSubscriberCleanup verifies the basic
// cleanup chain: MapWithError creates an intermediate validatedFromRefreshable
// that subscribes to the root config, and MapFromValidated creates derived
// refreshables that subscribe to the validated result. Dropping all derived
// references should cascade cleanup back to the root.
func TestMapWithErrorMapFromValidatedFanOutSubscriberCleanup(t *testing.T) {
	ctx := context.Background()
	config := refreshable.New(42)

	baselineSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Baseline subscriber count: %d", baselineSubscribers)

	// MapWithError: config -> validatedFromRefreshable (subscribes to config) -> validParams
	validParams, err := refreshable.MapWithErrorAuto(ctx, config, func(_ context.Context, v int) (int, error) {
		return v * 2, nil
	})
	require.NoError(t, err)
	require.Equal(t, 84, validParams.Unvalidated())

	// Fan out: 3 MapFromValidated derived refreshables
	r1 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 1 })
	r2 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 2 })
	r3 := refreshable.MapFromValidatedAuto(validParams, func(v int) int { return v + 3 })

	require.Equal(t, 85, r1.Current())
	require.Equal(t, 86, r2.Current())
	require.Equal(t, 87, r3.Current())

	afterCreateSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Subscriber count after creating chain: %d", afterCreateSubscribers)
	require.Greater(t, afterCreateSubscribers, baselineSubscribers,
		"creating the chain should add subscribers to config")

	// Drop everything
	r1 = nil          //nolint:ineffassign
	r2 = nil          //nolint:ineffassign
	r3 = nil          //nolint:ineffassign
	validParams = nil //nolint:ineffassign

	forceGCAndCleanup()

	finalSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Subscriber count after GC cleanup: %d", finalSubscribers)

	assert.Equal(t, baselineSubscribers, finalSubscribers,
		"subscriber count should return to baseline after dropping all derived refreshables; "+
			"got growth of %d", finalSubscribers-baselineSubscribers)
}

// ---------------------------------------------------------------------------
// httpclient-like pattern: full NewClientFromRefreshableConfig simulation.
// ---------------------------------------------------------------------------

// TestHTTPClientLikePatternSubscriberCleanup mimics the refreshable chain
// created by httpclient.NewClientFromRefreshableConfig:
//
//  1. Root Updatable `config` (like the ClientConfig refreshable)
//  2. MapWithError(config, validateFn) -> validParams (like validated client params)
//  3. Multiple MapFromValidated(validParams, fn) calls (~10 derived refreshables,
//     like ServiceName, DialerParams, TransportParams, MetricsConfig, etc.)
//  4. One derived refreshable used in Map to create a "transport-like" chain
//     that captures several other derived refreshables (simulating the HTTP
//     transport that holds ServiceName, DisableMetrics, metricsTags)
//  5. Another derived refreshable used in Map for a "uri scorer" (simulating
//     the URI scorer that previously captured the builder)
//  6. A "client" struct holding some of these refreshables plus the Map results
//  7. Drop the client, GC, verify config subscriber count returns to 0
//
// This is the pattern that caused unbounded memory growth in production when
// refreshable subscriptions were not cleaned up after the HTTP client was
// garbage collected.
func TestHTTPClientLikePatternSubscriberCleanup(t *testing.T) {
	ctx := context.Background()

	// clientConfig simulates httpclient.ClientConfig
	type clientConfig struct {
		ServiceName    string
		URIs           []string
		DialTimeout    int
		DisableMetrics bool
		MaxRetries     int
	}

	config := refreshable.New(clientConfig{
		ServiceName:    "my-service",
		URIs:           []string{"https://localhost:8080"},
		DialTimeout:    30,
		DisableMetrics: false,
		MaxRetries:     3,
	})

	baselineSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Baseline subscriber count: %d", baselineSubscribers)

	// Simulate creating a client from a refreshable config, like
	// NewClientFromRefreshableConfig does.
	createClient := func() any {
		// Step 2: MapWithError validates config and produces validated params.
		// In the real code this is: validParams, _, _ := MapWithError(ctx, config, validateFn)
		validParams, err := refreshable.MapWithErrorAuto(ctx, config, func(_ context.Context, c clientConfig) (clientConfig, error) {
			if c.ServiceName == "" {
				return clientConfig{}, errors.New("service name required")
			}
			return c, nil
		})
		require.NoError(t, err)

		// Step 3: Fan out MapFromValidated for various config fields.
		// In the real code these are things like:
		//   b.ServiceName = MapFromValidated(validParams, func(p) string { return p.ServiceName })
		//   b.DialerParams = MapFromValidated(validParams, func(p) DialerParams { ... })
		//   etc.
		serviceName := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) string {
			return c.ServiceName
		})
		uris := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) []string {
			return c.URIs
		})
		dialTimeout := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.DialTimeout
		})
		disableMetrics := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) bool {
			return c.DisableMetrics
		})
		maxRetries := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.MaxRetries
		})
		// Additional derived refreshables to simulate the ~10 that the real code creates
		metricsTags := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) string {
			return fmt.Sprintf("service:%s", c.ServiceName)
		})
		dialerConfig := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.DialTimeout * 2
		})
		idleTimeout := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.DialTimeout * 3
		})
		tlsHandshakeTimeout := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.DialTimeout / 2
		})
		expectContinueTimeout := refreshable.MapFromValidatedAuto(validParams, func(c clientConfig) int {
			return c.DialTimeout / 3
		})

		// Step 4: "transport-like" chain. In the real httpclient code, the
		// transport captures several of the above refreshables. Here we simulate
		// that with a Map that reads from multiple captured refreshables.
		transportRefreshable := refreshable.MapAuto(dialerConfig, func(dc int) string {
			// In the real code, the transport builder closure captures
			// serviceName, disableMetrics, metricsTags, etc.
			return fmt.Sprintf("transport(dialer=%d, service=%s, metrics=%v, tags=%s)",
				dc, serviceName.Current(), disableMetrics.Current(), metricsTags.Current())
		})

		// Step 5: "uri scorer" chain. In the real code, the URI scorer was
		// previously capturing the builder which held all refreshables.
		uriScorerRefreshable := refreshable.MapAuto(uris, func(u []string) string {
			return fmt.Sprintf("scorer(%v, retries=%d)", u, maxRetries.Current())
		})

		// Step 6: Assemble the "client" struct.
		type client struct {
			ServiceName           refreshable.Refreshable[string]
			URIs                  refreshable.Refreshable[[]string]
			DialTimeout           refreshable.Refreshable[int]
			DisableMetrics        refreshable.Refreshable[bool]
			MaxRetries            refreshable.Refreshable[int]
			IdleTimeout           refreshable.Refreshable[int]
			TLSHandshakeTimeout   refreshable.Refreshable[int]
			ExpectContinueTimeout refreshable.Refreshable[int]
			Transport             refreshable.Refreshable[string]
			URIScorer             refreshable.Refreshable[string]
		}
		return &client{
			ServiceName:           serviceName,
			URIs:                  uris,
			DialTimeout:           dialTimeout,
			DisableMetrics:        disableMetrics,
			MaxRetries:            maxRetries,
			IdleTimeout:           idleTimeout,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ExpectContinueTimeout: expectContinueTimeout,
			Transport:             transportRefreshable,
			URIScorer:             uriScorerRefreshable,
		}
	}

	// Create and discard a warmup client, similar to the apollo-unbundler test.
	_ = createClient()
	forceGCAndCleanup()
	warmupSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Subscriber count after warmup + GC: %d", warmupSubscribers)

	// Create and discard multiple clients, simulating repeated calls to
	// serviceToImportClient without caching.
	const iterations = 50
	for range iterations {
		_ = createClient()
	}

	afterCreateSubscribers := updatableSubscriberCount(t, config)
	t.Logf("Subscriber count after creating %d clients (before GC): %d", iterations, afterCreateSubscribers)

	forceGCAndCleanup()

	finalSubscribers := updatableSubscriberCount(t, config)
	subscriberGrowth := finalSubscribers - baselineSubscribers
	t.Logf("Final subscriber count: %d (growth from baseline: %d)", finalSubscribers, subscriberGrowth)

	// After GC cleanup, the subscriber count on the root config should return
	// to baseline. Each NewClientFromRefreshableConfig-like call adds 1
	// subscriber to config via validatedFromRefreshable. Without cleanup
	// propagation, this would grow by 1 per client (i.e. growth == iterations).
	assert.Less(t, subscriberGrowth, 5,
		"subscriber count grew by %d after discarding %d clients; "+
			"the cleanup chain from derived refreshables back to the root config is not working",
		subscriberGrowth, iterations)
}
