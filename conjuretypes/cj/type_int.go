// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// int32Codec provides JSON marshaling and unmarshaling for integer types (signed).
// Encodes values as JSON numbers, and decodes JSON numbers into the underlying type.
type int32Codec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{}

func (int32Codec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if int64(receiver) < math.MinInt32 || int64(receiver) > math.MaxInt32 {
		return WrapEncodeError(enc, "invalid int32", fmt.Errorf("value %d is out of range for a 32-bit signed integer", int64(receiver)))
	}
	return enc.WriteToken(jsontext.Int(int64(receiver)))
}

func (int32Codec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
	}
	if tok.Kind() != jsontext.KindNumber {
		return newKindMismatchTokenError(dec, tok, "json int")
	}
	num, err := strconv.ParseInt(tok.String(), 10, 32)
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid int32", err)
	}
	*receiver = T(num)
	return nil
}

func (int32Codec[T]) Equal(a, b T) bool {
	return a == b
}

// int32MapKeyCodec provides JSON marshaling and unmarshaling for signed integer types used as map keys.
// Encodes integer keys as JSON strings to comply with JSON map key requirements.
type int32MapKeyCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{}

func (int32MapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if int64(receiver) < math.MinInt32 || int64(receiver) > math.MaxInt32 {
		return WrapEncodeError(enc, "invalid int32", fmt.Errorf("value %d is out of range for a 32-bit signed integer", int64(receiver)))
	}
	return enc.WriteToken(jsontext.String(strconv.FormatInt(int64(receiver), 10)))
}

func (int32MapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	i, err := strconv.ParseInt(tok.String(), 10, 32)
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid int32", err)
	}
	*receiver = T(i)
	return nil
}

func (int32MapKeyCodec[T]) Equal(a, b T) bool {
	return a == b
}
