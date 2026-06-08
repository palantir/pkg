// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"maps"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// mapCodec provides JSON marshaling for Conjure maps. Encodes maps as JSON
// objects; when deterministic encoding is enabled, keys are sorted in place by
// the key codec's Sort method (see MapKeyCodec) before writing.
//
// Disable sorting with json.Deterministic(false).
// Format nil maps as 'null' with json.FormatNilMapAsNull(true).
type mapCodec[T ~map[K]V, K comparable, V any, KEY MapKeyCodec[K], VAL Codec[V]] struct{}

// When deterministic encoding is enabled, keys are materialized and ordered with the key codec's Sort
// before writing; otherwise the map is iterated in Go's native order.
func (mapCodec[T, K, V, KEY, VAL]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if formatNull, _ := json.GetOption(enc.Options(), json.FormatNilMapAsNull); formatNull {
			if err := enc.WriteToken(jsontext.Null); err != nil {
				return WrapEncodeError(enc, "", err)
			}
			return nil
		}
	}
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	if deterministic, ok := json.GetOption(enc.Options(), json.Deterministic); deterministic || !ok {
		sortedKeys := make([]K, 0, len(receiver))
		for k := range receiver {
			sortedKeys = append(sortedKeys, k)
		}
		(*new(KEY)).Sort(sortedKeys)
		for _, k := range sortedKeys {
			if err := (*new(KEY)).MarshalJSONTo(enc, k); err != nil {
				return err
			}
			if err := (*new(VAL)).MarshalJSONTo(enc, receiver[k]); err != nil {
				return err
			}
		}
	} else {
		for k, v := range receiver {
			if err := (*new(KEY)).MarshalJSONTo(enc, k); err != nil {
				return err
			}
			if err := (*new(VAL)).MarshalJSONTo(enc, v); err != nil {
				return err
			}
		}
	}
	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	return nil
}

func (mapCodec[T, K, V, KEY, VAL]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if *receiver == nil {
		*receiver = make(T)
	} else {
		clear(*receiver)
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
	if tok.Kind() != jsontext.KindBeginObject {
		return newKindMismatchTokenError(dec, tok, "object opening brace")
	}
	// key and val are hoisted out of the loop: &key / &val escape to the heap when
	// passed to the nested decoders, so declaring them once pays that escape once
	// rather than per entry. The reset each iteration is required, not just tidy: a
	// reused reference-typed val (slice, map, pointer) would otherwise be mutated in
	// place by the nested decoder, corrupting the value already stored in the map.
	var key K
	var val V
	for {
		if dec.PeekKind() == jsontext.KindEndObject {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, err)
			}
			return nil
		}
		key, val = *new(K), *new(V)
		if err := (*new(KEY)).UnmarshalJSONFrom(dec, &key); err != nil {
			return err
		}
		if err := (*new(VAL)).UnmarshalJSONFrom(dec, &val); err != nil {
			return err
		}
		if _, ok := (*receiver)[key]; ok {
			return NewDuplicateMapKeyError(dec)
		}
		(*receiver)[key] = val
	}
}

func (mapCodec[T, K, V, KEY, VAL]) Equal(a, b T) bool {
	return maps.EqualFunc(a, b, (*new(VAL)).Equal)
}

func (c mapCodec[T, K, V, KEY, VAL]) Contains(set []T, item T) bool {
	return slices.ContainsFunc(set, func(x T) bool { return c.Equal(item, x) })
}
