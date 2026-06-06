// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

var (
	ErrMissingFields    = errors.New("missing required fields")
	ErrUnknownFields    = errors.New("unknown fields")
	ErrUnknownEnum      = errors.New("unknown enum value")
	ErrDuplicateField   = errors.New("duplicate field")
	ErrDuplicateMapKey  = errors.New("duplicate map key")
	ErrDuplicateSetItem = errors.New("duplicate set item")
)

// WrapSyntaxError annotates a syntax error with the current decoder position.
//
// If cause is already a jsontext.SyntacticError, it is returned as-is when no
// additional message is provided so the original parser offset and pointer are
// preserved.
func WrapSyntaxError(dec *jsontext.Decoder, message string, cause error) error {
	if cause == nil {
		cause = errors.New(message)
	} else if message != "" {
		cause = fmt.Errorf("%s: %w", message, cause)
	}
	var syntactic *jsontext.SyntacticError
	if message == "" && errors.As(cause, &syntactic) {
		return cause
	}
	return &jsontext.SyntacticError{
		ByteOffset:  dec.InputOffset(),
		JSONPointer: dec.StackPointer(),
		Err:         cause,
	}
}

// WrapEncodeError annotates an encode error with the current encoder position.
func WrapEncodeError(enc *jsontext.Encoder, message string, cause error) error {
	return &json.SemanticError{
		ByteOffset:  enc.OutputOffset(),
		JSONPointer: enc.StackPointer(),
		Err:         errorWithMessage(message, cause),
	}
}

// NewKindMismatchError returns an error for a JSON kind mismatch.
func NewKindMismatchError(dec *jsontext.Decoder, got jsontext.Kind, want string) error {
	return semanticDecodeError(dec, got, nil, fmt.Errorf("want %s, got %s", want, got.String()))
}

// NewInvalidValueError returns an error for a JSON value with the right kind but invalid contents.
func NewInvalidValueError(dec *jsontext.Decoder, message string, err error) error {
	return semanticDecodeError(dec, 0, nil, errorWithMessage(message, err))
}

func newKindMismatchTokenError(dec *jsontext.Decoder, tok jsontext.Token, want string) error {
	kind := tok.Kind()
	return semanticDecodeTokenError(dec, tok, fmt.Errorf("want %s, got %s", want, kind.String()))
}

func newInvalidTokenValueError(dec *jsontext.Decoder, tok jsontext.Token, message string, err error) error {
	return semanticDecodeTokenError(dec, tok, errorWithMessage(message, err))
}

func newInvalidJSONValueError(dec *jsontext.Decoder, value jsontext.Value, message string, err error) error {
	return semanticDecodeError(dec, value.Kind(), value, errorWithMessage(message, err))
}

// NewUnmarshalFieldError returns an error for a struct field that cannot be decoded.
func NewUnmarshalFieldError(dec *jsontext.Decoder, fieldDescriptor string, cause error) error {
	var semantic *json.SemanticError
	var syntactic *jsontext.SyntacticError
	if errors.As(cause, &semantic) || errors.As(cause, &syntactic) {
		return cause
	}
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("%s: %w", fieldDescriptor, cause))
}

// NewMissingFieldsError returns an error for missing required struct fields.
func NewMissingFieldsError(dec *jsontext.Decoder, typeName string, fields []string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("type %s missing required fields %v: %w", typeName, fields, ErrMissingFields))
}

// NewUnknownFieldsError returns an error for unknown struct fields.
func NewUnknownFieldsError(dec *jsontext.Decoder, typeName string, fields []string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("type %s has unknown fields %v: %w", typeName, fields, ErrUnknownFields))
}

// NewUnknownEnumError returns an error for an unrecognized enum value.
func NewUnknownEnumError(dec *jsontext.Decoder, typeName string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("type %s has unrecognized enum value: %w", typeName, ErrUnknownEnum))
}

// NewDuplicateFieldKeyError returns an error for a duplicate struct field key.
func NewDuplicateFieldKeyError(dec *jsontext.Decoder, fieldDescriptor string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("field %s duplicated: %w", fieldDescriptor, ErrDuplicateField))
}

// NewDuplicateMapKeyError returns an error for a duplicate decoded map key.
func NewDuplicateMapKeyError(dec *jsontext.Decoder, typeName string) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("type %s has duplicate map keys: %w", typeName, ErrDuplicateMapKey))
}

// NewDuplicateSetItemError returns an error for a duplicate decoded set item.
func NewDuplicateSetItemError(dec *jsontext.Decoder, typeName string, index int) error {
	return semanticDecodeError(dec, 0, nil, fmt.Errorf("type %s has a duplicate set item at index %d: %w", typeName, index, ErrDuplicateSetItem))
}

func semanticDecodeError(dec *jsontext.Decoder, kind jsontext.Kind, value jsontext.Value, err error) error {
	return &json.SemanticError{
		ByteOffset:  dec.InputOffset(),
		JSONPointer: dec.StackPointer(),
		JSONKind:    kind,
		JSONValue:   value.Clone(),
		Err:         err,
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
