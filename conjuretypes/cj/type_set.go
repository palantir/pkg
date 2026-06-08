// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// setCodec provides json marshaling for sets of type T using a nested encoder ITEM.
// It writes a json array, delegating encoding of each element to ITEM.
//
// Encoded duplicate items in the receiver slice (as determined by ITEM.Contains) are skipped.
// The emitted JSON list's elements will otherwise be in the same order as the original.
//
// Decoded duplicate items (as determined by ITEM.Contains) result in an error wrapping ErrDuplicateSetItem.
//
// If the receiver is nil and json.FormatNilSliceAsNull is true, the JSON null value is written.
type setCodec[T ~[]U, U any, ITEM SetItemCodec[U]] struct{}

func (setCodec[T, U, ITEM]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if *receiver == nil {
		*receiver = make(T, 0)
	} else {
		*receiver = (*receiver)[:0]
	}
	if dec.PeekKind() == jsontext.KindNull {
		if _, err := dec.ReadToken(); err != nil {
			return WrapSyntaxError(dec, err)
		}
		return nil
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() != jsontext.KindBeginArray {
		return newKindMismatchTokenError(dec, tok, "set opening bracket")
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
		item := &(*receiver)[len(*receiver)-1]
		if err := (*new(ITEM)).UnmarshalJSONFrom(dec, item); err != nil {
			return err
		}
		if (*new(ITEM)).Contains((*receiver)[:len(*receiver)-1], *item) {
			return NewDuplicateSetItemError(dec)
		}
	}
}

func (setCodec[T, U, ITEM]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if formatNull, _ := json.GetOption(enc.Options(), json.FormatNilSliceAsNull); formatNull {
			if err := enc.WriteToken(jsontext.Null); err != nil {
				return WrapEncodeError(enc, "", err)
			}
			return nil
		}
	}
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	for i, item := range receiver {
		if (*new(ITEM)).Contains(receiver[0:i], item) {
			continue
		}
		if err := (*new(ITEM)).MarshalJSONTo(enc, item); err != nil {
			return err
		}
	}
	if err := enc.WriteToken(jsontext.EndArray); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	return nil
}

func (setCodec[T, U, ITEM]) Equal(a, b T) bool {
	// Set equality is independent of order and multiplicity: two sets are equal
	// iff they contain the same distinct elements. This matches the codec, which
	// drops duplicates on marshal and rejects them on decode, so a duplicate-bearing
	// slice represents the same set as its deduplicated form.
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
