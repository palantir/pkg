// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json/jsontext"
)

// optionalCodec is the codec for optional (pointer) values of type T, mapping
// nil pointers to JSON null and delegating non-nil values to ITEM.
type optionalCodec[T *U, U any, ITEM Codec[U]] struct{}

func (optionalCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		return enc.WriteToken(jsontext.Null)
	}
	return (*new(ITEM)).MarshalJSONTo(enc, *receiver)
}

func (optionalCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if dec.PeekKind() == jsontext.KindNull {
		if _, err := dec.ReadToken(); err != nil {
			return WrapSyntaxError(dec, err)
		}
		*receiver = nil
		return nil
	}
	*receiver = new(U)
	return (*new(ITEM)).UnmarshalJSONFrom(dec, *receiver)
}

func (optionalCodec[T, U, ITEM]) Equal(a, b T) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && (*new(ITEM)).Equal(*a, *b))
}

func (c optionalCodec[T, U, ITEM]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
