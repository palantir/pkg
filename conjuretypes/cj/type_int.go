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

// int32Codec encodes signed integers as JSON numbers, enforcing the 32-bit signed range.
type int32Codec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ comparableCodec[T] }

func (int32Codec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if err := checkInt32Range(enc, receiver); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.Int(int64(receiver)))
}

func (int32Codec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindNumber {
		return NewKindError[T](dec, tok.Kind(), "json int")
	}
	num, err := strconv.ParseInt(tok.String(), 10, 32)
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(num)
	return nil
}

// int32MapKeyCodec encodes signed integers as JSON strings, since JSON object keys must be strings.
type int32MapKeyCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ orderedKeyCodec[T] }

func (int32MapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if err := checkInt32Range(enc, receiver); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.String(strconv.FormatInt(int64(receiver), 10)))
}

// checkInt32Range reports an encode error if receiver falls outside the 32-bit
// signed range, shared by the value and map-key codecs.
func checkInt32Range[T ~int | ~int8 | ~int16 | ~int32 | ~int64](enc *jsontext.Encoder, receiver T) error {
	if int64(receiver) < math.MinInt32 || int64(receiver) > math.MaxInt32 {
		return WrapEncodeError[T](enc, fmt.Errorf("value %d is out of range for a 32-bit signed integer", int64(receiver)))
	}
	return nil
}

func (int32MapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	i, err := strconv.ParseInt(tok.String(), 10, 32)
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(i)
	return nil
}
