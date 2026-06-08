// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"maps"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// orderedMapCodec provides JSON marshaling for maps with ordered keys (strings and numbers).
// Encodes maps as JSON objects, sorting keys using Go's cmp.Ordered rules, and delegates encoding of keys and values.
//
// Disable sorting with json.Deterministic(false).
// Format nil maps as 'null' with json.FormatNilMapAsNull(true).
type orderedMapCodec[T ~map[K]V, K cmp.Ordered, V any, KEY Codec[K], VAL Codec[V]] struct{}

func (orderedMapCodec[T, K, V, KEY, VAL]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return mapMarshalJSONTo[T, K, V, KEY, VAL](enc, receiver, func(keys []K) {
		slices.Sort(keys)
	})
}

func (orderedMapCodec[T, K, V, KEY, VAL]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	return mapUnmarshalJSONFrom[T, K, V, KEY, VAL](dec, receiver)
}

func (orderedMapCodec[T, K, V, KEY, VAL]) Equal(a, b T) bool {
	return maps.EqualFunc(a, b, (*new(VAL)).Equal)
}

// comparableMapCodec provides JSON marshaling for maps using a custom key comparison function.
// Encodes maps as JSON objects, sorting keys using cj.MapKeyCodec's Compare method from KEY,
// and delegates encoding of keys and values.
//
// Types compatible with OrderedMap should likely use that unless non-standard sorting is required.
//
// Disable sorting with json.Deterministic(false).
// Format nil maps as 'null' with json.FormatNilMapAsNull(true).
type comparableMapCodec[T ~map[K]V, K comparable, V any, KEY MapKeyCodec[K], VAL Codec[V]] struct{}

func (comparableMapCodec[T, K, V, KEY, VAL]) MarshalJSONTo(enc *jsontext.Encoder, receiver T) error {
	return mapMarshalJSONTo[T, K, V, KEY, VAL](enc, receiver, func(keys []K) {
		slices.SortFunc(keys, (*new(KEY)).Compare)
	})
}

func (comparableMapCodec[T, K, V, KEY, VAL]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	return mapUnmarshalJSONFrom[T, K, V, KEY, VAL](dec, receiver)
}

func (comparableMapCodec[T, K, V, KEY, VAL]) Equal(a, b T) bool {
	return maps.EqualFunc(a, b, (*new(VAL)).Equal)
}

// mapMarshalJSONTo provides JSON marshaling for maps, using nested KEY and VAL encoders for keys and values.
// When deterministic encoding is enabled, keys are materialized and ordered with sortKeys before writing;
// otherwise the map is iterated in Go's native order.
func mapMarshalJSONTo[T ~map[K]V, K comparable, V any, KEY Codec[K], VAL Codec[V]](enc *jsontext.Encoder, receiver T, sortKeys func([]K)) error {
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
		sortKeys(sortedKeys)
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

// mapUnmarshalJSONFrom provides JSON unmarshaling for maps, using nested KEY and VAL decoders for keys and values.
// Decodes JSON objects into Go maps of the specified types.
func mapUnmarshalJSONFrom[T ~map[K]V, K comparable, V any, KEY Codec[K], VAL Codec[V]](dec *jsontext.Decoder, receiver *T) error {
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
	for {
		if dec.PeekKind() == jsontext.KindEndObject {
			if _, err := dec.ReadToken(); err != nil {
				return WrapSyntaxError(dec, err)
			}
			return nil
		}
		key := *new(K)
		if err := (*new(KEY)).UnmarshalJSONFrom(dec, &key); err != nil {
			return err
		}
		val := *new(V)
		if err := (*new(VAL)).UnmarshalJSONFrom(dec, &val); err != nil {
			return err
		}
		if _, ok := (*receiver)[key]; ok {
			return NewDuplicateMapKeyError(dec, reflect.TypeFor[T]().String())
		}
		(*receiver)[key] = val
	}
}
