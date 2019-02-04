// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
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
