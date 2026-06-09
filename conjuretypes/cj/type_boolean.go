// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// booleanCodec provides JSON marshaling and unmarshaling for bool-like types.
type booleanCodec[T ~bool] struct{ comparableCodec[T] }

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
		return NewKindError[T](dec, tok.Kind(), "json boolean")
	}
	return nil
}

// booleanMapKeyCodec encodes bool map keys as the JSON strings "true"/"false", since JSON object keys must be strings.
type booleanMapKeyCodec[T ~bool] struct{ comparableCodec[T] }

func (booleanMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(strconv.FormatBool(bool(receiver))))
}

func (booleanMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	b, err := strconv.ParseBool(tok.String())
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
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

func (c booleanMapKeyCodec[T]) Sort(keys []T) { slices.SortFunc(keys, c.Compare) }
