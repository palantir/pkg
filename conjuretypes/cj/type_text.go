// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"bytes"
	"encoding"

	"github.com/go-json-experiment/json/jsontext"
)

// textCodec provides JSON marshaling for types implementing encoding.TextMarshaler.
// Encodes values as JSON strings using the MarshalText method, and supports comparison by text value.
//
// textCodec provides JSON unmarshaling for types implementing encoding.TextUnmarshaler.
// Decodes JSON strings by calling UnmarshalText on the target type.
// Type U is the pointer to T that implements encoding.TextUnmarshaler.
type textCodec[T encoding.TextMarshaler, U interface {
	*T
	encoding.TextUnmarshaler
}] struct{}

func (textCodec[T, U]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	text, err := receiver.MarshalText()
	if err != nil {
		return err
	}
	return enc.WriteToken(jsontext.String(string(text)))
}

func (textCodec[T, U]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver U) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	if err := receiver.UnmarshalText([]byte(tok.String())); err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid text", err)
	}
	return nil
}

func (textCodec[T, U]) Compare(a, b T) int {
	aText, errA := a.MarshalText()
	bText, errB := b.MarshalText()
	switch {
	case errA != nil && errB != nil:
		return 0
	case errA != nil:
		return 1 // erroring values sort after marshalable ones
	case errB != nil:
		return -1
	}
	return bytes.Compare(aText, bText)
}

func (textCodec[T, U]) Equal(a, b T) bool {
	aText, errA := a.MarshalText()
	bText, errB := b.MarshalText()
	if errA != nil || errB != nil {
		return false
	}
	return bytes.Equal(aText, bText)
}
