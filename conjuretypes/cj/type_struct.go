// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// structCodec marshals and unmarshals types that implement the JSON v2
// MarshalerTo/UnmarshalerFrom methods, delegating to them. U is the pointer to
// T that implements UnmarshalerFrom.
type structCodec[T json.MarshalerTo, U interface {
	*T
	json.UnmarshalerFrom
}] struct{}

func (structCodec[T, U]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return receiver.MarshalJSONTo(enc)
}

func (structCodec[T, U]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver U) error {
	return receiver.UnmarshalJSONFrom(dec)
}

// Equal compares via reflect.DeepEqual (generated structs expose no field-wise
// Equal). This compares Go representation, not Conjure semantic value, so it can
// diverge from the component codecs and report false inequality: same-instant
// datetimes with a different monotonic reading or *Location, and a nil vs empty
// []byte, are unequal here though dateTimeCodec/binaryCodec treat them as equal.
func (structCodec[T, U]) Equal(a, b T) bool {
	return reflect.DeepEqual(a, b)
}

func (c structCodec[T, U]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
