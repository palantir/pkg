// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"bytes"
	"slices"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/uuid"
)

// uuidCodec provides JSON marshaling and unmarshaling for uuid.UUID.
type uuidCodec[T ~[16]byte] struct{ comparableCodec[T] }

func (uuidCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(uuid.UUID(receiver).String()))
}

func (uuidCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	parsed, err := uuid.ParseUUID(tok.String())
	if err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(parsed)
	return nil
}

func (uuidCodec[T]) Compare(a, b T) int {
	return bytes.Compare(a[:], b[:])
}

func (c uuidCodec[T]) Sort(keys []T) { slices.SortFunc(keys, c.Compare) }
