// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typenames

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct{}

func TestNames(t *testing.T) {
	for _, test := range []struct {
		objs                                []interface{}
		expectedShortName, expectedLongName string
	}{
		{
			objs: []interface{}{
				nil,
			},
			expectedShortName: "null",
			expectedLongName:  "null",
		},
		{
			objs: []interface{}{
				true,
			},
			expectedShortName: "boolean",
			expectedLongName:  "boolean",
		},
		{
			objs: []interface{}{
				int(0),
				int8(0),
				int16(0),
				int32(0),
				int64(0),
				big.NewInt(0),
			},
			expectedShortName: "integer",
			expectedLongName:  "integer",
		},
		{
			objs: []interface{}{
				uint(0),
				uint8(0),
				uint16(0),
				uint32(0),
				uint64(0),
				uintptr(0),
			},
			expectedShortName: "unsigned integer",
			expectedLongName:  "unsigned integer",
		},
		{
			objs: []interface{}{
				float32(0),
				float64(0),
				big.NewFloat(0),
			},
			expectedShortName: "float",
			expectedLongName:  "float",
		},
		{
			objs: []interface{}{
				complex(float32(0), float32(0)),
				complex(float64(0), float64(0)),
			},
			expectedShortName: "complex number",
			expectedLongName:  "complex number",
		},
		{
			objs: []interface{}{
				[]interface{}{},
				[...]interface{}{},
			},
			expectedShortName: "array",
			expectedLongName:  "array of object",
		},
		{
			objs: []interface{}{
				[]*string{},
				[...]*string{},
			},
			expectedShortName: "array",
			expectedLongName:  "array of *string",
		},
		{
			objs: []interface{}{
				make(chan fmt.Stringer),
			},
			expectedShortName: "Go channel",
			expectedLongName:  "Go channel of Stringer",
		},
		{
			objs: []interface{}{
				TestNames,
				(*big.Int).String,
			},
			expectedShortName: "function",
			expectedLongName:  "function",
		},
		{
			objs: []interface{}{
				map[string]string{},
			},
			expectedShortName: "map",
			expectedLongName:  "map of string to string",
		},
		{
			objs: []interface{}{
				map[string]interface{}{},
			},
			expectedShortName: "map",
			expectedLongName:  "map of string to object",
		},
		{
			objs: []interface{}{
				&[]string{""}[0],
			},
			expectedShortName: "*string",
			expectedLongName:  "*string",
		},
		{
			objs: []interface{}{
				&[]*string{&[]string{""}[0]}[0],
			},
			expectedShortName: "**string",
			expectedLongName:  "**string",
		},
		{
			objs: []interface{}{
				"",
			},
			expectedShortName: "string",
			expectedLongName:  "string",
		},
		{
			objs: []interface{}{
				testStruct{},
			},
			expectedShortName: "struct testStruct",
			expectedLongName:  "struct testStruct",
		},
	} {
		for _, obj := range test.objs {
			for _, v := range []interface{}{reflect.ValueOf(obj), reflect.TypeOf(obj), obj} {
				assert.Equal(t, test.expectedShortName, ShortName(v), "short name of %v", v)
				assert.Equal(t, test.expectedLongName, LongName(v), "long name of %v", v)
			}
		}
	}
}
