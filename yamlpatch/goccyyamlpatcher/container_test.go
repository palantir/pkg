// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/palantir/pkg/yamlpatch/internal/testhelpers"
)

func TestContainers(t *testing.T) {
	testhelpers.RunContainerTests(
		t,
		"goccy",
		newGoccyYAMLLibrary(
			UseNonFlowWhenModifyingEmptyContainer(false),
			YAMLEncodeOption(yaml.Indent(testhelpers.ContainerTestIndentSpaces)),
			YAMLEncodeOption(yaml.IndentSequence(false)),
		),
	)
}
