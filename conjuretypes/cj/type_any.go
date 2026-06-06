// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	stdjson "encoding/json"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// anyCodec provides generic JSON marshaling and unmarshaling for any Go type T.
// It is a fallback encoder/decoder for types not otherwise handled by more specific
// implementations. Use this when you want to delegate to the default Go JSON logic,
// but still participate in the MarshalerTo/UnmarshalerFrom interfaces.
type anyCodec[T any] struct{}

func (anyCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return json.MarshalEncode(enc, receiver, DefaultOptions)
}

func (anyCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if dec.PeekKind() == jsontext.KindNull {
		// Consume a single token so the decoder advances past the null,
		// matching json/v2's default kind-mismatch behavior and the
		// read-then-validate scalar codecs.
		tok, err := dec.ReadToken()
		if err != nil {
			return WrapSyntaxError(dec, "", err)
		}
		return newKindMismatchTokenError(dec, tok, "non-optional value")
	}
	return json.UnmarshalDecode(dec, receiver, DefaultOptions)
}

func (anyCodec[T]) Equal(a, b T) bool {
	return reflect.DeepEqual(a, b)
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
