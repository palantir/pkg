// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMicroSecondsTimerUpdate(t *testing.T) {
	microSecTimer := newMicroSecondsTimer()
	microSecTimer.Update(100 * time.Microsecond)

	assert.Equal(t, int64(1), microSecTimer.Count())
	assert.Equal(t, int64(100), microSecTimer.Max())
	assert.InDelta(t, int64(100), microSecTimer.Mean(), 1)
	assert.Equal(t, int64(100), microSecTimer.Min())
	assert.Equal(t, float64(100), microSecTimer.Percentile(5))
	assert.Equal(t, float64(0), microSecTimer.Rate1())
	assert.Equal(t, float64(0), microSecTimer.Rate5())
	assert.Equal(t, float64(0), microSecTimer.Rate15())
	assert.Equal(t, float64(0), microSecTimer.StdDev())
	assert.Equal(t, int64(100), microSecTimer.Sum())
}

func TestMicroSecondsTimerUpdateSince(t *testing.T) {
	microSecTimer := newMicroSecondsTimer()
	microSecTimer.UpdateSince(time.Now().Add(-100 * time.Microsecond))

	assert.Equal(t, int64(1), microSecTimer.Count())
	assert.Equal(t, int64(100), microSecTimer.Max())
	assert.InDelta(t, int64(100), microSecTimer.Mean(), 1)
	assert.Equal(t, int64(100), microSecTimer.Min())
	assert.Equal(t, float64(100), microSecTimer.Percentile(5))
	assert.Equal(t, float64(0), microSecTimer.Rate1())
	assert.Equal(t, float64(0), microSecTimer.Rate5())
	assert.Equal(t, float64(0), microSecTimer.Rate15())
	assert.Equal(t, float64(0), microSecTimer.StdDev())
	assert.Equal(t, int64(100), microSecTimer.Sum())
}

func TestMicroSecondsTimerTime(t *testing.T) {
	microSecTimer := newMicroSecondsTimer()
	microSecTimer.Time(func() {
		time.Sleep(10 * time.Microsecond)
	})

	assert.Equal(t, int64(1), microSecTimer.Count())
	assert.InDelta(t, int64(10), microSecTimer.Max(), 1000)
	assert.InDelta(t, int64(10), microSecTimer.Mean(), 1000)
	assert.InDelta(t, int64(10), microSecTimer.Min(), 1000)
	assert.InDelta(t, float64(10), microSecTimer.Percentile(5), 1000)
	assert.Equal(t, float64(0), microSecTimer.Rate1())
	assert.Equal(t, float64(0), microSecTimer.Rate5())
	assert.Equal(t, float64(0), microSecTimer.Rate15())
	assert.Equal(t, float64(0), microSecTimer.StdDev())
	assert.InDelta(t, int64(10), microSecTimer.Sum(), 1000)
}
