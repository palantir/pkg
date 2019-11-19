// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics_test

import (
	"testing"
	"time"

	metricspkg "github.com/palantir/go-metrics"
	"github.com/palantir/pkg/metrics"
	"github.com/palantir/pkg/objmatcher"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	val := metricspkg.NewCounter()
	val.Inc(13)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "counter", mv.Type())
	assert.Equal(t, map[string]interface{}{"count": int64(13)}, mv.Values())
}

func TestGauge(t *testing.T) {
	val := metricspkg.NewGauge()
	val.Update(13)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "gauge", mv.Type())
	assert.Equal(t, map[string]interface{}{"value": int64(13)}, mv.Values())
}

func TestGaugeFloat64(t *testing.T) {
	val := metricspkg.NewGaugeFloat64()
	val.Update(13.13)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "gauge", mv.Type())
	assert.Equal(t, map[string]interface{}{"value": float64(13.13)}, mv.Values())
}

func TestHistogram(t *testing.T) {
	val := metricspkg.NewHistogram(metricspkg.NewExpDecaySample(1028, 0.015))
	val.Update(1)
	val.Update(13)
	val.Update(100)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "histogram", mv.Type())
	assert.Equal(t, map[string]interface{}{
		"count":  int64(3),
		"min":    int64(1),
		"max":    int64(100),
		"mean":   float64(38),
		"stddev": float64(44.11349000022555),
		"p50":    float64(13),
		"p95":    float64(100),
		"p99":    float64(100),
	}, mv.Values())
}

func TestMeter(t *testing.T) {
	val := metricspkg.NewMeter()
	val.Mark(13)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "meter", mv.Type())
	assert.NoError(t, objmatcher.MapMatcher(map[string]objmatcher.Matcher{
		"count": objmatcher.NewEqualsMatcher(int64(13)),
		"1m":    objmatcher.NewEqualsMatcher(float64(0)),
		"5m":    objmatcher.NewEqualsMatcher(float64(0)),
		"15m":   objmatcher.NewEqualsMatcher(float64(0)),
		"mean":  objmatcher.NewAnyMatcher(),
	}).Matches(mv.Values()))
}

func TestTimer(t *testing.T) {
	val := metricspkg.NewTimer()
	val.Update(time.Second)
	val.Update(2 * time.Minute)

	mv := metrics.ToMetricVal(val)

	assert.Equal(t, "timer", mv.Type())
	assert.NoError(t, objmatcher.MapMatcher(map[string]objmatcher.Matcher{
		"count":    objmatcher.NewEqualsMatcher(int64(2)),
		"min":      objmatcher.NewEqualsMatcher(int64(1000000000)),
		"max":      objmatcher.NewEqualsMatcher(int64(120000000000)),
		"mean":     objmatcher.NewEqualsMatcher(float64(6.05e+10)),
		"stddev":   objmatcher.NewEqualsMatcher(float64(5.95e+10)),
		"1m":       objmatcher.NewEqualsMatcher(float64(0)),
		"5m":       objmatcher.NewEqualsMatcher(float64(0)),
		"15m":      objmatcher.NewEqualsMatcher(float64(0)),
		"meanRate": objmatcher.NewAnyMatcher(),
		"p50":      objmatcher.NewEqualsMatcher(float64(6.05e+10)),
		"p95":      objmatcher.NewEqualsMatcher(float64(1.2e+11)),
		"p99":      objmatcher.NewEqualsMatcher(float64(1.2e+11)),
	}).Matches(mv.Values()))
}
