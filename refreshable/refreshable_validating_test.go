// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatingRefreshable(t *testing.T) {
	type container struct{ Value string }
	r := refreshable.New(container{Value: "value"})
	vr, _, err := refreshable.Validate[container](r, func(i container) error {
		if len(i.Value) == 0 {
			return errors.New("empty")
		}
		return nil
	})
	require.NoError(t, err)
	v, err := vr.Validation()
	require.NoError(t, err)
	require.Equal(t, "value", v.Value)
	require.Equal(t, "value", r.Current().Value)
	require.Equal(t, "value", vr.Current().Value)

	// attempt bad update
	r.Update(container{})
	require.Equal(t, r.Current().Value, "")
	v, err = vr.Validation()
	require.EqualError(t, err, "empty", "expected validation error")
	require.Equal(t, "", v.Value, "expected invalid value from Validation")
	require.Equal(t, vr.Current().Value, "value", "expected unchanged validating refreshable")

	// attempt good update
	r.Update(container{Value: "value2"})
	v, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, "value2", v.Value)
	require.Equal(t, "value2", vr.Current().Value)
	require.Equal(t, "value2", r.Current().Value)
}

func TestMapValidatingRefreshable(t *testing.T) {
	r := refreshable.New("https://palantir.com:443")
	vr, _, err := refreshable.MapWithError[string, *url.URL](r, url.Parse)
	require.NoError(t, err)
	_, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, r.Current(), "https://palantir.com:443")
	require.Equal(t, vr.Current().Hostname(), "palantir.com")

	// attempt bad update
	r.Update(":::error.com")
	assert.Equal(t, r.Current(), ":::error.com")
	_, err = vr.Validation()
	require.EqualError(t, err, "parse \":::error.com\": missing protocol scheme", "expected err from validating refreshable")
	assert.Equal(t, vr.Current().Hostname(), "palantir.com", "expected unchanged validating refreshable")

	// attempt good update
	r.Update("https://example.com")
	_, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, r.Current(), "https://example.com")
	require.Equal(t, vr.Current().Hostname(), "example.com")
}

// TestValidatingRefreshable_SubscriptionRaceCondition tests that the ValidatingRefreshable stays current
// if the underlying refreshable updates during the creation process.
func TestValidatingRefreshable_SubscriptionRaceCondition(t *testing.T) {
	//r := &updateImmediatelyRefreshable{r: refreshable.New(1), newValue: 2}
	r := refreshable.New(1)
	var seen1, seen2 bool
	vr, _, err := refreshable.Validate[int](r, func(i int) error {
		go r.Update(2)
		switch i {
		case 1:
			seen1 = true
		case 2:
			seen2 = true
		}
		return nil
	})
	require.NoError(t, err)
	// If this returns 1, it is likely because the VR contains a stale value
	assert.Eventually(t, func() bool {
		return vr.Current() == 2
	}, time.Second, time.Millisecond)

	assert.True(t, seen1, "expected to process 1 value")
	assert.True(t, seen2, "expected to process 2 value")
}

// updateImmediatelyRefreshable is a mock implementation which updates to newValue immediately when Current() is called
type updateImmediatelyRefreshable struct {
	r        refreshable.Updatable[int]
	newValue int
}

func (r *updateImmediatelyRefreshable) Current() int {
	c := r.r.Current()
	r.r.Update(r.newValue)
	return c
}

func (r *updateImmediatelyRefreshable) Subscribe(f func(int)) refreshable.UnsubscribeFunc {
	return r.r.Subscribe(f)
}
