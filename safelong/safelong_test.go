// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safelong_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/palantir/pkg/safelong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var safeLongJSONs = []struct {
	val  int64
	json string
}{
	{
		val:  13,
		json: `13`,
	},
	{
		val:  -42,
		json: `-42`,
	},
	{
		val:  0,
		json: `0`,
	},
}

func TestSafeLongMarshal(t *testing.T) {
	for i, currCase := range safeLongJSONs {
		currSafeLong, err := safelong.NewSafeLong(currCase.val)
		require.NoError(t, err, "Case %d", i)
		bytes, err := json.Marshal(currSafeLong)
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, currCase.json, string(bytes), "Case %d", i)
	}
}

func TestSafeLongUnmarshal(t *testing.T) {
	for i, currCase := range safeLongJSONs {
		wantSafeLong, err := safelong.NewSafeLong(currCase.val)
		require.NoError(t, err, "Case %d", i)

		var gotSafeLong safelong.SafeLong
		err = json.Unmarshal([]byte(currCase.json), &gotSafeLong)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, wantSafeLong, gotSafeLong, "Case %d", i)
	}
}

func TestSafeLongBoundsEnforcedByMarshal(t *testing.T) {
	wantErrFmt := "json: error calling MarshalJSON for type safelong.SafeLong: %d is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991"

	for i, currVal := range []int64{
		int64(1) << 53,
		-(int64(1) << 53),
	} {
		currSafeLong := safelong.SafeLong(currVal)
		_, err := json.Marshal(currSafeLong)
		assert.EqualError(t, err, fmt.Sprintf(wantErrFmt, currVal), "Case %d", i)
	}
}

func TestSafeLongBoundsEnforcedByUnmarshal(t *testing.T) {
	wantErrFmt := "%d is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991"

	for i, currVal := range []int64{
		int64(1) << 53,
		-(int64(1) << 53),
	} {
		var gotSafeLong *safelong.SafeLong
		err := json.Unmarshal([]byte(fmt.Sprintf("%d", currVal)), &gotSafeLong)
		assert.EqualError(t, err, fmt.Sprintf(wantErrFmt, currVal), "Case %d", i)
	}
}

func TestBoundsEnforcedByNewSafeLong(t *testing.T) {
	wantErrFmt := "%d is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991"

	for i, currVal := range []int64{
		int64(1) << 53,
		-(int64(1) << 53),
	} {
		_, err := safelong.NewSafeLong(currVal)
		assert.EqualError(t, err, fmt.Sprintf(wantErrFmt, currVal), "Case %d", i)
	}
}
