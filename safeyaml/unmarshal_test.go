// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safeyaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var data = `top-level:
  unquoted_0: unquoted
  "quoted_0": unquoted
  unquoted_1: "quoted"
  quoted_1: "quoted"
  int: 0
  bool: true`

func TestUnmarshal(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		input    string
		expected interface{}
		errMsg   string
	}{
		{
			name:  "test unmarshal",
			input: data,
			expected: map[string]interface{}{
				"top-level": map[string]interface{}{
					"unquoted_0": "unquoted",
					"quoted_0":   "unquoted",
					"unquoted_1": "quoted",
					"quoted_1":   "quoted",
					"int":        0,
					"bool":       true,
				},
			},
		},
		{
			name: "conflicting keys",
			input: `top-level:
  "5": ""
  5: ""`,
			errMsg: "conflicting key",
		},
	} {
		var got interface{}
		err := Unmarshal([]byte(testcase.input), &got)
		if testcase.errMsg == "" {
			if assert.NoError(t, err, testcase.name) {
				assert.Equal(t, testcase.expected, got, testcase.name)
			}
		} else {
			if assert.Error(t, err, testcase.name) {
				assert.Contains(t, err.Error(), testcase.errMsg, testcase.name)
			}
		}
	}
}
