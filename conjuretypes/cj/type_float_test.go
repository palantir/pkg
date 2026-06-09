// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"math"
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestFloat(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: 0.0, JSON: "0"},
		},
		{
			Name: "positive",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: 123.456, JSON: "123.456"},
		},
		{
			Name: "negative",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: -42.5, JSON: "-42.5"},
		},
		{
			Name: "large",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: 1e30, JSON: "1e+30"},
		},
		{
			Name: "small",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: 1e-18, JSON: "1e-18"},
		},
		{
			Name: "nan",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: math.NaN(), JSON: "\"NaN\""},
		},
		{
			Name: "+inf",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: math.Inf(1), JSON: "\"Infinity\""},
		},
		{
			Name: "-inf",
			Test: typeTestCase[float64]{Codec: cj.Float[float64](), Value: math.Inf(-1), JSON: "\"-Infinity\""},
		},
		{
			Name: "map",
			Test: typeTestCase[map[float64]float64]{Codec: cj.Map[map[float64]float64](cj.FloatMapKey[float64](), cj.Float[float64]()), Value: map[float64]float64{0.0: 0.0, 123.456: 123.456, -42.5: -42.5, 1e30: 1e30, 1e-18: 1e-18, math.Inf(1): math.Inf(1), math.Inf(-1): math.Inf(-1)},
				JSON: `{"-Infinity":"-Infinity","-42.5":-42.5,"0":0,"1e-18":1e-18,"123.456":123.456,"1e+30":1e+30,"Infinity":"Infinity"}`,
			},
		},
		{
			// Conjure permits NaN as a double map key; see receiveMapDoubleAliasExample
			// in the conjure verifier.
			Name: "nan_as_map_key",
			Test: typeTestCase[float64]{Codec: cj.FloatMapKey[float64](), Value: math.NaN(), JSON: `"NaN"`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
