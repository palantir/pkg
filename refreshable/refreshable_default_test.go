// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"testing"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRefreshable(t *testing.T) {
	type container struct{ Value string }

	v := &container{Value: "original"}
	r := refreshable.New(v)
	assert.Equal(t, r.Current(), v)

	t.Run("Update", func(t *testing.T) {
		v2 := &container{Value: "updated"}
		r.Update(v2)
		assert.Equal(t, r.Current(), v2)
	})

	t.Run("Subscribe", func(t *testing.T) {
		var v1, v2 container
		unsub1 := r.Subscribe(func(i *container) {
			v1 = *i
		})
		_ = r.Subscribe(func(i *container) {
			v2 = *i
		})
		assert.Equal(t, v1.Value, "")
		assert.Equal(t, v2.Value, "")
		r.Update(&container{Value: "value"})
		assert.Equal(t, v1.Value, "value")
		assert.Equal(t, v2.Value, "value")

		unsub1()
		r.Update(&container{Value: "value2"})
		assert.Equal(t, v1.Value, "value", "should be unchanged after unsubscribing")
		assert.Equal(t, v2.Value, "value2", "should be updated after unsubscribing other")
	})

	t.Run("Map", func(t *testing.T) {
		r.Update(&container{Value: "value"})
		m, _ := refreshable.Map[*container, int](r, func(i *container) int {
			return len(i.Value)
		})
		assert.Equal(t, m.Current(), 5)

		r.Update(&container{Value: "updated"})
		assert.Equal(t, m.Current(), 7)
	})

}
