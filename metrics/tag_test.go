// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustTag(t *testing.T) {
	t.Run("must convert uppercase to lowercase", func(t *testing.T) {
		actual, err := NewTag("keyWithUpper", "valueWithUpper")
		assert.NoError(t, err)
		assert.Equal(t, Tag{
			key:   "keywithupper",
			value: "valuewithupper",
		},
			actual)
	})

	t.Run("must allow special chars", func(t *testing.T) {
		actual, err := NewTag("a_-./0", "a_-:./0")
		assert.NoError(t, err)
		assert.Equal(t, Tag{
			key:   "a_-./0",
			value: "a_-:./0",
		},
			actual)
	})

	t.Run("must convert invalid chars", func(t *testing.T) {
		actual, err := NewTag("a(❌)", "a(❌)")
		assert.NoError(t, err)
		assert.Equal(t, Tag{
			key:   "a___",
			value: "a___",
		},
			actual)
	})

	t.Run("must error when key+value empty", func(t *testing.T) {
		_, err := NewTag("", "")
		assert.Error(t, err)
	})

	t.Run("must error when key+value pair is longer that 200 chars", func(t *testing.T) {
		s := strings.Repeat("a", 100)
		_, err := NewTag(s, s)
		assert.Error(t, err)
	})
}
