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
	if receiver == nil && getOptionOrFalse(enc.Options(), json.FormatNilMapAsNull) {
		if err := enc.WriteToken(jsontext.Null); err != nil {
			return WrapEncodeError(enc, "", err)
		}
		return nil
	}
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	if getOptionOrTrue(enc.Options(), json.Deterministic) {
		sortedKeys := make([]K, 0, len(receiver))
		for k := range receiver {
			sortedKeys = append(sortedKeys, k)
		}
		slices.Sort(sortedKeys)
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
	if receiver == nil && getOptionOrFalse(enc.Options(), json.FormatNilMapAsNull) {
		if err := enc.WriteToken(jsontext.Null); err != nil {
			return WrapEncodeError(enc, "", err)
		}
		return nil
	}
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return WrapEncodeError(enc, "", err)
	}
	if getOptionOrTrue(enc.Options(), json.Deterministic) {
		sortedKeys := make([]K, 0, len(receiver))
		for k := range receiver {
			sortedKeys = append(sortedKeys, k)
		}
		slices.SortFunc(sortedKeys, (*new(KEY)).Compare)
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

func (comparableMapCodec[T, K, V, KEY, VAL]) UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error {
	return mapUnmarshalJSONFrom[T, K, V, KEY, VAL](dec, receiver)
}

func (comparableMapCodec[T, K, V, KEY, VAL]) Equal(a, b T) bool {
	return maps.EqualFunc(a, b, (*new(VAL)).Equal)
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
			return WrapSyntaxError(dec, "", err)
		}
		return nil
	}
	return VisitJSONObjectFields(dec, func(dec *jsontext.Decoder) error {
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
		return nil
	})
}
