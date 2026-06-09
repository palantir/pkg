// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"errors"
	"io"
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
	dec := jsontext.NewDecoder(strings.NewReader(`["a","b"]`))
	_, _ = dec.ReadToken() // advance the decoder so it has a position to offer
	_, _ = dec.ReadToken()

	// A syntactic error with no position of its own picks up the decoder's, and
	// its cause gains a stacktrace; the *jsontext.SyntacticError type survives.
	cause := &jsontext.SyntacticError{Err: errors.New("bad syntax")}
	err := cj.WrapSyntaxError(dec, cause)

	var syntactic *jsontext.SyntacticError
	require.ErrorAs(t, err, &syntactic)
	assert.NotZero(t, syntactic.ByteOffset)
	assert.ErrorContains(t, err, "bad syntax")
	assert.ErrorIs(t, err, cause)

	var stackTracer werror.StackTracer
	require.ErrorAs(t, err, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

type failReader struct{ err error }

func (r failReader) Read([]byte) (int, error) { return 0, r.err }

// A transport read failure is not a syntax error: WrapSyntaxError leaves it
// untouched so it stays an I/O error rather than being dressed as a
// *jsontext.SyntacticError, and the underlying cause stays detectable.
func TestWrapSyntaxErrorPassesThroughIOError(t *testing.T) {
	var s string
	cause := errors.New("conn reset")

	err := cj.ClientDecoder(cj.String[string]()).Decode(failReader{cause}, &s)
	require.Error(t, err)

	var syntactic *jsontext.SyntacticError
	assert.False(t, errors.As(err, &syntactic), "I/O read error must not be classified as syntactic")
	assert.ErrorIs(t, err, cause)
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

	// Outer type and parser position are preserved; the cause gains a stacktrace.
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

	err := cj.NewKindError[string](dec, jsontext.KindNumber, "json string")

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
			err:  cj.NewMissingFieldsError[testStruct](dec, []string{"name"}),
			is:   cj.ErrMissingFields,
		},
		{
			name: "unknown fields",
			err:  cj.NewUnknownFieldsError[testStruct](dec, []string{"extra"}),
			is:   cj.ErrUnknownFields,
		},
		{
			name: "unknown enum",
			err:  cj.NewUnknownEnumError[string](dec),
			is:   cj.ErrUnknownEnum,
		},
		{
			name: "duplicate field",
			err:  cj.NewDuplicateFieldKeyError[testStruct](dec),
			is:   cj.ErrDuplicateField,
		},
		{
			name: "duplicate map key",
			err:  cj.NewDuplicateMapKeyError[map[string]int](dec),
			is:   cj.ErrDuplicateMapKey,
		},
		{
			name: "duplicate set item",
			err:  cj.NewDuplicateSetItemError[[]string](dec),
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

	err := cj.WrapEncodeError[string](enc, cause)

	var semantic *json.SemanticError
	require.ErrorAs(t, err, &semantic)
	assert.Equal(t, int64(0), semantic.ByteOffset)
	assert.Equal(t, jsontext.Pointer(""), semantic.JSONPointer)
	assert.ErrorContains(t, semantic.Err, "encode failed")
	assert.ErrorIs(t, err, cause)

	var stackTracer werror.StackTracer
	require.ErrorAs(t, err, &stackTracer)
	assert.NotNil(t, stackTracer.StackTrace())
}

// A decode error keeps its json-package outer type for errors.As detection
// while its cause carries a werror stacktrace.
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

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

// A transport failure while encoding surfaces as a bare I/O error, never dressed
// as a *json.SemanticError: json/v2 classifies writer failures as I/O and we
// return them unwrapped, so callers can still detect io.EOF and friends.
func TestEncodeIOErrorIsNotSemantic(t *testing.T) {
	err := cj.ClientEncoder(cj.String[string]()).Encode(errWriter{io.EOF}, "value")
	require.Error(t, err)

	var semantic *json.SemanticError
	assert.False(t, errors.As(err, &semantic), "I/O write error must not be classified as semantic")
	assert.ErrorIs(t, err, io.EOF)
}

func TestErrorIntegrationWithCodec(t *testing.T) {
	t.Run("wrong_type_for_string", func(t *testing.T) {
		var result string
		err := cj.ClientDecoder(cj.String[string]()).Unmarshal([]byte(`123`), &result)
		require.Error(t, err)

		var semantic *json.SemanticError
		require.ErrorAs(t, err, &semantic)
		assert.Equal(t, jsontext.KindNumber, semantic.JSONKind)
		// Only the kind is recorded, never the value, so sensitive content cannot leak.
		assert.Empty(t, semantic.JSONValue)
		assert.Equal(t, jsontext.Pointer(""), semantic.JSONPointer)
		// The entry point records the decoded Go type, not the internal codec
		// wrapper json/v2 would otherwise fill in (see fillGoType).
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
		assert.Empty(t, semantic.JSONValue)
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
		assert.Empty(t, semantic.JSONValue)
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
