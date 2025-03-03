// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

import (
	"testing"

	"github.com/palantir/pkg/yamlpatch/internal/testhelpers"
)

func TestApplyYAMLPatch(t *testing.T) {
	testhelpers.RunApplyYAMLPatchTests(t, "goyaml", newGoyamlYAMLLibrary())
}

func TestApplyYAMLPatch_CustomObjectTest(t *testing.T) {
	testhelpers.RunApplyYAMLPatchCustomObjectTests(t, "goyaml", New())
}
