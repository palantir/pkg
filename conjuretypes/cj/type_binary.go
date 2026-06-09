// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"bytes"
	"encoding/base64"
	"slices"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/binary"
)

// binaryCodec marshals byte slice types as base64-encoded (StdEncoding) JSON strings.
type binaryCodec[T ~[]byte] struct{}

func (binaryCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	dst := slices.Grow(enc.AvailableBuffer(), base64.StdEncoding.EncodedLen(len(receiver))+2)
	dst = append(dst, '"')
	dst = base64.StdEncoding.AppendEncode(dst, receiver)
	dst = append(dst, '"')
	return enc.WriteValue(dst)
}

func (binaryCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if dec.PeekKind() != jsontext.KindString {
		// Consume the token so the decoder advances past the offending scalar.
		tok, err := dec.ReadToken()
		if err != nil {
			return WrapSyntaxError(dec, err)
		}
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	val, err := dec.ReadValue()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	unquoted, err := jsontext.AppendUnquote(nil, val)
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if len(unquoted) == 0 {
		*receiver = make(T, 0)
		return nil
	}
	decodedLen := base64.StdEncoding.DecodedLen(len(unquoted))
	if cap(*receiver) < decodedLen {
		*receiver = make(T, 0, decodedLen)
	} else {
		*receiver = (*receiver)[:0]
	}
	decoded, err := base64.StdEncoding.AppendDecode(*receiver, unquoted)
	if err != nil {
		return WrapDecodeError[T](dec, val.Kind(), err)
	}
	*receiver = decoded
	return nil
}

func (binaryCodec[T]) Compare(a, b T) int {
	return bytes.Compare(a, b)
}

func (binaryCodec[T]) Equal(a, b T) bool {
	return bytes.Equal(a, b)
}

func (c binaryCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}

type binaryMapKeyCodec[T binary.Binary] struct{ orderedKeyCodec[T] }

func (binaryMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	// Binary map keys are already base64 text; emit them as a plain JSON string.
	return enc.WriteToken(jsontext.String(string(receiver)))
}

func (binaryMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	b64 := tok.String()
	if _, err := base64.StdEncoding.DecodeString(b64); err != nil {
		return WrapDecodeError[T](dec, tok.Kind(), err)
	}
	*receiver = T(b64)
	return nil
}
