// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"bytes"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/uuid"
)

// uuidCodec provides JSON marshaling and unmarshaling for uuid.UUID.
type uuidCodec[T ~[16]byte] struct{}

func (uuidCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(uuid.UUID(receiver).String()))
}

func (uuidCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	parsed, err := uuid.ParseUUID(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "", err)
	}
	*receiver = T(parsed)
	return nil
}

func (uuidCodec[T]) Compare(a, b T) int {
	return bytes.Compare(a[:], b[:])
}

func (uuidCodec[T]) Equal(a, b T) bool {
	return a == b
}
