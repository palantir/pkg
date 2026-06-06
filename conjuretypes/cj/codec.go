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

func NewAnonymousType[T any, E Codec[T]](receiver *T, _ E) anonymousCodec[T, E] {
	return anonymousCodec[T, E]{receiver: receiver}
}

type anonymousCodec[T any, E Codec[T]] struct {
	receiver *T
}

func (a anonymousCodec[T, E]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if a.receiver == nil {
		return enc.WriteToken(jsontext.Null)
	}
	return (*new(E)).MarshalJSONTo(enc, *a.receiver)
}

func (a anonymousCodec[T, E]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if a.receiver == nil {
		return fmt.Errorf("cj.NewAnonymousType: cannot unmarshal into nil receiver")
	}
	return (*new(E)).UnmarshalJSONFrom(dec, a.receiver)
}

// Unmarshaling Utils //

func Unmarshal[T any, D Codec[T]](data []byte, v *T, _ D, opts ...json.Options) error {
	return json.Unmarshal(data, &anonymousCodec[T, D]{v}, DefaultOptions, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func UnmarshalRead[T any, D Codec[T]](r io.Reader, v *T, _ D, opts ...json.Options) error {
	return json.UnmarshalRead(r, &anonymousCodec[T, D]{v}, DefaultOptions, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func unmarshalOptions[T any, D Codec[T]](opts ...json.Options) json.Options {
	return json.JoinOptions(
		DefaultOptions,
		jsontext.AllowDuplicateNames(true),
		json.WithUnmarshalers(json.JoinUnmarshalers(
			defaultUnmarshalers,
			json.UnmarshalFromFunc((*new(D)).UnmarshalJSONFrom),
		)),
		json.JoinOptions(opts...),
	)
}

type clientDecoder[T any, D Codec[T]] struct{}

func ClientDecoder[T any, D Codec[T]](_ D) clientDecoder[T, D] {
	return clientDecoder[T, D]{}
}

func (clientDecoder[T, D]) Accept() string {
	return "application/json"
}

func (clientDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D))
	}
	return json.UnmarshalRead(r, v, unmarshalOptions[T, D]())
}

func (clientDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D))
	}
	return json.Unmarshal(data, v, unmarshalOptions[T, D]())
}

type serverDecoder[T any, D Codec[T]] struct{}

func ServerDecoder[T any, D Codec[T]](_ D) serverDecoder[T, D] {
	return serverDecoder[T, D]{}
}

func (serverDecoder[T, D]) Accept() string {
	return "application/json"
}

func (serverDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.UnmarshalRead(r, v, unmarshalOptions[T, D](json.RejectUnknownMembers(true)))
}

func (serverDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.Unmarshal(data, v, unmarshalOptions[T, D](json.RejectUnknownMembers(true)))
}

// Marshaling Utils //

func Marshal[T any, E Codec[T]](v *T, codec E, opts ...json.Options) ([]byte, error) {
	return json.Marshal(NewAnonymousType(v, codec), DefaultOptions, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func MarshalWrite[T any, E Codec[T]](out io.Writer, v *T, codec E, opts ...json.Options) error {
	return json.MarshalWrite(out, NewAnonymousType(v, codec), DefaultOptions, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func marshalOptions[T any, E Codec[T]]() json.Options {
	return json.JoinOptions(
		DefaultOptions,
		jsontext.AllowDuplicateNames(true),
		json.WithMarshalers(json.JoinMarshalers(
			defaultMarshalers,
			json.MarshalToFunc((*new(E)).MarshalJSONTo),
		)),
	)
}

type clientEncoder[T any, E Codec[T]] struct{}

func ClientEncoder[T any, E Codec[T]](_ E) clientEncoder[T, E] {
	return clientEncoder[T, E]{}
}

func (clientEncoder[T, E]) ContentType() string {
	return "application/json"
}

func (clientEncoder[T, E]) Encode(w io.Writer, v any) error {
	return json.MarshalWrite(w, v, marshalOptions[T, E]())
}

func (clientEncoder[T, E]) Marshal(v any) ([]byte, error) {
	return json.Marshal(v, marshalOptions[T, E]())
}

type serverEncoder[T any, E Codec[T]] struct{}

func ServerEncoder[T any, E Codec[T]](_ E) serverEncoder[T, E] {
	return serverEncoder[T, E]{}
}

func (serverEncoder[T, E]) ContentType() string {
	return "application/json"
}

func (serverEncoder[T, E]) Encode(w io.Writer, v any) error {
	return json.MarshalWrite(w, v, marshalOptions[T, E]())
}

func (serverEncoder[T, E]) Marshal(v any) ([]byte, error) {
	return json.Marshal(v, marshalOptions[T, E]())
}
