// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"iter"

	"github.com/palantir/go-metrics"
)

type MetricVal interface {
	Type() string

	// Keys implements iter.Seq[string] by returning the keys that can be used to retrieve values from the metric.
	// Example:
	//
	//	for key := range val.Keys() {
	//		if !skip(key) {
	//			fmt.Println(key, mv.Value(key))
	// 		}
	//	}
	Keys() iter.Seq[string]

	// KeySlice returns the keys that can be used to retrieve values from the metric.
	// The returned slice must not be modified.
	KeySlice() []string

	// Value returns the computed value for the given key. If the key is not recognized, returns nil.
	Value(key string) interface{}

	// Deprecated: use Keys and Value to iterate through values and avoid eager computation of skipped keys.
	Values() map[string]interface{}
}

func ToMetricVal(in interface{}) MetricVal {
	switch val := in.(type) {
	case metrics.Counter:
		return &counterVal{Counter: val}
	case metrics.Gauge:
		return &gaugeVal{Gauge: val}
	case metrics.GaugeFloat64:
		return &gaugeFloat64Val{GaugeFloat64: val}
	case metrics.Histogram:
		return &histogramVal{Histogram: val}
	case metrics.Meter:
		return &meterVal{Meter: val}
	case metrics.Timer:
		return &timerVal{Timer: val}
	}
	return nil
}

type counterVal struct {
	metrics.Counter
}

func (v *counterVal) Type() string {
	return "counter"
}

var counterKeys = []string{"count"}

func (v *counterVal) KeySlice() []string {
	return counterKeys
}

func (v *counterVal) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"count"} {
			if !yield(key) {
				return
			}
		}
	}
}

func (v *counterVal) Value(key string) interface{} {
	switch key {
	case "count":
		return v.Counter.Count()
	default:
		return nil
	}
}

func (v *counterVal) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

type gaugeVal struct {
	metrics.Gauge
}

func (v *gaugeVal) Type() string {
	return "gauge"
}

func (v *gaugeVal) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"value"} {
			if !yield(key) {
				return
			}
		}
	}
}

var gaugeKeys = []string{"value"}

func (v *gaugeVal) KeySlice() []string {
	return gaugeKeys
}

func (v *gaugeVal) Value(key string) interface{} {
	switch key {
	case "value":
		return v.Gauge.Value()
	default:
		return nil
	}
}

func (v *gaugeVal) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

type gaugeFloat64Val struct {
	metrics.GaugeFloat64
}

func (v *gaugeFloat64Val) Type() string {
	return "gauge"
}

func (v *gaugeFloat64Val) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"value"} {
			if !yield(key) {
				return
			}
		}
	}
}

func (v *gaugeFloat64Val) KeySlice() []string {
	return gaugeKeys
}

func (v *gaugeFloat64Val) Value(key string) interface{} {
	switch key {
	case "value":
		return v.GaugeFloat64.Value()
	default:
		return nil
	}
}

func (v *gaugeFloat64Val) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

type histogramVal struct {
	metrics.Histogram
}

func (v *histogramVal) Type() string {
	return "histogram"
}

func (v *histogramVal) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"count", "min", "max", "mean", "stddev", "p50", "p95", "p99"} {
			if !yield(key) {
				return
			}
		}
	}
}

var histogramKeys = []string{"count", "min", "max", "mean", "stddev", "p50", "p95", "p99"}

func (v *histogramVal) KeySlice() []string {
	return histogramKeys
}

func (v *histogramVal) Value(key string) interface{} {
	switch key {
	case "count":
		return v.Histogram.Count()
	case "min":
		return v.Histogram.Min()
	case "max":
		return v.Histogram.Max()
	case "mean":
		return v.Histogram.Mean()
	case "stddev":
		return v.Histogram.StdDev()
	case "p50":
		return v.Histogram.Percentile(0.5)
	case "p95":
		return v.Histogram.Percentile(0.95)
	case "p99":
		return v.Histogram.Percentile(0.99)
	default:
		return nil
	}
}

func (v *histogramVal) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

type meterVal struct {
	metrics.Meter
}

func (v *meterVal) Type() string {
	return "meter"
}

func (v *meterVal) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"count", "1m", "5m", "15m", "mean"} {
			if !yield(key) {
				return
			}
		}
	}
}

var meterKeys = []string{"count", "1m", "5m", "15m", "mean"}

func (v *meterVal) KeySlice() []string {
	return meterKeys
}

func (v *meterVal) Value(key string) interface{} {
	switch key {
	case "count":
		return v.Meter.Count()
	case "1m":
		return v.Meter.Rate1()
	case "5m":
		return v.Meter.Rate5()
	case "15m":
		return v.Meter.Rate15()
	case "mean":
		return v.Meter.RateMean()
	default:
		return nil
	}
}

func (v *meterVal) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

type timerVal struct {
	metrics.Timer
}

func (v *timerVal) Type() string {
	return "timer"
}

func (v *timerVal) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, key := range []string{"count", "1m", "5m", "15m", "meanRate", "min", "max", "mean", "stddev", "p50", "p95", "p99"} {
			if !yield(key) {
				return
			}
		}
	}
}

var timerKeys = []string{"count", "1m", "5m", "15m", "meanRate", "min", "max", "mean", "stddev", "p50", "p95", "p99"}

func (v *timerVal) KeySlice() []string {
	return timerKeys
}

func (v *timerVal) Value(key string) interface{} {
	switch key {
	case "count":
		return v.Timer.Count()
	case "1m":
		return v.Timer.Rate1()
	case "5m":
		return v.Timer.Rate5()
	case "15m":
		return v.Timer.Rate15()
	case "meanRate":
		return v.Timer.RateMean()
	case "min":
		return v.Timer.Min()
	case "max":
		return v.Timer.Max()
	case "mean":
		return v.Timer.Mean()
	case "stddev":
		return v.Timer.StdDev()
	case "p50":
		return v.Timer.Percentile(0.5)
	case "p95":
		return v.Timer.Percentile(0.95)
	case "p99":
		return v.Timer.Percentile(0.99)
	default:
		return nil
	}
}

func (v *timerVal) Values() map[string]interface{} {
	return collectValuesByKey(v)
}

func collectValuesByKey(mv MetricVal) map[string]interface{} {
	keys := mv.KeySlice()
	values := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		values[key] = mv.Value(key)
	}
	return values
}
