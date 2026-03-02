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
	require.Equal(t, "value", vr.Current().Validated.Value)

	// attempt bad update
	r.Update(container{})
	require.Equal(t, r.Current().Value, "")
	v, err = vr.Validation()
	require.EqualError(t, err, "empty", "expected validation error")
	require.Equal(t, "", v.Value, "expected invalid value from Validation")
	require.Equal(t, vr.Current().Validated.Value, "value", "expected unchanged validating refreshable")

	// attempt good update
	r.Update(container{Value: "value2"})
	v, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, "value2", v.Value)
	require.Equal(t, "value2", vr.Current().Validated.Value)
	require.Equal(t, "value2", r.Current().Value)
}

func TestMapValidatingRefreshable(t *testing.T) {
	r := refreshable.New("https://palantir.com:443")
	vr, _, err := refreshable.MapWithError[string, *url.URL](r, url.Parse)
	require.NoError(t, err)
	_, err = vr.Validation()
	require.NoError(t, err)
	host, _, err := refreshable.MapWithError(vr, func(c refreshable.ValidRefreshableContainer[*url.URL]) (string, error) {
		return c.Validated.Hostname(), c.LastErr
	})
	require.NoError(t, err)
	_, err = host.Validation()
	require.NoError(t, err)
	char, _, err := refreshable.MapWithError(host, func(c refreshable.ValidRefreshableContainer[string]) (string, error) {
		return c.Validated[0:1], c.LastErr
	})
	require.NoError(t, err)
	_, err = char.Validation()
	require.NoError(t, err)

	require.Equal(t, r.Current(), "https://palantir.com:443")
	require.Equal(t, vr.Current().Validated.Hostname(), "palantir.com")
	require.Equal(t, "palantir.com", host.Current().Validated)
	require.Equal(t, "p", char.Current().Validated)

	// attempt bad update
	r.Update(":::error.com")
	assert.Equal(t, r.Current(), ":::error.com")
	_, err = vr.Validation()
	require.EqualError(t, err, "parse \":::error.com\": missing protocol scheme", "expected err from validating refreshable")
	assert.Equal(t, vr.Current().Validated.Hostname(), "palantir.com", "expected unchanged validating refreshable")
	_, err = host.Validation()
	assert.Error(t, err)
	require.Equal(t, "palantir.com", host.Current().Validated)
	_, err = char.Validation()
	require.Equal(t, "p", char.Current().Validated)
	assert.Error(t, err)

	// attempt good update
	r.Update("https://example.com")
	_, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, r.Current(), "https://example.com")
	require.Equal(t, vr.Current().Validated.Hostname(), "example.com")
}

func TestMapValidatingRefreshableChained(t *testing.T) {
	r := refreshable.New("https://palantir.com:443")
	vr, _, err := refreshable.MapWithError[string, *url.URL](r, url.Parse)
	require.NoError(t, err)
	host, _, err := refreshable.MapWithError(vr, func(c refreshable.ValidRefreshableContainer[*url.URL]) (string, error) {
		return c.Validated.Hostname(), c.LastErr
	})
	require.NoError(t, err)
	hostLen, _, err := refreshable.MapWithError(host, func(c refreshable.ValidRefreshableContainer[string]) (int, error) {
		return len(c.Validated), c.LastErr
	})
	require.NoError(t, err)
	require.Equal(t, "palantir.com", host.Current().Validated)
	require.Equal(t, 12, hostLen.Current().Validated)
	// attempt bad update — error propagates through the entire chain
	r.Update(":::error.com")
	_, err = host.Validation()
	assert.Error(t, err)
	_, err = hostLen.Validation()
	assert.Error(t, err)
	assert.Equal(t, "palantir.com", host.Current().Validated, "expected unchanged validated value")
	assert.Equal(t, 12, hostLen.Current().Validated, "expected unchanged validated value")
	// attempt good update — recovery propagates through the entire chain
	r.Update("https://example.com")
	_, err = host.Validation()
	assert.NoError(t, err)
	_, err = hostLen.Validation()
	assert.NoError(t, err)
	assert.Equal(t, "example.com", host.Current().Validated)
	assert.Equal(t, 11, hostLen.Current().Validated)
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
		return vr.Current().Validated == 2
	}, time.Second, time.Millisecond)

	assert.True(t, seen1, "expected to process 1 value")
	assert.True(t, seen2, "expected to process 2 value")
}
