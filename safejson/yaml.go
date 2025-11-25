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
	return fromYAMLValue(reflect.ValueOf(y), "")
}

func fromYAMLValue(v reflect.Value, path string) (any, error) {
	switch v.Kind() {
	case reflect.Map:
		return fromYAMLMap(v, path)
	case reflect.Slice, reflect.Array:
		return fromYAMLArray(v, path)
	case reflect.Interface, reflect.Ptr:
		return fromYAMLValue(v.Elem(), path)
	case reflect.Invalid:
		return nil, nil
	default:
		return v.Interface(), nil
	}
}

func fromYAMLMap(v reflect.Value, path string) (any, error) {
	m := make(map[string]any, v.Len())
	for _, entry := range v.MapKeys() {
		k, err := fromYAMLKey(entry, path)
		if err != nil {
			return nil, err
		}
		v, err := fromYAMLValue(v.MapIndex(entry), fmt.Sprintf("%s.%s", path, k))
		if err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

func fromYAMLArray(v reflect.Value, path string) (any, error) {
	a := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		v, err := fromYAMLValue(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return nil, err
		}
		a[i] = v
	}
	return a, nil
}

// Based on logic in Go standard library's encoding/json/encode.go: https://cs.opensource.google/go/go/+/refs/tags/go1.25.4:src/encoding/json/encode.go;l=416
var textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()

func fromYAMLKey(k reflect.Value, path string) (string, error) {
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
		return fromYAMLKey(k.Elem(), path)
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
