// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"github.com/go-json-experiment/json/jsontext"
)

// Codec is implemented by types that can encode and decode a Go value of type T to JSON using the
// provided jsontext.Encoder and jsontext.Decoder. Constructors for each Conjure type
// (e.g., Boolean, Int32, String, List, Map, etc.) are provided in this package.
// Each implementation ensures correct marshaling of the corresponding Go type to the appropriate JSON representation.
// Implementations' zero values must be valid for use by container encoders.
type Codec[T any] interface {
	// MarshalJSONTo can be passed to json.MarshalToFunc or used to implement json.MarshalerTo.
	MarshalJSONTo(enc *jsontext.Encoder, receiver T) error
	// UnmarshalJSONFrom can be passed to json.UnmarshalFromFunc or used to implement json.UnmarshalerFrom.
	UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error
	// Equal returns whether the two values can be considered equal for the purposes of map and set uniqueness.
	Equal(T, T) bool
}

// MapKeyCodec is implemented by types that can be used as map keys in Conjure types
// but do not implement cmp.Ordered. The encoder's Compare method is used to sort map keys in a deterministic order.
// This is used for types like UUID, binary, datetime, etc. that need custom comparison logic for ordering.
type MapKeyCodec[K comparable] interface {
	Codec[K]
	// Compare returns -1 if a < b, 0 if a == b, and 1 if a > b.
	// This is used to sort map keys and set elements in a deterministic order for types that are comparable
	// but don't support ordering operators like <, >, ==.
	Compare(K, K) int
}
