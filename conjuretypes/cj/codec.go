// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
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

// MapKeyCodec is implemented by codecs for types used as Conjure map keys.
// Sort orders a slice of keys in place to produce deterministic map output.
//
// Sort lives on the codec so a single map codec can serve every key type without
// losing performance: cmp.Ordered keys (integer, safelong, double, string-backed
// binary) sort via the fast slices.Sort path, while keys that are comparable but
// not ordered (uuid, datetime, boolean) sort via their own comparison. A map
// codec parameterized only by K comparable could not call slices.Sort itself, so
// it would otherwise be forced onto the slower comparator path for every key.
type MapKeyCodec[K comparable] interface {
	Codec[K]
	// Sort orders keys in place so that map encoding is deterministic.
	Sort(keys []K)
}

// SetItemCodec is implemented by codecs for types used as Conjure set elements.
// Sets deduplicate on both encode and decode, an O(n^2) scan, so Contains lives
// on the codec rather than the set: an element type whose equality is Go's ==
// (every comparableCodec) deduplicates via slices.Contains with an inlined ==,
// while types with a custom Equal (datetime, binary, struct, ...) fall back to a
// scan over that Equal. A set codec parameterized only by U could not reach
// slices.Contains itself, so it would be forced onto the slower Equal path for
// every element. This mirrors MapKeyCodec.Sort.
type SetItemCodec[T any] interface {
	Codec[T]
	// Contains reports whether set already holds item, used to drop duplicates.
	Contains(set []T, item T) bool
}

// fillGoType records T on a SemanticError that does not yet name a Go type.
// Without it, json/v2 would populate GoType with this internal codec wrapper
// rather than the value being (un)marshaled. T is the root type passed to the
// entry point; the error's JSON pointer locates failures within nested values.
func fillGoType[T any](err error) error {
	if serr, ok := errors.AsType[*json.SemanticError](err); ok && serr.GoType == nil {
		serr.GoType = reflect.TypeFor[T]()
	}
	return err
}

type anonymousMarshaler[T any, E Codec[T]] struct {
	receiver T
}

func (a *anonymousMarshaler[T, E]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return fillGoType[T]((*new(E)).MarshalJSONTo(enc, a.receiver))
}

type anonymousUnmarshaler[T any, E Codec[T]] struct {
	receiver *T
}

func (a anonymousUnmarshaler[T, E]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if a.receiver == nil {
		return fmt.Errorf("cj.NewAnonymousType: cannot unmarshal into nil receiver")
	}
	return fillGoType[T]((*new(E)).UnmarshalJSONFrom(dec, a.receiver))
}

func Unmarshal[T any, D Codec[T]](data []byte, v *T, _ D, opts ...json.Options) error {
	// AllowDuplicateNames lets the codec (not the jsontext syntax layer) detect
	// duplicate object members, so map codecs can report the richer
	// ErrDuplicateMapKey and catch canonicalized duplicates (e.g. "01" and "1").
	if len(opts) == 0 {
		return json.Unmarshal(data, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.Unmarshal(data, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func UnmarshalRead[T any, D Codec[T]](r io.Reader, v *T, _ D, opts ...json.Options) error {
	if len(opts) == 0 {
		return json.UnmarshalRead(r, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.UnmarshalRead(r, &anonymousUnmarshaler[T, D]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func ClientDecoder[T any, D Codec[T]](_ D) clientDecoder[T, D] { return clientDecoder[T, D]{} }

func ServerDecoder[T any, D Codec[T]](_ D) serverDecoder[T, D] { return serverDecoder[T, D]{} }

func Marshal[T any, E Codec[T]](v T, _ E, opts ...json.Options) ([]byte, error) {
	// Pass a pointer: json/v2 makes a reflect.New copy of any non-pointer value to
	// obtain an addressable value, costing an extra allocation per call.
	//
	// AllowDuplicateNames skips json/v2's per-member duplicate-name validation,
	// which the codecs do not need: map keys are unique by construction and
	// generated object codecs emit each field name once.
	if len(opts) == 0 {
		return json.Marshal(&anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.Marshal(&anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func MarshalWrite[T any, E Codec[T]](out io.Writer, v T, _ E, opts ...json.Options) error {
	if len(opts) == 0 {
		return json.MarshalWrite(out, &anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true))
	}
	return json.MarshalWrite(out, &anonymousMarshaler[T, E]{v}, jsontext.AllowDuplicateNames(true), json.JoinOptions(opts...))
}

func ClientEncoder[T any, E Codec[T]](_ E) defaultEncoder[T, E] { return defaultEncoder[T, E]{} }

func ServerEncoder[T any, E Codec[T]](_ E) defaultEncoder[T, E] { return defaultEncoder[T, E]{} }

type clientDecoder[T any, D Codec[T]] struct{}

func (clientDecoder[T, D]) Accept() string {
	return "application/json"
}

func (clientDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D))
	}
	return json.UnmarshalRead(r, v, json.WithUnmarshalers(defaultUnmarshalers))
}

func (clientDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D))
	}
	return json.Unmarshal(data, v, json.WithUnmarshalers(defaultUnmarshalers))
}

type serverDecoder[T any, D Codec[T]] struct{}

func (serverDecoder[T, D]) Accept() string {
	return "application/json"
}

func (serverDecoder[T, D]) Decode(r io.Reader, v any) error {
	if receiver, ok := v.(*T); ok {
		return UnmarshalRead(r, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.UnmarshalRead(r, v, json.WithUnmarshalers(defaultUnmarshalers), json.RejectUnknownMembers(true))
}

func (serverDecoder[T, D]) Unmarshal(data []byte, v any) error {
	if receiver, ok := v.(*T); ok {
		return Unmarshal(data, receiver, *new(D), json.RejectUnknownMembers(true))
	}
	return json.Unmarshal(data, v, json.WithUnmarshalers(defaultUnmarshalers), json.RejectUnknownMembers(true))
}

type defaultEncoder[T any, E Codec[T]] struct{}

func (defaultEncoder[T, E]) ContentType() string {
	return "application/json"
}

func (defaultEncoder[T, E]) Encode(w io.Writer, v any) error {
	if value, ok := v.(T); ok {
		return MarshalWrite(w, value, *new(E))
	}
	return json.MarshalWrite(w, v, json.WithMarshalers(defaultMarshalers))
}

func (defaultEncoder[T, E]) Marshal(v any) ([]byte, error) {
	if value, ok := v.(T); ok {
		return Marshal(value, *new(E))
	}
	return json.Marshal(v, json.WithMarshalers(defaultMarshalers))
}

// comparableCodec is embedded by scalar codecs whose values are equal exactly
// when Go's == says so (numbers, bool, string, and the like). It supplies the
// Equal method (==) and the SetItemCodec.Contains method (slices.Contains, an
// inlined == scan) so each codec does not repeat the same one-liners.
//
// There is deliberately no cmp.Ordered-based Compare mixin: ordered keys sort
// through orderedKeyCodec's slices.Sort path (see MapKeyCodec), and the key
// types that do need a Compare method (datetime, uuid, boolean, ...) all define
// a custom one, so a cmp.Compare wrapper would never be used.
type comparableCodec[T comparable] struct{}

func (comparableCodec[T]) Equal(a, b T) bool { return a == b }

func (comparableCodec[T]) Contains(set []T, item T) bool { return slices.Contains(set, item) }

// orderedKeyCodec is embedded by map-key codecs whose key type is cmp.Ordered.
// It supplies Equal (via comparableCodec) and the MapKeyCodec.Sort method backed
// by slices.Sort, which is faster than a comparator-driven sort. Key types that
// are comparable but not ordered implement Sort themselves over their Compare.
type orderedKeyCodec[K cmp.Ordered] struct{ comparableCodec[K] }

func (orderedKeyCodec[K]) Sort(keys []K) { slices.Sort(keys) }
