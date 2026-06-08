// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"github.com/go-json-experiment/json/jsontext"
)

// stringCodec provides JSON marshaling and unmarshaling for string-like types.
// Encodes values as JSON strings, and decodes JSON strings into the underlying type.
type stringCodec[T ~string] struct{ orderedKeyCodec[T] }

func (stringCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(string(receiver)))
}

func (stringCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	*receiver = T(tok.String())
	return nil
}
