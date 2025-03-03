// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

import (
	"testing"

	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
)

func TestApplyYAMLPatch_goyaml(t *testing.T) {
	yamlpatchcommon.RunApplyYAMLPatchTests(t, "goyaml", newGoyamlYAMLLibrary())
}

func TestApplyYAMLPatch_CustomObjectTest_goyaml(t *testing.T) {
	yamlpatchcommon.RunApplyYAMLPatchCustomObjectTests(t, "goyaml", New())
}
