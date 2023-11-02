// Copyright (c) 2023 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToV2(t *testing.T) {
	v1 := NewDefaultRefreshable("original")
	v2 := ToV2[string](v1)
	assert.Equal(t, v2.Current(), "original")
	assert.NoError(t, v1.Update("updated"))
	assert.Equal(t, v2.Current(), "updated")
}
