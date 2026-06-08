// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"unicode/utf8"

	"github.com/go-json-experiment/json/jsontext"
)

type bearerTokenCodec[T ~string] struct{}

func (bearerTokenCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(string(receiver)))
}

func (bearerTokenCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	str := tok.String()
	if len(str) == 0 {
		return newInvalidTokenValueError(dec, tok, "empty bearer token", nil)
	}
	// Per RFC 6750 section 2.1, a bearer token is one or more base64 characters
	// followed by optional '=' padding: 1*( b64char ) *"=". The '=' padding is
	// only valid as a trailing suffix, not interspersed within the token.
	i := 0
	for i < len(str) && str[i] < utf8.RuneSelf && validBearerTokenChars[str[i]] {
		i++
	}
	if i == 0 {
		// the token must begin with at least one base64 character
		return newInvalidTokenValueError(dec, tok, "invalid character for bearer token", nil)
	}
	for ; i < len(str); i++ {
		if str[i] != '=' {
			return newInvalidTokenValueError(dec, tok, "invalid character for bearer token", nil)
		}
	}
	*receiver = T(str)
	return nil
}

func (bearerTokenCodec[T]) Equal(a, b T) bool {
	return a == b
}

// validBearerTokenChars marks the ASCII bytes permitted in a bearer token: the
// base64 alphabet plus the RFC 6750 extras. It is indexed directly by byte value.
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
