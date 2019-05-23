// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merge

import (
	"fmt"
	"reflect"
)

// Maps returns a new map that is the result of merging the two provided inputs, which must both be maps. Returns an
// error if either of the inputs are not maps. If the types of the input values differ, an error is returned.
// Merging is performed by creating a new map, setting its contents to be "dest", and then setting the key/value pairs in
// "src" on the new map (unless the value is a map, in which case a merge is performed recursively).
func Maps(dest, src interface{}) (interface{}, error) {
	result, err := mergeMaps(reflect.ValueOf(dest), reflect.ValueOf(src))
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

// mergeMaps requires both inputs to be maps; if not, an error is returned. If both input maps have the same type,
// the returned map has the same type as well. If the input maps have different
// types, an error is returned. Otherwise, a new map is created and populated
// with the merge result for the return value. For map entries with the same key,
// the determineValue helper method is used to determine the resulting value for the key.
// Entries with nil values are preserved in the map.
func mergeMaps(dest, src reflect.Value) (reflect.Value, error) {
	if dest.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("expected destination to be a map")
	}
	if src.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("expected source be a map")
	}

	if dest.Type() != src.Type() {
		return reflect.Value{}, fmt.Errorf("expected maps of same type")
	}
	result := reflect.MakeMap(dest.Type())
	for _, destKey := range dest.MapKeys() {
		result.SetMapIndex(destKey, dest.MapIndex(destKey))
	}
	for _, srcKey := range src.MapKeys() {
		srcVal := src.MapIndex(srcKey)
		destVal := dest.MapIndex(srcKey)
		var resultVal reflect.Value
		var err error
		if !destVal.IsValid() {
			if safeIsNil(srcVal) {
				result.SetMapIndex(srcKey, srcVal)
				continue
			}
			resultVal = srcVal
		} else {
			if safeIsNil(srcVal) {
				result.SetMapIndex(srcKey, srcVal)
				continue
			}
			if resultVal, err = determineValue(destVal, srcVal); err != nil {
				return reflect.Value{}, err
			}
		}
		result.SetMapIndex(srcKey, resultVal)
	}
	return result, nil
}

// determineValue inspects the 'dest' and 'src' values and follows these rules:
// 1. If the values have different kinds, the value of 'src' is returned.
// 2. If the values are maps with the same type, the maps are recursively merged using the mergeMaps helper method.
// 3. If the values are interfaces, determineValue is called with the element values that the interfaces contain.
// 4. If the values are pointers, determineValue is called with the pointer's elements, and the address of the result is returned.
// 5. If the values are any other kind, the value of 'src' is returned.
func determineValue(destVal, srcVal reflect.Value) (reflect.Value, error) {
	if destVal.Kind() != srcVal.Kind() {
		return srcVal, nil
	}
	switch srcVal.Kind() {
	case reflect.Map:
		return mergeMaps(destVal, srcVal)
	case reflect.Interface:
		return determineValue(destVal.Elem(), srcVal.Elem())
	default:
		return srcVal, nil
	}
}

// safeIsNil only calls IsNil if the value is an interface, pointer, map, or slice (IsNil will not panic in these cases)
func safeIsNil(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Interface, reflect.Ptr:
		return val.IsNil() || safeIsNil(val.Elem())
	case reflect.Slice, reflect.Map:
		return val.IsNil()
	default:
		return false
	}
}
