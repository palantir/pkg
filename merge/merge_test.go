// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merge_test

import (
	"testing"

	"github.com/palantir/pkg/merge"
	"github.com/stretchr/testify/assert"
)

type TestStruct1 struct {
	Foo string
}

func TestMergeMaps(t *testing.T) {
	srcVal := "src"
	destVal := "dest"
	for _, test := range []struct {
		name                string
		src, dest, expected interface{}
		expectedErr         string
	}{
		{
			name: "config maps",
			src: map[string]interface{}{
				"conf": map[string]interface{}{
					"map": map[string]interface{}{
						"value1": 1,
						"value2": 2,
					},
					"string": "What number am I thinking of?",
					"array":  []string{"one", "two", "three"},
				},
				"location": "src location",
			},
			dest: map[string]interface{}{
				"conf": map[string]interface{}{
					"map": map[string]interface{}{
						"value1": 5,
					},
					"array": map[string]string{"key": "four", "key2": "five"},
				},
				"string": "What letter am I thinking of?",
			},
			expected: map[string]interface{}{
				"conf": map[string]interface{}{
					"map": map[string]interface{}{
						"value1": 1,
						"value2": 2,
					},
					"string": "What number am I thinking of?",
					"array":  []string{"one", "two", "three"},
				},
				"string":   "What letter am I thinking of?",
				"location": "src location",
			},
		},
		{
			name: "no overlap",
			src: map[string]interface{}{
				"b": &srcVal,
			},
			dest: map[string]interface{}{
				"c": &destVal,
			},
			expected: map[string]interface{}{
				"b": &srcVal,
				"c": &destVal,
			},
		},
		{
			name: "pointers",
			src: map[string]*string{
				"a": &srcVal,
				"b": &srcVal,
			},
			dest: map[string]*string{
				"a": &destVal,
				"c": &destVal,
			},
			expected: map[string]*string{
				"a": &srcVal,
				"b": &srcVal,
				"c": &destVal,
			},
		},
		{
			name: "different map types returns error",
			src: map[string]interface{}{
				"a": "a",
				"b": "b",
			},
			dest: map[string]string{
				"a": "a",
				"c": "c",
			},
			expectedErr: "expected maps of same type",
		},
		{
			name: "different map entry value types return the value from src",
			src: map[string]interface{}{
				"a": "a string",
			},
			dest: map[string]interface{}{
				"a": []string{"a string in a slice that will be overridden"},
				"b": "c",
			},
			expected: map[string]interface{}{
				"a": "a string",
				"b": "c",
			},
		},
		{
			name: "typed nil value for a src map entry results in a typed nil entry for that key",
			src: map[string]interface{}{
				"a": (*string)(nil),
			},
			dest: map[string]interface{}{
				"a": "foo",
				"b": "c",
			},
			expected: map[string]interface{}{
				"a": (*string)(nil),
				"b": "c",
			},
		},
		{
			name: "untyped nil value for a src map entry results in a nil entry for that key",
			src: map[string]interface{}{
				"a": nil,
				"c": nil,
			},
			dest: map[string]interface{}{
				"a": "foo",
				"b": "c",
			},
			expected: map[string]interface{}{
				"a": nil,
				"b": "c",
				"c": nil,
			},
		},
		{
			name: "src val for structs is used",
			src: map[string]interface{}{
				"a": TestStruct1{
					Foo: "src foo value",
				},
			},
			dest: map[string]interface{}{
				"a": "dest bar value",
			},
			expected: map[string]interface{}{
				"a": TestStruct1{
					Foo: "src foo value",
				},
			},
		},
		{
			name: "src value for pointers is used",
			src: map[string]interface{}{
				"a": &map[string]interface{}{
					"b": "c",
				},
				"b": (*string)(nil),
				"c": &[]string{"d"},
			},
			dest: map[string]interface{}{
				"a": &map[string]interface{}{
					"c": "d",
				},
				"b": &destVal,
				"c": "d",
				"d": "non-pointer type",
			},
			expected: map[string]interface{}{
				"a": &map[string]interface{}{
					"b": "c",
				},
				"b": (*string)(nil),
				"c": &[]string{"d"},
				"d": "non-pointer type",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged, err := merge.Maps(test.dest, test.src)
			if test.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, merged)
			} else {
				assert.EqualError(t, err, test.expectedErr)
				assert.Nil(t, merged)
			}
		})
	}
}
