// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkRegisterMetric(b *testing.B) {
	b.Run("1 tag", func(b *testing.B) {
		doBench(b, 1)
	})
	b.Run("10 tag", func(b *testing.B) {
		doBench(b, 10)
	})
	b.Run("100 tag", func(b *testing.B) {
		doBench(b, 100)
	})
}

func doBench(b *testing.B, n int) {
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
