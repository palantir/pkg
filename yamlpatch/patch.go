// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package yamlpatch is a library for applying RFC6902 JSON patches to yaml documents.
// It leverages go-yaml v3's Node type to preserve comments, ordering, and most formatting.
package yamlpatch

import (
	"bytes"
	"reflect"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	// defaultIndentSpaces configures the yaml encoder used to marshal patched objects.
	defaultIndentSpaces = 2
)

// Apply the patch to a yaml document provided in originalBytes and return the updated content.
// A best effort is made to minimize changes outside the patched paths but some whitespace changes are unavoidable.
func Apply(originalBytes []byte, patch Patch) ([]byte, error) {
	node := new(yaml.Node)
	if err := yaml.Unmarshal(originalBytes, node); err != nil {
		return nil, errors.Wrap(err, "unmarshal original yaml bytes")
	}

	for _, op := range patch {
		if op.Path == nil {
			return nil, errors.Errorf("op %s requires a path field", op.Type)
		}
		var err error
		switch op.Type {
		case OperationAdd:
			err = patchAdd(node, op.Path, op.Value, op.Comment)
		case OperationRemove:
			err = patchRemove(node, op.Path)
		case OperationReplace:
			err = patchReplace(node, op.Path, op.Value, op.Comment)
		case OperationMove:
			err = patchMove(node, op.Path, op.From)
		case OperationCopy:
			err = patchCopy(node, op.Path, op.From)
		case OperationTest:
			err = patchTest(node, op.Path, op.Value)
		default:
			err = errors.Errorf("unexpected op")
		}
		if err != nil {
			return nil, errors.Wrapf(err, "op %s %s", op.Type, op.Path.String())
		}
	}
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer func() {
		_ = enc.Close()
	}()
	enc.SetIndent(defaultIndentSpaces)
	if err := enc.Encode(node); err != nil {
		return nil, errors.Wrapf(err, "marshal patched node")
	}
	return buf.Bytes(), nil
}

func patchAdd(node *yaml.Node, path Path, value interface{}, comment string) error {
	valueNode, err := valueToYAMLNode(value, comment)
	if err != nil {
		return err
	}

	// special case creating the root node
	if path.String() == "/" {
		switch {
		case node.Kind == yaml.DocumentNode && node.Content[0].Kind == yaml.ScalarNode && node.Content[0].Tag == "!!null":
			// Empty document
			node.Content[0] = valueNode
			return nil
		case node.Kind == 0 && len(node.Content) == 0:
			// Create new document
			node.Kind = yaml.DocumentNode
			node.Content = []*yaml.Node{valueNode}
			return nil
		}
	}

	parent, _, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	return parent.Add(path.Key(), valueNode)
}

func patchRemove(node *yaml.Node, path Path) error {
	parent, _, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	return parent.Remove(path.Key())
}

func patchReplace(node *yaml.Node, path Path, value interface{}, comment string) error {
	parent, _, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	valueNode, err := valueToYAMLNode(value, comment)
	if err != nil {
		return err
	}
	return parent.Set(path.Key(), valueNode)
}

func patchMove(node *yaml.Node, path Path, from Path) error {
	fromParent, fromNode, err := getParentAndLeaf(node, from)
	if err != nil {
		return err
	}
	if fromNode == nil {
		return errors.Errorf("node not found for move patch at path %s", from)
	}

	toParent, _, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	if err := fromParent.Remove(from.Key()); err != nil {
		return err
	}
	return toParent.Add(path.Key(), fromNode)
}

func patchCopy(node *yaml.Node, path Path, from Path) error {
	_, fromNode, err := getParentAndLeaf(node, from)
	if err != nil {
		return err
	}
	if fromNode == nil {
		return errors.Errorf("node not found for copy patch at path %s", from)
	}

	// Create a deep copy of the value
	fromNodeYAMLBytes, err := yaml.Marshal(fromNode)
	if err != nil {
		return err
	}
	fromNodeClone, err := unmarshalNode(fromNodeYAMLBytes)
	if err != nil {
		return err
	}

	toParent, _, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	return toParent.Add(path.Key(), fromNodeClone)
}

func patchTest(node *yaml.Node, path Path, testValue interface{}) error {
	_, valueNode, err := getParentAndLeaf(node, path)
	if err != nil {
		return err
	}
	if valueNode == nil {
		return errors.Errorf("node not found for test patch at path %s", path)
	}
	valueObj, err := yamlNodeToValue(valueNode)
	if err != nil {
		return err
	}
	// roundtrip test value to use standard type, comment is unset since this operation only
	// looks at the existing value and compares it against the test value
	testValueNode, err := valueToYAMLNode(testValue, "")
	if err != nil {
		return err
	}
	testValueObj, err := yamlNodeToValue(testValueNode)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(valueObj, testValueObj) {
		return errors.Errorf("testing path %s value failed", path)
	}
	return nil
}

// getParentAndLeaf returns the Node and its parent container specified by the Path segments.
// If the parent path (excluding the final key) does not exist, an error is returned.
// If the key does not exist in the container, the Node return value will be nil.
func getParentAndLeaf(node *yaml.Node, path Path) (container, *yaml.Node, error) {
	var (
		currNode      = node
		currContainer container
		err           error
	)
	for i, part := range path {
		if currNode == nil {
			return nil, nil, errors.Errorf("parent path %s does not exist", path[:i].String())
		}
		currContainer, err = newContainer(currNode)
		if err != nil {
			return nil, nil, err
		}
		currNode, err = currContainer.Get(part)
		if err != nil {
			return nil, nil, err
		}
	}
	return currContainer, currNode, nil
}

// unmarshalNode unmarshals the text as a yaml.Node and removes the DocumentNode wrapping,
// returning the underlying content node.
func unmarshalNode(text []byte) (*yaml.Node, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(text, &node); err != nil {
		return nil, err
	}
	if node.Kind != yaml.DocumentNode {
		return nil, errors.Errorf("expected freshly unmarshaled Node to be DocumentNode, got %d", node.Kind)
	}
	return node.Content[0], nil
}

func valueToYAMLNode(value interface{}, comment string) (*yaml.Node, error) {
	yamlBytes, err := yaml.Marshal(value)
	if err != nil {
		return nil, err
	}
	node, err := unmarshalNode(yamlBytes)
	if err != nil {
		return nil, err
	}
	clearYAMLStyle(node)
	node.HeadComment = comment
	return node, nil
}

func yamlNodeToValue(node *yaml.Node) (interface{}, error) {
	yamlBytes, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}
	var obj interface{}
	if err := yaml.Unmarshal(yamlBytes, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func clearYAMLStyle(node *yaml.Node) {
	node.Style = 0
	node.Line = 0
	node.Column = 0
	for _, n := range node.Content {
		clearYAMLStyle(n)
	}
}
