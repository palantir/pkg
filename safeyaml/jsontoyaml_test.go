// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safeyaml

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestJSONtoYAML(t *testing.T) {
	for _, test := range []struct {
		Name string
		JSON string
		YAML string
	}{
		{
			Name: "object",
			JSON: `{"z":"a", "y":"b", "x": 1, "w": 1.2, "v": {"foo": "bar", "baz": "qux"}}`,
			YAML: "z: a\n\"y\": b\nx: 1\nw: 1.2\nv:\n  foo: bar\n  baz: qux\n",
		},
		{
			Name: "slice",
			JSON: `["a", 1, {"foo": "bar"}]`,
			YAML: "- a\n- 1\n- foo: bar\n",
		},
		{
			Name: "empty slice",
			JSON: `[]`,
			YAML: "[]\n",
		},
		{
			Name: "empty object",
			JSON: `{}`,
			YAML: "{}\n",
		},
		{
			Name: "nil",
			JSON: `{"foo": null}`,
			YAML: "foo: null\n",
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			obj, err := JSONtoYAML([]byte(test.JSON))
			require.NoError(t, err)
			out, err := yaml.Marshal(obj)
			require.NoError(t, err)
			require.Equal(t, test.YAML, string(out))
		})
	}
}