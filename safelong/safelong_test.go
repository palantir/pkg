// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safelong_test

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
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
	wantErrFmt := "json: error calling MarshalJSONTo for type *safelong.SafeLong: %d is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between -9007199254740991 and 9007199254740991"

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

func TestSafeLongMapKey(t *testing.T) {
	in := map[safelong.SafeLong]safelong.SafeLong{-42: 13, 0: 0}
	const want = `{"-42":13,"0":0}`
	for _, tc := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{name: "v1", marshal: json.Marshal, unmarshal: json.Unmarshal},
		{name: "v2",
			marshal:   func(v any) ([]byte, error) { return jsonv2.Marshal(v, jsonv2.Deterministic(true)) },
			unmarshal: func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.marshal(in)
			require.NoError(t, err)
			assert.Equal(t, want, string(out))

			var got map[safelong.SafeLong]safelong.SafeLong
			require.NoError(t, tc.unmarshal(out, &got))
			assert.Equal(t, in, got)
		})
	}
}

func TestSafeLongV2(t *testing.T) {
	for i, currCase := range safeLongJSONs {
		out, err := jsonv2.Marshal(safelong.SafeLong(currCase.val))
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, currCase.json, string(out), "Case %d", i)

		var got safelong.SafeLong
		require.NoError(t, jsonv2.Unmarshal(out, &got), "Case %d", i)
		assert.Equal(t, safelong.SafeLong(currCase.val), got, "Case %d", i)
	}
	for _, in := range []string{
		`9007199254740992`,
		`"13"`,
		`1.5`,
		`{"9007199254740992":0}`,
		`{"1.5":0}`,
	} {
		t.Run("rejects "+in, func(t *testing.T) {
			var v any = new(safelong.SafeLong)
			if in[0] == '{' {
				v = new(map[safelong.SafeLong]int)
			}
			assert.Error(t, jsonv2.Unmarshal([]byte(in), v))
		})
	}
	t.Run("rejects out of bounds map key", func(t *testing.T) {
		_, err := jsonv2.Marshal(map[safelong.SafeLong]int{1 << 53: 0})
		assert.Error(t, err)
	})
}
