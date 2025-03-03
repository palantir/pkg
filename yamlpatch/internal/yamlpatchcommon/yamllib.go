// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatchcommon

import (
	"github.com/palantir/pkg/yamlpatch/yamlpatch"
)

func NewYAMLPatchApplier[NodeT any](lib YAMLLibrary[NodeT]) yamlpatch.Patcher {
	return &yamlLibraryPatcher[NodeT]{
		yamllib: lib,
	}
}

type yamlLibraryPatcher[NodeT any] struct {
	yamllib YAMLLibrary[NodeT]
}

func (y *yamlLibraryPatcher[T]) Apply(originalBytes []byte, patch yamlpatch.Patch) ([]byte, error) {
	return ApplyUsingYAMLLibrary(y.yamllib, originalBytes, patch)
}

type YAMLLibrary[NodeT any] interface {
	// Unmarshal unmarshals the provided YAML bytes into the provided output value.
	Unmarshal(in []byte, out interface{}) (err error)

	// BytesToNode unmarshals the provided YAML bytes into a node. The unmarshalled node will typically be a node that
	// represents a YAML document (a YAML document node), but the specifics depend on the implementation.
	BytesToNode(in []byte) (node NodeT, err error)

	// BytesToContentNode unmarshals the provided YAML bytes into a node. This function is similar to BytesToNode, but
	// if the node returned by the initial unmarshal is a document node, this function will return the content node of
	// the document node (if effectively "unwraps" the document node).
	BytesToContentNode(in []byte) (node NodeT, err error)

	NodeToBytes(node NodeT) (out []byte, err error)

	NodeToValue(node NodeT) (out any, err error)

	ValueToNode(value any, comment string) (NodeT, error)

	SetDocumentNodeContent(documentNode NodeT, valueNode NodeT) error

	NewContainer(node NodeT) (YAMLContainer[NodeT], error)

	CopyNode(node NodeT) (NodeT, error)

	// NodeIsNil returns true if the provided NodeT is nil (or its conceptual equivalent). For most implementations,
	// this will simply be implemented as "return node == nil". However, this function is necessary because the type
	// constraint of "NodeT" is "any", so in generic code "node == nil" / "node == *new(NodeT)" does not compile because
	// "NodeT" is not guaranteed to be comparable or a pointer type.
	// See https://groups.google.com/g/golang-nuts/c/bloyX1Zxjaw, https://github.com/golang/go/issues/61372
	NodeIsNil(node NodeT) bool
}
