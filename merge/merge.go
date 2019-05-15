// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merge

import (
	"fmt"
	"reflect"
)

// Maps merges two interfaces using the mergeMaps method, and returns the resulting value.
func Maps(dest, src interface{}) (interface{}, error) {
	result, err := mergeMaps(reflect.ValueOf(dest), reflect.ValueOf(src))
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

// mergeMaps requires both inputs to be maps; if not, an error is returned. If both input maps have the same type,
// the returned map has the same type as well. If the input maps have different
// types, src is returned unchanged. Otherwise, a new map is created and populated
// with the merge result for the return value. For map entries with the same key, the following rules apply:
// 1. If the values have different types, the value from 'src' is used.
// 2. If the src value is nil, the entry is absent from the resulting map.
// 3. If the values are the same type and are structs, then src's value for each struct field is used, excepting map type fields, which are recursively merged.
// 4. If the values are the same type and are slices or primitives, the value from 'src' is used.
// 5. If the values are maps, the maps are recursively merged using the mergeMaps helper method.
func mergeMaps(dest, src reflect.Value) (reflect.Value, error) {
	if dest.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("expected destination to be a map")
	}
	if src.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("expected source be a map")
	}

	if dest.Type() != src.Type() {
		return src, nil
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
			resultVal = srcVal
		} else if resultVal, err = mergeValues(destVal, srcVal); err != nil {
			return reflect.Value{}, err
		}
		result.SetMapIndex(srcKey, resultVal)
	}
	return result, nil
}

func mergeValues(destVal, srcVal reflect.Value) (reflect.Value, error) {
	if destVal.Kind() != srcVal.Kind() {
		return srcVal, nil
	}
	switch srcVal.Kind() {
	case reflect.Map:
		return mergeMaps(destVal, srcVal)
	case reflect.Interface:
		return mergeValues(destVal.Elem(), srcVal.Elem())
	case reflect.Ptr:
		val, err := mergeValues(destVal.Elem(), srcVal.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		return val.Addr(), nil
	case reflect.Struct:
		if srcVal.Type() != destVal.Type() {
			return srcVal, nil
		}
		val := reflect.New(srcVal.Type())
		for i := 0; i < srcVal.NumField(); i++ {
			_ = destVal.Field(i)
			fieldVal, err := mergeValues(destVal.Field(i), srcVal.Field(i))
			if err != nil {
				return reflect.Value{}, err
			}
			if val.Elem().Field(i).CanSet() {
				val.Elem().Field(i).Set(fieldVal)
			}
		}
		return val.Elem(), nil
	default:
		return srcVal, nil
	}
}