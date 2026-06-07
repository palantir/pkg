// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"fmt"
	"io"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type anonymousMarshaler[T any, E Codec[T]] struct {
	receiver T
}

func (a anonymousMarshaler[T, E]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return (*new(E)).MarshalJSONTo(enc, a.receiver)
}

type anonymousUnmarshaler[T any, E Codec[T]] struct {
	receiver *T
}

func (a anonymousUnmarshaler[T, E]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if a.receiver == nil {
		return fmt.Errorf("cj.NewAnonymousType: cannot unmarshal into nil receiver")
	}
	return (*new(E)).UnmarshalJSONFrom(dec, a.receiver)
}

func Unmarshal[T any, D Codec[T]](data []byte, v *T, _ D, opts ...json.Options) error {
	// AllowDuplicateNames lets the codec (not the jsontext syntax layer) detect
	// duplicate object members, so map codecs can report the richer
	// ErrDuplicateMapKey and catch canonicalized duplicates (e.g. "01" and "1").
	return json.Unmarshal(data, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func UnmarshalRead[T any, D Codec[T]](r io.Reader, v *T, _ D, opts ...json.Options) error {
	return json.UnmarshalRead(r, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func ClientDecoder[T any, D Codec[T]](_ D) clientDecoder[T, D] { return clientDecoder[T, D]{} }

func ServerDecoder[T any, D Codec[T]](_ D) serverDecoder[T, D] { return serverDecoder[T, D]{} }

func Marshal[T any, E Codec[T]](v T, _ E, opts ...json.Options) ([]byte, error) {
	return json.Marshal(anonymousMarshaler[T, E]{v}, opts...)
}

func MarshalWrite[T any, E Codec[T]](out io.Writer, v T, _ E, opts ...json.Options) error {
	return json.MarshalWrite(out, anonymousMarshaler[T, E]{v}, opts...)
}

func ClientEncoder[T any, E Codec[T]](_ E) defaultEncoder[T, E] { return defaultEncoder[T, E]{} }

func ServerEncoder[T any, E Codec[T]](_ E) defaultEncoder[T, E] { return defaultEncoder[T, E]{} }

type clientDecoder[T any, D Codec[T]] struct{}

func (clientDecoder[T, D]) Accept() string {
	return "application/json"
}

func (clientDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D))
	}
	return json.UnmarshalRead(r, v, json.WithUnmarshalers(defaultUnmarshalers))
}

func (clientDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D))
	}
	return json.Unmarshal(data, v, json.WithUnmarshalers(defaultUnmarshalers))
}

type serverDecoder[T any, D Codec[T]] struct{}

func (serverDecoder[T, D]) Accept() string {
	return "application/json"
}

func (serverDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.UnmarshalRead(r, v, json.WithUnmarshalers(defaultUnmarshalers), json.RejectUnknownMembers(true))
}

func (serverDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.Unmarshal(data, v, json.WithUnmarshalers(defaultUnmarshalers), json.RejectUnknownMembers(true))
}

type defaultEncoder[T any, E Codec[T]] struct{}

func (defaultEncoder[T, E]) ContentType() string {
	return "application/json"
}

func (defaultEncoder[T, E]) Encode(w io.Writer, v any) error {
	return json.MarshalWrite(w, v, json.WithMarshalers(defaultMarshalers))
}

func (defaultEncoder[T, E]) Marshal(v any) ([]byte, error) {
	return json.Marshal(v, json.WithMarshalers(defaultMarshalers))
}
