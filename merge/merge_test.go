// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merge_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/merge"
)

func TestMergeMaps(t *testing.T) {
	srcVal := "src"
	destVal := "dest"
	for _, test := range []struct {
		name                string
		src, dest, expected interface{}
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
			name: "different types returns src unchanged",
			src: map[string]interface{}{
				"a": "a",
				"b": "b",
			},
			dest: map[string]string{
				"a": "a",
				"c": "c",
			},
			expected: map[string]interface{}{
				"a": "a",
				"b": "b",
			},
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
			name: "explicit nil value for a src map entry results in no entry for that key",
			src: map[string]interface{}{
				"a": nil,
			},
			dest: map[string]interface{}{
				"a": "foo",
				"b": "c",
			},
			expected: map[string]interface{}{
				"b": "c",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged, err := merge.Maps(test.dest, test.src)
			require.NoError(t, err)
			require.Equal(t, test.expected, merged)
		})
	}
}
