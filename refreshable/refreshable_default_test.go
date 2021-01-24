// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"testing"

	"github.com/palantir/pkg/refreshable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRefreshable(t *testing.T) {
	type container struct{ Value string }

	v := &container{Value: "original"}
	r := refreshable.NewDefaultRefreshable(v)
	assert.Equal(t, r.Current(), v)

	t.Run("Update", func(t *testing.T) {
		v2 := &container{Value: "updated"}
		err := r.Update(v2)
		require.NoError(t, err)
		assert.Equal(t, r.Current(), v2)
	})

	t.Run("Subscribe", func(t *testing.T) {
		var v1, v2 container
		unsub1 := r.Subscribe(func(i interface{}) {
			v1 = *(i.(*container))
		})
		_ = r.Subscribe(func(i interface{}) {
			v2 = *(i.(*container))
		})
		assert.Equal(t, v1.Value, "")
		assert.Equal(t, v2.Value, "")
		err := r.Update(&container{Value: "value"})
		require.NoError(t, err)
		assert.Equal(t, v1.Value, "value")
		assert.Equal(t, v2.Value, "value")

		unsub1()
		err = r.Update(&container{Value: "value2"})
		require.NoError(t, err)
		assert.Equal(t, v1.Value, "value", "should be unchanged after unsubscribing")
		assert.Equal(t, v2.Value, "value2", "should be updated after unsubscribing other")
	})

	t.Run("Map", func(t *testing.T) {
		err := r.Update(&container{Value: "value"})
		require.NoError(t, err)
		m := r.Map(func(i interface{}) interface{} {
			return len(i.(*container).Value)
		})
		assert.Equal(t, m.Current(), 5)

		err = r.Update(&container{Value: "updated"})
		require.NoError(t, err)
		assert.Equal(t, m.Current(), 7)
	})

}
