// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestOptional(t *testing.T) {
	type optStr = *string
	type optInt = *int

	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "nil string",
			Test: typeTestCase[optStr]{Codec: cj.Optional[optStr](cj.String[string]()), Value: nil, JSON: "null"},
		},
		{
			Name: "some string",
			Test: typeTestCase[optStr]{Codec: cj.Optional[optStr](cj.String[string]()), Value: new("foo"), JSON: "\"foo\""},
		},
		{
			Name: "nil int",
			Test: typeTestCase[optInt]{Codec: cj.Optional[optInt](cj.Int32[int]()), Value: nil, JSON: "null"},
		},
		{
			Name: "some int",
			Test: typeTestCase[optInt]{Codec: cj.Optional[optInt](cj.Int32[int]()), Value: new(123), JSON: "123"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
