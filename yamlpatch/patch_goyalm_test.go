// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"testing"
)

func TestApplyYAMLPatch_goyaml(t *testing.T) {
	runApplyYAMLPatchTests(t, "goyaml", NewGoyamlYAMLLibrary())
}

func TestApplyYAMLPatch_CustomObjectTest_goyaml(t *testing.T) {
	runApplyYAMLPatchCustomObjectTests(t, "goyaml", NewGoyamlYAMLLibrary())
}
