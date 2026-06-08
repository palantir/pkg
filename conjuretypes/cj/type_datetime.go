// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"slices"
	"time"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/datetime"
)

type dateTimeCodec[T time.Time | datetime.DateTime] struct{}

func (dateTimeCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(datetime.DateTime(receiver).String()))
}

func (dateTimeCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	parse, err := datetime.ParseDateTime(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid datetime", err)
	}
	*receiver = T(parse)
	return nil
}

// Equal compares by instant, so the same moment in different time zones is equal
// as a value. Map keys use dateTimeMapKeyCodec (see DateTimeMapKey) instead.
func (dateTimeCodec[T]) Equal(a, b T) bool {
	return time.Time(a).Equal(time.Time(b))
}

func (c dateTimeCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}

// dateTimeMapKeyCodec orders and compares datetime map keys by their wire string
// (RFC3339Nano), not by instant: a JSON object member's identity and order are
// defined by the emitted key, matching json/v2. Comparing by instant would tie
// same-instant keys that serialize differently (e.g. "...Z" vs "...+01:00").
type dateTimeMapKeyCodec[T time.Time | datetime.DateTime] struct{}

func (dateTimeMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return dateTimeCodec[T]{}.MarshalJSONTo(enc, receiver)
}

func (dateTimeMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	return dateTimeCodec[T]{}.UnmarshalJSONFrom(dec, receiver)
}

func (dateTimeMapKeyCodec[T]) Compare(a, b T) int {
	return cmp.Compare(datetime.DateTime(a).String(), datetime.DateTime(b).String())
}

func (c dateTimeMapKeyCodec[T]) Sort(keys []T) { slices.SortFunc(keys, c.Compare) }

func (dateTimeMapKeyCodec[T]) Equal(a, b T) bool {
	return datetime.DateTime(a).String() == datetime.DateTime(b).String()
}

func (c dateTimeMapKeyCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
