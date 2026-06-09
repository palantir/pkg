// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	werror "github.com/palantir/witchcraft-go-error"
)

var (
	ErrMissingFields    = errors.New("missing required fields")
	ErrUnknownFields    = errors.New("unknown fields")
	ErrUnknownEnum      = errors.New("unknown enum value")
	ErrDuplicateField   = errors.New("duplicate field")
	ErrDuplicateMapKey  = errors.New("duplicate map key")
	ErrDuplicateSetItem = errors.New("duplicate set item")
)

// WrapEncodeError annotates an encode error for a value of type T with the
// current encoder position and a stacktrace, recording T as the GoType.
func WrapEncodeError[T any](enc *jsontext.Encoder, cause error) error {
	return semanticError(enc.OutputOffset(), enc.StackPointer(), 0, reflect.TypeFor[T](), cause)
}

// WrapSyntaxError annotates a *jsontext.SyntacticError with a stacktrace and
// fills the decoder position into any envelope field the error left unset,
// mirroring semanticError. A non-syntactic cause — an I/O read failure or
// io.EOF — is returned untouched so json/v2's own I/O classification and
// errors.Is keep working, matching jsontext's own wrapSyntacticError.
func WrapSyntaxError(dec *jsontext.Decoder, cause error) error {
	syntactic, ok := errors.AsType[*jsontext.SyntacticError](cause)
	if !ok {
		return cause
	}
	syntactic.ByteOffset = cmp.Or(syntactic.ByteOffset, dec.InputOffset())
	syntactic.JSONPointer = cmp.Or(syntactic.JSONPointer, dec.StackPointer())
	syntactic.Err = cmp.Or(werror.Convert(syntactic.Err), werror.Error("invalid value"))
	return syntactic
}

// WrapDecodeError returns an error for a JSON value of the expected kind that is
// not a valid Conjure value of type T (an unparseable uuid, datetime, base64,
// and the like), recording T as the GoType. It is the general semantic decode
// error; the New* constructors below are presets over it.
func WrapDecodeError[T any](dec *jsontext.Decoder, kind jsontext.Kind, cause error) error {
	return semanticError(dec.InputOffset(), dec.StackPointer(), kind, reflect.TypeFor[T](), cause)
}

// NewKindError returns an error for a JSON kind mismatch decoding a value of
// type T. got is recorded as the JSONKind, so the message only names what was
// wanted.
func NewKindError[T any](dec *jsontext.Decoder, got jsontext.Kind, want string) error {
	return WrapDecodeError[T](dec, got, fmt.Errorf("want %s", want))
}

// NewMissingFieldsError returns an error for required fields missing from a
// struct of type T.
func NewMissingFieldsError[T any](dec *jsontext.Decoder, fields []string) error {
	return WrapDecodeError[T](dec, jsontext.KindBeginObject, fmt.Errorf("%w %v", ErrMissingFields, fields))
}

// NewUnknownFieldsError returns an error for unknown fields on a struct of type T.
func NewUnknownFieldsError[T any](dec *jsontext.Decoder, fields []string) error {
	return WrapDecodeError[T](dec, jsontext.KindBeginObject, fmt.Errorf("%w %v", ErrUnknownFields, fields))
}

// NewUnknownEnumError returns an error for an unrecognized value of enum type T.
func NewUnknownEnumError[T any](dec *jsontext.Decoder) error {
	return WrapDecodeError[T](dec, jsontext.KindString, ErrUnknownEnum)
}

// NewDuplicateFieldKeyError returns an error for a duplicate field key on a
// struct of type T.
func NewDuplicateFieldKeyError[T any](dec *jsontext.Decoder) error {
	return WrapDecodeError[T](dec, jsontext.KindBeginObject, ErrDuplicateField)
}

// NewDuplicateMapKeyError returns an error for a duplicate decoded key in a map
// of type T. The kind is left unset because the decoder is positioned on the
// duplicate entry, which the JSON pointer already locates, not the enclosing
// object.
func NewDuplicateMapKeyError[T any](dec *jsontext.Decoder) error {
	return WrapDecodeError[T](dec, 0, ErrDuplicateMapKey)
}

// NewDuplicateSetItemError returns an error for a duplicate decoded element in a
// set of type T. The kind is left unset because the decoder is positioned on the
// duplicate element, which the JSON pointer already locates, not the enclosing
// array.
func NewDuplicateSetItemError[T any](dec *jsontext.Decoder) error {
	return WrapDecodeError[T](dec, 0, ErrDuplicateSetItem)
}

// semanticError builds a *json.SemanticError at the given position. If cause is
// already a *json.SemanticError (one surfaced by a nested codec), only its unset
// envelope fields are filled, mirroring json/v2's own augmentation rather than
// overwriting more specific context. Its underlying error is still run through
// convertCause so the result always carries a stacktrace, even if the nested
// codec produced the SemanticError without one.
func semanticError(offset int64, pointer jsontext.Pointer, kind jsontext.Kind, goType reflect.Type, cause error) error {
	serr, ok := errors.AsType[*json.SemanticError](cause)
	if !ok {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: pointer,
			JSONKind:    kind,
			GoType:      goType,
			Err:         cmp.Or(werror.Convert(cause), werror.Error("invalid value")),
		}
	}
	serr.ByteOffset = cmp.Or(serr.ByteOffset, offset)
	serr.JSONPointer = cmp.Or(serr.JSONPointer, pointer)
	serr.JSONKind = cmp.Or(serr.JSONKind, kind)
	serr.GoType = cmp.Or(serr.GoType, goType)
	serr.Err = cmp.Or(werror.Convert(serr.Err), werror.Error("invalid value"))
	return serr
}
