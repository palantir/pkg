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

// binaryCodec provides JSON marshaling and unmarshaling for byte slice types (e.g., []byte).
// Encodes values as base64-encoded JSON strings using base64.StdEncoding.
// Implements comparison for equality and ordering based on byte content
// so values can be used as map keys.
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
		// Consume a single token so the decoder advances past the offending
		// scalar, matching json/v2's default kind-mismatch behavior and the
		// read-then-validate scalar codecs.
		tok, err := dec.ReadToken()
		if err != nil {
			return WrapSyntaxError(dec, err)
		}
		return newKindMismatchTokenError(dec, tok, "json string")
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
		return newInvalidJSONValueError(dec, val, err)
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

type binaryMapKeyCodec[T binary.Binary] struct{ comparableCodec[T] }

func (binaryMapKeyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return (stringCodec[binary.Binary]{}).MarshalJSONTo(enc, binary.Binary(receiver))
}

func (binaryMapKeyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	b64 := tok.String()
	if _, err := base64.StdEncoding.DecodeString(b64); err != nil {
		return newInvalidTokenValueError(dec, tok, "", err)
	}
	*receiver = T(b64)
	return nil
}
