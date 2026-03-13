// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"testing"
	"time"

	refreshable "github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
)

// equalLenString implements selfEqual for use with NewEqualMethod.
// Its Equal method compares string lengths rather than contents, so two
// strings of the same length are "equal" even if they differ—making the
// behavior clearly distinct from reflect.DeepEqual.
type equalLenString struct{ val string }

func (e equalLenString) Equal(other equalLenString) bool { return len(e.val) == len(other.val) }

// testUpdatable verifies debouncing: updating with an equal value should not
// notify subscribers, while updating with a different value should.
func testUpdatable[T any](t *testing.T, r refreshable.Updatable[T], same, different T) {
	t.Helper()
	updates := 0
	r.Subscribe(func(T) { updates++ })
	assert.Equal(t, 1, updates, "subscribe should fire immediately")

	r.Update(same)
	assert.Equal(t, 1, updates, "equal value should be debounced")

	r.Update(different)
	assert.Equal(t, 2, updates, "different value should notify")
}

func TestNewComparable(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		r := refreshable.NewComparable("hello")
		assert.Equal(t, "hello", r.Current())
		testUpdatable(t, r, "hello", "world")
	})
	t.Run("int", func(t *testing.T) {
		r := refreshable.NewComparable(42)
		assert.Equal(t, 42, r.Current())
		testUpdatable(t, r, 42, 99)
	})
	t.Run("bool", func(t *testing.T) {
		testUpdatable(t, refreshable.NewComparable(true), true, false)
	})
	t.Run("struct", func(t *testing.T) {
		type kv struct{ K, V string }
		testUpdatable(t, refreshable.NewComparable(kv{"a", "b"}), kv{"a", "b"}, kv{"c", "d"})
	})
}

func TestNewComparableMap(t *testing.T) {
	r := refreshable.NewComparableMap(map[string]int{"a": 1})
	assert.Equal(t, map[string]int{"a": 1}, r.Current())
	testUpdatable(t, r, map[string]int{"a": 1}, map[string]int{"b": 2})
}

func TestNewComparableSlice(t *testing.T) {
	r := refreshable.NewComparableSlice([]string{"a", "b"})
	assert.Equal(t, []string{"a", "b"}, r.Current())
	testUpdatable(t, r, []string{"a", "b"}, []string{"c"})
}

func TestNewEqualMethod(t *testing.T) {
	t.Run("time.Time", func(t *testing.T) {
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		// time.Equal treats the same instant in different zones as equal.
		sameInstant := now.In(time.FixedZone("UTC+1", 3600))
		testUpdatable(t, refreshable.NewEqualMethod(now), sameInstant, now.Add(time.Second))
	})
	t.Run("custom", func(t *testing.T) {
		// "hi" and "ab" have the same length (Equal returns true), but "bye" has a different length.
		testUpdatable(t, refreshable.NewEqualMethod(equalLenString{"hi"}), equalLenString{"ab"}, equalLenString{"bye"})
	})
}

func TestNewEqualFunc(t *testing.T) {
	// NewEqualFunc works with any type given a custom equality function.
	type point struct{ X, Y int }
	r := refreshable.NewEqualFunc(point{1, 2}, func(a, b point) bool { return a == b })
	assert.Equal(t, point{1, 2}, r.Current())
	testUpdatable(t, r, point{1, 2}, point{3, 4})
}

func TestNewEqualMethodMap(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sameInstant := now.In(time.FixedZone("UTC+1", 3600))
	r := refreshable.NewEqualMethodMap(map[string]time.Time{"t": now})
	testUpdatable(t, r, map[string]time.Time{"t": sameInstant}, map[string]time.Time{"t": now.Add(time.Hour)})
}

func TestNewEqualMethodSlice(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sameInstant := now.In(time.FixedZone("UTC+1", 3600))
	r := refreshable.NewEqualMethodSlice([]time.Time{now})
	testUpdatable(t, r, []time.Time{sameInstant}, []time.Time{now.Add(time.Hour)})
}

func TestNewBytes(t *testing.T) {
	r := refreshable.NewBytes([]byte("hello"))
	assert.Equal(t, []byte("hello"), r.Current())
	testUpdatable(t, r, []byte("hello"), []byte("world"))
}

func TestNewBytes_NamedType(t *testing.T) {
	type blob []byte
	r := refreshable.NewBytes(blob("data"))
	assert.Equal(t, blob("data"), r.Current())
	testUpdatable(t, r, blob("data"), blob("other"))
}

func TestCacheWith(t *testing.T) {
	t.Run("propagates values from source", func(t *testing.T) {
		source := refreshable.NewComparable("hello")
		cached := refreshable.CacheWith[string](source, refreshable.NewComparable)
		assert.Equal(t, "hello", cached.Current())

		source.Update("world")
		assert.Equal(t, "world", cached.Current())
	})

	t.Run("debounces with constructor equality", func(t *testing.T) {
		// Use NewEqualMethod so that time.Time.Equal is used for debouncing,
		// which treats the same instant in different zones as equal.
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		source := refreshable.New(now)
		sourceUpdates := 0
		source.Subscribe(func(t time.Time) { sourceUpdates++ })
		assert.Equal(t, 1, sourceUpdates, "subscribe should fire immediately")

		cached := refreshable.CacheWith[time.Time](source, refreshable.NewEqualMethod)
		cacheUpdates := 0
		cached.Subscribe(func(time.Time) { cacheUpdates++ })
		assert.Equal(t, 1, cacheUpdates, "subscribe should fire immediately")

		// Same instant in a different zone: time.Equal considers them equal.
		sameInstant := now.In(time.FixedZone("UTC+1", 3600))
		source.Update(sameInstant)
		assert.Equal(t, 1, cacheUpdates, "equal time should be debounced")
		assert.Equal(t, 2, sourceUpdates, "expected reflect-based source not to debounce equal time")

		source.Update(now.Add(time.Second))
		assert.Equal(t, 2, cacheUpdates, "different time should notify")
	})

	t.Run("debounces map with element equality", func(t *testing.T) {
		// Source uses reflect.DeepEqual, which compares time.Time zone pointers.
		// CacheWith uses NewEqualMethodMap, which compares values with time.Time.Equal.
		// An update with the same instant in a different zone passes through the
		// source (not DeepEqual) but is debounced by the cached refreshable.
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		source := refreshable.New(map[string]time.Time{"t": now})
		sourceUpdates := 0
		source.Subscribe(func(t map[string]time.Time) { sourceUpdates++ })
		cached := refreshable.CacheWith[map[string]time.Time](source, refreshable.NewEqualMethodMap)
		assert.Equal(t, map[string]time.Time{"t": now}, cached.Current())

		cacheUpdates := 0
		cached.Subscribe(func(map[string]time.Time) { cacheUpdates++ })
		assert.Equal(t, 1, cacheUpdates)

		sameInstant := now.In(time.FixedZone("UTC+1", 3600))
		source.Update(map[string]time.Time{"t": sameInstant})
		assert.Equal(t, 1, cacheUpdates, "same instant in different zone should be debounced by CacheWith")
		assert.Equal(t, 2, sourceUpdates, "expected reflect-based source not to debounce equal time")

		source.Update(map[string]time.Time{"t": now.Add(time.Hour)})
		assert.Equal(t, 2, cacheUpdates, "different time should notify")
	})
}
