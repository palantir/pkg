// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// fillGoType records T on a SemanticError that does not yet name a Go type.
// Without it, json/v2 would populate GoType with this internal codec wrapper
// rather than the value being (un)marshaled. T is the root type passed to the
// entry point; the error's JSON pointer locates failures within nested values.
func fillGoType[T any](err error) error {
	if serr, ok := errors.AsType[*json.SemanticError](err); ok && serr.GoType == nil {
		serr.GoType = reflect.TypeFor[T]()
	}
	return err
}

type anonymousMarshaler[T any, E Codec[T]] struct {
	receiver T
}

func (a *anonymousMarshaler[T, E]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return fillGoType[T]((*new(E)).MarshalJSONTo(enc, a.receiver))
}

type anonymousUnmarshaler[T any, E Codec[T]] struct {
	receiver *T
}

func (a anonymousUnmarshaler[T, E]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if a.receiver == nil {
		return fmt.Errorf("cj.NewAnonymousType: cannot unmarshal into nil receiver")
	}
	return fillGoType[T]((*new(E)).UnmarshalJSONFrom(dec, a.receiver))
}

func Unmarshal[T any, D Codec[T]](data []byte, v *T, _ D, opts ...json.Options) error {
	// AllowDuplicateNames lets the codec (not the jsontext syntax layer) detect
	// duplicate object members, so map codecs can report the richer
	// ErrDuplicateMapKey and catch canonicalized duplicates (e.g. "01" and "1").
	if len(opts) == 0 {
		return json.Unmarshal(data, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.Unmarshal(data, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func UnmarshalRead[T any, D Codec[T]](r io.Reader, v *T, _ D, opts ...json.Options) error {
	if len(opts) == 0 {
		return json.UnmarshalRead(r, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.UnmarshalRead(r, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func ClientDecoder[T any, D Codec[T]](_ D) clientDecoder[T, D] { return clientDecoder[T, D]{} }

func ServerDecoder[T any, D Codec[T]](_ D) serverDecoder[T, D] { return serverDecoder[T, D]{} }

func Marshal[T any, E Codec[T]](v T, _ E, opts ...json.Options) ([]byte, error) {
	// Pass a pointer: json/v2 makes a reflect.New copy of any non-pointer value to
	// obtain an addressable value, costing an extra allocation per call.
	//
	// AllowDuplicateNames skips json/v2's per-member duplicate-name validation,
	// which the codecs do not need: map keys are unique by construction and
	// generated object codecs emit each field name once.
	if len(opts) == 0 {
		return json.Marshal(&anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.Marshal(&anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func MarshalWrite[T any, E Codec[T]](out io.Writer, v T, _ E, opts ...json.Options) error {
	if len(opts) == 0 {
		return json.MarshalWrite(out, &anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.MarshalWrite(out, &anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
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
	if value, ok := v.(T); ok {
		return MarshalWrite(w, value, *new(E))
	}
	return json.MarshalWrite(w, v, json.WithMarshalers(defaultMarshalers))
}

func (defaultEncoder[T, E]) Marshal(v any) ([]byte, error) {
	if value, ok := v.(T); ok {
		return Marshal(value, *new(E))
	}
	return json.Marshal(v, json.WithMarshalers(defaultMarshalers))
}
