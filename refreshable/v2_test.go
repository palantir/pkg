// Copyright (c) 2023 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToFromV2(t *testing.T) {
	base := NewDefaultRefreshable("original")
	v2 := ToV2[string](base)
	v1 := FromV2(v2)
	assert.Equal(t, base.Current(), "original", "base missing original value")
	assert.Equal(t, v2.Current(), "original", "v2 missing original value")
	assert.Equal(t, v1.Current(), "original", "v1 missing original value")

	assert.NoError(t, base.Update("updated"))
	assert.Equal(t, base.Current(), "updated", "base missing updated value")
	assert.Equal(t, v2.Current(), "updated", "v2 missing updated value")
	assert.Equal(t, v1.Current(), "updated", "v1 missing updated value")
}

func TestToV2RespectsUpdates(t *testing.T) {
	for i := 0; i < 10000; i++ {
		v1 := NewDefaultRefreshable(1)
		updateCalled := make(chan struct{})
		go func() {
			require.NoError(t, v1.Update(2))
			close(updateCalled)
		}()
		v2 := ToV2[int](v1)
		<-updateCalled
		assert.Equal(t, 2, v2.Current(),
			"v2 should always be updated after Update is called")
	}
}
