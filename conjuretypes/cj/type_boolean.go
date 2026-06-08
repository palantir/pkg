// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// booleanCodec provides JSON marshaling and unmarshaling for Go bool-like types.
// Encodes values as JSON true/false, and decodes JSON booleans into the underlying type.
type booleanCodec[T ~bool] struct{}

func (booleanCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver {
		return enc.WriteToken(jsontext.True)
	}
	return enc.WriteToken(jsontext.False)
}

func (booleanCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	switch tok.Kind() {
	case jsontext.KindTrue:
		*receiver = true
	case jsontext.KindFalse:
		*receiver = false
	default:
		return newKindMismatchTokenError(dec, tok, "json boolean")
	}
	return nil
}

func (booleanCodec[T]) Equal(a, b T) bool {
	return a == b
}

// booleanMapKeyCodec provides JSON marshaling for bool-like types used as map keys.
// Encodes bool keys as the JSON strings "true" or "false" (not as booleans), to comply with JSON map key requirements.
type booleanMapKeyCodec[T ~bool] struct{}

func (booleanMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver {
		return enc.WriteToken(jsontext.String("true"))
	}
	return enc.WriteToken(jsontext.String("false"))
}

func (booleanMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	b, err := strconv.ParseBool(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid boolean", err)
	}
	*receiver = T(b)
	return nil
}

func (booleanMapKeyCodec[T]) Compare(a, b T) int {
	if a == b {
		return 0
	}
	if a {
		return 1
	}
	return -1
}

func (booleanMapKeyCodec[T]) Equal(a, b T) bool {
	return a == b
}
