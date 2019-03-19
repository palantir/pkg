// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
	"testing"

	"github.com/palantir/go-metrics"
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

func TestFromContextAddTags(t *testing.T) {
	ctx := context.Background()
	reg := &rootRegistry{
		registry: metrics.NewPrefixedRegistry("foo"),
	}

	ctx = WithRegistry(ctx, reg)

	assert.Equal(t, reg, FromContext(ctx))
	assert.Equal(t, Tags(nil), TagsFromContext(ctx))

	ctx = AddTags(ctx, MustNewTag("bar", "baz"))

	assert.Equal(t, &childRegistry{
		root: reg,
		tags: Tags{MustNewTag("bar", "baz")},
	}, FromContext(ctx))
	assert.Equal(t, Tags{MustNewTag("bar", "baz")}, TagsFromContext(ctx))

	ctx1 := AddTags(ctx, MustNewTag("qux", "quux"))

	// Verify original context was not modified
	assert.Equal(t, Tags{MustNewTag("bar", "baz")}, TagsFromContext(ctx))

	assert.Equal(t, &childRegistry{
		root: reg,
		tags: Tags{MustNewTag("bar", "baz"), MustNewTag("qux", "quux")},
	}, FromContext(ctx1))

	assert.Equal(t, Tags{MustNewTag("bar", "baz"), MustNewTag("qux", "quux")}, TagsFromContext(ctx1))

}

func TestDefaultFromContext(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, FromContext(ctx), DefaultMetricsRegistry)
}
