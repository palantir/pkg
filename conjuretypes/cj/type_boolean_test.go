// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	type boolAlias bool
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "true",
			Test: typeTestCase[bool]{Codec: cj.Boolean[bool](), Value: true, JSON: "true"},
		},
		{
			Name: "boolAlias(true)",
			Test: typeTestCase[boolAlias]{Codec: cj.Boolean[boolAlias](), Value: true, JSON: "true"},
		},
		{
			Name: "false",
			Test: typeTestCase[bool]{Codec: cj.Boolean[bool](), Value: false, JSON: "false"},
		},
		{
			Name: "boolAlias(false)",
			Test: typeTestCase[boolAlias]{Codec: cj.Boolean[boolAlias](), Value: false, JSON: "false"},
		},
		{
			Name: "null",
			Test: typeTestCase[bool]{Codec: cj.Boolean[bool](), JSON: "null",
				SkipTestMarshal:      true,
				ErrUnmarshalJSONFrom: "KindMismatchError at offset 4: want json boolean, got null",
			},
		},
		{
			Name: "map_keys",
			Test: typeTestCase[map[bool]bool]{Codec: cj.ComparableMap[map[bool]bool](cj.BooleanMapKey[bool](), cj.Boolean[bool]()), Value: map[bool]bool{true: true, false: false},
				JSON: "{\"false\":false,\"true\":true}",
			},
		},
		{
			Name: "map_key_invalid",
			Test: typeTestCase[bool]{Codec: cj.BooleanMapKey[bool](), JSON: `"invalid"`,
				SkipTestMarshal:      true,
				ErrUnmarshalJSONFrom: "InvalidValueError at offset 9: invalid boolean: strconv.ParseBool: parsing \"invalid\": invalid syntax",
			},
		},
		{
			Name: "map_key_not_string",
			Test: typeTestCase[bool]{Codec: cj.BooleanMapKey[bool](), JSON: `true`,
				SkipTestMarshal:      true,
				ErrUnmarshalJSONFrom: "KindMismatchError at offset 4: want json string, got true",
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

func TestBooleanMapKeyCompare(t *testing.T) {
	encoder := cj.BooleanMapKey[bool]()

	tests := []struct {
		name     string
		a, b     bool
		expected int
	}{
		{"both false", false, false, 0},
		{"both true", true, true, 0},
		{"false less than true", false, true, -1},
		{"true greater than false", true, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Compare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
