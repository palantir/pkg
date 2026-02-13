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

func BenchmarkFromContext(b *testing.B) {
	b.ReportAllocs()
	registry := NewRootMetricsRegistry()
	ctx := context.Background()
	ctx = WithRegistry(ctx, registry)
	ctx = AddTags(ctx, MustNewTag("foo", "bar"))
	ctx = AddTags(ctx, MustNewTag("bar", "baz"))
	ctx = AddTags(ctx, MustNewTag("qux", "quux"))
	ctx = context.WithValue(ctx, "other1", "other1")
	ctx = context.WithValue(ctx, "other2", "other2")

	for b.Loop() {
		r := FromContext(ctx)
		_ = r
	}

	tags := FromContext(ctx).(*childRegistry).tags
	assert.Equal(b, Tags{MustNewTag("foo", "bar"), MustNewTag("bar", "baz"), MustNewTag("qux", "quux")}, tags)
}
