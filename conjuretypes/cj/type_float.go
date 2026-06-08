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

// floatCodec provides JSON marshaling and unmarshaling for float64-like types.
// Encodes values as JSON numbers, and decodes JSON numbers into the underlying type.
// Special values like "NaN", "Infinity", and "-Infinity" are handled by jsontext.Float.
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
			return newInvalidTokenValueError(dec, tok, "invalid float", err)
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
			return newKindMismatchTokenError(dec, tok, "json float")
		}
	default:
		return newKindMismatchTokenError(dec, tok, "json float")
	}
	return nil
}

// floatMapKeyCodec provides JSON marshaling and unmarshaling for float64-like types used as map keys.
// Encodes float keys as JSON strings to comply with JSON map key requirements, supporting
// special values like "NaN", "Infinity", and "-Infinity".
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
		// Append the quoted shortest representation directly into the encoder's
		// buffer to avoid an intermediate []byte and string allocation. A float64
		// in shortest form is at most 24 bytes, plus the two surrounding quotes.
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
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	switch s := tok.String(); s {
	case "NaN":
		// Conjure permits NaN as a double map key on the wire (matching the value
		// codec and the marshal side, which writes "NaN"). A Go map cannot dedup
		// NaN keys (NaN != NaN), so a pathological duplicate "NaN" member will not
		// raise DuplicateMapKeyError.
		*receiver = T(math.NaN())
	case "Infinity":
		*receiver = T(math.Inf(1))
	case "-Infinity":
		*receiver = T(math.Inf(-1))
	default:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newInvalidTokenValueError(dec, tok, "invalid float", err)
		}
		*receiver = T(f)
	}
	return nil
}
