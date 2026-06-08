// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"strings"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/rid"
)

// ridConstraint provides a common interface for types based on rid.ResourceIdentifier.
type ridConstraint interface {
	~struct {
		Service  string
		Instance string
		Type     string
		Locator  string
	}
}

// ridCodec provides JSON marshaling and unmarshaling for types based on rid.ResourceIdentifier.
// Encodes values as JSON strings using the canonical string representation of the resource identifier.
// Implements comparison based on all ridCodec fields for use as a map key.
type ridCodec[T ridConstraint] struct{ comparableCodec[T] }

func (ridCodec[T]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return enc.WriteToken(jsontext.String(rid.ResourceIdentifier(receiver).String()))
}

func (ridCodec[T]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindString {
		return newKindMismatchTokenError(dec, tok, "json string")
	}
	parsed, err := rid.ParseRID(tok.String())
	if err != nil {
		return newInvalidTokenValueError(dec, tok, "invalid resource identifier", err)
	}
	*receiver = T(parsed)
	return nil
}

func (ridCodec[T]) Compare(a, b T) int {
	ra, rb := rid.ResourceIdentifier(a), rid.ResourceIdentifier(b)
	return cmp.Or(
		strings.Compare(ra.Service, rb.Service),
		strings.Compare(ra.Instance, rb.Instance),
		strings.Compare(ra.Type, rb.Type),
		strings.Compare(ra.Locator, rb.Locator),
	)
}
