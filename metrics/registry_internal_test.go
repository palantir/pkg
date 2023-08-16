// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootRegistry_UnregisterReleasesResources(t *testing.T) {
	// register root metrics
	root := NewRootMetricsRegistry().(*rootRegistry)

	id := toMetricTagsID("my-counter", Tags{})

	// register metric
	root.Counter("my-counter").Inc(1)
	assert.Contains(t, root.idToMetricWithTags, id)

	// unregister metric
	root.Unregister("my-counter")
	assert.NotContains(t, root.idToMetricWithTags, id)
}

func TestChildRegistryDoesNotReuseSlice(t *testing.T) {
	root := NewRootMetricsRegistry()
	ctx := WithRegistry(context.Background(), root)
	ctx = AddTags(ctx, MustNewTag("foo", "bar"))

	myTags := make(Tags, 0, 4)
	myTags = append(myTags, MustNewTag("baz", "qux"))

	child := FromContext(ctx).(*childRegistry)
	child.Counter("my-counter", myTags...).Inc(1)
	root.Each(func(name string, tags Tags, value MetricVal) {
		assert.Equal(t, "my-counter", name)
		assert.Equal(t, Tags{MustNewTag("baz", "qux"), MustNewTag("foo", "bar")}, tags)
		assert.EqualValues(t, 1, value.(*counterVal).Counter.Count())
	})
	assert.Len(t, myTags, 1)
}
