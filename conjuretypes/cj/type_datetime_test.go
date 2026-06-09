// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"
	"time"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/datetime"
	"github.com/stretchr/testify/assert"
)

func TestDateTime(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "zero_time",
			Test: typeTestCase[time.Time]{Codec: cj.DateTime[time.Time](), Value: time.Time{}, JSON: "\"0001-01-01T00:00:00Z\""},
		},
		{
			Name: "iso8601_time",
			Test: typeTestCase[time.Time]{Codec: cj.DateTime[time.Time](), Value: time.Date(2025, 5, 12, 19, 26, 0, 0, time.UTC), JSON: "\"2025-05-12T19:26:00Z\""},
		},
		{
			Name: "zero_datetime",
			Test: typeTestCase[datetime.DateTime]{Codec: cj.DateTime[datetime.DateTime](), Value: datetime.DateTime{}, JSON: "\"0001-01-01T00:00:00Z\""},
		},
		{
			Name: "iso8601_datetime",
			Test: typeTestCase[datetime.DateTime]{Codec: cj.DateTime[datetime.DateTime](), Value: must(datetime.ParseDateTime("2025-05-12T19:26:00Z")), JSON: "\"2025-05-12T19:26:00Z\""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

// TestDateTimeEqual covers equality by instant AND wire string: equal instants in
// different zones serialize to distinct object names, so they are distinct set
// elements and map keys.
func TestDateTimeEqual(t *testing.T) {
	codec := cj.DateTime[datetime.DateTime]()
	utc := datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	utcAgain := datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	plusOne := datetime.DateTime(time.Date(2024, 1, 1, 1, 0, 0, 0, time.FixedZone("", 3600))) // same instant as utc
	later := datetime.DateTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	assert.True(t, codec.Equal(utc, utcAgain), "same instant and zone are equal")
	assert.False(t, codec.Equal(utc, plusOne), "same instant, different zone is distinct (wire strings differ)")
	assert.False(t, codec.Equal(utc, later), "different instant is not equal")
}

// TestDateTimeSort covers map-key ordering: keys sort by instant, ties at the same
// instant in different zones broken deterministically by wire string.
func TestDateTimeSort(t *testing.T) {
	utc := datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	plusOne := datetime.DateTime(time.Date(2024, 1, 1, 1, 0, 0, 0, time.FixedZone("", 3600))) // same instant as utc
	later := datetime.DateTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	keys := []datetime.DateTime{later, plusOne, utc}
	cj.DateTime[datetime.DateTime]().Sort(keys)

	got := make([]string, len(keys))
	for i, k := range keys {
		got[i] = k.String()
	}
	// utc and plusOne share an instant; tie broken by wire string ("...Z" < "...+01:00").
	assert.Equal(t, []string{"2024-01-01T00:00:00Z", "2024-01-01T01:00:00+01:00", "2025-01-01T00:00:00Z"}, got)
}
