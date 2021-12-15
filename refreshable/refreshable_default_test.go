// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRefreshable(t *testing.T) {
	type container struct{ Value string }

	v := &container{Value: "original"}
	r := refreshable.NewDefaultRefreshable[*container](v)
	assert.Equal(t, r.Current(), v)

	t.Run("Update", func(t *testing.T) {
		v2 := &container{Value: "updated"}
		err := r.Update(v2)
		require.NoError(t, err)
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
		m := r.Map(func(i *container) any {
			return len(i.Value)
		})
		assert.Equal(t, m.Current(), 5)

		err = r.Update(&container{Value: "updated"})
		require.NoError(t, err)
		assert.Equal(t, m.Current(), 7)
	})
}

func TestTypes(t *testing.T) {
	// bool
	b := refreshable.NewDefaultRefreshable(true)
	assert.Equal(t, true, b.Current())
	assert.NoError(t, b.Update(false))
	assert.Equal(t, false, b.Current())
	nr := b.Map(func(t bool) any {
		if t {
			return "true"
		} else {
			return "false"
		}
	})
	assert.Equal(t, "false", nr.Current())

	// *bool
	bPtr := refreshable.NewDefaultRefreshable(ptr(true))
	assert.Equal(t, ptr(true), bPtr.Current())
	assert.NoError(t, bPtr.Update(ptr(false)))
	assert.Equal(t, ptr(false), bPtr.Current())
	bnr := bPtr.Map(func(t *bool) any {
		if *t {
			return "true"
		} else {
			return "false"
		}
	})
	assert.Equal(t, "false", bnr.Current())

	// duration
	d1 := refreshable.NewDefaultRefreshable(time.Second)
	assert.Equal(t, time.Second, d1.Current())
	assert.NoError(t, d1.Update(10*time.Second))
	assert.Equal(t, 10*time.Second, d1.Current())
	d1nr := d1.Map(func(d time.Duration) any {
		return d.String()
	})
	assert.Equal(t, (10 * time.Second).String(), d1nr.Current())

	// *duration
	d2 := refreshable.NewDefaultRefreshable(ptr(time.Second))
	assert.Equal(t, ptr(time.Second), d2.Current())
	assert.NoError(t, d2.Update(ptr(10*time.Second)))
	assert.Equal(t, ptr(10*time.Second), d2.Current())
	d2nr := d2.Map(func(d *time.Duration) any {
		return d.String()
	})
	assert.Equal(t, (10 * time.Second).String(), d2nr.Current())

	// []string
	v1 := []string{"hello"}
	v2 := []string{"world"}
	ls := refreshable.NewDefaultRefreshable(v1)
	assert.Equal(t, v1, ls.Current())
	assert.NoError(t, ls.Update(v2))
	assert.Equal(t, v2, ls.Current())
	lsnr := ls.Map(func(t []string) any {
		return "test"
	})
	assert.Equal(t, "test", lsnr.Current())

	// custom type
	type myVal struct {
		first string
		last  string
	}
	mv1 := myVal{
		first: "hello",
		last:  "world",
	}
	mv2 := myVal{
		first: "another",
		last:  "world",
	}
	m1 := refreshable.NewDefaultRefreshable(mv1)
	assert.Equal(t, mv1, m1.Current())
	assert.NoError(t, m1.Update(mv2))
	assert.Equal(t, mv2, m1.Current())
	mnr := m1.Map(func(t myVal) any {
		return t.first
	})
	assert.Equal(t, mv2.first, mnr.Current())

	// TODO(tabboud): Add all types
	// float64
	// *float64
	// int
	// *int
	// int64
	// *int64
	// string
	// *string
}

func ptr[T any](val T) *T {
	return &val
}
