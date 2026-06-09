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

// mapCodec encodes Conjure maps as JSON objects. With json.Deterministic
// (the default), keys are sorted via the key codec's Sort method (see
// MapKeyCodec); disable with json.Deterministic(false). Set
// json.FormatNilMapAsNull(true) to encode nil maps as 'null'.
type mapCodec[T ~map[K]V, K comparable, V any, KEY MapKeyCodec[K], VAL Codec[V]] struct{}

func (mapCodec[T, K, V, KEY, VAL]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	if receiver == nil {
		if formatNull, _ := json.GetOption(enc.Options(), json.FormatNilMapAsNull); formatNull {
			return enc.WriteToken(jsontext.Null)
		}
	}
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
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
	return enc.WriteToken(jsontext.EndObject)
}

func (mapCodec[T, K, V, KEY, VAL]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	if *receiver == nil {
		*receiver = make(T)
	} else {
		clear(*receiver)
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return WrapSyntaxError(dec, err)
	}
	if tok.Kind() == jsontext.KindNull {
		return nil
	}
	if tok.Kind() != jsontext.KindBeginObject {
		return NewKindError[T](dec, tok.Kind(), "object opening brace")
	}
	// Hoisted out of the loop so the heap escape from passing &key/&val to the
	// nested decoders is paid once. The per-iteration reset is required: reusing a
	// reference-typed val would let the decoder mutate a value already stored.
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
			return NewDuplicateMapKeyError[T](dec)
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
