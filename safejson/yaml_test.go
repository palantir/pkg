// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson_test

import (
	"encoding"
	"testing"

	"github.com/palantir/pkg/safejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var _ encoding.TextMarshaler = (*structWithMarshalText)(nil)

type structWithMarshalText struct{}

func (structWithMarshalText) MarshalText() ([]byte, error) {
	return []byte("test-text"), nil
}

func TestFromYAMLValue_convertsInputs(t *testing.T) {
	for _, test := range []struct {
		input  map[any]any
		output map[string]any
	}{
		{
			input: map[any]any{
				"hello": "world",
				"123":   123,
				"foo": map[string]any{
					"bar": 1,
					"baz": 2,
				},
			},
			output: map[string]any{
				"hello": "world",
				"123":   123,
				"foo": map[string]any{
					"bar": 1,
					"baz": 2,
				},
			},
		},
		{
			input: map[any]any{
				"1":   "one",
				"two": "2",
				"other_map": map[any]any{
					"sky":   "blue",
					"grass": "green",
				},
			},
			output: map[string]any{
				"1":   "one",
				"two": "2",
				"other_map": map[string]any{
					"sky":   "blue",
					"grass": "green",
				},
			},
		}, {
			input: map[any]any{
				"array": []any{
					map[any]any{
						"a": "b",
						"b": "c",
						"c": "d",
					},
				},
			},
			output: map[string]any{
				"array": []any{
					map[string]any{
						"a": "b",
						"b": "c",
						"c": "d",
					},
				},
			},
		},
		{
			input: map[any]any{
				"array": nil,
			},
			output: map[string]any{
				"array": nil,
			},
		},
		{
			input: map[any]any{
				6: "six",
				7: "seven",
			},
			output: map[string]any{
				"6": "six",
				"7": "seven",
			},
		},
		{
			input: map[any]any{
				structWithMarshalText{}: "test-value",
			},
			output: map[string]any{
				"test-text": "test-value",
			},
		},
		{
			input: map[any]any{
				"inner-map": map[int]string{
					6: "seven",
				},
			},
			output: map[string]any{
				"inner-map": map[string]any{
					"6": "seven",
				},
			},
		},
		{
			input: map[any]any{
				"string-key":            "string-value",
				toPtr("string-ptr-key"): "string-ptr-value",
				6:                       "int-value",
				toPtr(7):                "int-ptr-value",
			},
			output: map[string]any{
				"string-key":     "string-value",
				"string-ptr-key": "string-ptr-value",
				"6":              "int-value",
				"7":              "int-ptr-value",
			},
		},
	} {
		out, err := safejson.FromYAMLValue(test.input)
		require.NoError(t, err)
		assert.Equal(t, test.output, out)
	}
}

func TestFromYAML_ErrorCases(t *testing.T) {
	type structWithNoMarshalText struct{}

	for _, tc := range []struct {
		name     string
		inputMap map[any]any
		wantErr  string
	}{
		{
			name: "map with struct key that does not implement TextMarshaler",
			inputMap: map[any]any{
				"two": "2",
				"other_map": map[any]any{
					1:                         "one",
					"sky":                     "blue",
					"grass":                   "green",
					structWithNoMarshalText{}: "value-with-struct-key",
				},
			},
			wantErr: "expected map key inside other_map to be a valid key type (string, number, TextMarshaler) but was safejson_test.structWithNoMarshalText: {}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := safejson.FromYAMLValue(tc.inputMap)
			assert.EqualError(t, err, tc.wantErr)
			assert.Nil(t, out)
		})
	}
}

func TestFromYAML(t *testing.T) {
	for _, test := range []struct {
		input  string
		output map[string]any
	}{
		{
			input: `---
x:
  z: 0
`,
			output: map[string]any{
				"x": map[string]any{
					"z": 0,
				},
			},
		},
	} {
		var y any
		err := yaml.Unmarshal([]byte(test.input), &y)
		if assert.NoError(t, err) {
			j, err := safejson.FromYAMLValue(y)
			assert.NoError(t, err)
			assert.Equal(t, test.output, j)
		}
	}
}

func TestMapInStructsNotConverted(t *testing.T) {
	val := struct {
		v map[any]string
	}{
		v: map[any]string{
			13: "thirteen",
		},
	}

	res, err := safejson.FromYAMLValue(val)
	require.NoError(t, err)
	assert.Equal(t, val, res)
}

func toPtr[T any](in T) *T {
	return &in
}
