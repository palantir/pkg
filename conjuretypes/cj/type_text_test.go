// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"fmt"
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "text",
			Test: typeTestCase[uuid.UUID]{Codec: cj.Text[uuid.UUID](), Value: must(uuid.ParseUUID("10101010-1010-1010-1010-101010101010")), JSON: "\"10101010-1010-1010-1010-101010101010\""},
		},
		{
			Name: "map",
			Test: typeTestCase[map[uuid.UUID]string]{Codec: cj.Map[map[uuid.UUID]string](cj.Text[uuid.UUID](), cj.String[string]()), Value: map[uuid.UUID]string{
				must(uuid.ParseUUID("00101010-1010-1010-1010-101010101010")): "foo",
				must(uuid.ParseUUID("00202020-2020-2020-2020-202020202020")): "bar",
			},
				JSON: `{"00101010-1010-1010-1010-101010101010":"foo","00202020-2020-2020-2020-202020202020":"bar"}`,
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

func TestTextCompare(t *testing.T) {
	encoder := cj.Text[textString]()

	tests := []struct {
		name     string
		a, b     textString
		expected int
	}{
		{"equal", textString("hello"), textString("hello"), 0},
		{"a less than b", textString("apple"), textString("banana"), -1},
		{"a greater than b", textString("zebra"), textString("apple"), 1},
		{"empty strings", textString(""), textString(""), 0},
		{"empty vs non-empty", textString(""), textString("hello"), -1},
		{"non-empty vs empty", textString("hello"), textString(""), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Compare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTextCompareWithError(t *testing.T) {
	encoder := cj.Text[textWithError]()

	ok := textWithError{value: "other", shouldError: false}
	bad := textWithError{value: "test", shouldError: true}
	bad2 := textWithError{value: "test2", shouldError: true}

	// Erroring values sort after marshalable ones; two erroring values are unordered (0).
	assert.Equal(t, 1, encoder.Compare(bad, ok))
	assert.Equal(t, -1, encoder.Compare(ok, bad))
	assert.Equal(t, 0, encoder.Compare(bad, bad2))
}

type textWithError struct {
	value       string
	shouldError bool
}

func (t textWithError) MarshalText() ([]byte, error) {
	if t.shouldError {
		return nil, fmt.Errorf("marshal error")
	}
	return []byte(t.value), nil
}

func (t *textWithError) UnmarshalText(text []byte) error {
	t.value = string(text)
	return nil
}
