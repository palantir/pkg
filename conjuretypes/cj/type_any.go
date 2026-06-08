// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	stdjson "encoding/json"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// anyCodec delegates JSON encoding and decoding to the default json/v2 logic.
// Use it as a fallback for types without a more specific codec.
type anyCodec[T any] struct{}

func (anyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return json.MarshalEncode(enc, receiver, DefaultOptions)
}

func (anyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if dec.PeekKind() == jsontext.KindNull {
		// Conjure's null-coercion rule (wire spec 5.6.1) initializes only
		// optional, list, set, and map from a null/absent value; coercing null
		// to any other type -- including a non-optional any -- is an error.
		// Model a nullable any as optional<any>. (Nulls nested within an any
		// value are still accepted by the delegated decode below; this guard
		// rejects only a top-level null.)
		tok, err := dec.ReadToken()
		if err != nil {
			return WrapSyntaxError(dec, err)
		}
		return newKindMismatchTokenError(dec, tok, "non-optional value")
	}
	return json.UnmarshalDecode(dec, receiver, DefaultOptions)
}

func (anyCodec[T]) Equal(a, b T) bool {
	return reflect.DeepEqual(a, b)
}

func (c anyCodec[T]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}

var (
	defaultMarshalers = json.JoinMarshalers(
		json.MarshalFunc(func(number stdjson.Number) ([]byte, error) {
			return []byte(number), nil
		}),
		json.MarshalFunc(func(rawMessage stdjson.RawMessage) ([]byte, error) {
			return rawMessage, nil
		}),
	)
	defaultUnmarshalers = json.JoinUnmarshalers(
		json.UnmarshalFunc(func(data []byte, number *stdjson.Number) error {
			*number = stdjson.Number(data)
			return nil
		}),
		json.UnmarshalFunc(func(data []byte, message *stdjson.RawMessage) error {
			*message = append((*message)[:0], data...)
			return nil
		}),
	)

	DefaultOptions = json.JoinOptions(
		json.WithMarshalers(defaultMarshalers),
		json.WithUnmarshalers(defaultUnmarshalers),
	)
)
