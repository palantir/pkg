// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

import (
	"testing"

	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
)

func TestContainers_goyaml(t *testing.T) {
	yamlpatchcommon.RunContainerTests(t, "goyaml", newGoyamlYAMLLibrary(IndentSpaces(yamlpatchcommon.ContainerTestIndentSpaces)))
}
