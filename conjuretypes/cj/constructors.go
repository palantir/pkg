// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"encoding"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/palantir/pkg/binary"
	"github.com/palantir/pkg/datetime"
)

// Any returns a codec that delegates JSON encoding and decoding to the default JSON implementation.
func Any[T any]() (_ anyCodec[T]) { return }

// BearerToken returns a codec for Conjure bearer token values.
func BearerToken[T ~string]() (_ bearerTokenCodec[T]) { return }

// Binary returns a codec for Conjure binary values.
func Binary[T ~[]byte]() (_ binaryCodec[T]) { return }

// BinaryMapKey returns a codec for Conjure binary values used as map keys.
func BinaryMapKey[T binary.Binary]() (_ binaryMapKeyCodec[T]) { return }

// Boolean returns a codec for Conjure boolean values.
func Boolean[T ~bool]() (_ booleanCodec[T]) { return }

// BooleanMapKey returns a codec for Conjure boolean values used as map keys.
func BooleanMapKey[T ~bool]() (_ booleanMapKeyCodec[T]) { return }

// DateTime returns a codec for Conjure datetime values. It also serves as a map
// key codec: keys sort by instant, with same-instant keys broken by their wire
// string so deterministic output is stable even across differing time zones.
func DateTime[T time.Time | datetime.DateTime]() (_ dateTimeCodec[T]) { return }

// Float returns a codec for Conjure double values.
func Float[T ~float64]() (_ floatCodec[T]) { return }

// FloatMapKey returns a codec for Conjure double values used as map keys.
func FloatMapKey[T ~float64]() (_ floatMapKeyCodec[T]) { return }

// Int32 returns a codec for Conjure integer values.
func Int32[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() (_ int32Codec[T]) { return }

// Int32MapKey returns a codec for Conjure integer values used as map keys.
func Int32MapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() (_ int32MapKeyCodec[T]) { return }

// List returns a codec for Conjure list values.
func List[T ~[]U, U any, ITEM Codec[U]](_ ITEM) (_ listCodec[T, U, ITEM]) { return }

// Map returns a codec for maps.
func Map[T ~map[K]V, K comparable, V any, KEY MapKeyCodec[K], VAL Codec[V]](_ KEY, _ VAL) (_ mapCodec[T, K, V, KEY, VAL]) {
	return
}

// Optional returns a codec for Conjure optional values represented as pointers.
func Optional[T *U, U any, ITEM Codec[U]](_ ITEM) (_ optionalCodec[T, U, ITEM]) { return }

// RID returns a codec for Conjure resource identifiers.
func RID[T ridConstraint]() (_ ridCodec[T]) { return }

// SafeLong returns a codec for Conjure safelong values.
func SafeLong[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() (_ safeLongCodec[T]) { return }

// SafeLongMapKey returns a codec for Conjure safelong values used as map keys.
func SafeLongMapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() (_ safeLongMapKeyCodec[T]) { return }

// Set returns a codec for Conjure set values represented as slices.
func Set[T ~[]U, U any, ITEM SetItemCodec[U]](_ ITEM) (_ setCodec[T, U, ITEM]) { return }

// String returns a codec for Conjure string values.
func String[T ~string]() (_ stringCodec[T]) { return }

// Struct returns a codec for generated structs that implement the JSON v2 methods directly.
func Struct[T json.MarshalerTo, U interface {
	*T
	json.UnmarshalerFrom
}]() (_ structCodec[T, U]) {
	return
}

// Text returns a codec for text-marshaled values.
func Text[T encoding.TextMarshaler, U interface {
	*T
	encoding.TextUnmarshaler
}]() (_ textCodec[T, U]) {
	return
}

// UUID returns a codec for Conjure UUID values.
func UUID[T ~[16]byte]() (_ uuidCodec[T]) { return }
