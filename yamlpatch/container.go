// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// container is an interface to abstract away indexing into sequence (list) and mapping (object) nodes.
// Keys are strings for compatibility with JSONPatch and JSON map keys.
// Sequence containers parse integer indices from the string representation.
type container interface {
	// Get returns the child node at the key index. If the key does not exist, it returns nil, nil.
	Get(key string) (*yaml.Node, error)
	// Set overwrites the key with val. It returns an error if the key does not already exist (or index out of bounds, for sequences).
	Set(key string, val *yaml.Node) error
	// Add adds a new node to the container. It returns an error if the key already exists.
	Add(key string, val *yaml.Node) error
	// Remove removes a node from the container. It returns an error if the key does not exist.
	Remove(key string) error
}

// newContainer returns the container impl matching node.Kind.
// If the node is not a Map or Sequence, an error is returned.
func newContainer(node *yaml.Node) (container, error) {
	if node == nil {
		return nil, errors.Errorf("unexpected nil yaml node")
	}
	switch node.Kind {
	case yaml.MappingNode:
		return mappingContainer{node: node}, nil
	case yaml.SequenceNode:
		return sequenceContainer{node: node}, nil
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return nil, errors.Errorf("unexpected yaml node: expected DocumentNode to have 1 child, got %v", node.Content)
		}
		return documentContainer{node: node}, nil
	case yaml.AliasNode:
		// Recursive call to bypass alias wrapping
		// TODO(maybe): Block writes to nodes accessed via alias since they may have unintended side effects.
		//  When generating the JSONPatch for a diff, the values are fully dealiased so if two paths that share an alias
		//  begin to differ, a change will be produced that ends up changing the alias target. This will change the
		//  resolved value(s) for the path that was supposed to remain unchanged. In this case the "best" approach is
		//  probably to copy the alias target to the original path then edit the copy and remove the alias reference.
		return newContainer(node.Alias)
	case yaml.ScalarNode:
		return nil, errors.Errorf("unexpected yaml node: scalar can not be a container")
	default:
		return nil, errors.Errorf("unexpected yaml node: kind %d tag %s", node.Kind, node.Tag)
	}
}

type mappingContainer struct {
	node *yaml.Node
}

func (c mappingContainer) Get(key string) (*yaml.Node, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	_, _, valNode := c.find(key)
	return valNode, nil
}

func (c mappingContainer) Set(key string, val *yaml.Node) error {
	if err := c.validate(); err != nil {
		return err
	}

	keyIdx, _, _ := c.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be replaced", key)
	}
	c.node.Content[keyIdx+1] = val
	return nil
}

func (c mappingContainer) Add(key string, val *yaml.Node) error {
	if err := c.validate(); err != nil {
		return err
	}
	if _, _, existingValue := c.find(key); existingValue != nil {
		return errors.Errorf("key %s already exists and can not be added", key)
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	c.node.Content = append(c.node.Content, keyNode, val)
	return nil
}

func (c mappingContainer) Remove(key string) error {
	if err := c.validate(); err != nil {
		return err
	}

	keyIdx, _, _ := c.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be removed", key)
	}
	ary := make([]*yaml.Node, len(c.node.Content)-2)
	copy(ary[0:keyIdx], c.node.Content[0:keyIdx])
	copy(ary[keyIdx:], c.node.Content[keyIdx+2:])
	// overwrite Content with new array
	c.node.Content = ary
	return nil
}

func (c mappingContainer) find(key string) (keyIdx int, keyNode, valNode *yaml.Node) {
	for i := 0; i < len(c.node.Content); i += 2 {
		keyNode, valNode := c.node.Content[i], c.node.Content[i+1]
		if keyNode.Value == key {
			return i, keyNode, valNode
		}
	}
	return -1, nil, nil
}

func (c mappingContainer) validate() error {
	if len(c.node.Content)%2 != 0 {
		return errors.Errorf("expected MappingNode to have even number of children, got %d", len(c.node.Content))
	}
	// Mapping nodes are stored as [k0,v0,k1,v1...] so we iterate two at a time.
	for i := 0; i < len(c.node.Content); i += 2 {
		keyNode := c.node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return errors.Errorf("jsonpatch only supports scalar mapping keys, got %d %s", keyNode.Kind, keyNode.Tag)
		}
	}
	return nil
}

type sequenceContainer struct {
	node *yaml.Node
}

func (c sequenceContainer) Get(key string) (*yaml.Node, error) {
	if key == "-" {
		return nil, nil
	}
	// Parse key into integer and index into array
	idx, err := parseSeqIndex(key)
	if err != nil {
		return nil, err
	}
	if idx > len(c.node.Content)-1 {
		// key is out of bounds
		return nil, nil
	}
	return c.node.Content[idx], nil
}

func (c sequenceContainer) Set(key string, val *yaml.Node) error {
	idx, err := parseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(c.node.Content)-1 {
		return errors.Errorf("set index key out of bounds (idx %d, len %d)", idx, len(c.node.Content))
	}
	c.node.Content[idx] = val
	return nil
}

func (c sequenceContainer) Add(key string, val *yaml.Node) error {
	if key == "-" {
		c.node.Content = append(c.node.Content, val)
		return nil
	}
	idx, err := parseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(c.node.Content) {
		return errors.Errorf("add index key out of bounds (idx %d, len %d)", idx, len(c.node.Content))
	}
	// create new array ary and insert val at idx
	ary := make([]*yaml.Node, len(c.node.Content)+1)
	copy(ary[0:idx], c.node.Content[0:idx])
	ary[idx] = val
	copy(ary[idx+1:], c.node.Content[idx:])
	// overwrite Content with new array
	c.node.Content = ary
	return nil
}

func (c sequenceContainer) Remove(key string) error {
	idx, err := parseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(c.node.Content)-1 {
		return errors.Errorf("remove index key out of bounds (idx %d, len %d)", idx, len(c.node.Content))
	}
	ary := make([]*yaml.Node, len(c.node.Content)-1)
	copy(ary[0:idx], c.node.Content[0:idx])
	copy(ary[idx:], c.node.Content[idx+1:])
	// overwrite Content with new array
	c.node.Content = ary
	return nil
}

func parseSeqIndex(indexStr string) (int, error) {
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, errors.Wrapf(err, "index into SequenceNode with non-integer %q key", indexStr)
	}
	if idx < 0 {
		return 0, errors.Errorf("index into SequenceNode with negative %q key", indexStr)
	}
	return idx, nil
}

// documentContainer is a special container that wraps a yaml.Document.
// Since documents have a single element, the 'key' argument in all methods must be the empty string "".
// An error is returned if any other key is provided, since the intention is likely not to be accessing a document node.
type documentContainer struct {
	node *yaml.Node
}

var errIllegalDocumentAccess = fmt.Errorf("documentContainer does not allow non-empty key access")

func (c documentContainer) Get(key string) (*yaml.Node, error) {
	if key != "" {
		return nil, errIllegalDocumentAccess
	}
	if c.isEmpty() {
		return nil, nil
	}
	return c.node.Content[0], nil
}

func (c documentContainer) Set(key string, val *yaml.Node) error {
	if key != "" {
		return errIllegalDocumentAccess
	}
	if c.isEmpty() {
		return errors.Errorf("document value does not exist and can not be replaced")
	}
	c.node.Content[0] = val
	return nil
}

func (c documentContainer) Add(key string, val *yaml.Node) error {
	if key != "" {
		return errIllegalDocumentAccess
	}
	// If we have a 'null' node, we can overwrite it.
	if !c.isEmpty() {
		return errors.Errorf("document value already exists and can not be added")
	}
	c.node.Content[0] = val
	return nil
}

func (c documentContainer) Remove(key string) error {
	if key != "" {
		return errIllegalDocumentAccess
	}
	return errors.Errorf("document does not implement Remove()")
}

func (c documentContainer) isEmpty() bool {
	return c.node.Content[0].Kind == yaml.ScalarNode && c.node.Content[0].Tag == "!!null"
}
