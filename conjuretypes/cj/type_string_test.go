// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestString(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[string]{Codec: cj.String[string](), Value: "", JSON: "\"\""},
		},
		{
			Name: "ascii",
			Test: typeTestCase[string]{Codec: cj.String[string](), Value: "hello", JSON: "\"hello\""},
		},
		{
			Name: "unicode",
			Test: typeTestCase[string]{Codec: cj.String[string](), Value: "héllo 世界", JSON: "\"héllo 世界\""},
		},
		{
			Name: "escaped",
			Test: typeTestCase[string]{Codec: cj.String[string](), Value: "foo\nbar", JSON: "\"foo\\nbar\""},
		},
		{
			Name: "null",
			Test: typeTestCase[string]{Codec: cj.String[string](), JSON: "null", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "KindMismatchError at offset 4: want json string, got null"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

type textString string

func (t textString) MarshalText() ([]byte, error) {
	return []byte(t), nil
}

func (t *textString) UnmarshalText(text []byte) error {
	*t = textString(text)
	return nil
}
