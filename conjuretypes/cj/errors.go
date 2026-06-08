// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"errors"
	"fmt"

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

// WrapSyntaxError annotates a syntax error with the current decoder position and
// attaches a stacktrace, preserving the *jsontext.SyntacticError type so callers
// can still detect it via errors.As. If cause is already positioned by jsontext,
// that offset and pointer are reused and only its underlying error is converted.
func WrapSyntaxError(dec *jsontext.Decoder, cause error) error {
	if syntactic, ok := errors.AsType[*jsontext.SyntacticError](cause); ok {
		return &jsontext.SyntacticError{
			ByteOffset:  syntactic.ByteOffset,
			JSONPointer: syntactic.JSONPointer,
			Err:         werror.Convert(syntactic.Err),
		}
	}
	return &jsontext.SyntacticError{
		ByteOffset:  dec.InputOffset(),
		JSONPointer: dec.StackPointer(),
		Err:         werror.Convert(cause),
	}
}

// WrapEncodeError annotates an encode error with the current encoder position
// and a stacktrace.
func WrapEncodeError(enc *jsontext.Encoder, message string, cause error) error {
	return &json.SemanticError{
		ByteOffset:  enc.OutputOffset(),
		JSONPointer: enc.StackPointer(),
		Err:         werror.Convert(errorWithMessage(message, cause)),
	}
}

// NewKindMismatchError returns an error for a JSON kind mismatch. The kind that
// was actually found is recorded on the SemanticError's JSONKind field, so the
// message only names what was wanted.
func NewKindMismatchError(dec *jsontext.Decoder, got jsontext.Kind, want string) error {
	return semanticDecodeError(dec, got, nil, fmt.Errorf("want %s", want))
}

func newKindMismatchTokenError(dec *jsontext.Decoder, tok jsontext.Token, want string) error {
	return semanticDecodeTokenError(dec, tok, fmt.Errorf("want %s", want))
}

func newInvalidTokenValueError(dec *jsontext.Decoder, tok jsontext.Token, message string, err error) error {
	return semanticDecodeTokenError(dec, tok, errorWithMessage(message, err))
}

func newInvalidJSONValueError(dec *jsontext.Decoder, value jsontext.Value, message string, err error) error {
	return semanticDecodeError(dec, value.Kind(), value, errorWithMessage(message, err))
}

// The constructors below leave the offending Go type and location to the
// SemanticError envelope: GoType is filled in by the codec entry point (see
// fillGoType) and JSONPointer is set from the decoder, so the messages carry
// only the information that the envelope does not.

// NewMissingFieldsError returns an error for missing required struct fields.
func NewMissingFieldsError(dec *jsontext.Decoder, fields []string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("%w %v", ErrMissingFields, fields))
}

// NewUnknownFieldsError returns an error for unknown struct fields.
func NewUnknownFieldsError(dec *jsontext.Decoder, fields []string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("%w %v", ErrUnknownFields, fields))
}

// NewUnknownEnumError returns an error for an unrecognized enum value.
func NewUnknownEnumError(dec *jsontext.Decoder) error {
	return semanticDecodeError(dec, 0, nil, ErrUnknownEnum)
}

// NewDuplicateFieldKeyError returns an error for a duplicate struct field key.
func NewDuplicateFieldKeyError(dec *jsontext.Decoder) error {
	return semanticDecodeError(dec, 0, nil, ErrDuplicateField)
}

// NewDuplicateMapKeyError returns an error for a duplicate decoded map key.
func NewDuplicateMapKeyError(dec *jsontext.Decoder) error {
	return semanticDecodeError(dec, 0, nil, ErrDuplicateMapKey)
}

// NewDuplicateSetItemError returns an error for a duplicate decoded set item.
func NewDuplicateSetItemError(dec *jsontext.Decoder) error {
	return semanticDecodeError(dec, 0, nil, ErrDuplicateSetItem)
}

func semanticDecodeError(dec *jsontext.Decoder, kind jsontext.Kind, value jsontext.Value, err error) error {
	return &json.SemanticError{
		ByteOffset:  dec.InputOffset(),
		JSONPointer: dec.StackPointer(),
		JSONKind:    kind,
		JSONValue:   value.Clone(),
		Err:         werror.Convert(err),
	}
}

func semanticDecodeTokenError(dec *jsontext.Decoder, tok jsontext.Token, err error) error {
	return semanticDecodeError(dec, tok.Kind(), tokenJSONValue(tok), err)
}

func tokenJSONValue(tok jsontext.Token) jsontext.Value {
	switch tok.Kind() {
	case jsontext.KindString:
		value, err := jsontext.AppendQuote(nil, tok.String())
		if err != nil {
			return nil
		}
		return value
	case jsontext.KindNumber:
		return jsontext.Value(tok.String())
	default:
		return nil
	}
}

func errorWithMessage(message string, cause error) error {
	switch {
	case cause == nil && message == "":
		return nil
	case cause == nil:
		return errors.New(message)
	case message == "":
		return cause
	default:
		return fmt.Errorf("%s: %w", message, cause)
	}
}
