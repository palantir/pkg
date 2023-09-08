// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func BenchmarkNewTag(b *testing.B) {
	for _, tc := range []struct {
		tagLen int
	}{
		{
			tagLen: 2,
		},
		{
			tagLen: 10,
		},
		{
			tagLen: 100,
		},
		{
			tagLen: 199,
		},
	} {
		tagKeyValue := strings.Repeat("a", tc.tagLen/2)
		b.Run(fmt.Sprintf("tagLen:%d", tc.tagLen), newTagBenchFunc(tagKeyValue, tagKeyValue))
	}
}

func BenchmarkNormalizeTag(b *testing.B) {
	for _, val := range []string{
		"a❌Long❌Tag❌With❌Emoji❌Chars",
		"a Long Tag With Space Chars",
		"aLongTagValueWithUpperChars",
		"alongtagvaluewithnospecials",
		"UPPER",
		"lower",
	} {
		b.Run(val, newTagBenchFunc("key", val))
	}
}

func newTagBenchFunc(tagKey, tagValue string) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			MustNewTag(tagKey, tagValue)
		}
	}
}

func BenchmarkRegisterMetric(b *testing.B) {
	b.Run("1 tag", func(b *testing.B) {
		doRegisterBench(b, 1)
	})
	b.Run("10 tag", func(b *testing.B) {
		doRegisterBench(b, 10)
	})
	b.Run("100 tag", func(b *testing.B) {
		doRegisterBench(b, 100)
	})
}

func doRegisterBench(b *testing.B, n int) {
	var tags Tags
	for i := 0; i < n; i++ {
		tags = append(tags, MustNewTag(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i)))
	}
	ctx := AddTags(WithRegistry(context.Background(), NewRootMetricsRegistry()), tags...)
	b.ResetTimer()
	b.ReportAllocs()
	reg := FromContext(ctx)
	for i := 0; i < b.N; i++ {
		reg.Counter("metricName").Inc(1)
	}
}

func BenchmarkHistogram(b *testing.B) {
	reg := NewRootMetricsRegistry()
	b.Run("HistogramWithSample with cached Tag", func(b *testing.B) {
		s := DefaultSample()
		t := MustNewTag("key", "value")
		for i := 0; i < b.N; i++ {
			reg.HistogramWithSample(b.Name(), s, t).Update(int64(i))
		}
		b.ReportAllocs()
	})
	b.Run("Histogram with cached Tag", func(b *testing.B) {
		t := MustNewTag("key", "value")
		for i := 0; i < b.N; i++ {
			reg.Histogram(b.Name(), t).Update(int64(i))
		}
		b.ReportAllocs()
	})
	b.Run("HistogramWithSample with NewTag", func(b *testing.B) {
		s := DefaultSample()
		for i := 0; i < b.N; i++ {
			reg.HistogramWithSample(b.Name(), s, MustNewTag("key", "value")).Update(int64(i))
		}
		b.ReportAllocs()
	})
	b.Run("Histogram with NewTag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reg.Histogram(b.Name(), MustNewTag("key", "value")).Update(int64(i))
		}
		b.ReportAllocs()
	})
	b.Run("cached Histogram", func(b *testing.B) {
		h := reg.Histogram(b.Name(), MustNewTag("key", "value"))
		for i := 0; i < b.N; i++ {
			h.Update(int64(i))
		}
		b.ReportAllocs()
	})
}
