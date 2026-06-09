// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type structMarshaler[T structMarshaler[T]] interface {
	Equal(other T) bool
	json.MarshalerTo
}

type structUnmarshaler[T structMarshaler[T]] interface {
	*T
	json.UnmarshalerFrom
}

// structCodec adapts generated struct types that already implement the json/v2
// and Equal methods into the Codec framework, with no per-type code.
type structCodec[T structMarshaler[T], U structUnmarshaler[T]] struct{}

func (structCodec[T, U]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return receiver.MarshalJSONTo(enc)
}

func (structCodec[T, U]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver U) error {
	return receiver.UnmarshalJSONFrom(dec)
}

func (structCodec[T, U]) Equal(a, b T) bool {
	return a.Equal(b)
}

func (c structCodec[T, U]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
