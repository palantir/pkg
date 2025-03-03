// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testhelpers

import (
	"testing"

	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ContainerTestIndentSpaces = 4

func RunContainerTests[NodeT any](t *testing.T, testNamePrefix string, yamllib yamlpatchcommon.YAMLLibrary[NodeT]) {
	for _, test := range []struct {
		Name     string
		Doc      string
		Patch    func(t *testing.T, node NodeT)
		Expected string
	}{
		{
			Name: "map: get complex value",
			Doc: `foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				fooNode, err := c.Get("foo")
				require.NoError(t, err)
				fooValue, err := yamllib.NodeToValue(fooNode)
				require.NoError(t, err)
				expected := map[string]interface{}{"bar": "val"}
				assert.Equal(t, expected, fooValue)
			},
			Expected: `foo:
    bar: val
`,
		},
		{
			Name: "map: get inner value",
			Doc: `foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				fooNode, err := c.Get("foo")
				require.NoError(t, err)
				fooContainer, err := yamllib.NewContainer(fooNode)
				require.NoError(t, err)
				barNode, err := fooContainer.Get("bar")
				require.NoError(t, err)
				barValue, err := yamllib.NodeToValue(barNode)
				require.NoError(t, err)
				assert.Equal(t, "val", barValue)
			},
			Expected: `foo:
    bar: val
`,
		},
		{
			Name: "map: get does not exist",
			Doc: `foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				missingNode, err := c.Get("notfound")
				require.NoError(t, err)
				assert.Nil(t, missingNode, "expected missing node to be nil")
			},
			Expected: `foo:
    bar: val
`,
		},
		{
			Name: "map: add key",
			Doc: `foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("newkey", valNode)
				require.NoError(t, err)
			},
			Expected: `foo:
    bar: val
newkey: newvalue
`,
		},
		{
			Name: "map: add complex value",
			Doc:  `key: value`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode(map[string]interface{}{"bar": "val"}, "")
				require.NoError(t, err)
				err = c.Add("foo", valNode)
				require.NoError(t, err)
			},
			Expected: `key: value
foo:
    bar: val
`,
		},
		{
			Name: "map: add key already exists error",
			Doc:  "foo: bar\n",
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("foo", valNode)
				require.EqualError(t, err, "key foo already exists and can not be added")
			},
			Expected: "foo: bar\n",
		},
		{
			Name: "map: set key",
			Doc: `foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Set("foo", valNode)
				require.NoError(t, err)
			},
			Expected: "foo: newvalue\n",
		},
		{
			Name: "map: set complex value",
			Doc:  `foo: value`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode(map[string]interface{}{"bar": "update", "baz": 2}, "")
				require.NoError(t, err)
				err = c.Set("foo", valNode)
				require.NoError(t, err)
			},
			Expected: `foo:
    bar: update
    baz: 2
`,
		},
		{
			Name: "map: set key does not exist error",
			Doc:  "foo: bar\n",
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Set("notfound", valNode)
				require.EqualError(t, err, "key notfound does not exist and can not be replaced")
			},
			Expected: "foo: bar\n",
		},
		{
			Name: "map: remove key",
			Doc: `
newkey: newvalue
foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("foo")
				require.NoError(t, err)
			},
			Expected: "newkey: newvalue\n",
		},
		{
			Name: "map: remove key extra whitespace",
			Doc: `


newkey: newvalue


foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("foo")
				require.NoError(t, err)
			},
			Expected: "newkey: newvalue\n",
		},
		{
			Name: "map: remove key does not exist error",
			Doc:  `foo: bar`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("notfound")
				require.EqualError(t, err, "key notfound does not exist and can not be removed")
			},
			Expected: "foo: bar\n",
		},
		{
			Name: "seq: get complex value",
			Doc: `
- 0
- foo: bar
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				fooNode, err := c.Get("1")
				require.NoError(t, err)
				fooValue, err := yamllib.NodeToValue(fooNode)
				require.NoError(t, err)
				expected := map[string]interface{}{"foo": "bar"}
				assert.Equal(t, expected, fooValue)
			},
			Expected: `- 0
- foo: bar
- 2
`,
		},
		{
			Name: "seq: get inner value",
			Doc: `
- 0
- foo: bar
- list: [ 2, 3 ]
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				elemNode, err := c.Get("2")
				require.NoError(t, err)
				elemContainer, err := yamllib.NewContainer(elemNode)
				require.NoError(t, err)
				listNode, err := elemContainer.Get("list")
				require.NoError(t, err)
				listContainer, err := yamllib.NewContainer(listNode)
				require.NoError(t, err)
				innerElemNode, err := listContainer.Get("0")
				require.NoError(t, err)

				innerElemValue, err := yamllib.NodeToValue(innerElemNode)
				require.NoError(t, err)
				assert.Equal(t, 2, convertToInt(innerElemValue))
			},
			Expected: `- 0
- foo: bar
- list: [2, 3]
`,
		},
		{
			Name: "seq: get does not exist",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				missingNode, err := c.Get("3")
				require.NoError(t, err)
				assert.Nil(t, missingNode, "expected missing node to be nil")
			},
			Expected: `- 0
- 1
- 2
`,
		},
		{
			Name: "seq: add key to end of sequence",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("-", valNode)
				require.NoError(t, err)
			},
			Expected: `- 0
- 1
- 2
- newvalue
`,
		},
		{
			Name: "seq: add key at last key of sequence",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("3", valNode)
				require.NoError(t, err)
			},
			Expected: `- 0
- 1
- 2
- newvalue
`,
		},
		{
			Name: "seq: add key to start of sequence",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("0", valNode)
				require.NoError(t, err)
			},
			Expected: `- newvalue
- 0
- 1
- 2
`,
		},
		{
			Name: "seq: add key to center of sequence",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("1", valNode)
				require.NoError(t, err)
			},
			Expected: `- 0
- newvalue
- 1
- 2
`,
		},
		{
			Name: "seq: add key to empty sequence",
			Doc:  `[]`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("-", valNode)
				require.NoError(t, err)
			},
			Expected: `[newvalue]
`,
		},
		{
			Name: "seq: add key out of bounds error",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Add("4", valNode)
				require.EqualError(t, err, "add index key out of bounds (idx 4, len 3)")
			},
			Expected: `- 0
- 1
- 2
`,
		},
		{
			Name: "seq: set key",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Set("1", valNode)
				require.NoError(t, err)
			},
			Expected: `- 0
- newvalue
- 2
`,
		},
		{
			Name: "seq: set complex value",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode(map[string]interface{}{"bar": "update", "baz": 2}, "")
				require.NoError(t, err)
				err = c.Set("1", valNode)
				require.NoError(t, err)
			},
			Expected: `- 0
- bar: update
  baz: 2
- 2
`,
		},
		{
			Name: "seq: set key does not exist error",
			Doc: `- 0
- 1
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				valNode, err := yamllib.ValueToNode("newvalue", "")
				require.NoError(t, err)
				err = c.Set("2", valNode)
				require.EqualError(t, err, "set index key out of bounds (idx 2, len 2)")
			},
			Expected: `- 0
- 1
`,
		},
		{
			Name: "seq: remove key",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("1")
				require.NoError(t, err)
			},
			Expected: `- 0
- 2
`,
		},
		{
			Name: "seq: remove first key",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("0")
				require.NoError(t, err)
			},
			Expected: `- 1
- 2
`,
		},
		{
			Name: "seq: remove last key",
			Doc: `- 0
- 1
- 2
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("2")
				require.NoError(t, err)
			},
			Expected: `- 0
- 1
`,
		},
		{
			Name: "seq: remove key extra whitespace",
			Doc: `


newkey: newvalue


foo:
    bar: val
`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("foo")
				require.NoError(t, err)
			},
			Expected: "newkey: newvalue\n",
		},
		{
			Name: "seq: remove key does not exist error",
			Doc:  `foo: bar`,
			Patch: func(t *testing.T, node NodeT) {
				c, err := yamllib.NewContainer(node)
				require.NoError(t, err)
				err = c.Remove("notfound")
				require.EqualError(t, err, "key notfound does not exist and can not be removed")
			},
			Expected: "foo: bar\n",
		},
	} {
		t.Run(testNamePrefix+" "+test.Name, func(t *testing.T) {
			node, err := yamllib.BytesToContentNode([]byte(test.Doc))
			require.NoError(t, err)

			test.Patch(t, node)
			out, err := yamllib.NodeToBytes(node)
			require.NoError(t, err)

			assertYAMLEqual(t, yamllib, []byte(test.Expected), out, true)
		})
	}
}

func assertYAMLEqual[NodeT any](t *testing.T, yamllib yamlpatchcommon.YAMLLibrary[NodeT], a, b []byte, testTextEqual bool) {
	var objA interface{}
	require.NoError(t, yamllib.Unmarshal(a, &objA))
	var objB interface{}
	require.NoError(t, yamllib.Unmarshal(b, &objB))
	if assert.Equal(t, objA, objB) && testTextEqual {
		assert.Equal(t, string(a), string(b), "YAML objects had equal data but differing text")
	}
}

// necessary because YAML/JSON representation of integers is not strictly defined. As such, if a YAML value like
// "num: 23" is loaded into an arbitrary data structure/map, the value could be an int, int64, uint, etc. depending on
// the library/implementation/configuration. Because these tests are meant to be generic and applicable across different
// library implementations, this function should be used to normalize integers before comparison/assertion.
func convertToInt(in any) int {
	switch v := in.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	}
	// this will panic
	return in.(int)
}
