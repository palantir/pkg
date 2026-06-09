// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"math"
	"slices"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// floatCodec encodes float64-like values as JSON numbers. Per the Conjure spec it also
// decodes the special strings "NaN", "Infinity", and "-Infinity".
type floatCodec[T ~float64] struct{ comparableCodec[T] }

func (floatCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.Float(float64(receiver)))
}

func (floatCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	switch kind := tok.Kind(); kind {
	case jsontext.KindNumber:
		f, err := tok.Float()
		if err != nil {
			return WrapDecodeError[T](dec, tok.Kind(), err)
		}
		*receiver = T(f)
	case jsontext.KindString:
		switch tok.String() {
		case "NaN":
			*receiver = T(math.NaN())
		case "Infinity":
			*receiver = T(math.Inf(+1))
		case "-Infinity":
			*receiver = T(math.Inf(-1))
		default:
			return NewKindError[T](dec, tok.Kind(), "json float")
		}
	default:
		return NewKindError[T](dec, tok.Kind(), "json float")
	}
	return nil
}

// floatMapKeyCodec encodes float64-like map keys as JSON strings, including the special
// values "NaN", "Infinity", and "-Infinity" per the Conjure spec.
type floatMapKeyCodec[T ~float64] struct{ orderedKeyCodec[T] }

func (floatMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	switch {
	case math.IsNaN(float64(receiver)):
		return enc.WriteToken(jsontext.String("NaN"))
	case math.IsInf(float64(receiver), +1):
		return enc.WriteToken(jsontext.String("Infinity"))
	case math.IsInf(float64(receiver), -1):
		return enc.WriteToken(jsontext.String("-Infinity"))
	default:
		// Append the quoted value directly into the encoder's buffer to avoid an
		// intermediate allocation; 32 fits the shortest float64 form plus quotes.
		dst := slices.Grow(enc.AvailableBuffer(), 32)
		dst = append(dst, '"')
		dst = jsontext.AppendFloat(dst, float64(receiver), 64)
		dst = append(dst, '"')
		return enc.WriteValue(dst)
	}
}

func (floatMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	switch s := tok.String(); s {
	case "NaN":
		// Duplicate "NaN" keys do not trigger DuplicateMapKeyError, since NaN != NaN.
		*receiver = T(math.NaN())
	case "Infinity":
		*receiver = T(math.Inf(1))
	case "-Infinity":
		*receiver = T(math.Inf(-1))
	default:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return WrapDecodeError[T](dec, tok.Kind(), err)
		}
		*receiver = T(f)
	}
	return nil
}
