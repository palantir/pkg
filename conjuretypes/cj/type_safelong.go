// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/safelong"
)

// safeLongCodec encodes signed integers as JSON numbers, enforcing the SafeLong range
// (±9007199254740991) so values remain exactly representable as JavaScript doubles.
type safeLongCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ comparableCodec[T] }

func (safeLongCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if err := checkSafeLongRange(enc, receiver); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.Int(int64(receiver)))
}

func (safeLongCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindNumber {
		return NewKindError[T](dec, tok.Kind(), "json int")
	}
	num, err := safelong.ParseSafeLong(tok.String())
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(num)
	return nil
}

// safeLongMapKeyCodec is safeLongCodec for map keys, encoding the SafeLong as a JSON string
// since JSON object keys must be strings.
type safeLongMapKeyCodec[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct{ orderedKeyCodec[T] }

func (safeLongMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if err := checkSafeLongRange(enc, receiver); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.String(strconv.FormatInt(int64(receiver), 10)))
}

// checkSafeLongRange reports an encode error if receiver falls outside the
// SafeLong range, shared by the value and map-key codecs.
func checkSafeLongRange[T ~int | ~int8 | ~int16 | ~int32 | ~int64](enc *jsontext.Encoder, receiver T) error {
	if _, err := safelong.NewSafeLong(int64(receiver)); err != nil {
		return WrapEncodeError[T](enc, err)
	}
	return nil
}

func (safeLongMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	i, err := safelong.ParseSafeLong(tok.String())
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(i)
	return nil
}
