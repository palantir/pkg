// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import (
	"encoding/json"
	"testing"

	"github.com/palantir/pkg/binary"
	"github.com/stretchr/testify/assert"
)

func TestBinary_Marshal(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Input  []byte
		Output []byte
	}{
		{
			Name:   "hello world",
			Input:  []byte(`hello world`),
			Output: []byte(`"aGVsbG8gd29ybGQ="`),
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			out, err := json.Marshal(binary.New(test.Input))
			assert.NoError(t, err)
			assert.Equal(t, string(test.Output), string(out))
		})
	}
}

func TestBinary_Unmarshal(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Input  []byte
		Output []byte
	}{
		{
			Name:   "hello world",
			Input:  []byte(`"aGVsbG8gd29ybGQ="`),
			Output: []byte(`hello world`),
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			var bin binary.Binary
			err := json.Unmarshal(test.Input, &bin)
			assert.NoError(t, err)
			bytes, err := bin.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, string(test.Output), string(bytes))
		})
	}
}
