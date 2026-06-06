// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"github.com/go-json-experiment/json/jsontext"
)

// optionalCodec provides JSON marshaling for optional (pointer) values of type T.
// Encodes nil pointers as JSON null, otherwise delegates encoding to ITEM.
// Decodes JSON null as nil, otherwise delegates decoding to ITEM.
type optionalCodec[T *U, U any, ITEM Codec[U]] struct{}

func (optionalCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if err := enc.WriteToken(jsontext.Null); err != nil {
			return WrapEncodeError(enc, "", err)
		}
		return nil
	}
	return (*new(ITEM)).MarshalJSONTo(enc, *receiver)
}

func (optionalCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if dec.PeekKind() == jsontext.KindNull {
		// still have to consume 'null' token
		if _, err := dec.ReadToken(); err != nil {
			return WrapSyntaxError(dec, "", err)
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
