// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// structCodec provides JSON marshaling for types that implement json.MarshalerTo.
// Delegates marshaling to the type's MarshalJSONTo method.
//
// structCodec provides JSON unmarshaling for types that implement json.UnmarshalerFrom.
// Delegates unmarshaling to the type's UnmarshalJSONFrom method.
// Type U is the pointer to T that implements UnmarshalerFrom.
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

func (structCodec[T, U]) Equal(a, b T) bool {
	return reflect.DeepEqual(a, b)
}
