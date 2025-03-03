// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package yamlpatch is a library for applying RFC6902 JSON patches to yaml documents.
// It leverages go-yaml v3's Node type to preserve comments, ordering, and most formatting.
package yamlpatch

import (
	"github.com/palantir/pkg/yamlpatch/gopkgv3yamlpatcher"
)

// Apply calls ApplyUsingYAMLLibrary using the "gopkg.in/yaml.v3" YAML library.
func Apply(originalBytes []byte, patch Patch) ([]byte, error) {
	return gopkgv3yamlpatcher.New().Apply(originalBytes, patch)
}
