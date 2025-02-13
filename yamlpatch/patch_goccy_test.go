// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestApplyYAMLPatch_goccy(t *testing.T) {
	runApplyYAMLPatchTests(
		t,
		"goccy",
		NewGoccyYAMLLibrary(
			yaml.IndentSequence(false),
		),
	)
}

func TestApplyYAMLPatch_CustomObjectTest_goccy(t *testing.T) {
	runApplyYAMLPatchCustomObjectTests(
		t,
		"goccy",
		NewGoccyYAMLLibrary(
			yaml.IndentSequence(false),
		),
	)
}
