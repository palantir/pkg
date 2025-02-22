// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestContainers_goccy(t *testing.T) {
	runContainerTests(
		t,
		"goccy",
		NewGoccyYAMLLibrary(
			GoccyUseNonFlowWhenModifyingEmptyContainer(false),
			GoccyYAMLEncodeOption(yaml.Indent(containerTestIndentSpaces)),
			GoccyYAMLEncodeOption(yaml.IndentSequence(false)),
		),
	)
}
