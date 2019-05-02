// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merge

import (
	"fmt"
	"reflect"
)

// Both inputs are required to be maps. If both input maps have the same type,
// the returned map has the same type as well. If the input maps have different
// types, src is returned unchanged.
func Maps(dest, src interface{}) (interface{}, error) {
	result, err := mergeMaps(reflect.ValueOf(dest), reflect.ValueOf(src))
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

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
	default:
		return srcVal, nil
	}
}
