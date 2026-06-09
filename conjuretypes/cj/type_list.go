// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// listCodec encodes slices of type T as JSON arrays, delegating each element to ITEM.
// A nil receiver is written as null when json.FormatNilSliceAsNull is true.
type listCodec[T ~[]U, U any, ITEM Codec[U]] struct{}

func (listCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if formatNull, _ := json.GetOption(enc.Options(), json.FormatNilSliceAsNull); formatNull {
			return enc.WriteToken(jsontext.Null)
		}
	}
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	for _, item := range receiver {
		if err := (*new(ITEM)).MarshalJSONTo(enc, item); err != nil {
			return err
		}
	}
	return enc.WriteToken(jsontext.EndArray)
}

func (listCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if *receiver == nil {
		*receiver = make(T, 0)
	} else {
		*receiver = (*receiver)[:0]
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() == jsontext.KindNull {
		return nil
	}
	if tok.Kind() != jsontext.KindBeginArray {
		return NewKindError[T](dec, tok.Kind(), "list opening bracket")
	}
	for {
		if dec.PeekKind() == jsontext.KindEndArray {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, err)
			}
			return nil
		}
		// Grow into the backing array and decode in place so the element does
		// not escape to the heap via the &item passed to the nested decoder.
		*receiver = append(*receiver, *new(U))
		if err := (*new(ITEM)).UnmarshalJSONFrom(dec, &(*receiver)[len(*receiver)-1]); err != nil {
			return err
		}
	}
}

func (listCodec[T, U, ITEM]) Equal(a, b T) bool {
	return slices.EqualFunc(a, b, (*new(ITEM)).Equal)
}

func (c listCodec[T, U, ITEM]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
