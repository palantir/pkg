// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// VisitJSONObjectFields asserts that dec's next value is a json object, reads the opening and closing braces,
// and calls visit(dec) for each key in the object. The function is expected to read two values from the decoder, the key and the value.
func VisitJSONObjectFields(dec *jsontext.Decoder, visit func(dec *jsontext.Decoder) error) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
	}
	if tok.Kind() != jsontext.KindBeginObject {
		return newKindMismatchTokenError(dec, tok, "object opening brace")
	}
	for {
		if dec.PeekKind() == jsontext.KindEndObject {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, "", err)
			}
			return nil
		}
		if err := visit(dec); err != nil {
			return err
		}
	}
}

// VisitJSONListFields asserts that dec's next value is a json list, reads the opening and closing brackets,
// and calls visit(dec) for each item in the list. The function is expected to read one value from the decoder.
func VisitJSONListFields(dec *jsontext.Decoder, visit func(dec *jsontext.Decoder) error) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, "", err)
	}
	if tok.Kind() != jsontext.KindBeginArray {
		return newKindMismatchTokenError(dec, tok, "list opening bracket")
	}
	for {
		if dec.PeekKind() == jsontext.KindEndArray {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, "", err)
			}
			return nil
		}
		if err := visit(dec); err != nil {
			return err
		}
	}
}

func getOptionOrTrue(options jsontext.Options, setter func(bool) jsontext.Options) bool {
	value, ok := json.GetOption(options, setter)
	if !ok {
		return true
	}
	return value
}

func getOptionOrFalse(options jsontext.Options, setter func(bool) jsontext.Options) bool {
	value, ok := json.GetOption(options, setter)
	if !ok {
		return false
	}
	return value
}
