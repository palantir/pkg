// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero",
			Test: typeTestCase[[16]byte]{Codec: cj.UUID[[16]byte](), Value: uuid.UUID{},
				JSON: "\"00000000-0000-0000-0000-000000000000\"",
			},
		},
		{
			Name: "value",
			Test: typeTestCase[uuid.UUID]{Codec: cj.UUID[uuid.UUID](), Value: must(uuid.ParseUUID("10101010-1010-1010-1010-101010101010")),
				JSON: "\"10101010-1010-1010-1010-101010101010\"",
			},
		},
		{
			Name: "null",
			Test: typeTestCase[uuid.UUID]{Codec: cj.UUID[uuid.UUID](), Value: must(uuid.ParseUUID("10101010-1010-1010-1010-101010101010")),
				JSON:                 "null",
				SkipTestMarshal:      true,
				ErrUnmarshalJSONFrom: "KindMismatchError at offset 4: want json string, got null",
			},
		},
		{
			Name: "map",
			Test: typeTestCase[map[uuid.UUID]string]{Codec: cj.ComparableMap[map[uuid.UUID]string](cj.UUID[uuid.UUID](), cj.String[string]()), Value: map[uuid.UUID]string{
				must(uuid.ParseUUID("00101010-1010-1010-1010-101010101010")): "foo",
				must(uuid.ParseUUID("00202020-2020-2020-2020-202020202020")): "bar",
			},
				JSON: `{"00101010-1010-1010-1010-101010101010":"foo","00202020-2020-2020-2020-202020202020":"bar"}`,
			},
		},
		{
			Name: "invalid",
			Test: typeTestCase[uuid.UUID]{Codec: cj.UUID[uuid.UUID](), JSON: "\"0000\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "InvalidValueError at offset 6: invalid UUID length: 4"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

func TestUUIDCompare(t *testing.T) {
	encoder := cj.UUID[[16]byte]()

	uuid1 := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	uuid2 := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 16}
	uuid3 := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	tests := []struct {
		name     string
		a, b     [16]byte
		expected int
	}{
		{"equal uuids", uuid1, uuid3, 0},
		{"a less than b", uuid1, uuid2, -1},
		{"a greater than b", uuid2, uuid1, 1},
		{"zero uuids", [16]byte{}, [16]byte{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Compare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
