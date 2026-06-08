// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"strings"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/safelong"
	"github.com/stretchr/testify/require"
)

const (
	minSafeLong int64 = -9007199254740991
	maxSafeLong int64 = 9007199254740991
)

func TestSafeLongValues(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: 0, JSON: "0"},
		},
		{
			Name: "max",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: maxSafeLong, JSON: "9007199254740991"},
		},
		{
			Name: "min",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), Value: minSafeLong, JSON: "-9007199254740991"},
		},
		{
			Name: "named_type",
			Test: typeTestCase[safelong.SafeLong]{Codec: cj.SafeLong[safelong.SafeLong](), Value: safelong.SafeLong(maxSafeLong), JSON: "9007199254740991"},
		},
		{
			Name: "string_rejected",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), JSON: `"42"`, SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"42\" into Go int64 after offset 4: want json int"},
		},
		{
			Name: "too_large_unmarshal",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), JSON: "9007199254740992", SkipTestMarshal: true,
				ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON number 9007199254740992 into Go int64 after offset 16: invalid safelong: 9007199254740992 is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991",
			},
		},
		{
			Name: "too_small_unmarshal",
			Test: typeTestCase[int64]{Codec: cj.SafeLong[int64](), JSON: "-9007199254740992", SkipTestMarshal: true,
				ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON number -9007199254740992 into Go int64 after offset 17: invalid safelong: -9007199254740992 is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991",
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

func TestSafeLongMaps(t *testing.T) {
	type SL = safelong.SafeLong
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "map_value",
			Test: typeTestCase[map[string]SL]{Codec: cj.OrderedMap[map[string]SL](cj.String[string](), cj.SafeLong[SL]()), Value: map[string]SL{"a": 42, "b": -42}, JSON: `{"a":42,"b":-42}`},
		},
		{
			Name: "map_key",
			Test: typeTestCase[map[SL]int]{Codec: cj.OrderedMap[map[SL]int](cj.SafeLongMapKey[SL](), cj.Int32[int]()), Value: map[SL]int{100: 1, -200: -2}, JSON: `{"-200":-2,"100":1}`},
		},
		{
			Name: "map_key_rejects_out_of_range",
			Test: typeTestCase[map[SL]int]{Codec: cj.OrderedMap[map[SL]int](cj.SafeLongMapKey[SL](), cj.Int32[int]()), JSON: `{"9007199254740992":1}`, SkipTestMarshal: true,
				ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"9007199254740992\" into Go map[safelong.SafeLong]int within \"/9007199254740992\": invalid safelong: 9007199254740992 is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991",
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

func TestSafeLongMapKeyRequiresString(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`42`), cj.DefaultOptions)
	var got int64

	err := cj.SafeLongMapKey[int64]().UnmarshalJSONFrom(dec, &got)

	require.ErrorContains(t, err, "want json string")
}
