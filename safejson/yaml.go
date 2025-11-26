// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ToMapWithJSONMarshalableKeys returns a version of the provided map where the map keys are known to be safe to marshal
// as JSON. Specifically, for both the provided map and any map contained within it recursively, the returned map types
// will be map[string]any, where the keys in the map are the JSON string representations of the previous keys. Returns
// an error if any of the maps have a key type that is not supported as a JSON object key type.
//
// Assumes that the input consists of only maps, slices, arrays, primitives, and pointers to these types. Structs are
// assumed to have been converted into a map representation: if a struct value is encountered, it will be treated as a
// primitive and left as-is (in particular, any maps in the struct will not be converted).
//
// The most common use case for this function is to convert the output of a yaml.Unmarshal function (which may unmarshal
// all maps into map types of map[any]any) to a form that can be marshalled using json.Marshal. This is necessary
// because json.Marshal will fail when encountering any map[any]any types, even if the actual keys that are present in
// the map are all of a type that could be used as JSON object keys.
//
// This function is logically equivalent to the FromYAMLValue function, but its naming, types, and documentation are
// improved to reflect the actual functionality.
func ToMapWithJSONMarshalableKeys[K comparable, V any](in map[K]V) (map[string]any, error) {
	value, err := fromValue(reflect.ValueOf(in), "")
	if err != nil {
		return nil, err
	}
	mapValue, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", value)
	}
	return mapValue, nil
}

// FromYAMLValue returns a version of the provided input where all nested map[any]any values are
// converted to map[string]any. The supported key types are the same as those supported by encoding/json: string, int
// variants (int/uint/etc.), pointers to such values, and types that implement the encoding.TextMarshaler interface. For
// any map where the key type is one of the supported non-string types, the returned map's key will be the string value
// for the type obtained in the same manner used by encoding/json.
//
// The input should be the representation of an object in map[any]any form (as opposed to a string or []byte of the
// YAML itself). Returns an error if the conversion fails because any of the map keys are not representable as strings.
//
// Assumes that the input consists of only maps, slices, arrays, primitives, and pointers to these types. Structs are
// assumed to have been converted into a map representation: if a struct value is encountered, it will be treated as a
// primitive and left as-is (in particular, any maps in the struct will not be converted).
//
// Many YAML libraries unmarshal YAML content as map[any]any, but the Go JSON library requires JSON maps to be
// represented as map[string]any, so this function is helpful in converting the former to the latter.
func FromYAMLValue(y any) (any, error) {
	return fromValue(reflect.ValueOf(y), "")
}

// fromValue takes the provided value and returns a logically equivalent value where maps in the value are converted to
// be map[string]any, where the key is the JSON string representation of the previous key. This is done recursively for
// all map values, elements of slices or arrays, and for pointer types. Returns an error if there are any reachable maps
// where the key type is not a supported type for JSON object keys. Note that this function does *not* apply this
// transformation to fields of structs: the intended input to this function is a structure that has been decoded from
// a representation like YAML into a series of maps, slices, and primitives.
func fromValue(v reflect.Value, path string) (any, error) {
	switch v.Kind() {
	case reflect.Map:
		return fromMap(v, path)
	case reflect.Slice, reflect.Array:
		return fromSliceOrArray(v, path)
	case reflect.Interface, reflect.Ptr:
		return fromValue(v.Elem(), path)
	case reflect.Invalid:
		return nil, nil
	default:
		return v.Interface(), nil
	}
}

// fromMap takes the provided value (which must be known to be a map type: map[K]V) and returns a logically equivalent
// map[string]any where the keys in the map (and in all maps contained in the values of the map, recursively) are the
// JSON string representation of the previous key value. Returns an error if any of the maps have a key type that is
// not supported as a JSON object key type.
func fromMap(v reflect.Value, path string) (map[string]any, error) {
	m := make(map[string]any, v.Len())
	for _, entry := range v.MapKeys() {
		k, err := convertObjectKeyToString(entry, path)
		if err != nil {
			return nil, err
		}
		v, err := fromValue(v.MapIndex(entry), fmt.Sprintf("%s.%s", path, k))
		if err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

// fromMap takes the provided value (which must be known to be a slice or array type: []V) and returns a logically
// equivalent []any where, for any maps in the values (and in all maps contained in the values of such maps,
// recursively) are the JSON string representation of the previous key value. Returns an error if any of the maps have a
// key type that is not supported as a JSON object key type.
func fromSliceOrArray(v reflect.Value, path string) ([]any, error) {
	a := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		v, err := fromValue(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return nil, err
		}
		a[i] = v
	}
	return a, nil
}

// Based on logic in Go standard library's encoding/json/encode.go: https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=416
var textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()

// convertObjectKeyToString converts the provided value to its string representation. The type of the provided value
// must be a type that is valid to use as a key in a JSON object: string, int/uint variants, or a type that implements
// the encoding.TextMarshaler interface. The conversion of the value to a string is done in the same manner that it
// would be performed by json.Marshal if the value were a key in an object. Returns an error if the provided value is
// not of a type that can be used as a key in a JSON object. The path argument is only used in error output.
func convertObjectKeyToString(k reflect.Value, path string) (string, error) {
	switch k.Kind() {
	case reflect.String:
		return k.String(), nil
	// types match those checked in logic in Go standard library's encoding/json/encode.go: https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=823-825
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// matches logic in https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=556
		return strconv.FormatInt(k.Int(), 10), nil
	// types match those checked in logic in Go standard library's encoding/json/encode.go: https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=823-825
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// matches logic in https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=564
		return strconv.FormatUint(k.Uint(), 10), nil
	case reflect.Interface, reflect.Ptr:
		return convertObjectKeyToString(k.Elem(), path)
	default:
		// if type implements encoding.TextMarshaler, use it.
		// Matches logic at https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=994-1000
		if k.Type().Implements(textMarshalerType) {
			if m, ok := k.Interface().(encoding.TextMarshaler); ok {
				text, err := m.MarshalText()
				if err != nil {
					return "", fmt.Errorf("failed to marshal value %v at path %s as test using MarshalText: %w", k, path, err)
				}
				return string(text), nil
			}
		}
		return "", expectedString(k, path)
	}
}

func expectedString(k reflect.Value, path string) error {
	var valStr string
	if k.IsValid() {
		valStr = fmt.Sprintf("%v: %v", k.Type(), k.Interface())
	} else {
		valStr = "null"
	}
	path = strings.TrimPrefix(path, ".") // no leading dot

	msg := "expected map key"
	if path != "" {
		msg += fmt.Sprintf(" inside %s", path)

	}
	return fmt.Errorf("%s to be a valid key type (string, number, TextMarshaler) but was %s", msg, valStr)
}
