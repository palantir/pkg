// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshal(t *testing.T) {
	value := []string{"a", "b", "c"}

	data, err := cj.Marshal(value, cj.List[[]string](cj.String[string]()))
	require.NoError(t, err)
	assert.Equal(t, `["a","b","c"]`, string(data))

	var decoded []string
	err = cj.Unmarshal(data, &decoded, cj.List[[]string](cj.String[string]()))
	require.NoError(t, err)
	assert.Equal(t, value, decoded)
}

func TestUnmarshalAllowsCodecDuplicateDetection(t *testing.T) {
	var decoded map[string]int

	err := cj.Unmarshal(
		[]byte(`{"a":1,"a":2}`),
		&decoded,
		cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate map key")
}

func TestClientEncoderAndDecoder(t *testing.T) {
	encoder := cj.ClientEncoder(cj.List[[]int](cj.Int32[int]()))
	data, err := encoder.Marshal([]int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, `[1,2,3]`, string(data))

	var decoded []int
	decoder := cj.ClientDecoder(cj.List[[]int](cj.Int32[int]()))
	err = decoder.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, decoded)
}

func TestClientDecoderAllowsCodecDuplicateDetection(t *testing.T) {
	var decoded map[string]int
	decoder := cj.ClientDecoder(cj.Map[map[string]int](cj.String[string](), cj.Int32[int]()))

	err := decoder.Unmarshal([]byte(`{"a":1,"a":2}`), &decoded)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate map key")
}

func TestServerDecoderRejectsUnknownMembers(t *testing.T) {
	decoder := cj.ServerDecoder(cj.Struct[testStruct]())
	var decoded testStruct

	err := decoder.Unmarshal([]byte(`{"name":"John","age":25,"unknown":"field"}`), &decoded)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown fields")
}

func TestClientDecoderAllowsUnknownMembers(t *testing.T) {
	decoder := cj.ClientDecoder(cj.Struct[testStruct]())
	var decoded testStruct

	err := decoder.Unmarshal([]byte(`{"name":"John","age":25,"unknown":"field"}`), &decoded)

	require.NoError(t, err)
	assert.Equal(t, testStruct{Name: "John", Age: 25}, decoded)
}

func TestClientEncoderWrite(t *testing.T) {
	var buf bytes.Buffer
	encoder := cj.ClientEncoder(cj.String[string]())

	err := encoder.Encode(&buf, "hello")

	require.NoError(t, err)
	assert.Equal(t, `"hello"`, buf.String())
}

// TestClientEncoderUsesCodec ensures the encoder routes through the provided
// codec rather than falling back to plain json: the set codec drops duplicate
// items on marshal, which json.Marshal of a []int would not do.
func TestClientEncoderUsesCodec(t *testing.T) {
	encoder := cj.ClientEncoder(cj.Set[[]int](cj.Int32[int]()))

	data, err := encoder.Marshal([]int{1, 1, 2})
	require.NoError(t, err)
	assert.Equal(t, `[1,2]`, string(data))

	var buf bytes.Buffer
	err = encoder.Encode(&buf, []int{3, 3, 4})
	require.NoError(t, err)
	assert.Equal(t, `[3,4]`, buf.String())
}
