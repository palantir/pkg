// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	for _, test := range []struct {
		In  string
		Out Path
	}{
		{
			In:  "/",
			Out: Path{""},
		},
		{
			In:  "/foo",
			Out: Path{"", "foo"},
		},
		{
			In:  "/foo/bar",
			Out: Path{"", "foo", "bar"},
		},
	} {
		t.Run(test.In, func(t *testing.T) {
			var p Path
			err := p.UnmarshalText([]byte(test.In))
			require.NoError(t, err)
			if assert.Equal(t, test.Out, p) {
				assert.Equal(t, test.In, p.String(), "roundtrip equality failed")
			}
		})
	}
}
