// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatchcommon

import (
	"reflect"

	"github.com/palantir/pkg/yamlpatch/yamlpatch"
	"github.com/pkg/errors"
)

var (
	// DefaultIndentSpaces configures the yaml encoder used to marshal patched objects.
	DefaultIndentSpaces = 2
)

// ApplyUsingYAMLLibrary applies the provided patch to a yaml document provided in originalBytes using the provided YAML
// library and returns the updated content. A best effort is made to minimize changes outside the patched paths but some
// whitespace changes are unavoidable.
func ApplyUsingYAMLLibrary[NodeT any](yamllib YAMLLibrary[NodeT], originalBytes []byte, patch yamlpatch.Patch) ([]byte, error) {
	node, err := yamllib.BytesToNode(originalBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert YAML bytes to node")
	}

	for _, op := range patch {
		if op.Path == nil {
			return nil, errors.Errorf("op %s requires a path field", op.Type)
		}
		var err error
		switch op.Type {
		case yamlpatch.OperationAdd:
			err = patchAdd(yamllib, node, op.Path, op.Value, op.Comment)
		case yamlpatch.OperationRemove:
			err = patchRemove(yamllib, node, op.Path)
		case yamlpatch.OperationReplace:
			err = patchReplace(yamllib, node, op.Path, op.Value, op.Comment)
		case yamlpatch.OperationMove:
			err = patchMove(yamllib, node, op.Path, op.From)
		case yamlpatch.OperationCopy:
			err = patchCopy(yamllib, node, op.Path, op.From)
		case yamlpatch.OperationTest:
			err = patchTest(yamllib, node, op.Path, op.Value)
		default:
			err = errors.Errorf("unexpected op")
		}
		if err != nil {
			return nil, errors.Wrapf(err, "op %s %s", op.Type, op.Path.String())
		}
	}

	patchedBytes, err := yamllib.NodeToBytes(node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert patched node to bytes")
	}
	return patchedBytes, nil
}

// getParentAndLeaf returns the Node and its parent container specified by the Path segments.
// If the parent path (excluding the final key) does not exist, an error is returned.
// If the key does not exist in the container, the Node return value will be nil.
func getParentAndLeaf[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path) (YAMLContainer[NodeT], NodeT, error) {
	var (
		currNode      = node
		currContainer YAMLContainer[NodeT]
		err           error
	)
	for i, part := range path {
		if yamllib.NodeIsNil(currNode) {
			return nil, *new(NodeT), errors.Errorf("parent path %s does not exist", path[:i].String())
		}
		currContainer, err = yamllib.NewContainer(currNode)
		if err != nil {
			return nil, *new(NodeT), err
		}
		currNode, err = currContainer.Get(part)
		if err != nil {
			return nil, *new(NodeT), err
		}
	}
	return currContainer, currNode, nil
}

func patchAdd[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path, value interface{}, comment string) error {
	valueNode, err := yamllib.ValueToNode(value, comment)
	if err != nil {
		return err
	}

	// special case creating the root node
	if path.String() == "/" {
		return yamllib.SetDocumentNodeContent(node, valueNode)
	}

	parent, _, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	return parent.Add(path.Key(), valueNode)
}

func patchRemove[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path) error {
	parent, _, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	return parent.Remove(path.Key())
}

func patchReplace[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path, value interface{}, comment string) error {
	parent, _, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	valueNode, err := yamllib.ValueToNode(value, comment)
	if err != nil {
		return err
	}
	return parent.Set(path.Key(), valueNode)
}

func patchMove[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path, from yamlpatch.Path) error {
	fromParent, fromNode, err := getParentAndLeaf(yamllib, node, from)
	if err != nil {
		return err
	}
	if yamllib.NodeIsNil(fromNode) {
		return errors.Errorf("node not found for move patch at path %s", from)
	}

	toParent, _, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	if err := fromParent.Remove(from.Key()); err != nil {
		return err
	}
	return toParent.Add(path.Key(), fromNode)
}

func patchCopy[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path, from yamlpatch.Path) error {
	_, fromNode, err := getParentAndLeaf(yamllib, node, from)
	if err != nil {
		return err
	}
	if yamllib.NodeIsNil(fromNode) {
		return errors.Errorf("node not found for copy patch at path %s", from)
	}

	fromNodeCopy, err := yamllib.CopyNode(fromNode)
	if err != nil {
		return errors.Wrapf(err, "failed to create copy of node")
	}

	toParent, _, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	return toParent.Add(path.Key(), fromNodeCopy)
}

func patchTest[NodeT any](yamllib YAMLLibrary[NodeT], node NodeT, path yamlpatch.Path, testValue interface{}) error {
	_, valueNode, err := getParentAndLeaf(yamllib, node, path)
	if err != nil {
		return err
	}
	if yamllib.NodeIsNil(valueNode) {
		return errors.Errorf("node not found for test patch at path %s", path)
	}
	valueObj, err := yamllib.NodeToValue(valueNode)
	if err != nil {
		return err
	}
	// roundtrip test value to use standard type, comment is unset since this operation only
	// looks at the existing value and compares it against the test value
	testValueNode, err := yamllib.ValueToNode(testValue, "")
	if err != nil {
		return err
	}
	testValueObj, err := yamllib.NodeToValue(testValueNode)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(valueObj, testValueObj) {
		return errors.Errorf("testing path %s value failed", path)
	}
	return nil
}
