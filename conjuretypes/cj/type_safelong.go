// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/safelong"
)

// safeLongCodec provides JSON marshaling and unmarshaling for integer types (signed).
// Encodes values as JSON numbers, and decodes JSON numbers into the underlying type.
type safeLongCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ comparableCodec[T] }

func (safeLongCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if _, err := safelong.NewSafeLong(int64(receiver)); err != nil {
		return WrapEncodeError(enc, "invalid safelong", err)
	}
	return enc.WriteToken(jsontext.Int(int64(receiver)))
}

func (safeLongCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindNumber {
		return newKindMismatchTokenError(dec, tok, "json int")
	}
	num, err := safelong.ParseSafeLong(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid safelong", err)
	}
	*receiver = T(num)
	return nil
}

// safeLongMapKeyCodec provides JSON marshaling and unmarshaling for signed integer types used as map keys.
// Encodes integer keys as JSON strings to comply with JSON map key requirements.
type safeLongMapKeyCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ comparableCodec[T] }

func (safeLongMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if _, err := safelong.NewSafeLong(int64(receiver)); err != nil {
		return WrapEncodeError(enc, "invalid safelong", err)
	}
	return enc.WriteToken(jsontext.String(strconv.FormatInt(int64(receiver), 10)))
}

func (safeLongMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	i, err := safelong.ParseSafeLong(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid safelong", err)
	}
	*receiver = T(i)
	return nil
}
