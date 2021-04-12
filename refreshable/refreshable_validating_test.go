// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/palantir/pkg/refreshable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatingRefreshable(t *testing.T) {
	type container struct{ Value string }
	r := refreshable.NewDefaultRefreshable(container{Value: "value"})
	vr, err := refreshable.NewValidatingRefreshable(r, func(i interface{}) error {
		if len(i.(container).Value) == 0 {
			return errors.New("empty")
		}
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current().(container).Value, "value")
	require.Equal(t, vr.Current().(container).Value, "value")

	// attempt bad update
	err = r.Update(container{})
	require.NoError(t, err, "no err expected from default refreshable")
	require.Equal(t, r.Current().(container).Value, "")

	require.EqualError(t, vr.LastValidateErr(), "empty", "expected err from validating refreshable")
	require.Equal(t, vr.Current().(container).Value, "value", "expected unchanged validating refreshable")

	// attempt good update
	require.NoError(t, r.Update(container{Value: "value2"}))
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, "value2", vr.Current().(container).Value)
	require.Equal(t, "value2", r.Current().(container).Value)
}

func TestMapValidatingRefreshable(t *testing.T) {
	r := refreshable.NewDefaultRefreshable("https://palantir.com:443")
	vr, err := refreshable.NewMapValidatingRefreshable(r, func(i interface{}) (interface{}, error) {
		return url.Parse(i.(string))
	})
	require.NoError(t, err)
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current().(string), "https://palantir.com:443")
	require.Equal(t, vr.Current().(*url.URL).Hostname(), "palantir.com")

	// attempt bad update
	err = r.Update(":::error.com")
	require.NoError(t, err, "no err expected from default refreshable")
	assert.Equal(t, r.Current().(string), ":::error.com")
	require.EqualError(t, vr.LastValidateErr(), "parse \":::error.com\": missing protocol scheme", "expected err from validating refreshable")
	assert.Equal(t, vr.Current().(*url.URL).Hostname(), "palantir.com", "expected unchanged validating refreshable")

	// attempt good update
	require.NoError(t, r.Update("https://example.com"))
	require.NoError(t, vr.LastValidateErr())
	require.Equal(t, r.Current().(string), "https://example.com")
	require.Equal(t, vr.Current().(*url.URL).Hostname(), "example.com")
}

// TestValidatingRefreshable_SubscriptionRaceCondition tests that the ValidatingRefreshable stays current
// if the underlying refreshable updates during the creation process.
func TestValidatingRefreshable_SubscriptionRaceCondition(t *testing.T) {
	r := &updateImmediatelyRefreshable{r: refreshable.NewDefaultRefreshable(1), newValue: 2}
	vr, err := refreshable.NewValidatingRefreshable(r, func(i interface{}) error { return nil })
	require.NoError(t, err)
	// If this returns 1, it is likely because the VR contains a stale value
	assert.Equal(t, 2, vr.Current())
}

// updateImmediatelyRefreshable is a mock implementation which updates to newValue immediately when Current() is called
type updateImmediatelyRefreshable struct {
	r        *refreshable.DefaultRefreshable
	newValue interface{}
}

func (r *updateImmediatelyRefreshable) Current() interface{} {
	c := r.r.Current()
	_ = r.r.Update(r.newValue)
	return c
}

func (r *updateImmediatelyRefreshable) Subscribe(f func(interface{})) func() {
	return r.r.Subscribe(f)
}

func (r *updateImmediatelyRefreshable) Map(f func(interface{}) interface{}) refreshable.Refreshable {
	return r.r.Map(f)
}
