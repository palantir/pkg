// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// setCodec marshals a set of type T as a JSON array, delegating each element to ITEM.
//
// Duplicates (per ITEM.Contains) are dropped on marshal, preserving the order of the
// first occurrence, and rejected on unmarshal with an error wrapping ErrDuplicateSetItem.
//
// A nil receiver is written as JSON null when json.FormatNilSliceAsNull is set.
type setCodec[T ~[]U, U any, ITEM SetItemCodec[U]] struct{}

func (setCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
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
		return NewKindError[T](dec, tok.Kind(), "set opening bracket")
	}
	for {
		if dec.PeekKind() == jsontext.KindEndArray {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, err)
			}
			return nil
		}
		// Decode in place so &item does not force the element to escape to the heap.
		*receiver = append(*receiver, *new(U))
		item := &(*receiver)[len(*receiver)-1]
		if err := (*new(ITEM)).UnmarshalJSONFrom(dec, item); err != nil {
			return err
		}
		if (*new(ITEM)).Contains((*receiver)[:len(*receiver)-1], *item) {
			return NewDuplicateSetItemError[T](dec)
		}
	}
}

func (setCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if formatNull, _ := json.GetOption(enc.Options(), json.FormatNilSliceAsNull); formatNull {
			return enc.WriteToken(jsontext.Null)
		}
	}
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	for i, item := range receiver {
		if (*new(ITEM)).Contains(receiver[0:i], item) {
			continue
		}
		if err := (*new(ITEM)).MarshalJSONTo(enc, item); err != nil {
			return err
		}
	}
	return enc.WriteToken(jsontext.EndArray)
}

func (setCodec[T, U, ITEM]) Equal(a, b T) bool {
	// Sets are equal iff they hold the same distinct elements, ignoring order and
	// multiplicity.
	containsAll := func(haystack, needles T) bool {
		for _, n := range needles {
			if !(*new(ITEM)).Contains(haystack, n) {
				return false
			}
		}
		return true
	}
	return containsAll(a, b) && containsAll(b, a)
}

func (c setCodec[T, U, ITEM]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
