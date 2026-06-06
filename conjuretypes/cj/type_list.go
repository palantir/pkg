// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// listCodec provides JSON marshaling for slices of type T using a nested encoder ITEM.
// Encodes slices as JSON arrays, delegating encoding of each element to ITEM.
// Format nil slices as 'null' with json.FormatNilSliceAsNull(true).
type listCodec[T ~[]U, U any, ITEM Codec[U]] struct{}

func (listCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil && getOptionOrFalse(enc.Options(), json.FormatNilSliceAsNull) {
		if err := enc.WriteToken(jsontext.Null); err != nil {
			return WrapEncodeError(enc, "", err)
		}
		return nil
	}
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	for _, item := range receiver {
		if err := (*new(ITEM)).MarshalJSONTo(enc, item); err != nil {
			return err
		}
	}
	if err := enc.WriteToken(jsontext.EndArray); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	return nil
}

func (listCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if *receiver == nil {
		*receiver = make(T, 0)
	} else {
		*receiver = (*receiver)[:0]
	}
	if dec.PeekKind() == jsontext.KindNull {
		if _, err := dec.ReadToken(); err != nil {
			return WrapSyntaxError(dec, "", err)
		}
		return nil
	}
	return VisitJSONListFields(dec, func(dec *jsontext.Decoder) error {
		item := *new(U)
		if err := (*new(ITEM)).UnmarshalJSONFrom(dec, &item); err != nil {
			return err
		}
		*receiver = append(*receiver, item)
		return nil
	})
}

func (listCodec[T, U, ITEM]) Equal(a, b T) bool {
	return slices.EqualFunc(a, b, (*new(ITEM)).Equal)
}
