// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datetime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/palantir/pkg/datetime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dateTimeJSONs = []struct {
	sec        int64
	zoneOffset int
	str        string
	json       string
}{
	{
		sec:  1483326245,
		str:  `2017-01-02T03:04:05Z`,
		json: `"2017-01-02T03:04:05Z"`,
	},
	{
		sec:  1483326245,
		str:  `2017-01-02T03:04:05Z`,
		json: `"2017-01-02T03:04:05.000Z"`,
	},
	{
		sec:  1483326245,
		str:  `2017-01-02T03:04:05Z`,
		json: `"2017-01-02T03:04:05.000000000Z"`,
	},
	{
		sec:        1483326245,
		zoneOffset: 3600,
		str:        `2017-01-02T04:04:05+01:00`,
		json:       `"2017-01-02T04:04:05.000000000+01:00"`,
	},
	{
		sec:        1483326245,
		zoneOffset: 7200,
		str:        `2017-01-02T05:04:05+02:00`,
		json:       `"2017-01-02T05:04:05.000000000+02:00"`,
	},
	{
		sec:        1483326245,
		zoneOffset: 3600,
		str:        `2017-01-02T04:04:05+01:00`,
		json:       `"2017-01-02T04:04:05.000000000+01:00[Europe/Berlin]"`,
	},
}

func TestDateTimeString(t *testing.T) {
	for i, currCase := range dateTimeJSONs {
		currDateTime := datetime.DateTime(time.Unix(currCase.sec, 0).In(time.FixedZone("", currCase.zoneOffset)))
		assert.Equal(t, currCase.str, currDateTime.String(), "Case %d", i)
	}
}

func TestDateTimeMarshal(t *testing.T) {
	for i, currCase := range dateTimeJSONs {
		currDateTime := datetime.DateTime(time.Unix(currCase.sec, 0).In(time.FixedZone("", currCase.zoneOffset)))
		bytes, err := json.Marshal(currDateTime)
		require.NoError(t, err, "Case %d: marshal %q", i, currDateTime.String())

		var unmarshaledFromMarshal datetime.DateTime
		err = json.Unmarshal(bytes, &unmarshaledFromMarshal)
		require.NoError(t, err, "Case %d: unmarshal %q", i, string(bytes))

		var unmarshaledFromCase datetime.DateTime
		err = json.Unmarshal([]byte(currCase.json), &unmarshaledFromCase)
		require.NoError(t, err, "Case %d: unmarshal %q", i, currCase.json)

		assert.Equal(t, unmarshaledFromCase, unmarshaledFromMarshal, "Case %d", i)
	}
}

func TestDateTimeUnmarshal(t *testing.T) {
	for i, currCase := range dateTimeJSONs {
		wantDateTime := time.Unix(currCase.sec, 0).UTC()
		if currCase.zoneOffset != 0 {
			wantDateTime = wantDateTime.In(time.FixedZone("", currCase.zoneOffset))
		}

		var gotDateTime datetime.DateTime
		err := json.Unmarshal([]byte(currCase.json), &gotDateTime)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, wantDateTime, time.Time(gotDateTime), "Case %d", i)
	}
}
