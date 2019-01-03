// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package boolean_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/boolean"
)

func TestBoolean(t *testing.T) {
	for _, test := range []struct {
		Name    string
		mapJSON string
		mapVal  map[boolean.Boolean]bool
	}{
		{
			Name:    "basic",
			mapJSON: `{"true":true,"false":false}`,
			mapVal:  map[boolean.Boolean]bool{true: true, false: false},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			var decodedMap map[boolean.Boolean]bool
			err := json.Unmarshal([]byte(test.mapJSON), &decodedMap)
			require.NoError(t, err)
			require.Equal(t, test.mapVal, decodedMap)
			encodedJSON, err := json.Marshal(decodedMap)
			require.NoError(t, err)
			require.JSONEq(t, test.mapJSON, string(encodedJSON))
		})
	}
}
