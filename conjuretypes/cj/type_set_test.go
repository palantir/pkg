// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/datetime"
	"github.com/palantir/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{}, JSON: "[]"},
		},
		{
			Name: "one",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{42}, JSON: "[42]"},
		},
		{
			Name: "several",
			Test: typeTestCase[[]string]{Codec: cj.Set[[]string](cj.String[string]()), Value: []string{"a", "b", "c"}, JSON: "[\"a\",\"b\",\"c\"]"},
		},
		{
			Name: "null_marshal",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), JSON: "[]", SkipTestUnmarshal: true, Value: []int(nil)},
		},
		{
			Name: "null_unmarshal",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), JSON: "null", SkipTestMarshal: true, Value: []int{}},
		},
		{
			Name: "comparable",
			Test: typeTestCase[[]uuid.UUID]{Codec: cj.Set[[]uuid.UUID](cj.UUID[uuid.UUID]()), Value: []uuid.UUID{must(uuid.ParseUUID("10101010-1010-1010-1010-101010101010")), must(uuid.ParseUUID("20202020-2020-2020-2020-202020202020"))}, JSON: "[\"10101010-1010-1010-1010-101010101010\",\"20202020-2020-2020-2020-202020202020\"]"},
		},
		{
			Name: "marshal_ordered_dedupe",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{5, 4, 3, 2, 1, 2, 3, 4, 5}, JSON: "[5,4,3,2,1]", SkipTestUnmarshal: true},
		},
		{
			Name: "marshal_dupes",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{42, 42, 1}, JSON: "[42,1]", SkipTestUnmarshal: true},
		},
		{
			Name: "ordered",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{5, 4, 3, 2, 1}, JSON: "[5,4,3,2,1]"},
		},
		{
			Name: "unmarshal_dupes",
			Test: typeTestCase[[]int]{Codec: cj.Set[[]int](cj.Int32[int]()), Value: []int{1, 42}, JSON: "[42, 42, 1]", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal into Go []int within \"/1\": duplicate set item"},
		},
		{
			Name: "set_of_structs",
			Test: typeTestCase[[]testStruct]{Codec: cj.Set[[]testStruct](cj.Struct[testStruct]()), Value: []testStruct{{Name: "c", Age: 3}, {Name: "b", Age: 2}, {Name: "a", Age: 1}},
				JSON: "[{\"name\":\"c\",\"age\":3},{\"name\":\"b\",\"age\":2},{\"name\":\"a\",\"age\":1}]",
			},
		},
		{
			Name: "set_of_structs_dedupes",
			Test: typeTestCase[[]testStruct]{Codec: cj.Set[[]testStruct](cj.Struct[testStruct]()), Value: []testStruct{{Name: "c", Age: 3}, {Name: "b", Age: 2}, {Name: "a", Age: 1}, {Name: "b", Age: 2}, {Name: "b", Age: 2}},
				JSON:              "[{\"name\":\"c\",\"age\":3},{\"name\":\"b\",\"age\":2},{\"name\":\"a\",\"age\":1}]",
				SkipTestUnmarshal: true,
			},
		},
		{
			Name: "set_of_structs_errors",
			Test: typeTestCase[[]testStruct]{Codec: cj.Set[[]testStruct](cj.Struct[testStruct]()), JSON: "[{\"name\":\"c\",\"age\":3},{\"name\":\"b\",\"age\":2},{\"name\":\"a\",\"age\":1},{\"name\":\"b\",\"age\":2},{\"name\":\"b\",\"age\":2}]",
				ErrUnmarshalJSONFrom: "json: cannot unmarshal into Go []cj_test.testStruct within \"/3\": duplicate set item",
				SkipTestMarshal:      true,
			},
		},
		{
			Name: "datetimes",
			Test: typeTestCase[[]datetime.DateTime]{Codec: cj.Set[[]datetime.DateTime](cj.Text[datetime.DateTime]()), JSON: "[\"2025-05-12T19:26:00Z\",\"2001-01-01T19:26:00Z\",\"0001-01-01T00:00:00Z\"]",
				Value: []datetime.DateTime{must(datetime.ParseDateTime("2025-05-12T19:26:00Z")), must(datetime.ParseDateTime("2001-01-01T19:26:00Z")), must(datetime.ParseDateTime("0001-01-01T00:00:00Z"))},
			},
		},
		{
			Name: "datetimes_errors",
			Test: typeTestCase[[]datetime.DateTime]{Codec: cj.Set[[]datetime.DateTime](cj.Text[datetime.DateTime]()), JSON: "[\"2025-05-12T19:26:00Z\",\"2001-01-01T19:26:00Z\",\"0001-01-01T00:00:00Z\",\"0001-01-01T00:00:00.00Z\"]",
				Value:                []datetime.DateTime{must(datetime.ParseDateTime("2025-05-12T19:26:00Z")), must(datetime.ParseDateTime("2001-01-01T19:26:00Z")), must(datetime.ParseDateTime("0001-01-01T00:00:00Z"))},
				ErrUnmarshalJSONFrom: "json: cannot unmarshal into Go []datetime.DateTime within \"/3\": duplicate set item",
				SkipTestMarshal:      true,
			},
		},
		{
			Name: "datetimes_dedupes",
			Test: typeTestCase[[]datetime.DateTime]{Codec: cj.Set[[]datetime.DateTime](cj.Text[datetime.DateTime]()), JSON: "[\"2025-05-12T19:26:00Z\",\"2001-01-01T19:26:00Z\",\"0001-01-01T00:00:00Z\"]",
				Value: []datetime.DateTime{
					must(datetime.ParseDateTime("2025-05-12T19:26:00Z")),
					must(datetime.ParseDateTime("2001-01-01T19:26:00Z")),
					must(datetime.ParseDateTime("2001-01-01T19:26:00.000Z")),
					must(datetime.ParseDateTime("0001-01-01T00:00:00Z")),
					must(datetime.ParseDateTime("0001-01-01T00:00:00Z")),
				},
				SkipTestUnmarshal: true,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestSetEqual(t *testing.T) {
	// Set equality ignores order and multiplicity: two sets are equal iff they
	// contain the same distinct elements (consistent with marshal dropping duplicates).
	codec := cj.Set[[]int](cj.Int32[int]())
	for _, tc := range []struct {
		Name  string
		A, B  []int
		Equal bool
	}{
		{Name: "same elements same order", A: []int{1, 2, 3}, B: []int{1, 2, 3}, Equal: true},
		{Name: "same elements different order", A: []int{1, 2}, B: []int{2, 1}, Equal: true},
		{Name: "different elements", A: []int{1, 2}, B: []int{1, 3}, Equal: false},
		{Name: "differing multiplicities, same set", A: []int{1, 1, 2}, B: []int{1, 2, 2}, Equal: true},
		{Name: "duplicate vs distinct, same set", A: []int{1, 1, 2}, B: []int{1, 2}, Equal: true},
		{Name: "proper subset is not equal", A: []int{1, 2}, B: []int{1, 2, 3}, Equal: false},
		{Name: "empty sets", A: []int{}, B: []int{}, Equal: true},
		{Name: "empty vs non-empty", A: []int{}, B: []int{1}, Equal: false},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Equal, codec.Equal(tc.A, tc.B))
			assert.Equal(t, tc.Equal, codec.Equal(tc.B, tc.A), "Equal should be symmetric")
		})
	}
}
