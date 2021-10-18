// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bearertoken

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	for _, test := range []struct {
		Input string
		Error string
	}{
		{"", "empty bearer token"},
		{"valid", ""},
		{"valid__underscores", ""},
		{"valid=", ""},
		{"=invalid", "invalid character '=' for bearer token"},
		{"in valid", "invalid character ' ' for bearer token"},
		{"in!valid", "invalid character '!' for bearer token"},
	} {
		t.Run(test.Input, func(t *testing.T) {
			tok, err := New(test.Input)
			if test.Error != "" {
				require.EqualError(t, err, test.Error)
			} else {
				require.NoError(t, err)
				require.Equal(t, Token(test.Input), tok)
			}
		})
	}
}
