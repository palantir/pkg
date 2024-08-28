// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyJSONPatch(t *testing.T) {
	for _, test := range []struct {
		Name      string
		Body      string
		Patch     []string
		Expected  string
		ExpectErr string
	}{
		{
			Name:     "add to empty string",
			Patch:    []string{`{"op":"add","path":"/","value":{"key":"value"}}`},
			Body:     "",
			Expected: "key: value",
		},
		{
			Name:     "add scalar to empty string",
			Patch:    []string{`{"op":"add","path":"/","value":"value"}`},
			Body:     "",
			Expected: "value",
		},
		{
			Name:     "add sequence to empty string",
			Patch:    []string{`{"op":"add","path":"/","value":["v1", "v2"]}`},
			Body:     "",
			Expected: "- v1\n- v2",
		},
		{
			Name:     "add to empty doc",
			Patch:    []string{`{"op":"add","path":"/","value":{"key":"value"}}`},
			Body:     "---",
			Expected: "key: value",
		},
		{
			Name:  "add to map",
			Patch: []string{`{"op":"add","path":"/foo/baz","value":"value"}`},
			Body: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar`,
			Expected: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar
  baz: value`,
		},
		{
			Name:  "add to root doc",
			Patch: []string{`{"op":"add","path":"/baz","value":"value"}`},
			Body: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar`,
			Expected: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar
baz: value`,
		},
		{
			Name:  "add to end of array",
			Patch: []string{`{"op":"add","path":"/foo/arr/-","value":4}`},
			Body: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar`,
			Expected: `# my foo
foo:
  arr: [1, 2, 3, 4]
  bar: 1 # my bar`,
		},
		{
			Name:  "add into array",
			Patch: []string{`{"op":"add","path":"/foo/arr/0","value":0}`},
			Body: `# my foo
foo:
  arr: [1, 2, 3]
  bar: 1 # my bar`,
			Expected: `# my foo
foo:
  arr: [0, 1, 2, 3]
  bar: 1 # my bar`,
		},
		{
			Name:  "add to array with comment",
			Patch: []string{`{"op":"add","path":"/foo/arr/-","value":4,"comment":"the number 4"}`},
			Body: `# my foo
foo:
  arr:
    # numbers 1 through 3
    - 1
    - 2
    - 3`,
			Expected: `# my foo
foo:
  arr:
    # numbers 1 through 3
    - 1
    - 2
    - 3
    # the number 4
    - 4`,
		},
		{
			Name:  "replace object with comment",
			Patch: []string{`{"op":"replace","path":"/foo/bar","value":{"key":"value"},"comment":"key value pair"}`},
			Body: `# my foo
foo:
  bar: hello-world # previous comment`,
			Expected: `# my foo
foo:
  bar:
    # key value pair
    key: value`,
		},
		// Test cases from json-patch: https://github.com/evanphx/json-patch/blob/master/patch_test.go to verify JSON Patch correctness.
		{
			Name:     "jsonpatch add",
			Body:     `{"foo": "bar"}`,
			Patch:    []string{`{"op": "add", "path": "/baz", "value": "qux"}`},
			Expected: `{"foo": "bar", baz: qux}`,
		},
		{
			Name:     "jsonpatch add",
			Body:     `{"foo": [ "bar", "baz" ]}`,
			Patch:    []string{`{"op": "add", "path": "/foo/1", "value": "qux"}`},
			Expected: `{"foo": ["bar", qux, "baz"]}`,
		},
		{
			Name:     "jsonpatch remove",
			Body:     `{"baz": "qux", "foo": "bar"}`,
			Patch:    []string{`{"op": "remove", "path": "/baz"}`},
			Expected: `{"foo": "bar"}`,
		},
		{
			Name:     "jsonpatch remove",
			Body:     `{"foo": [ "bar", "qux", "baz" ]}`,
			Patch:    []string{`{"op": "remove", "path": "/foo/1"}`},
			Expected: `{"foo": ["bar", "baz"]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"baz": "qux", "foo": "bar"}`,
			Patch:    []string{`{"op": "replace", "path": "/baz", "value": "boo"}`},
			Expected: `{"baz": boo, "foo": "bar"}`,
		},
		{
			Name:     "jsonpatch move",
			Body:     `{"foo":{"bar":"baz","waldo":"fred"},"qux":{"corge":"grault"}}`,
			Patch:    []string{`{"op": "move", "from": "/foo/waldo", "path": "/qux/thud"}`},
			Expected: `{"foo": {"bar": "baz"}, "qux": {"corge": "grault", thud: "fred"}}`,
		},
		{
			Name:     "jsonpatch move",
			Body:     `{"foo":["all","grass","cows","eat"]}`,
			Patch:    []string{`{"op": "move", "from": "/foo/1", "path": "/foo/3"}`},
			Expected: `{"foo": ["all", "cows", "eat", "grass"]}`,
		},

		{
			Name:     "jsonpatch move",
			Body:     `{"foo": [ "all", "grass", "cows", "eat" ] }`,
			Patch:    []string{`{"op": "move", "from": "/foo/1", "path": "/foo/2"}`},
			Expected: `{"foo": ["all", "cows", "grass", "eat"]}`,
		},
		{
			Name:     "jsonpatch add",
			Body:     `{"foo": "bar" }`,
			Patch:    []string{`{"op": "add", "path": "/child", "value": { "grandchild": { } }}`},
			Expected: `{"foo": "bar", child: {grandchild: {}}}`,
		},
		{
			Name:     "jsonpatch add",
			Body:     `{"foo": ["bar"] }`,
			Patch:    []string{`{"op": "add", "path": "/foo/-", "value": ["abc", "def"]}`},
			Expected: `{"foo": ["bar", [abc, def]]}`,
		},
		{
			Name:     "jsonpatch remove",
			Body:     `{"foo": "bar", "qux": { "baz": 1, "bar": null } }`,
			Patch:    []string{`{"op": "remove", "path": "/qux/bar"}`},
			Expected: `{"foo": "bar", "qux": {"baz": 1}}`,
		},
		{
			Name:     "jsonpatch add",
			Body:     `{"foo": "bar" }`,
			Patch:    []string{`{"op": "add", "path": "/baz", "value": null}`},
			Expected: `{"foo": "bar", baz: null}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"foo": ["bar"]}`,
			Patch:    []string{`{"op": "replace", "path": "/foo/0", "value": "baz"}`},
			Expected: `{"foo": [baz]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"foo": ["bar", "baz"]}`,
			Patch:    []string{`{"op": "replace", "path": "/foo/0", "value": "bum"}`},
			Expected: `{"foo": [bum, "baz"]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"foo": ["bar","qux","baz"]}`,
			Patch:    []string{`{"op": "replace", "path": "/foo/1", "value": "bum"}`},
			Expected: `{"foo": ["bar", bum, "baz"]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `[ {"foo": ["bar","qux","baz"]}]`,
			Patch:    []string{`{"op": "replace", "path": "/0/foo/0", "value": "bum"}`},
			Expected: `[{"foo": [bum, "qux", "baz"]}]`,
		},
		{
			Name:     "jsonpatch copy",
			Body:     `[ {"foo": ["bar","qux","baz"], "bar": ["qux","baz"]}]`,
			Patch:    []string{`{"op": "copy", "from": "/0/foo/0", "path": "/0/bar/0"}`},
			Expected: `[{"foo": ["bar", "qux", "baz"], "bar": ["bar", "qux", "baz"]}]`,
		},
		{
			Name:     "jsonpatch copy",
			Body:     `[ { "foo": {"bar": ["qux","baz"]}, "baz": {"qux": "bum"}}]`,
			Patch:    []string{`{"op": "copy", "from": "/0/foo/bar", "path": "/0/baz/bar"}`},
			Expected: `[{"foo": {"bar": ["qux", "baz"]}, "baz": {"qux": "bum", bar: ["qux", "baz"]}}]`,
		},
		{
			Name:     "jsonpatch copy",
			Body:     `{"foo": ["bar"]}`,
			Patch:    []string{`{"op": "copy", "path": "/foo/0", "from": "/foo"}`},
			Expected: `{"foo": [["bar"], "bar"]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"bar": [{"baz": null}]}`,
			Patch:    []string{`{"op": "replace", "path": "/bar/0/baz", "value": 1}`},
			Expected: `{"bar": [{"baz": 1}]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"bar": [{"baz": 1}]}`,
			Patch:    []string{`{"op": "replace", "path": "/bar/0/baz", "value": null}`},
			Expected: `{"bar": [{"baz": null}]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"bar": [null]}`,
			Patch:    []string{`{"op": "replace", "path": "/bar/0", "value": 1}`},
			Expected: `{"bar": [1]}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{"bar": [1]}`,
			Patch:    []string{`{"op": "replace", "path": "/bar/0", "value": null }`},
			Expected: `{"bar": [null]}`,
		},
		{
			Name:     "jsonpatch remove",
			Body:     `{ "foo": "bar", "qux": { "baz": 1, "bar": null } }`,
			Patch:    []string{`{ "op": "remove", "path": "/qux/bar" }`},
			Expected: `{"foo": "bar", "qux": {"baz": 1}}`,
		},
		{
			Name:     "jsonpatch replace",
			Body:     `{ "foo": "bar", "qux": { "baz": 1, "bar": null } }`,
			Patch:    []string{`{ "op": "replace", "path": "/qux/baz", "value": null }`},
			Expected: `{"foo": "bar", "qux": {"baz": null, "bar": null}}`,
		},
		{
			Name:      "jsonpatch add err",
			Body:      `{ "foo": "bar" }`,
			Patch:     []string{`{ "op": "add", "path": "/baz/bat", "value": "qux" }`},
			ExpectErr: "op add /baz/bat: parent path /baz does not exist",
		},
		{
			Name:      "jsonpatch remove err",
			Body:      `{ "a": { "b": { "d": 1 } } }`,
			Patch:     []string{`{ "op": "remove", "path": "/a/b/c" }`},
			ExpectErr: "op remove /a/b/c: key c does not exist and can not be removed",
		},
		{
			Name:      "jsonpatch move err",
			Body:      `{ "a": { "b": { "d": 1 } } }`,
			Patch:     []string{`{ "op": "move", "from": "/a/b/c", "path": "/a/b/e" }`},
			ExpectErr: "op move /a/b/e: node not found for move patch at path /a/b/c",
		},
		{
			Name:      "jsonpatch move err",
			Body:      `{ "a": { "b": { "d": 1 } } }`,
			Patch:     []string{`{ "op": "move", "from": "/a/b/c", "path": "/a/b/e" }`},
			ExpectErr: "op move /a/b/e: node not found for move patch at path /a/b/c",
		},
		{
			Name:      "jsonpatch remove err",
			Body:      `{ "a": { "b": [1] } }`,
			Patch:     []string{`{ "op": "remove", "path": "/a/b/1" }`},
			ExpectErr: "op remove /a/b/1: remove index key out of bounds (idx 1, len 1)",
		},
		{
			Name:      "jsonpatch move err",
			Body:      `{ "a": { "b": [1] } }`,
			Patch:     []string{`{ "op": "move", "from": "/a/b/1", "path": "/a/b/2" }`},
			ExpectErr: "op move /a/b/2: node not found for move patch at path /a/b/1",
		},
		{
			Name:      "jsonpatch add err",
			Body:      `{ "foo": "bar" }`,
			Patch:     []string{`{ "op": "add", "pathz": "/baz", "value": "qux" }`},
			ExpectErr: "op add requires a path field",
		},
		{
			Name:      "jsonpatch replace err",
			Body:      `{ "foo": ["bar","baz"]}`,
			Patch:     []string{`{ "op": "replace", "path": "/foo/2", "value": "bum" }`},
			ExpectErr: "op replace /foo/2: set index key out of bounds (idx 2, len 2)",
		},
		{
			Name:      "jsonpatch add err",
			Body:      `{ "foo": ["bar","baz"]}`,
			Patch:     []string{`{ "op": "add", "path": "/foo/-4", "value": "bum"}`},
			ExpectErr: "op add /foo/-4: index into SequenceNode with negative \"-4\" key",
		},
		{
			Name:      "jsonpatch replace err",
			Body:      `{ "name":{ "foo": "bat", "qux": "bum"}}`,
			Patch:     []string{`{ "op": "replace", "path": "/foo/bar", "value":"baz"}`},
			ExpectErr: "op replace /foo/bar: parent path /foo does not exist",
		},
		{
			Name:      "jsonpatch add err",
			Body:      `{ "foo": ["bar"]}`,
			Patch:     []string{`{"op": "add", "path": "/foo/2", "value": "bum"}`},
			ExpectErr: "op add /foo/2: add index key out of bounds (idx 2, len 1)",
		},
		{
			Name:      "jsonpatch remove err",
			Body:      `{ "foo": []}`,
			Patch:     []string{`{"op": "remove", "path": "/foo/-"}`},
			ExpectErr: "op remove /foo/-: index into SequenceNode with non-integer \"-\" key: strconv.Atoi: parsing \"-\": invalid syntax",
		},
		{
			Name:      "jsonpatch remove err",
			Body:      `{ "foo": []}`,
			Patch:     []string{`{"op": "remove", "path": "/foo/-1"}`},
			ExpectErr: "op remove /foo/-1: index into SequenceNode with negative \"-1\" key",
		},
		{
			Name:      "jsonpatch null err",
			Body:      `{}`,
			Patch:     []string{`{"op":null,"path":"/"}`},
			ExpectErr: "op  /: unexpected op",
		},
		{
			Name:      "jsonpatch add err",
			Body:      `{}`,
			Patch:     []string{`{"op":"add","path":null}`},
			ExpectErr: "op add requires a path field",
		},
		{
			Name:      "jsonpatch copy err",
			Body:      `{}`,
			Patch:     []string{`{ "op": "copy", "from": null }`},
			ExpectErr: "op copy requires a path field",
		},
		{
			Name:      "jsonpatch copy err",
			Body:      `{ "foo": ["bar"]}`,
			Patch:     []string{`{"op": "copy", "path": "/foo/6666666666", "from": "/"}`},
			ExpectErr: "op copy /foo/6666666666: add index key out of bounds (idx 6666666666, len 1)",
		},
		{
			Name:      "jsonpatch copy err",
			Body:      `{ "foo": ["bar"]}`,
			Patch:     []string{`{"op": "copy", "path": "/foo/2", "from": "/foo/0"}`},
			ExpectErr: "op copy /foo/2: add index key out of bounds (idx 2, len 1)",
		},
		{
			Name:      "jsonpatch move err",
			Body:      `{ "foo": [ "all", "grass", "cows", "eat" ] }`,
			Patch:     []string{`{ "op": "move", "from": "/foo/1", "path": "/foo/4" }`},
			ExpectErr: "op move /foo/4: add index key out of bounds (idx 4, len 3)",
		},
		{
			Name: "jsonpatch test",
			Body: `{"baz":"qux","foo":["a",2,"c"]}`,
			Patch: []string{
				`{ "op": "test", "path": "/baz", "value": "qux" }`,
				`{ "op": "test", "path": "/foo/1", "value": 2 }`,
			},
			Expected: `{"baz": "qux", "foo": ["a", 2, "c"]}`,
		},
		{
			Name:      "jsonpatch test err",
			Body:      `{ "baz": "qux" }`,
			Patch:     []string{`{ "op": "test", "path": "/baz", "value": "bar" }`},
			ExpectErr: "op test /baz: testing path /baz value failed",
		},
		{
			Name: "jsonpatch test err",
			Body: `{"baz":"qux","foo":["a",2,"c"]}`,
			Patch: []string{
				`{ "op": "test", "path": "/baz", "value": "qux" }`,
				`{ "op": "test", "path": "/foo/1", "value": "c" }`,
			},
			ExpectErr: "op test /foo/1: testing path /foo/1 value failed",
		},
		{
			Name:      "jsonpatch test err",
			Body:      `{ "baz": "qux" }`,
			Patch:     []string{`{ "op": "test", "path": "/foo", "value": 42 }`},
			ExpectErr: "op test /foo: node not found for test patch at path /foo",
		},
		{
			Name:      "jsonpatch test err",
			Body:      `{ "baz": "qux" }`,
			Patch:     []string{`{ "op": "test", "path": "/foo", "value": null }`},
			ExpectErr: "op test /foo: node not found for test patch at path /foo",
		},
		{
			Name:     "jsonpatch test",
			Body:     `{ "foo": null }`,
			Patch:    []string{`{ "op": "test", "path": "/foo", "value": null }`},
			Expected: `{"foo": null}`,
		},
		{
			Name:      "jsonpatch test err",
			Body:      `{ "foo": {} }`,
			Patch:     []string{`{ "op": "test", "path": "/foo", "value": null }`},
			ExpectErr: "op test /foo: testing path /foo value failed",
		},
		{
			Name:      "jsonpatch test err",
			Body:      `{ "foo": [] }`,
			Patch:     []string{`{ "op": "test", "path": "/foo", "value": null }`},
			ExpectErr: "op test /foo: testing path /foo value failed",
		},
		{
			Name:     "jsonpatch test",
			Body:     `{ "baz/foo": "qux" }`,
			Patch:    []string{`{ "op": "test", "path": "/baz~1foo", "value": "qux"}`},
			Expected: `{"baz/foo": "qux"}`,
		},
		{
			Name:      "jsonpatch test",
			Body:      `{ "foo": [] }`,
			Patch:     []string{`{ "op": "test", "path": "/foo"}`},
			ExpectErr: "op test /foo: testing path /foo value failed",
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			patch := make(Patch, len(test.Patch))
			for i, patchStr := range test.Patch {
				var op Operation
				err := json.Unmarshal([]byte(patchStr), &op)
				require.NoError(t, err)
				patch[i] = op
			}
			out, err := Apply([]byte(test.Body), patch)
			if test.ExpectErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.Expected, strings.TrimSpace(string(out)))
			} else {
				assert.EqualError(t, err, test.ExpectErr)
			}
		})
	}
}

func TestApplyJSONPatch_CustomObject(t *testing.T) {
	// Tests adding an object where the Value has custom yaml serialization
	type MyObject struct {
		FieldA string `yaml:"custom-tag"`
	}

	originalBytes := []byte(`existing-field: true`)

	out, err := Apply(originalBytes, Patch{
		{
			Type:  "add",
			Path:  MustParsePath("/custom-path"),
			Value: MyObject{FieldA: "custom-value"},
		},
	})
	require.NoError(t, err)

	expected := `existing-field: true
custom-path:
  custom-tag: custom-value
`
	require.Equal(t, expected, string(out))
}
