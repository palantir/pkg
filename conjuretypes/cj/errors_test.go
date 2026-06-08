// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
	werror "github.com/palantir/witchcraft-go-error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapSyntaxError(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`invalid`))
	cause := errors.New("bad syntax")

	err := cj.WrapSyntaxError(dec, cause)

	var syntactic *jsontext.SyntacticError
	require.ErrorAs(t, err, &syntactic)
	assert.Equal(t, int64(0), syntactic.ByteOffset)
	assert.Equal(t, jsontext.Pointer(""), syntactic.JSONPointer)
	assert.ErrorContains(t, err, "bad syntax")
	assert.ErrorIs(t, err, cause)

	// The underlying error is converted to carry a stacktrace.
	var stackTracer werror.StackTracer
	require.ErrorAs(t, err, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

func TestWrapSyntaxErrorPreservesParserError(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`{ 1:2 }`))
	_, err := dec.ReadToken()
	require.NoError(t, err)
	_, err = dec.ReadToken()
	require.Error(t, err)

	var original *jsontext.SyntacticError
	require.ErrorAs(t, err, &original)

	wrapped := cj.WrapSyntaxError(dec, err)

	// The outer type and parser position are preserved so callers can still
	// detect a SyntacticError, but the underlying error gains a stacktrace.
	var syntactic *jsontext.SyntacticError
	require.ErrorAs(t, wrapped, &syntactic)
	assert.Equal(t, original.ByteOffset, syntactic.ByteOffset)
	assert.Equal(t, original.JSONPointer, syntactic.JSONPointer)
	assert.NotZero(t, syntactic.ByteOffset)

	var stackTracer werror.StackTracer
	require.ErrorAs(t, wrapped, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

func TestSemanticConstructors(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`123`))

	err := cj.NewKindMismatchError(dec, jsontext.KindNumber, "json string")

	var semantic *json.SemanticError
	require.ErrorAs(t, err, &semantic)
	assert.Equal(t, int64(0), semantic.ByteOffset)
	assert.Equal(t, jsontext.Pointer(""), semantic.JSONPointer)
	assert.Equal(t, jsontext.KindNumber, semantic.JSONKind)
	assert.ErrorContains(t, semantic.Err, "want json string")
}

func TestSemanticConstructorsPreserveCategory(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`{}`))

	tests := []struct {
		name string
		err  error
		is   error
	}{
		{
			name: "missing fields",
			err:  cj.NewMissingFieldsError(dec, []string{"name"}),
			is:   cj.ErrMissingFields,
		},
		{
			name: "unknown fields",
			err:  cj.NewUnknownFieldsError(dec, []string{"extra"}),
			is:   cj.ErrUnknownFields,
		},
		{
			name: "unknown enum",
			err:  cj.NewUnknownEnumError(dec),
			is:   cj.ErrUnknownEnum,
		},
		{
			name: "duplicate field",
			err:  cj.NewDuplicateFieldKeyError(dec),
			is:   cj.ErrDuplicateField,
		},
		{
			name: "duplicate map key",
			err:  cj.NewDuplicateMapKeyError(dec),
			is:   cj.ErrDuplicateMapKey,
		},
		{
			name: "duplicate set item",
			err:  cj.NewDuplicateSetItemError(dec),
			is:   cj.ErrDuplicateSetItem,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var semantic *json.SemanticError
			require.ErrorAs(t, tc.err, &semantic)
			assert.ErrorIs(t, tc.err, tc.is)
		})
	}
}

func TestWrapEncodeError(t *testing.T) {
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)
	cause := errors.New("encode failed")

	err := cj.WrapEncodeError(enc, "while writing", cause)

	var semantic *json.SemanticError
	require.ErrorAs(t, err, &semantic)
	assert.Equal(t, int64(0), semantic.ByteOffset)
	assert.Equal(t, jsontext.Pointer(""), semantic.JSONPointer)
	assert.ErrorContains(t, semantic.Err, "while writing")
	assert.ErrorIs(t, err, cause)

	var stackTracer werror.StackTracer
	require.ErrorAs(t, err, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

// TestErrorsCarryStacktrace verifies that a decode error surfaced through the
// codec keeps the json-package outer type (so callers' errors.As detection still
// works) while also carrying a werror stacktrace on its underlying cause.
func TestErrorsCarryStacktrace(t *testing.T) {
	var result string
	err := cj.ClientDecoder(cj.String[string]()).Unmarshal([]byte(`123`), &result)
	require.Error(t, err)

	var semantic *json.SemanticError
	require.ErrorAs(t, err, &semantic)

	var stackTracer werror.StackTracer
	require.ErrorAs(t, err, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

func TestErrorIntegrationWithCodec(t *testing.T) {
	t.Run("wrong_type_for_string", func(t *testing.T) {
		var result string
		err := cj.ClientDecoder(cj.String[string]()).Unmarshal([]byte(`123`), &result)
		require.Error(t, err)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.Equal(t, jsontext.KindNumber, semantic.JSONKind)
		assert.Equal(t, jsontext.Value("123"), semantic.JSONValue)
		assert.Equal(t, jsontext.Pointer(""), semantic.JSONPointer)
		// The entry point records the decoded Go type rather than the internal
		// codec wrapper that json/v2 would otherwise fill in (see fillGoType).
		assert.Equal(t, reflect.TypeFor[string](), semantic.GoType)
		assert.ErrorContains(t, semantic.Err, "want json string")
	})

	t.Run("invalid_base64", func(t *testing.T) {
		var result []byte
		err := cj.ClientDecoder(cj.Binary[[]byte]()).Unmarshal([]byte(`"not_base64!@#"`), &result)
		require.Error(t, err)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.Equal(t, jsontext.KindString, semantic.JSONKind)
		assert.Equal(t, jsontext.Value(`"not_base64!@#"`), semantic.JSONValue)
		assert.ErrorContains(t, semantic.Err, "illegal base64 data")
	})

	t.Run("duplicate_map_key", func(t *testing.T) {
		var result map[int]int
		err := cj.ClientDecoder(cj.Map[map[int]int](cj.Int32MapKey[int](), cj.Int32[int]())).Unmarshal([]byte(`{"01":1,"1":2}`), &result)
		require.Error(t, err)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.ErrorIs(t, err, cj.ErrDuplicateMapKey)
		assert.NotEmpty(t, semantic.JSONPointer)
	})
}

func TestDirectCodecSemanticErrorPointer(t *testing.T) {
	t.Run("list item", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1,"bad"]`), cj.DefaultOptions)
		var result []int

		err := cj.List[[]int](cj.Int32[int]()).UnmarshalJSONFrom(dec, &result)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.Equal(t, jsontext.Pointer("/1"), semantic.JSONPointer)
		assert.Equal(t, jsontext.KindString, semantic.JSONKind)
		assert.Equal(t, jsontext.Value(`"bad"`), semantic.JSONValue)
	})

	t.Run("map key collision", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"01":1,"1":2}`), cj.DefaultOptions, jsontext.AllowDuplicateNames(true))
		var result map[int]int

		err := cj.Map[map[int]int](cj.Int32MapKey[int](), cj.Int32[int]()).UnmarshalJSONFrom(dec, &result)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.Equal(t, jsontext.Pointer("/1"), semantic.JSONPointer)
		assert.ErrorIs(t, err, cj.ErrDuplicateMapKey)
	})
}
