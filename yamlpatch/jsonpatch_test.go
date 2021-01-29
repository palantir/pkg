// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//// This file adapted from https://github.com/evanphx/json-patch/blob/master/patch_test.go to verify JSON Patch correctness.
//
package yamlpatch

//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"reflect"
//	"testing"
//)
//
//func reformatJSON(j string) string {
//	buf := new(bytes.Buffer)
//
//	json.Indent(buf, []byte(j), "", "  ")
//
//	return buf.String()
//}
//
//func compareJSON(a, b string) bool {
//	// return Equal([]byte(a), []byte(b))
//
//	var objA, objB map[string]interface{}
//	json.Unmarshal([]byte(a), &objA)
//	json.Unmarshal([]byte(b), &objB)
//
//	// fmt.Printf("Comparing %#v\nagainst %#v\n", objA, objB)
//	return reflect.DeepEqual(objA, objB)
//}
//
//func applyPatch(docStr, patchStr string) (string, error) {
//	var patch Patch
//	if err := json.Unmarshal([]byte(patchStr), &patch); err != nil {
//		return "", err
//	}
//	out, err := Apply([]byte(docStr), patch)
//	if err != nil {
//		return "", err
//	}
//	return string(out), nil
//}
//
//type Case struct {
//	doc, patch, result string
//}
//
//func repeatedA(r int) string {
//	var s string
//	for i := 0; i < r; i++ {
//		s += "A"
//	}
//	return s
//}
//

//var Cases = []Case{
//	{
//		doc: `{ "foo": "bar"}`,
//		patch: `[
//        { "op": "add", "path": "/baz", "value": "qux" }
//    ]`,
//		result: `{
//      "baz": "qux",
//      "foo": "bar"
//    }`,
//	},
//	{
//		doc: `{ "foo": [ "bar", "baz" ] }`,
//		patch: `[
//    { "op": "add", "path": "/foo/1", "value": "qux" }
//   ]`,
//		result: `{ "foo": [ "bar", "qux", "baz" ] }`,
//	},
//	{
//		doc: `{ "foo": [ "bar", "baz" ] }`,
//		patch: `[
//     { "op": "add", "path": "/foo/-1", "value": "qux" }
//    ]`,
//		result: `{ "foo": [ "bar", "baz", "qux" ] }`,
//	},
//	{
//		doc:    `{ "baz": "qux", "foo": "bar" }`,
//		patch:  `[ { "op": "remove", "path": "/baz" } ]`,
//		result: `{ "foo": "bar" }`,
//	},
//	{
//		doc:    `{ "foo": [ "bar", "qux", "baz" ] }`,
//		patch:  `[ { "op": "remove", "path": "/foo/1" } ]`,
//		result: `{ "foo": [ "bar", "baz" ] }`,
//	},
//	{
//		doc:    `{ "baz": "qux", "foo": "bar" }`,
//		patch:  `[ { "op": "replace", "path": "/baz", "value": "boo" } ]`,
//		result: `{ "baz": "boo", "foo": "bar" }`,
//	},
//	{
//		doc: `{
//	"foo": {
//		"bar": "baz",
//		"waldo": "fred"
//	},
//	"qux": {
//		"corge": "grault"
//	}
//}`,
//		patch: `[ { "op": "move", "from": "/foo/waldo", "path": "/qux/thud" } ]`,
//		result: `{
//    "foo": {
//      "bar": "baz"
//    },
//    "qux": {
//      "corge": "grault",
//      "thud": "fred"
//    }
//  }`,
//	},
//	{
//		doc:    `{ "foo": [ "all", "grass", "cows", "eat" ] }`,
//		patch:  `[ { "op": "move", "from": "/foo/1", "path": "/foo/3" } ]`,
//		result: `{ "foo": [ "all", "cows", "eat", "grass" ] }`,
//	},
//	{
//		doc:    `{ "foo": [ "all", "grass", "cows", "eat" ] }`,
//		patch:  `[ { "op": "move", "from": "/foo/1", "path": "/foo/2" } ]`,
//		result: `{ "foo": [ "all", "cows", "grass", "eat" ] }`,
//	},
//	{
//		doc:    `{ "foo": "bar" }`,
//		patch:  `[ { "op": "add", "path": "/child", "value": { "grandchild": { } } } ]`,
//		result: `{ "foo": "bar", "child": { "grandchild": { } } }`,
//	},
//	{
//		doc:    `{ "foo": ["bar"] }`,
//		patch:  `[ { "op": "add", "path": "/foo/-", "value": ["abc", "def"] } ]`,
//		result: `{ "foo": ["bar", ["abc", "def"]] }`,
//	},
//	{
//		doc:    `{ "foo": "bar", "qux": { "baz": 1, "bar": null } }`,
//		patch:  `[ { "op": "remove", "path": "/qux/bar" } ]`,
//		result: `{ "foo": "bar", "qux": { "baz": 1 } }`,
//	},
//	{
//		doc:    `{ "foo": "bar" }`,
//		patch:  `[ { "op": "add", "path": "/baz", "value": null } ]`,
//		result: `{ "baz": null, "foo": "bar" }`,
//	},
//	{
//		doc:    `{ "foo": ["bar"]}`,
//		patch:  `[ { "op": "replace", "path": "/foo/0", "value": "baz"}]`,
//		result: `{ "foo": ["baz"]}`,
//	},
//	{
//		doc:    `{ "foo": ["bar","baz"]}`,
//		patch:  `[ { "op": "replace", "path": "/foo/0", "value": "bum"}]`,
//		result: `{ "foo": ["bum","baz"]}`,
//	},
//	{
//		doc:    `{ "foo": ["bar","qux","baz"]}`,
//		patch:  `[ { "op": "replace", "path": "/foo/1", "value": "bum"}]`,
//		result: `{ "foo": ["bar", "bum","baz"]}`,
//	},
//	{
//		doc:    `[ {"foo": ["bar","qux","baz"]}]`,
//		patch:  `[ { "op": "replace", "path": "/0/foo/0", "value": "bum"}]`,
//		result: `[ {"foo": ["bum","qux","baz"]}]`,
//	},
//	{
//		doc:    `[ {"foo": ["bar","qux","baz"], "bar": ["qux","baz"]}]`,
//		patch:  `[ { "op": "copy", "from": "/0/foo/0", "path": "/0/bar/0"}]`,
//		result: `[ {"foo": ["bar","qux","baz"], "bar": ["bar", "baz"]}]`,
//	},
//	{
//		doc:    `[ {"foo": ["bar","qux","baz"], "bar": ["qux","baz"]}]`,
//		patch:  `[ { "op": "copy", "from": "/0/foo/0", "path": "/0/bar"}]`,
//		result: `[ {"foo": ["bar","qux","baz"], "bar": ["bar", "qux", "baz"]}]`,
//	},
//	{
//		doc:    `[ { "foo": {"bar": ["qux","baz"]}, "baz": {"qux": "bum"}}]`,
//		patch:  `[ { "op": "copy", "from": "/0/foo/bar", "path": "/0/baz/bar"}]`,
//		result: `[ { "baz": {"bar": ["qux","baz"], "qux":"bum"}, "foo": {"bar": ["qux","baz"]}}]`,
//	},
//	{
//		doc:    `{ "foo": ["bar"]}`,
//		patch:  `[{"op": "copy", "path": "/foo/0", "from": "/foo"}]`,
//		result: `{ "foo": [["bar"], "bar"]}`,
//	},
//	{
//		doc:    `{ "foo": ["bar","qux","baz"]}`,
//		patch:  `[ { "op": "remove", "path": "/foo/-2"}]`,
//		result: `{ "foo": ["bar", "baz"]}`,
//	},
//	{
//		doc:    `{ "foo": []}`,
//		patch:  `[ { "op": "add", "path": "/foo/-1", "value": "qux"}]`,
//		result: `{ "foo": ["qux"]}`,
//	},
//	{
//		doc:    `{ "bar": [{"baz": null}]}`,
//		patch:  `[ { "op": "replace", "path": "/bar/0/baz", "value": 1 } ]`,
//		result: `{ "bar": [{"baz": 1}]}`,
//	},
//	{
//		doc:    `{ "bar": [{"baz": 1}]}`,
//		patch:  `[ { "op": "replace", "path": "/bar/0/baz", "value": null } ]`,
//		result: `{ "bar": [{"baz": null}]}`,
//	},
//	{
//		doc:    `{ "bar": [null]}`,
//		patch:  `[ { "op": "replace", "path": "/bar/0", "value": 1 } ]`,
//		result: `{ "bar": [1]}`,
//	},
//	{
//		doc:    `{ "bar": [1]}`,
//		patch:  `[ { "op": "replace", "path": "/bar/0", "value": null } ]`,
//		result: `{ "bar": [null]}`,
//	},
//	{
//		doc: fmt.Sprintf(`{ "foo": ["A", %q] }`, repeatedA(48)),
//		// The wrapping quotes around 'A's are included in the copy
//		// size, so each copy operation increases the size by 50 bytes.
//		patch: `[ { "op": "copy", "path": "/foo/-", "from": "/foo/1" },
//		   { "op": "copy", "path": "/foo/-", "from": "/foo/1" }]`,
//		result: fmt.Sprintf(`{ "foo": ["A", %q, %q, %q] }`, repeatedA(48), repeatedA(48), repeatedA(48)),
//	},
//}

//
//type BadCase struct {
//	doc, patch string
//}
//
//var MutationTestCases = []BadCase{
//	{
//		doc:   `{ "foo": "bar", "qux": { "baz": 1, "bar": null } }`,
//		patch: `[ { "op": "remove", "path": "/qux/bar" } ]`,
//	},
//	{
//		doc:   `{ "foo": "bar", "qux": { "baz": 1, "bar": null } }`,
//		patch: `[ { "op": "replace", "path": "/qux/baz", "value": null } ]`,
//	},
//}
//
//var BadCases = []BadCase{
//	{
//		doc:   `{ "foo": "bar" }`,
//		patch: `[ { "op": "add", "path": "/baz/bat", "value": "qux" } ]`,
//	},
//	{
//		doc:   `{ "a": { "b": { "d": 1 } } }`,
//		patch: `[ { "op": "remove", "path": "/a/b/c" } ]`,
//	},
//	{
//		doc:   `{ "a": { "b": { "d": 1 } } }`,
//		patch: `[ { "op": "move", "from": "/a/b/c", "path": "/a/b/e" } ]`,
//	},
//	{
//		doc:   `{ "a": { "b": [1] } }`,
//		patch: `[ { "op": "remove", "path": "/a/b/1" } ]`,
//	},

//	{
//		doc:   `{ "a": { "b": [1] } }`,
//		patch: `[ { "op": "move", "from": "/a/b/1", "path": "/a/b/2" } ]`,
//	},
//	{
//		doc:   `{ "foo": "bar" }`,
//		patch: `[ { "op": "add", "pathz": "/baz", "value": "qux" } ]`,
//	},
//	{
//		doc:   `{ "foo": "bar" }`,
//		patch: `[ { "op": "add", "path": "", "value": "qux" } ]`,
//	},
//	{
//		doc:   `{ "foo": ["bar","baz"]}`,
//		patch: `[ { "op": "replace", "path": "/foo/2", "value": "bum"}]`,
//	},
//	{
//		doc:   `{ "foo": ["bar","baz"]}`,
//		patch: `[ { "op": "add", "path": "/foo/-4", "value": "bum"}]`,
//	},
//	{
//		doc:   `{ "name":{ "foo": "bat", "qux": "bum"}}`,
//		patch: `[ { "op": "replace", "path": "/foo/bar", "value":"baz"}]`,
//	},
//	{
//		doc:   `{ "foo": ["bar"]}`,
//		patch: `[ {"op": "add", "path": "/foo/2", "value": "bum"}]`,
//	},
//	{
//		doc:   `{ "foo": []}`,
//		patch: `[ {"op": "remove", "path": "/foo/-"}]`,
//	},
//	{
//		doc:   `{ "foo": []}`,
//		patch: `[ {"op": "remove", "path": "/foo/-1"}]`,
//	},
//	{
//		doc:   `{ "foo": ["bar"]}`,
//		patch: `[ {"op": "remove", "path": "/foo/-2"}]`,
//	},
//	{
//		doc:   `{}`,
//		patch: `[ {"op":null,"path":""} ]`,
//	},
//	{
//		doc:   `{}`,
//		patch: `[ {"op":"add","path":null} ]`,
//	},
//	{
//		doc:   `{}`,
//		patch: `[ { "op": "copy", "from": null }]`,
//	},
//	{
//		doc:   `{ "foo": ["bar"]}`,
//		patch: `[{"op": "copy", "path": "/foo/6666666666", "from": "/"}]`,
//	},
//	// Can't copy into an index greater than the size of the array
//	{
//		doc:   `{ "foo": ["bar"]}`,
//		patch: `[{"op": "copy", "path": "/foo/2", "from": "/foo/0"}]`,
//	},
//	// Accumulated copy size cannot exceed AccumulatedCopySizeLimit.
//	{
//		doc: fmt.Sprintf(`{ "foo": ["A", %q] }`, repeatedA(49)),
//		// The wrapping quotes around 'A's are included in the copy
//		// size, so each copy operation increases the size by 51 bytes.
//		patch: `[ { "op": "copy", "path": "/foo/-", "from": "/foo/1" },
//		   { "op": "copy", "path": "/foo/-", "from": "/foo/1" }]`,
//	},
//	// Can't move into an index greater than or equal to the size of the array
//	{
//		doc:   `{ "foo": [ "all", "grass", "cows", "eat" ] }`,
//		patch: `[ { "op": "move", "from": "/foo/1", "path": "/foo/4" } ]`,
//	},
//}

//
//func TestAllCases(t *testing.T) {
//	for _, c := range Cases {
//		out, err := applyPatch(c.doc, c.patch)
//
//		if err != nil {
//			t.Errorf("Unable to apply patch: %s", err)
//		}
//
//		if !compareJSON(out, c.result) {
//			t.Errorf("Patch did not apply. Expected:\n%s\n\nActual:\n%s",
//				reformatJSON(c.result), reformatJSON(out))
//		}
//	}
//
//	for _, c := range MutationTestCases {
//		out, err := applyPatch(c.doc, c.patch)
//
//		if err != nil {
//			t.Errorf("Unable to apply patch: %s", err)
//		}
//
//		if compareJSON(out, c.doc) {
//			t.Errorf("Patch did not apply. Original:\n%s\n\nPatched:\n%s",
//				reformatJSON(c.doc), reformatJSON(out))
//		}
//	}
//
//	for _, c := range BadCases {
//		_, err := applyPatch(c.doc, c.patch)
//
//		if err == nil {
//			t.Errorf("Patch %q should have failed to apply but it did not", c.patch)
//		}
//	}
//}
//

//type TestCase struct {
//	doc, patch string
//	result     bool
//	failedPath string
//}
//
//var TestCases = []TestCase{
//	{
//		`{
//      "baz": "qux",
//      "foo": [ "a", 2, "c" ]
//    }`,
//		`[
//      { "op": "test", "path": "/baz", "value": "qux" },
//      { "op": "test", "path": "/foo/1", "value": 2 }
//    ]`,
//		true,
//		"",
//	},
//	{
//		`{ "baz": "qux" }`,
//		`[ { "op": "test", "path": "/baz", "value": "bar" } ]`,
//		false,
//		"/baz",
//	},
//	{
//		`{
//      "baz": "qux",
//      "foo": ["a", 2, "c"]
//    }`,
//		`[
//      { "op": "test", "path": "/baz", "value": "qux" },
//      { "op": "test", "path": "/foo/1", "value": "c" }
//    ]`,
//		false,
//		"/foo/1",
//	},
//	{
//		`{ "baz": "qux" }`,
//		`[ { "op": "test", "path": "/foo", "value": 42 } ]`,
//		false,
//		"/foo",
//	},
//	{
//		`{ "baz": "qux" }`,
//		`[ { "op": "test", "path": "/foo", "value": null } ]`,
//		true,
//		"",
//	},
//	{
//		`{ "foo": null }`,
//		`[ { "op": "test", "path": "/foo", "value": null } ]`,
//		true,
//		"",
//	},
//	{
//		`{ "foo": {} }`,
//		`[ { "op": "test", "path": "/foo", "value": null } ]`,
//		false,
//		"/foo",
//	},
//	{
//		`{ "foo": [] }`,
//		`[ { "op": "test", "path": "/foo", "value": null } ]`,
//		false,
//		"/foo",
//	},
//	{
//		`{ "baz/foo": "qux" }`,
//		`[ { "op": "test", "path": "/baz~1foo", "value": "qux"} ]`,
//		true,
//		"",
//	},
//	{
//		`{ "foo": [] }`,
//		`[ { "op": "test", "path": "/foo"} ]`,
//		false,
//		"/foo",
//	},
//}

//
//func TestAllTest(t *testing.T) {
//	for _, c := range TestCases {
//		_, err := applyPatch(c.doc, c.patch)
//
//		if c.result && err != nil {
//			t.Errorf("Testing failed when it should have passed: %s", err)
//		} else if !c.result && err == nil {
//			t.Errorf("Testing passed when it should have faild: %s", err)
//		} else if !c.result {
//			expected := fmt.Sprintf("testing value %s failed: test failed", c.failedPath)
//			if err.Error() != expected {
//				t.Errorf("Testing failed as expected but invalid message: expected [%s], got [%s]", expected, err)
//			}
//		}
//	}
//}
//

////func TestAdd(t *testing.T) {
////	testCases := []struct {
////		name string
////		key  string
////		val  interface{}
////		arr  interface{}
////		err  string
////	}{
////		{
////			name: "should work",
////			key:  "0",
////			val:  lazyNode{},
////			arr:  partialArray{},
////		},
////		{
////			name: "index too large",
////			key:  "1",
////			val:  lazyNode{},
////			arr:  partialArray{},
////			err:  "Unable to access invalid index: 1: invalid index referenced",
////		},
////		{
////			name: "negative should work",
////			key:  "-1",
////			val:  lazyNode{},
////			arr:  partialArray{},
////		},
////		{
////			name: "negative too small",
////			key:  "-2",
////			val:  lazyNode{},
////			arr:  partialArray{},
////			err:  "Unable to access invalid index: -2: invalid index referenced",
////		},
////		{
////			name: "negative but negative disabled",
////			key:  "-1",
////			val:  lazyNode{},
////			arr:  partialArray{},
////			err:  "Unable to access invalid index: -1: invalid index referenced",
////		},
////	}
////	for _, tc := range testCases {
////		t.Run(tc.name, func(t *testing.T) {
////			key := tc.key
////			arr := &tc.arr
////			val := &tc.val
////			err := arr.add(key, val)
////			if err == nil && tc.err != "" {
////				t.Errorf("Expected error but got none! %v", tc.err)
////			} else if err != nil && tc.err == "" {
////				t.Errorf("Did not expect error but go: %v", err)
////			} else if err != nil && err.Error() != tc.err {
////				t.Errorf("Expected error %v but got error %v", tc.err, err)
////			}
////		})
////	}
////}
