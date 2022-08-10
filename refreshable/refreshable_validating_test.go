// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatingRefreshable(t *testing.T) {
	type container struct{ Value string }
	r := refreshable.New(container{Value: "value"})
	vr, err := refreshable.NewValidatingRefreshable[container](r, func(i container) error {
		if len(i.Value) == 0 {
			return errors.New("empty")
		}
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current().Value, "value")
	require.Equal(t, vr.Current().Value, "value")

	// attempt bad update
	r.Update(container{})
	require.Equal(t, r.Current().Value, "")

	require.EqualError(t, vr.LastValidateErr(), "empty", "expected err from validating refreshable")
	require.Equal(t, vr.Current().Value, "value", "expected unchanged validating refreshable")

	// attempt good update
	r.Update(container{Value: "value2"})
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, "value2", vr.Current().Value)
	require.Equal(t, "value2", r.Current().Value)
}

func TestMapValidatingRefreshable(t *testing.T) {
	r := refreshable.New("https://palantir.com:443")
	vr, err := refreshable.NewMapValidatingRefreshable[string, *url.URL](r, url.Parse)
	require.NoError(t, err)
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current(), "https://palantir.com:443")
	require.Equal(t, vr.Current().Hostname(), "palantir.com")

	// attempt bad update
	r.Update(":::error.com")
	assert.Equal(t, r.Current(), ":::error.com")
	require.EqualError(t, vr.LastValidateErr(), "parse \":::error.com\": missing protocol scheme", "expected err from validating refreshable")
	assert.Equal(t, vr.Current().Hostname(), "palantir.com", "expected unchanged validating refreshable")

	// attempt good update
	r.Update("https://example.com")
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current(), "https://example.com")
	require.Equal(t, vr.Current().Hostname(), "example.com")
}

// TestValidatingRefreshable_SubscriptionRaceCondition tests that the ValidatingRefreshable stays current
// if the underlying refreshable updates during the creation process.
func TestValidatingRefreshable_SubscriptionRaceCondition(t *testing.T) {
	r := &updateImmediatelyRefreshable{r: refreshable.New(1), newValue: 2}
	vr, err := refreshable.NewValidatingRefreshable[int](r, func(i int) error { return nil })
	require.NoError(t, err)
	// If this returns 1, it is likely because the VR contains a stale value
	assert.Equal(t, 2, vr.Current())
}

// updateImmediatelyRefreshable is a mock implementation which updates to newValue immediately when Current() is called
type updateImmediatelyRefreshable struct {
	r        *refreshable.DefaultRefreshable[int]
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
