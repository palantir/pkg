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

func TestDateTimeCompare(t *testing.T) {
	encoder := cj.DateTimeMapKey[time.Time]()

	time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)
	time3 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		a, b     time.Time
		expected int
	}{
		{"equal times", time1, time3, 0},
		{"a before b", time1, time2, -1},
		{"a after b", time2, time1, 1},
		{"zero times", time.Time{}, time.Time{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Compare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDateTimeCompareWithDateTime(t *testing.T) {
	encoder := cj.DateTimeMapKey[datetime.DateTime]()

	dt1 := datetime.DateTime(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	dt2 := datetime.DateTime(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC))
	dt3 := datetime.DateTime(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		a, b     datetime.DateTime
		expected int
	}{
		{"equal datetimes", dt1, dt3, 0},
		{"a before b", dt1, dt2, -1},
		{"a after b", dt2, dt1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Compare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDateTimeMapKeyEquality contrasts the value codec (equality by instant) with
// the map-key codec (equality and ordering by wire string), using two datetimes
// that denote the same instant in different time zones.
func TestDateTimeMapKeyEquality(t *testing.T) {
	utc := datetime.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))                     // "2024-01-01T00:00:00Z"
	plusOne := datetime.DateTime(time.Date(2024, 1, 1, 1, 0, 0, 0, time.FixedZone("", 3600))) // "2024-01-01T01:00:00+01:00"

	// The value codec compares by instant, so the two are equal.
	assert.True(t, cj.DateTime[datetime.DateTime]().Equal(utc, plusOne), "value codec compares by instant")

	// The map-key codec compares by the emitted object name, so they are distinct
	// and "...Z" sorts before "...+01:00".
	mapKey := cj.DateTimeMapKey[datetime.DateTime]()
	assert.False(t, mapKey.Equal(utc, plusOne), "map-key codec compares by wire string")
	assert.Equal(t, -1, mapKey.Compare(utc, plusOne))
	assert.Equal(t, 1, mapKey.Compare(plusOne, utc))

	// Identical wire strings are equal and compare as 0.
	assert.True(t, mapKey.Equal(utc, utc))
	assert.Equal(t, 0, mapKey.Compare(utc, utc))
}
