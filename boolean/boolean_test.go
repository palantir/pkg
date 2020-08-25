// Copyright (c) 2020 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package boolean

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoolean_Marshal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  bool
		output []byte
	}{
		{
			name:   "false input",
			input:  false,
			output: []byte(`"false"`),
		},
		{
			name:   "true input",
			input:  true,
			output: []byte(`"true"`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(Boolean(tc.input))
			require.NoError(t, err)
			require.Equal(t, out, tc.output)
		})
	}
}

func TestBoolean_Unmarshal(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     []byte
		output    Boolean
		expectErr bool
	}{
		{
			name:      "nil input",
			input:     nil,
			output:    false,
			expectErr: true,
		},
		{
			name:      "empty input",
			input:     []byte(`""`),
			output:    false,
			expectErr: true,
		},
		{
			name:      "false input",
			input:     []byte(`"false"`),
			output:    false,
			expectErr: false,
		},
		{
			name:      "true input",
			input:     []byte(`"true"`),
			output:    true,
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b Boolean
			err := json.Unmarshal(tc.input, &b)
			require.Equal(t, tc.expectErr, err != nil)
			require.Equal(t, b, tc.output)
		})
	}
}
