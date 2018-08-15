// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	goRuntimeMetricsToExclude = map[string]struct{}{
		"go.runtime.MemStats.BuckHashSys": {},
		"go.runtime.MemStats.DebugGC":     {},
		"go.runtime.MemStats.EnableGC":    {},
		"go.runtime.MemStats.NextGC":      {},
		"go.runtime.MemStats.LastGC":      {},
		"go.runtime.MemStats.Lookups":     {},
		"go.runtime.MemStats.TotalAlloc":  {}, // TotalAlloc increases as heap objects are allocated, but unlike Alloc and HeapAlloc, it does not decrease when objects are freed
		"go.runtime.MemStats.MCacheInuse": {},
		"go.runtime.MemStats.MCacheSys":   {},
		"go.runtime.MemStats.MSpanInuse":  {},
		"go.runtime.MemStats.MSpanSys":    {},
		"go.runtime.MemStats.Sys":         {},
	}

	_ Registry = &NoopRegistry{}
	_ Registry = &rootRegistry{}
	_ Registry = &childRegistry{}
)

// RootRegistry is the root metric registry for a product. A root registry has a prefix and a product name.
//
// Built-in Go metrics will be outputted as "<root-prefix>.<key>: <value>".
// Metrics registered on the root registry will be outputted as "<root-prefix>.<NAME>:<key>00: <value>".
// Metrics registered on subregistries of the root will be outputted as "<root-prefix>.<NAME>:<prefix>.<key>: <value>".
type RootRegistry interface {
	Registry

	// Subregistry returns a new subregistry of the root registry on which metrics can be registered.
	//
	// Specified tags will be always included in metrics emitted by a subregistry.
	Subregistry(prefix string, tags ...Tag) Registry
}

// MetricVisitor is a callback function type that can be passed into Registry.Each to report
// metrics into systems which consume metrics. An example use case is a MetricVisitor which
// writes its argument into a log file.
type MetricVisitor func(name string, tags Tags, value MetricVal)

const (
	defaultReservoirSize = 1028
	defaultAlpha         = 0.015
)

type Registry interface {
	Counter(name string, tags ...Tag) metrics.Counter
	Gauge(name string, tags ...Tag) metrics.Gauge
	GaugeFloat64(name string, tags ...Tag) metrics.GaugeFloat64
	Meter(name string, tags ...Tag) metrics.Meter
	Timer(name string, tags ...Tag) metrics.Timer
	Histogram(name string, tags ...Tag) metrics.Histogram
	HistogramWithSample(name string, sample metrics.Sample, tags ...Tag) metrics.Histogram
	// Each invokes the provided callback function on every user-defined metric registered on the router (including
	// those registered by sub-registries). Each is invoked on each metric in sorted order of the key.
	Each(MetricVisitor)
	// Unregister the metric with the given name and tags.
	Unregister(name string, tags ...Tag)
}

// NewRootMetricsRegistry creates a new root registry for metrics. This call also starts a goroutine that captures Go
// runtime information as metrics at the specified frequency.
func NewRootMetricsRegistry() RootRegistry {
	return &rootRegistry{
		registry:           metrics.NewRegistry(),
		idToMetricWithTags: make(map[metricTagsID]metricWithTags),
	}
}

var runtimeMemStats sync.Once

// CaptureRuntimeMemStats registers runtime memory metrics collectors and spawns
// a goroutine which collects them every collectionFreq.
func CaptureRuntimeMemStats(registry RootRegistry, collectionFreq time.Duration) {
	runtimeMemStats.Do(func() {
		if reg, ok := registry.(*rootRegistry); ok {
			goRegistry := metrics.NewPrefixedChildRegistry(reg.registry, "go.")
			metrics.RegisterRuntimeMemStats(goRegistry)
			go metrics.CaptureRuntimeMemStats(goRegistry, collectionFreq)
		}
	})
}

type rootRegistry struct {
	// the actual metrics.Registry on which all metrics are installed.
	registry metrics.Registry

	// map from metricTagsID to metricWithTags for all of the metrics in the userDefinedMetricsRegistry.
	idToMetricWithTags map[metricTagsID]metricWithTags

	// mutex lock to protect metric map concurrent writes
	idToMetricMutex sync.RWMutex
}

type childRegistry struct {
	prefix string
	tags   Tags
	root   *rootRegistry
}

// NoopRegistry is a "lightweight, high-speed implementation of Registry for when simplicity and performance
// matter above all else".
//
// Useful in testing infrastructure. Doesn't collect, store, or emit any metrics.
type NoopRegistry struct{}

func (r NoopRegistry) Counter(_ string, _ ...Tag) metrics.Counter {
	return metrics.NilCounter{}
}

func (r NoopRegistry) Gauge(_ string, _ ...Tag) metrics.Gauge {
	return metrics.NilGauge{}
}

func (r NoopRegistry) GaugeFloat64(_ string, _ ...Tag) metrics.GaugeFloat64 {
	return metrics.NilGaugeFloat64{}
}

func (r NoopRegistry) Meter(_ string, _ ...Tag) metrics.Meter {
	return metrics.NilMeter{}
}

func (r NoopRegistry) Timer(_ string, _ ...Tag) metrics.Timer {
	return metrics.NilTimer{}
}

func (r NoopRegistry) Histogram(_ string, _ ...Tag) metrics.Histogram {
	return metrics.NilHistogram{}
}

func (r NoopRegistry) HistogramWithSample(_ string, _ metrics.Sample, _ ...Tag) metrics.Histogram {
	return metrics.NilHistogram{}
}

func (r NoopRegistry) Each(MetricVisitor) {
	// no-op
}

func (r NoopRegistry) Unregister(name string, tags ...Tag) {
	// no-op
}

func (r *childRegistry) Counter(name string, tags ...Tag) metrics.Counter {
	return r.root.Counter(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) Gauge(name string, tags ...Tag) metrics.Gauge {
	return r.root.Gauge(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) GaugeFloat64(name string, tags ...Tag) metrics.GaugeFloat64 {
	return r.root.GaugeFloat64(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) Meter(name string, tags ...Tag) metrics.Meter {
	return r.root.Meter(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) Timer(name string, tags ...Tag) metrics.Timer {
	return r.root.Timer(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) Histogram(name string, tags ...Tag) metrics.Histogram {
	return r.root.Histogram(r.prefix+name, append(r.tags, tags...)...)
}

func (r *childRegistry) HistogramWithSample(name string, sample metrics.Sample, tags ...Tag) metrics.Histogram {
	return r.root.HistogramWithSample(r.prefix+name, sample, append(r.tags, tags...)...)
}

func (r *childRegistry) Each(f MetricVisitor) {
	r.root.Each(func(name string, tags Tags, metric MetricVal) {
		name = strings.TrimPrefix(name, r.prefix)
		f(name, tags, metric)
	})
}

func (r *childRegistry) Unregister(name string, tags ...Tag) {
	r.root.Unregister(r.prefix+name, append(r.tags, tags...)...)
}

func (r *rootRegistry) Subregistry(prefix string, tags ...Tag) Registry {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}
	return &childRegistry{
		prefix: prefix,
		tags:   Tags(tags),
		root:   r,
	}
}

func (r *rootRegistry) Each(f MetricVisitor) {
	// sort names so that iteration order is consistent
	var sortedNames []string
	r.registry.Each(func(name string, metric interface{}) {
		// filter out the runtime metrics that are defined in the exclude list
		if _, ok := goRuntimeMetricsToExclude[name]; ok {
			return
		}
		sortedNames = append(sortedNames, name)
	})
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		metric := r.registry.Get(name)

		var tags Tags
		r.idToMetricMutex.RLock()
		metricWithTags, ok := r.idToMetricWithTags[metricTagsID(name)]
		r.idToMetricMutex.RUnlock()
		if ok {
			name = metricWithTags.name
			for t := range metricWithTags.tags {
				tags = append(tags, t)
			}
			sort.Slice(tags, func(i, j int) bool {
				return tags[i].String() < tags[j].String()
			})
		} else {
			// if metric was not in idToMetricWithTags map, then it is a built-in metric. Trim the prefix since the
			// registry lookup will automatically prepend the prefix.
			metric = r.registry.Get(name)
		}
		val := ToMetricVal(metric)
		if val == nil {
			// this should never happen as all the things we put inside the registry can be turned into MetricVal
			panic("could not convert metric to MetricVal")
		}
		f(name, tags, val)
	}
}

func (r *rootRegistry) Unregister(name string, tags ...Tag) {
	metricID := toMetricTagsID(name, tags)
	r.registry.Unregister(string(metricID))
}

func (r *rootRegistry) Counter(name string, tags ...Tag) metrics.Counter {
	return metrics.GetOrRegisterCounter(r.registerMetric(name, tags), r.registry)
}

func (r *rootRegistry) Gauge(name string, tags ...Tag) metrics.Gauge {
	return metrics.GetOrRegisterGauge(r.registerMetric(name, tags), r.registry)
}

func (r *rootRegistry) GaugeFloat64(name string, tags ...Tag) metrics.GaugeFloat64 {
	return metrics.GetOrRegisterGaugeFloat64(r.registerMetric(name, tags), r.registry)
}

func (r *rootRegistry) Meter(name string, tags ...Tag) metrics.Meter {
	return metrics.GetOrRegisterMeter(r.registerMetric(name, tags), r.registry)
}

func (r *rootRegistry) Timer(name string, tags ...Tag) metrics.Timer {
	return metrics.GetOrRegisterTimer(r.registerMetric(name, tags), r.registry)
}

func (r *rootRegistry) Histogram(name string, tags ...Tag) metrics.Histogram {
	return r.HistogramWithSample(name, DefaultSample(), tags...)
}

func (r *rootRegistry) HistogramWithSample(name string, sample metrics.Sample, tags ...Tag) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(r.registerMetric(name, tags), r.registry, sample)
}

func DefaultSample() metrics.Sample {
	return metrics.NewExpDecaySample(defaultReservoirSize, defaultAlpha)
}

func (r *rootRegistry) registerMetric(name string, tags Tags) string {
	metricID := toMetricTagsID(name, tags)
	r.idToMetricMutex.Lock()
	r.idToMetricWithTags[metricID] = metricWithTags{
		name: name,
		tags: tags.ToSet(),
	}
	r.idToMetricMutex.Unlock()
	return string(metricID)
}

// metricWithTags stores a specific metric with its set of tags.
type metricWithTags struct {
	name string
	tags map[Tag]struct{}
}

// metricTagsID is the unique identifier for a given metric. Each {metricName, set<Tag>} pair is considered to be a
// unique metric. A metricTagsID is a string of the following form: "<name>|tags:|<tag1>|<tag2>|". The tags appear in
// ascending alphanumeric order. If a metric does not have any tags, its metricsTagsID is of the form: "<name>|tags:||".
type metricTagsID string

// toID generates the metricTagsID identifier for the metricWithTags. A unique {metricName, set<Tag>} input will
// generate a unique output.
func (m *metricWithTags) toID() metricTagsID {
	var sortedTags []string
	for t := range m.tags {
		sortedTags = append(sortedTags, t.String())
	}
	sort.Strings(sortedTags)

	return metricTagsID(fmt.Sprintf("%s|tags:|%s|", m.name, strings.Join(sortedTags, "|")))
}

func toMetricTagsID(name string, tags Tags) metricTagsID {
	return (&metricWithTags{
		name: name,
		tags: tags.ToSet(),
	}).toID()
}
