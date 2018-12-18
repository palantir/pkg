package metrics

import (
	"context"
	"testing"
)

func BenchmarkRegisterMetric(b *testing.B)  {
	ctx := AddTags(WithRegistry(context.Background(), NewRootMetricsRegistry()),
		MustNewTag("key1", "val1"),
		MustNewTag("key2", "val2"),
		MustNewTag("key3", "val3"),
		MustNewTag("key4", "val4"),
		MustNewTag("key5", "val5"),
		MustNewTag("key6", "val6"),
		MustNewTag("key7", "val7"),
		MustNewTag("key8", "val8"),
		)
	b.ResetTimer()
	b.ReportAllocs()
	reg := FromContext(ctx)
	for i := 0; i < b.N; i++ {
		reg.Counter("metricName").Inc(1)
	}
}
