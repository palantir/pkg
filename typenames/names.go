// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package typenames generates human-friendly names for go objects using reflection.
package typenames

import (
	"fmt"
	"reflect"
	"strings"
)

// LongName is like "map of string to object".
// Accepts reflect.Value, reflect.Type, or interface{}. You should pass
// whichever one is available. This is to avoid accidental panics from code that
// naively calls value.Type() or reflect.ValueOf(obj).Type().
func LongName(valueOrTypeOrObj interface{}) string {
	prefix, typ := simplifiedType(valueOrTypeOrObj)
	if typ == nil {
		return "null"
	}
	var name string
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		name = fmt.Sprintf("array of %v", ShortName(typ.Elem()))
	case reflect.Map:
		name = fmt.Sprintf("map of %v to %v", ShortName(typ.Key()), ShortName(typ.Elem()))
	case reflect.Chan:
		name = fmt.Sprintf("Go channel of %v", ShortName(typ.Elem()))
	default:
		name = typeNameHelper(typ)
	}
	return prefix + name
}

// ShortName is like "map".
// Like LongName, accepts reflect.Value, reflect.Type, or interface{}.
func ShortName(valueOrTypeOrObj interface{}) string {
	prefix, typ := simplifiedType(valueOrTypeOrObj)
	if typ == nil {
		return "null"
	}
	var name string
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		name = "array"
	case reflect.Map:
		name = "map"
	case reflect.Chan:
		name = "Go channel"
	default:
		name = typeNameHelper(typ)
	}
	return prefix + name
}

func simplifiedType(valueOrTypeOrObj interface{}) (prefix string, simple reflect.Type) {
	var typ reflect.Type
	switch v := valueOrTypeOrObj.(type) {
	case reflect.Value:
		if v.IsValid() {
			typ = v.Type()
		}
	case reflect.Type:
		typ = v
	default:
		typ = reflect.TypeOf(valueOrTypeOrObj)
	}

	if typ == nil {
		return "", nil
	}

	typ = reflect.PtrTo(typ) // so big.Int -> *big.Int
	for typ.Kind() == reflect.Ptr && !IsBigIntType(typ) && !IsBigFloatType(typ) {
		typ = typ.Elem()
		prefix += "*"
	}
	prefix = strings.TrimPrefix(prefix, "*") // to make up for reflect.PtrTo
	return prefix, typ
}

func typeNameHelper(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "unsigned integer"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Complex64, reflect.Complex128:
		return "complex number"
	case reflect.Func:
		return "function"
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return "object"
		}
	case reflect.Ptr:
		if IsBigIntType(typ) {
			return "integer"
		}
		if IsBigFloatType(typ) {
			return "float"
		}
		return "*" + ShortName(typ.Elem())
	case reflect.String:
		if IsJSONNumberType(typ) {
			return "number"
		}
	case reflect.Struct:
		if typ.Name() == "" {
			return "struct"
		}
		return fmt.Sprintf("struct %v", typ.Name())
	}
	if typ.Name() == "" {
		return typ.String()
	}
	return typ.Name()
}
