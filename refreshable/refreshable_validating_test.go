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
}
