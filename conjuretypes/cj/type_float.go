// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"math"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// floatCodec provides JSON marshaling and unmarshaling for float64-like types.
// Encodes values as JSON numbers, and decodes JSON numbers into the underlying type.
// Special values like "NaN", "Infinity", and "-Infinity" are handled by jsontext.Float.
type floatCodec[T ~float64] struct{}

func (floatCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.Float(float64(receiver)))
}

func (floatCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
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

func (floatCodec[T]) Equal(a, b T) bool {
	return a == b
}

// floatMapKeyCodec provides JSON marshaling and unmarshaling for float64-like types used as map keys.
// Encodes float keys as JSON strings to comply with JSON map key requirements, supporting
// special values like "NaN", "Infinity", and "-Infinity".
type floatMapKeyCodec[T ~float64] struct{}

func (floatMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	switch {
	case math.IsNaN(float64(receiver)):
		return enc.WriteToken(jsontext.String("NaN"))
	case math.IsInf(float64(receiver), +1):
		return enc.WriteToken(jsontext.String("Infinity"))
	case math.IsInf(float64(receiver), -1):
		return enc.WriteToken(jsontext.String("-Infinity"))
	default:
		return enc.WriteToken(jsontext.String(string(appendShortestFloat(nil, float64(receiver)))))
	}
}

// appendShortestFloat formats f using the same shortest round-trip representation
// as jsontext.Float (ECMA-262, 10th edition, section 7.1.12.1 and RFC 8785,
// section 3.2.2.3), so a float serializes identically whether used as a value or
// a map key. It mirrors the unexported jsonwire.AppendFloat for 64-bit floats.
func appendShortestFloat(dst []byte, f float64) []byte {
	format := byte('f')
	if abs := math.Abs(f); abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		format = 'e'
	}
	dst = strconv.AppendFloat(dst, f, format, -1, 64)
	if format == 'e' {
		// Clean up "e-09" to "e-9" to match jsontext formatting.
		if n := len(dst); n >= 4 && dst[n-4] == 'e' && dst[n-3] == '-' && dst[n-2] == '0' {
			dst[n-2] = dst[n-1]
			dst = dst[:n-1]
		}
	}
	return dst
}

func (floatMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	switch s := tok.String(); s {
	case "NaN":
		return newInvalidTokenValueError(dec, tok, "cannot use NaN as map key", nil)
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

func (floatMapKeyCodec[T]) Compare(a, b T) int {
	// Handle special cases first
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		// NaN comparison is undefined, but for sorting we need consistent behavior
		if math.IsNaN(float64(a)) && math.IsNaN(float64(b)) {
			return 0
		}
		if math.IsNaN(float64(a)) {
			return -1 // or 1, but be consistent
		}
		return 1
	}
	if float64(a) < float64(b) {
		return -1
	}
	if float64(a) > float64(b) {
		return 1
	}
	return 0
}

func (floatMapKeyCodec[T]) Equal(a, b T) bool {
	return a == b
}
