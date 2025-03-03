// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

import (
	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// newGoyamlContainer returns the container impl matching node.Kind.
// If the node is not a Map or Sequence, an error is returned.
func newGoyamlContainer(node *yaml.Node) (yamlpatchcommon.YAMLContainer[*yaml.Node], error) {
	if node == nil {
		return nil, errors.Errorf("unexpected nil yaml node")
	}
	switch node.Kind {
	case yaml.MappingNode:
		return &goyamlMappingContainer{node: node}, nil
	case yaml.SequenceNode:
		return &goyamlSequenceContainer{node: node}, nil
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return nil, errors.Errorf("unexpected yaml node: expected DocumentNode to have 1 child, got %v", node.Content)
		}
		return &goyamlDocumentContainer{node: node}, nil
	case yaml.AliasNode:
		// Recursive call to bypass alias wrapping
		// TODO(maybe): Block writes to nodes accessed via alias since they may have unintended side effects.
		//  When generating the JSONPatch for a diff, the values are fully dealiased so if two paths that share an alias
		//  begin to differ, a change will be produced that ends up changing the alias target. This will change the
		//  resolved value(s) for the path that was supposed to remain unchanged. In this case the "best" approach is
		//  probably to copy the alias target to the original path then edit the copy and remove the alias reference.
		return newGoyamlContainer(node.Alias)
	case yaml.ScalarNode:
		return nil, errors.Errorf("unexpected yaml node: scalar can not be a container")
	default:
		return nil, errors.Errorf("unexpected yaml node: kind %d tag %s", node.Kind, node.Tag)
	}
}

var _ yamlpatchcommon.YAMLContainer[*yaml.Node] = (*goyamlMappingContainer)(nil)

type goyamlMappingContainer struct {
	node *yaml.Node
}

func (g *goyamlMappingContainer) Get(key string) (*yaml.Node, error) {
	if err := g.validate(); err != nil {
		return nil, err
	}
	_, _, valNode := g.find(key)
	return valNode, nil
}

func (g *goyamlMappingContainer) Set(key string, val *yaml.Node) error {
	if err := g.validate(); err != nil {
		return err
	}

	keyIdx, _, _ := g.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be replaced", key)
	}
	g.node.Content[keyIdx+1] = val
	return nil
}

func (g *goyamlMappingContainer) Add(key string, val *yaml.Node) error {
	if err := g.validate(); err != nil {
		return err
	}
	if _, _, existingValue := g.find(key); existingValue != nil {
		return errors.Errorf("key %s already exists and can not be added", key)
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	g.node.Content = append(g.node.Content, keyNode, val)
	return nil
}

func (g *goyamlMappingContainer) Remove(key string) error {
	if err := g.validate(); err != nil {
		return err
	}

	keyIdx, _, _ := g.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be removed", key)
	}
	ary := make([]*yaml.Node, len(g.node.Content)-2)
	copy(ary[0:keyIdx], g.node.Content[0:keyIdx])
	copy(ary[keyIdx:], g.node.Content[keyIdx+2:])
	// overwrite Content with new array
	g.node.Content = ary
	return nil
}

func (g *goyamlMappingContainer) find(key string) (keyIdx int, keyNode, valNode *yaml.Node) {
	for i := 0; i < len(g.node.Content); i += 2 {
		keyNode, valNode := g.node.Content[i], g.node.Content[i+1]
		if keyNode.Value == key {
			return i, keyNode, valNode
		}
	}
	return -1, nil, nil
}

func (g *goyamlMappingContainer) validate() error {
	if len(g.node.Content)%2 != 0 {
		return errors.Errorf("expected MappingNode to have even number of children, got %d", len(g.node.Content))
	}
	// Mapping nodes are stored as [k0,v0,k1,v1...] so we iterate two at a time.
	for i := 0; i < len(g.node.Content); i += 2 {
		keyNode := g.node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return errors.Errorf("jsonpatch only supports scalar mapping keys, got %d %s", keyNode.Kind, keyNode.Tag)
		}
	}
	return nil
}

var _ yamlpatchcommon.YAMLContainer[*yaml.Node] = (*goyamlSequenceContainer)(nil)

type goyamlSequenceContainer struct {
	node *yaml.Node
}

func (g *goyamlSequenceContainer) Get(key string) (*yaml.Node, error) {
	if key == "-" {
		return nil, nil
	}
	// Parse key into integer and index into array
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return nil, err
	}
	if idx > len(g.node.Content)-1 {
		// key is out of bounds
		return nil, nil
	}
	return g.node.Content[idx], nil
}

func (g *goyamlSequenceContainer) Set(key string, val *yaml.Node) error {
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Content)-1 {
		return errors.Errorf("set index key out of bounds (idx %d, len %d)", idx, len(g.node.Content))
	}
	g.node.Content[idx] = val
	return nil
}

func (g *goyamlSequenceContainer) Add(key string, val *yaml.Node) error {
	if key == "-" {
		g.node.Content = append(g.node.Content, val)
		return nil
	}
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Content) {
		return errors.Errorf("add index key out of bounds (idx %d, len %d)", idx, len(g.node.Content))
	}
	// create new array ary and insert val at idx
	ary := make([]*yaml.Node, len(g.node.Content)+1)
	copy(ary[0:idx], g.node.Content[0:idx])
	ary[idx] = val
	copy(ary[idx+1:], g.node.Content[idx:])
	// overwrite Content with new array
	g.node.Content = ary
	return nil
}

func (g *goyamlSequenceContainer) Remove(key string) error {
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Content)-1 {
		return errors.Errorf("remove index key out of bounds (idx %d, len %d)", idx, len(g.node.Content))
	}
	ary := make([]*yaml.Node, len(g.node.Content)-1)
	copy(ary[0:idx], g.node.Content[0:idx])
	copy(ary[idx:], g.node.Content[idx+1:])
	// overwrite Content with new array
	g.node.Content = ary
	return nil
}

var _ yamlpatchcommon.YAMLContainer[*yaml.Node] = (*goyamlDocumentContainer)(nil)

// goyamlDocumentContainer is a special container that wraps a yaml.Document.
// Since documents have a single element, the 'key' argument in all methods must be the empty string "".
// An error is returned if any other key is provided, since the intention is likely not to be accessing a document node.
type goyamlDocumentContainer struct {
	node *yaml.Node
}

func (g *goyamlDocumentContainer) Get(key string) (*yaml.Node, error) {
	if key != "" {
		return nil, yamlpatchcommon.ErrIllegalDocumentAccess
	}
	if g.isEmpty() {
		return nil, nil
	}
	return g.node.Content[0], nil
}

func (g *goyamlDocumentContainer) Set(key string, val *yaml.Node) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	if g.isEmpty() {
		return errors.Errorf("document value does not exist and can not be replaced")
	}
	g.node.Content[0] = val
	return nil
}

func (g *goyamlDocumentContainer) Add(key string, val *yaml.Node) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	// If we have a 'null' node, we can overwrite it.
	if !g.isEmpty() {
		return errors.Errorf("document value already exists and can not be added")
	}
	g.node.Content[0] = val
	return nil
}

func (g *goyamlDocumentContainer) Remove(key string) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	return errors.Errorf("document does not implement Remove()")
}

func (g *goyamlDocumentContainer) isEmpty() bool {
	return g.node.Content[0].Kind == yaml.ScalarNode && g.node.Content[0].Tag == "!!null"
}
