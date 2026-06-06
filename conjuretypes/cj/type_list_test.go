// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestList(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[[]int]{Codec: cj.List[[]int](cj.Int32[int]()), Value: []int{}, JSON: "[]"},
		},
		{
			Name: "one",
			Test: typeTestCase[[]int]{Codec: cj.List[[]int](cj.Int32[int]()), Value: []int{42}, JSON: "[42]"},
		},
		{
			Name: "several",
			Test: typeTestCase[[]string]{Codec: cj.List[[]string](cj.String[string]()), Value: []string{"a", "b", "c"}, JSON: "[\"a\",\"b\",\"c\"]"},
		},
		{
			Name: "nested",
			Test: typeTestCase[[][]bool]{Codec: cj.List[[][]bool](cj.List[[]bool](cj.Boolean[bool]())), Value: [][]bool{{true, false}, {}}, JSON: "[[true,false],[]]"},
		},
		{
			Name: "null_marshal",
			Test: typeTestCase[[]int]{Codec: cj.List[[]int](cj.Int32[int]()), JSON: "[]", SkipTestUnmarshal: true, Value: []int(nil)},
		},
		{
			Name: "null_unmarshal",
			Test: typeTestCase[[]int]{Codec: cj.List[[]int](cj.Int32[int]()), JSON: "null", SkipTestMarshal: true, Value: []int{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
