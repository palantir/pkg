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
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	parse, err := datetime.ParseDateTime(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid datetime", err)
	}
	*receiver = T(parse)
	return nil
}

// Equal reports whether a and b are the same instant with the same wire string,
// so the same moment in different time zones is distinct: their emitted strings
// differ. This lets one codec serve datetime values, set elements, and map keys
// alike, since a JSON object member's identity is its emitted key string.
func (dateTimeCodec[T]) Equal(a, b T) bool {
	return time.Time(a).Equal(time.Time(b)) && datetime.DateTime(a).String() == datetime.DateTime(b).String()
}

func (c dateTimeCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}

// Sort orders keys by instant, breaking ties between same-instant keys by their
// wire string so that keys denoting the same moment in different time zones (e.g.
// "...Z" vs "...+01:00") get a deterministic, emitted-key order.
func (c dateTimeCodec[T]) Sort(keys []T) {
	slices.SortFunc(keys, func(a, b T) int {
		c := time.Time(a).Compare(time.Time(b))
		if c == 0 {
			return strings.Compare(datetime.DateTime(a).String(), datetime.DateTime(b).String())
		}
		return c
	})
}
