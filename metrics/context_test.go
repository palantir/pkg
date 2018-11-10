// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

func TestFromContextWithTags(t *testing.T) {
	ctx := context.Background()
	reg := &rootRegistry{
		registry: metrics.NewPrefixedRegistry("foo"),
	}

	ctx = WithRegistry(ctx, reg)
	ctx = AddTags(ctx, MustNewTag("bar", "baz"))

	assert.Equal(t, FromContext(ctx), &childRegistry{
		root: reg,
		tags: Tags{MustNewTag("bar", "baz")},
	})
}

func TestDefaultFromContext(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, FromContext(ctx), DefaultMetricsRegistry)
}
