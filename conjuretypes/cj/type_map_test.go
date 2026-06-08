// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"cmp"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/datetime"
)

func TestMap(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), Value: map[string]int{}, JSON: "{}"},
		},
		{
			Name: "one",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), Value: map[string]int{"foo": 1}, JSON: "{\"foo\":1}"},
		},
		{
			Name: "ordered",
			Test: typeTestCase[map[string]string]{Codec: cj.Map[map[string]string](cj.String[string](), cj.String[string]()), Value: map[string]string{"j": "10", "i": "9", "h": "8", "g": "7", "f": "6", "e": "5", "d": "4", "c": "3", "b": "2", "a": "1"}, JSON: "{\"a\":\"1\",\"b\":\"2\",\"c\":\"3\",\"d\":\"4\",\"e\":\"5\",\"f\":\"6\",\"g\":\"7\",\"h\":\"8\",\"i\":\"9\",\"j\":\"10\"}"},
		},
		{
			Name: "ordered_int_keys",
			Test: typeTestCase[map[int]int]{Codec: cj.Map[map[int]int](cj.Int32MapKey[int](), cj.Int32[int]()), Value: map[int]int{100: 100, 10: 10, 9: 9, 1: 1, 0: 0, -1: -1}, JSON: "{\"-1\":-1,\"0\":0,\"1\":1,\"9\":9,\"10\":10,\"100\":100}"},
		},
		{
			Name: "ordered_float_keys",
			Test: typeTestCase[map[float64]float64]{Codec: cj.Map[map[float64]float64](cj.FloatMapKey[float64](), cj.Float[float64]()), Value: map[float64]float64{100: 100, 10: 10, 9: 9, 1: 1, 0: 0, -1: -1, -0.10: -0.10, -0.9: -0.9, math.Inf(1): math.Inf(1), math.Inf(-1): math.Inf(-1)},
				JSON: "{\"-Infinity\":\"-Infinity\",\"-1\":-1,\"-0.9\":-0.9,\"-0.1\":-0.1,\"0\":0,\"1\":1,\"9\":9,\"10\":10,\"100\":100,\"Infinity\":\"Infinity\"}",
			},
		},
		{
			Name: "nested",
			Test: typeTestCase[map[string][]int]{Codec: cj.Map[map[string][]int](cj.String[string](), cj.List[[]int](cj.Int32[int]())), Value: map[string][]int{"nums": {1, 2, 3}}, JSON: "{\"nums\":[1,2,3]}"},
		},
		{
			// Regression: the map decoder reuses one value variable across
			// entries, so a reference-typed value (here a slice) must be reset
			// between entries or every key would alias the last decoded slice.
			Name: "nested multiple entries",
			Test: typeTestCase[map[string][]int]{Codec: cj.Map[map[string][]int](cj.String[string](), cj.List[[]int](cj.Int32[int]())), Value: map[string][]int{"a": {1, 2, 3}, "b": {4, 5, 6}, "c": {7, 8, 9}}, JSON: "{\"a\":[1,2,3],\"b\":[4,5,6],\"c\":[7,8,9]}"},
		},
		{
			Name: "boolean map key",
			Test: typeTestCase[map[bool]int]{Codec: cj.Map[map[bool]int](cj.BooleanMapKey[bool](), cj.Int32[int]()), Value: map[bool]int{true: 2, false: 2}, JSON: "{\"false\":2,\"true\":2}"},
		},
		{
			Name: "datetime map key",
			Test: typeTestCase[map[datetime.DateTime]string]{Codec: cj.Map[map[datetime.DateTime]string](cj.DateTime[datetime.DateTime](), cj.String[string]()), Value: map[datetime.DateTime]string{datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)): "2024-01-01T00:00:00Z", datetime.DateTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)): "2025-01-01T00:00:00Z"}, JSON: "{\"2024-01-01T00:00:00Z\":\"2024-01-01T00:00:00Z\",\"2025-01-01T00:00:00Z\":\"2025-01-01T00:00:00Z\"}"},
		},
		{
			// Two keys that denote the same instant in different zones serialize to
			// different object names; DateTime sorts by instant and breaks the tie by
			// that wire string, so the output is deterministic. (Marshal-only: parsing
			// "+01:00" back yields a different time.Time zone than the FixedZone written
			// here.)
			Name: "datetime map key same instant different zone",
			Test: typeTestCase[map[datetime.DateTime]int]{Codec: cj.Map[map[datetime.DateTime]int](cj.DateTime[datetime.DateTime](), cj.Int32[int]()), Value: map[datetime.DateTime]int{
				datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)):                 1,
				datetime.DateTime(time.Date(2024, 1, 1, 1, 0, 0, 0, time.FixedZone("", 3600))): 2,
			}, JSON: "{\"2024-01-01T00:00:00Z\":1,\"2024-01-01T01:00:00+01:00\":2}", SkipTestUnmarshal: true},
		},
		{
			Name: "null_marshal",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "{}", SkipTestUnmarshal: true, Value: map[string]int(nil)},
		},
		{
			Name: "null_unmarshal",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "null", SkipTestMarshal: true, Value: map[string]int{}},
		},
		{
			Name: "not an object",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "[]", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON array into Go map[string]int after offset 1: want object opening brace"},
		},
		{
			Name: "duplicate key",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "{\"a\":1,\"a\":2}", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "jsontext: duplicate object member name within \"/a\" after offset 7"},
		},
		{
			Name: "duplicate int key",
			Test: typeTestCase[map[int]int]{Codec: cj.Map[map[int]int](cj.Int32MapKey[int](), cj.Int32[int]()), JSON: "{\"01\":1,\"1\":2}", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal into Go map[int]int within \"/1\": duplicate map key"},
		},
		{
			Name: "key not string",
			Test: typeTestCase[map[string]int]{Codec: cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "{ 1:2 }", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "jsontext: object member name must be a string after offset 2"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", func(t *testing.T) {
				tc.Test.TestMarshal(t)
			})
			t.Run("Unmarshal", func(t *testing.T) {
				tc.Test.TestUnmarshal(t)
			})
		})
	}
}

// BenchmarkSort shows why MapKeyCodec.Sort specializes per key type rather than
// routing every key through a comparator: cmp.Ordered keys sort faster via
// slices.Sort than via slices.SortFunc. That gap is what lets a single map codec
// stay fast on ordered keys (see orderedKeyCodec).
func BenchmarkSort(b *testing.B) {
	const size = 1000
	data := make(map[int64]struct{}, size)
	for range size {
		data[rand.Int64()] = struct{}{}
	}
	b.Run("ordered loop", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sorted := make([]int64, 0, 1000)
			for k := range data {
				sorted = append(sorted, k)
			}
			slices.Sort(sorted)
			_ = sorted
		}
	})
	b.Run("compare loop", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sorted := make([]int64, 0, 1000)
			for k := range data {
				sorted = append(sorted, k)
			}
			slices.SortFunc(sorted, cmp.Compare)
			_ = sorted
		}
	})
	b.Run("ordered stream", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sorted := slices.Sorted(maps.Keys(data))
			_ = sorted
		}
	})
	b.Run("compare stream", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sorted := slices.SortedFunc(maps.Keys(data), cmp.Compare)
			_ = sorted
		}
	})
}
