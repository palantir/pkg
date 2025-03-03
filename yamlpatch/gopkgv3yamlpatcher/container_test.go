// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

import (
	"testing"

	"github.com/palantir/pkg/yamlpatch/internal/testhelpers"
)

func TestContainers(t *testing.T) {
	testhelpers.RunContainerTests(t, "goyaml", newGoyamlYAMLLibrary(IndentSpaces(testhelpers.ContainerTestIndentSpaces)))
}
