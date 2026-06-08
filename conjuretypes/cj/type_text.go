// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"bytes"
	"encoding"
	"slices"

	"github.com/go-json-experiment/json/jsontext"
)

type textMarshaler[T any] interface {
	encoding.TextMarshaler
}

type textUnmarshaler[T any] interface {
	*T
	encoding.TextUnmarshaler
}

// textCodec encodes T as a JSON string via MarshalText and decodes via
// UnmarshalText, comparing values by their text form. U is the pointer to T that
// implements encoding.TextUnmarshaler.
type textCodec[T textMarshaler[T], U textUnmarshaler[T]] struct{}

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

func (c textCodec[T, U]) Sort(keys []T) { slices.SortFunc(keys, c.Compare) }

func (textCodec[T, U]) Equal(a, b T) bool {
	aText, errA := a.MarshalText()
	bText, errB := b.MarshalText()
	if errA != nil || errB != nil {
		return false
	}
	return bytes.Equal(aText, bText)
}

func (c textCodec[T, U]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
