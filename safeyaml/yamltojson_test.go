// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safeyaml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYAMLUnmarshalerToJSONBytes(t *testing.T) {
	for _, test := range []struct {
		Name string
		YAML string
		JSON string
		Err  string
	}{
		{
			Name: "object",
			YAML: "z: a\n\"y\": true\nx: 1\nw: 1.2\nv:\n  foo: bar\n  baz: qux\n",
			JSON: `{"z":"a","y":true,"x":1,"w":1.2,"v":{"foo":"bar","baz":"qux"}}`,
		},
		{
			Name: "slice",
			YAML: "- a\n- 1\n- foo: bar\n",
			JSON: `["a",1,{"foo":"bar"}]`,
		},
		{
			Name: "empty slice",
			YAML: "[]\n",
			JSON: `[]`,
		},
		{
			Name: "empty object",
			YAML: "{}\n",
			JSON: `{}`,
		},
		{
			Name: "nil",
			YAML: "foo: null\n",
			JSON: `{"foo":null}`,
		},
		{
			Name: "object with number key",
			YAML: "1: 2",
			JSON: `{"1":2}`,
		},
		{
			Name: "object with invalid map key",
			YAML: "{1: 2}: 3",
			Err:  `yaml: invalid map key: map[interface {}]interface {}{1:2}`,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			out, err := YAMLtoJSONBytes([]byte(test.YAML))
			if test.Err == "" {
				require.NoError(t, err)
				require.JSONEq(t, test.JSON, string(out))
			} else {
				require.Error(t, err, "Expected error but got nil. JSON: %s", string(out))
				require.Contains(t, err.Error(), test.Err)
			}
		})
	}
}
