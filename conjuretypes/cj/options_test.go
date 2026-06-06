// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatNilSliceAsNull(t *testing.T) {
	t.Run("default_behavior_empty_array", func(t *testing.T) {
		data, err := cj.Marshal(new([]int), cj.List[[]int](cj.Int32[int]()))
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))
	})

	t.Run("format_as_null_enabled", func(t *testing.T) {
		data, err := cj.Marshal(new([]int), cj.List[[]int](cj.Int32[int]()), json.FormatNilSliceAsNull(true))
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))
	})

	t.Run("format_as_null_disabled", func(t *testing.T) {
		data, err := cj.Marshal(new([]int), cj.List[[]int](cj.Int32[int]()), json.FormatNilSliceAsNull(false))
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))
	})

	t.Run("non_nil_slice_unaffected", func(t *testing.T) {
		data, err := cj.Marshal(&[]int{}, cj.List[[]int](cj.Int32[int]()), json.FormatNilSliceAsNull(true))
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data))
	})
}

func TestFormatNilMapAsNull(t *testing.T) {
	t.Run("default_behavior_empty_object", func(t *testing.T) {
		data, err := cj.Marshal(new(map[string]int), cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()))
		require.NoError(t, err)
		assert.Equal(t, "{}", string(data))
	})

	t.Run("format_as_null_enabled", func(t *testing.T) {
		data, err := cj.Marshal(new(map[string]int), cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), json.FormatNilMapAsNull(true))
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))
	})

	t.Run("format_as_null_disabled", func(t *testing.T) {
		data, err := cj.Marshal(new(map[string]int), cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), json.FormatNilMapAsNull(false))
		require.NoError(t, err)
		assert.Equal(t, "{}", string(data))
	})

	t.Run("non_nil_map_unaffected", func(t *testing.T) {
		data, err := cj.Marshal(&map[string]int{}, cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), json.FormatNilMapAsNull(true))
		require.NoError(t, err)
		assert.Equal(t, "{}", string(data))
	})
}

func TestDeterministicOrdering(t *testing.T) {
	originalMap := map[string]int{"z": 1, "a": 2, "m": 3}

	t.Run("deterministic_by_default", func(t *testing.T) {
		data, err := cj.Marshal(&originalMap, cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()))
		require.NoError(t, err)
		assert.Equal(t, `{"a":2,"m":3,"z":1}`, string(data))
	})

	t.Run("deterministic_explicitly_enabled", func(t *testing.T) {
		data, err := cj.Marshal(&originalMap, cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), json.Deterministic(true))
		require.NoError(t, err)
		assert.Equal(t, `{"a":2,"m":3,"z":1}`, string(data))
	})

	t.Run("deterministic_disabled", func(t *testing.T) {
		data, err := cj.Marshal(&originalMap, cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]()), json.Deterministic(false))
		require.NoError(t, err)
		// Result should still be valid JSON, but order may vary
		// We can't assert exact order, but we can verify it parses correctly
		var result map[string]int
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, originalMap, result)
	})
}

func TestRejectUnknownMembers(t *testing.T) {

	// We can't easily test RejectUnknownMembers with our current type system
	// since it's typically used in generated code, but we can test the concept
	// using the anyCodec type and json options
	t.Run("strict_parsing_with_unknown_field", func(t *testing.T) {
		jsonWithUnknown := `{"name":"John","age":25,"unknown":"field"}`
		var result testStruct

		// This should work with lenient parsing (default)
		err := json.Unmarshal([]byte(jsonWithUnknown), &result)
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 25, result.Age)
	})

	t.Run("strict_parsing_rejects_unknown", func(t *testing.T) {
		jsonWithUnknown := `{"name":"John","age":25,"unknown":"field"}`
		var result testStruct

		// This should fail with strict parsing
		err := json.Unmarshal([]byte(jsonWithUnknown), &result, json.RejectUnknownMembers(true))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("strict_parsing_accepts_known_fields", func(t *testing.T) {
		validJSON := `{"name":"John","age":25}`
		var result testStruct

		// This should work even with strict parsing
		err := json.Unmarshal([]byte(validJSON), &result, json.RejectUnknownMembers(true))
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 25, result.Age)
	})
}
