// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"github.com/palantir/pkg/yamlpatch/yamlpatch"
)

const (
	OperationAdd     = yamlpatch.OperationAdd
	OperationReplace = yamlpatch.OperationReplace
	OperationRemove  = yamlpatch.OperationRemove
	OperationMove    = yamlpatch.OperationMove
	OperationCopy    = yamlpatch.OperationCopy
	OperationTest    = yamlpatch.OperationTest
)

type Patch = yamlpatch.Patch

// Operation represents a RFC6902 JSON Patch operation.
type Operation = yamlpatch.Operation

// Path represents a decoded JSON patch targeting a location within a document.
// Use ParsePath or UnmarshalText to construct a Path.
type Path = yamlpatch.Path

// MustParsePath is like ParsePath but panics in case of invalid input.
func MustParsePath(str string) Path {
	return yamlpatch.MustParsePath(str)
}

func ParsePath(str string) (Path, error) {
	return yamlpatch.ParsePath(str)
}
