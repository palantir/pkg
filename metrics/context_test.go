package metrics

import (
	"context"
	"testing"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	reg := &rootRegistry{
		registry: metrics.NewPrefixedRegistry("foo"),
	}

	ctx = WithRegistry(ctx, reg)

	assert.Equal(t, FromContext(ctx), reg)
}

func TestDefaultFromContext(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, FromContext(ctx), DefaultMetricsRegistry)
}
