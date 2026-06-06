// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/safelong"
)

func TestInt(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero",
			Test: typeTestCase[int]{Codec: cj.Int32[int](), Value: 0, JSON: "0"},
		},
		{
			Name: "positive",
			Test: typeTestCase[int]{Codec: cj.Int32[int](), Value: 42, JSON: "42"},
		},
		{
			Name: "negative",
			Test: typeTestCase[int]{Codec: cj.Int32[int](), Value: -7, JSON: "-7"},
		},
		{
			Name: "invalid",
			Test: typeTestCase[int]{Codec: cj.Int32[int](), JSON: "\"invalid\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "KindMismatchError at offset 9: want json int, got string"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestSafeLong(t *testing.T) {
	const maxSafe = 9007199254740991
	const minSafe = -9007199254740991
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: 0, JSON: "0"},
		},
		{
			Name: "positive",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: maxSafe, JSON: "9007199254740991"},
		},
		{
			Name: "negative",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: minSafe, JSON: "-9007199254740991"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestMapOfInt(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[map[string]int]{Codec: cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), Value: map[string]int{}, JSON: "{}"},
		},
		{
			Name: "simple",
			Test: typeTestCase[map[string]int]{Codec: cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), Value: map[string]int{"a": 1, "b": -2}, JSON: "{\"a\":1,\"b\":-2}"},
		},
		{
			Name: "null",
			Test: typeTestCase[map[string]int]{Codec: cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), JSON: "null", SkipTestMarshal: true, Value: map[string]int{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestMapOfSafeLong(t *testing.T) {
	type SL = safelong.SafeLong
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[map[string]SL]{Codec: cj.OrderedMap[map[string]SL](cj.String[string](), cj.SafeLong[SL]()), Value: map[string]SL{}, JSON: "{}"},
		},
		{
			Name: "simple",
			Test: typeTestCase[map[string]SL]{Codec: cj.OrderedMap[map[string]SL](cj.String[string](), cj.SafeLong[SL]()), Value: map[string]SL{"a": 42, "b": -42}, JSON: "{\"a\":42,\"b\":-42}"},
		},
		{
			Name: "null",
			Test: typeTestCase[map[string]SL]{Codec: cj.OrderedMap[map[string]SL](cj.String[string](), cj.SafeLong[SL]()), JSON: "null", SkipTestMarshal: true, Value: map[string]SL{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestMapIntKeySafeLongValue(t *testing.T) {
	type SL = safelong.SafeLong
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "simple",
			Test: typeTestCase[map[int]SL]{Codec: cj.OrderedMap[map[int]SL](cj.Int32MapKey[int](), cj.SafeLong[SL]()), Value: map[int]SL{1: 100, -2: -200}, JSON: "{\"-2\":-200,\"1\":100}"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestMapSafeLongKeyIntValue(t *testing.T) {
	type SL = safelong.SafeLong
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "simple",
			Test: typeTestCase[map[SL]int]{Codec: cj.OrderedMap[map[SL]int](cj.SafeLongMapKey[SL](), cj.Int32[int]()), Value: map[SL]int{100: 1, -200: -2}, JSON: "{\"-200\":-2,\"100\":1}"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
