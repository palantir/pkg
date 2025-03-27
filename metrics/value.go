// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"github.com/palantir/go-metrics"
)

type MetricVal interface {
	Type() string

	// Keys implements iter.Seq[string] by returning the keys that can be used to retrieve values from the metric.
	// Example:
	//
	//	for key := range val.Keys {
	//		if !skip(key) {
	//			fmt.Println(key, mv.Value(key))
	// 		}
	//	}
	Keys(yield func(string) bool)

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

func (v *counterVal) Keys(yield func(string) bool) {
	for _, key := range []string{"count"} {
		if !yield(key) {
			return
		}
	}
}

func (v *counterVal) Value(key string) interface{} {
	switch key {
	case "count":
		return v.Counter.Count()
	}
	return nil
}

func (v *counterVal) Values() map[string]interface{} {
	return map[string]interface{}{
		"count": v.Count(),
	}
}

type gaugeVal struct {
	metrics.Gauge
}

func (v *gaugeVal) Type() string {
	return "gauge"
}

func (v *gaugeVal) Keys(yield func(string) bool) {
	for _, key := range []string{"value"} {
		if !yield(key) {
			return
		}
	}
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
	return map[string]interface{}{
		"value": v.Gauge.Value(),
	}
}

type gaugeFloat64Val struct {
	metrics.GaugeFloat64
}

func (v *gaugeFloat64Val) Type() string {
	return "gauge"
}

func (v *gaugeFloat64Val) Keys(yield func(string) bool) {
	for _, key := range []string{"value"} {
		if !yield(key) {
			return
		}
	}
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
	return map[string]interface{}{
		"value": v.GaugeFloat64.Value(),
	}
}

type histogramVal struct {
	metrics.Histogram
}

func (v *histogramVal) Type() string {
	return "histogram"
}

func (v *histogramVal) Keys(yield func(string) bool) {
	for _, key := range []string{"count", "min", "max", "mean", "stddev", "p50", "p95", "p99"} {
		if !yield(key) {
			return
		}
	}
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
	return map[string]interface{}{
		"min":    v.Histogram.Min(),
		"max":    v.Histogram.Max(),
		"mean":   v.Histogram.Mean(),
		"stddev": v.Histogram.StdDev(),
		"p50":    v.Histogram.Percentile(0.5),
		"p95":    v.Histogram.Percentile(0.95),
		"p99":    v.Histogram.Percentile(0.99),
		"count":  v.Histogram.Count(),
	}
}

type meterVal struct {
	metrics.Meter
}

func (v *meterVal) Type() string {
	return "meter"
}

func (v *meterVal) Keys(yield func(string) bool) {
	for _, key := range []string{"count", "1m", "5m", "15m", "mean"} {
		if !yield(key) {
			return
		}
	}
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
	return map[string]interface{}{
		"count": v.Meter.Count(),
		"1m":    v.Meter.Rate1(),
		"5m":    v.Meter.Rate5(),
		"15m":   v.Meter.Rate15(),
		"mean":  v.Meter.RateMean(),
	}
}

type timerVal struct {
	metrics.Timer
}

func (v *timerVal) Type() string {
	return "timer"
}

func (v *timerVal) Keys(yield func(string) bool) {
	for _, key := range []string{"count", "1m", "5m", "15m", "meanRate", "min", "max", "mean", "stddev", "p50", "p95", "p99"} {
		if !yield(key) {
			return
		}
	}
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
	return map[string]interface{}{
		"count":    v.Timer.Count(),
		"1m":       v.Timer.Rate1(),
		"5m":       v.Timer.Rate5(),
		"15m":      v.Timer.Rate15(),
		"meanRate": v.Timer.RateMean(),
		"min":      v.Timer.Min(),
		"max":      v.Timer.Max(),
		"mean":     v.Timer.Mean(),
		"stddev":   v.Timer.StdDev(),
		"p50":      v.Timer.Percentile(0.5),
		"p95":      v.Timer.Percentile(0.95),
		"p99":      v.Timer.Percentile(0.99),
	}
}
