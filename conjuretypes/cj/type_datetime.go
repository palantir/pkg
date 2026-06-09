// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"
	"strings"
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
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	parse, err := datetime.ParseDateTime(tok.String())
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(parse)
	return nil
}

// Equal requires both the same instant and the same wire string, so the same
// moment in distinct time zones is unequal. This matches JSON map-key identity
// (the emitted key string), letting one codec serve values, set elements, and keys.
func (dateTimeCodec[T]) Equal(a, b T) bool {
	return time.Time(a).Equal(time.Time(b)) && datetime.DateTime(a).String() == datetime.DateTime(b).String()
}

func (c dateTimeCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}

// Sort orders by instant, then breaks ties on the wire string so same-instant
// keys in different zones (e.g. "...Z" vs "...+01:00") get a deterministic order.
func (dateTimeCodec[T]) Sort(keys []T) {
	slices.SortFunc(keys, func(a, b T) int {
		if c := time.Time(a).Compare(time.Time(b)); c != 0 {
			return c
		}
		return strings.Compare(datetime.DateTime(a).String(), datetime.DateTime(b).String())
	})
}
