// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/remoting"
)

var testEIID = remoting.ErrorInstanceID{
	0x0, 0x1, 0x2, 0x3,
	0x4, 0x5, 0x6, 0x7,
	0x8, 0x9, 0xA, 0xB,
	0xC, 0xD, 0xE, 0xF,
}

func TestErrorInstanceID_MarshalJSON(t *testing.T) {
	serialized, err := json.Marshal(testEIID)
	assert.NoError(t, err)
	assert.Equal(t, `"00010203-0405-0607-0809-0a0b0c0d0e0f"`, string(serialized))
}

func TestErrorInstanceID_UnmarshalJSON(t *testing.T) {
	t.Run("correct lower case", func(t *testing.T) {
		var actual remoting.ErrorInstanceID
		err := json.Unmarshal([]byte(`"00010203-0405-0607-0809-0a0b0c0d0e0f"`), &actual)
		assert.NoError(t, err)
		assert.Equal(t, testEIID, actual)
	})

	t.Run("correct upper case", func(t *testing.T) {
		var actual remoting.ErrorInstanceID
		err := json.Unmarshal([]byte(`"00010203-0405-0607-0809-0A0B0C0D0E0F"`), &actual)
		assert.NoError(t, err)
		assert.Equal(t, testEIID, actual)
	})

	t.Run("incorrect group", func(t *testing.T) {
		var actual remoting.ErrorInstanceID
		err := json.Unmarshal([]byte(`"00010203-04Z5-0607-0809-0A0B0C0D0E0F"`), &actual)
		assert.EqualError(t, err, "remoting: invalid UUID format")
	})
}
