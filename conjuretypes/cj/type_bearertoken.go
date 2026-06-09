// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"errors"
	"unicode/utf8"

	"github.com/go-json-experiment/json/jsontext"
)

var (
	errEmptyBearerToken       = errors.New("empty bearer token")
	errInvalidBearerTokenChar = errors.New("invalid character for bearer token")
)

type bearerTokenCodec[T ~string] struct{ comparableCodec[T] }

func (bearerTokenCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(string(receiver)))
}

func (bearerTokenCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return NewKindError[T](dec, tok.Kind(), "json string")
	}
	str := tok.String()
	if len(str) == 0 {
		return WrapDecodeError[T](dec, tok.Kind(), errEmptyBearerToken)
	}
	// RFC 6750 section 2.1 grammar: 1*( b64char ) *"=" (padding only as a trailing suffix).
	i := 0
	for i < len(str) && str[i] < utf8.RuneSelf && validBearerTokenChars[str[i]] {
		i++
	}
	if i == 0 {
		return WrapDecodeError[T](dec, tok.Kind(), errInvalidBearerTokenChar)
	}
	for ; i < len(str); i++ {
		if str[i] != '=' {
			return WrapDecodeError[T](dec, tok.Kind(), errInvalidBearerTokenChar)
		}
	}
	*receiver = T(str)
	return nil
}

// validBearerTokenChars is a byte-indexed lookup of the ASCII characters permitted
// in a bearer token: the base64 alphabet plus the RFC 6750 extras.
var validBearerTokenChars = func() [utf8.RuneSelf]bool {
	var chars [utf8.RuneSelf]bool
	for i := '0'; i <= '9'; i++ {
		chars[i] = true
	}
	for i := 'A'; i <= 'Z'; i++ {
		chars[i] = true
	}
	for i := 'a'; i <= 'z'; i++ {
		chars[i] = true
	}
	chars['+'] = true
	chars['-'] = true
	chars['.'] = true
	chars['/'] = true
	chars['_'] = true
	chars['~'] = true
	return chars
}()
