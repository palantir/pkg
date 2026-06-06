// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
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
		return WrapSyntaxError(dec, "", err)
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

func (dateTimeCodec[T]) Compare(a, b T) int {
	aTime, bTime := time.Time(a), time.Time(b)
	if aTime.After(bTime) {
		return 1
	}
	if aTime.Before(bTime) {
		return -1
	}
	return 0
}

func (dateTimeCodec[T]) Equal(a, b T) bool {
	return time.Time(a).Equal(time.Time(b))
}
