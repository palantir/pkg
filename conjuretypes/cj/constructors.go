// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"encoding"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/palantir/pkg/binary"
	"github.com/palantir/pkg/datetime"
)

// Any returns a codec that delegates JSON encoding and decoding to the default JSON implementation.
func Any[T any]() anyCodec[T] {
	return anyCodec[T]{}
}

// BearerToken returns a codec for Conjure bearer token values.
func BearerToken[T ~string]() bearerTokenCodec[T] {
	return bearerTokenCodec[T]{}
}

// Binary returns a codec for Conjure binary values.
func Binary[T ~[]byte]() binaryCodec[T] {
	return binaryCodec[T]{}
}

// BinaryMapKey returns a codec for Conjure binary values used as map keys.
func BinaryMapKey[T binary.Binary]() binaryMapKeyCodec[T] {
	return binaryMapKeyCodec[T]{}
}

// Boolean returns a codec for Conjure boolean values.
func Boolean[T ~bool]() booleanCodec[T] {
	return booleanCodec[T]{}
}

// BooleanMapKey returns a codec for Conjure boolean values used as map keys.
func BooleanMapKey[T ~bool]() booleanMapKeyCodec[T] {
	return booleanMapKeyCodec[T]{}
}

// ComparableMap returns a codec for maps whose keys need a custom deterministic ordering.
func ComparableMap[T ~map[K]V, K comparable, V any, KEY MapKeyCodec[K], VAL Codec[V]](_ KEY, _ VAL) comparableMapCodec[T, K, V, KEY, VAL] {
	return comparableMapCodec[T, K, V, KEY, VAL]{}
}

// DateTime returns a codec for Conjure datetime values.
func DateTime[T time.Time | datetime.DateTime]() dateTimeCodec[T] {
	return dateTimeCodec[T]{}
}

// Float returns a codec for Conjure double values.
func Float[T ~float64]() floatCodec[T] {
	return floatCodec[T]{}
}

// FloatMapKey returns a codec for Conjure double values used as map keys.
func FloatMapKey[T ~float64]() floatMapKeyCodec[T] {
	return floatMapKeyCodec[T]{}
}

// Int32 returns a codec for Conjure integer values.
func Int32[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() int32Codec[T] {
	return int32Codec[T]{}
}

// Int32MapKey returns a codec for Conjure integer values used as map keys.
func Int32MapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() int32MapKeyCodec[T] {
	return int32MapKeyCodec[T]{}
}

// List returns a codec for Conjure list values.
func List[T ~[]U, U any, ITEM Codec[U]](_ ITEM) listCodec[T, U, ITEM] {
	return listCodec[T, U, ITEM]{}
}

// Optional returns a codec for Conjure optional values represented as pointers.
func Optional[T *U, U any, ITEM Codec[U]](_ ITEM) optionalCodec[T, U, ITEM] {
	return optionalCodec[T, U, ITEM]{}
}

// OrderedMap returns a codec for maps whose keys can be sorted with cmp.Ordered.
func OrderedMap[T ~map[K]V, K cmp.Ordered, V any, KEY Codec[K], VAL Codec[V]](_ KEY, _ VAL) orderedMapCodec[T, K, V, KEY, VAL] {
	return orderedMapCodec[T, K, V, KEY, VAL]{}
}

// RID returns a codec for Conjure resource identifiers.
func RID[T ridConstraint]() ridCodec[T] {
	return ridCodec[T]{}
}

// SafeLong returns a codec for Conjure safelong values.
func SafeLong[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() safeLongCodec[T] {
	return safeLongCodec[T]{}
}

// SafeLongMapKey returns a codec for Conjure safelong values used as map keys.
func SafeLongMapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() safeLongMapKeyCodec[T] {
	return safeLongMapKeyCodec[T]{}
}

// Set returns a codec for Conjure set values represented as slices.
func Set[T ~[]U, U any, ITEM Codec[U]](_ ITEM) setCodec[T, U, ITEM] {
	return setCodec[T, U, ITEM]{}
}

// String returns a codec for Conjure string values.
func String[T ~string]() stringCodec[T] {
	return stringCodec[T]{}
}

// Struct returns a codec for generated structs that implement the JSON v2 methods directly.
func Struct[T json.MarshalerTo, U interface {
	*T
	json.UnmarshalerFrom
}]() structCodec[T, U] {
	return structCodec[T, U]{}
}

// Text returns a codec for text-marshaled values.
func Text[T encoding.TextMarshaler, U interface {
	*T
	encoding.TextUnmarshaler
}]() textCodec[T, U] {
	return textCodec[T, U]{}
}

// UUID returns a codec for Conjure UUID values.
func UUID[T ~[16]byte]() uuidCodec[T] {
	return uuidCodec[T]{}
}
