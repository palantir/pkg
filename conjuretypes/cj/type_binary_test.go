// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"encoding/base64"
	"testing"

	"github.com/palantir/pkg/binary"
	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestBinary(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "nil",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), Value: nil, JSON: `""`, SkipTestUnmarshal: true},
		},
		{
			Name: "empty_marshal",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), Value: []byte{}, JSON: `""`},
		},
		{
			Name: "null_unmarshal",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), JSON: "null", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON null into Go []uint8 after offset 4: want json string"},
		},
		{
			Name: "bytes",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), Value: []byte("hello 👋"), JSON: "\"aGVsbG8g8J+Riw==\""},
		},
		{
			Name: "invalid base64",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), JSON: "\"not_base64!@#\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"not_base64!@#\" into Go []uint8 after offset 15: illegal base64 data at input byte 3"},
		},
		{
			Name: "not a string",
			Test: typeTestCase[[]byte]{Codec: cj.Binary[[]byte](), JSON: "123", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON number 123 into Go []uint8 after offset 3: want json string"},
		},
		{
			Name: "map",
			Test: typeTestCase[map[binary.Binary]string]{Codec: cj.Map[map[binary.Binary]string](cj.BinaryMapKey[binary.Binary](), cj.String[string]()), Value: map[binary.Binary]string{
				binary.Binary(base64.StdEncoding.EncodeToString([]byte("a"))): "a",
				binary.Binary(base64.StdEncoding.EncodeToString([]byte("b"))): "b",
				binary.Binary(base64.StdEncoding.EncodeToString([]byte("c"))): "c",
			},
				JSON: `{"YQ==":"a","Yg==":"b","Yw==":"c"}`,
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
