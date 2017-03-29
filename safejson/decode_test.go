// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeNumber(t *testing.T) {
	in := map[string]interface{}{
		"a": json.Number("12"),
		"b": json.Number("34"),
		"c": json.Number("56"),
	}

	var out struct {
		First  int         `json:"a"`
		Second json.Number `json:"b"`
		Third  interface{} `json:"c"`
	}

	err := Decode(in, &out)
	assert.NoError(t, err)

	assert.Equal(t, 12, out.First)
	assert.Equal(t, json.Number("34"), out.Second)
	assert.Equal(t, json.Number("56"), out.Third)
}
