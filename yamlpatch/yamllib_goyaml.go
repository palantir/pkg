// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"bytes"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func NewGoyamlYAMLLibrary() YAMLLibrary[*yaml.Node] {
	return newGoyamlYAMLLibrary(defaultIndentSpaces)
}

func newGoyamlYAMLLibrary(indentSpaces int) YAMLLibrary[*yaml.Node] {
	return &goyamlYAMLLib{
		indentSpaces: indentSpaces,
	}
}

var _ YAMLLibrary[*yaml.Node] = (*goyamlYAMLLib)(nil)

type goyamlYAMLLib struct {
	indentSpaces int
}

func (g *goyamlYAMLLib) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

func (g *goyamlYAMLLib) BytesToNode(in []byte) (*yaml.Node, error) {
	node := new(yaml.Node)
	if err := yaml.Unmarshal(in, node); err != nil {
		return nil, errors.Wrap(err, "unmarshal original yaml bytes")
	}
	return node, nil
}

func (g *goyamlYAMLLib) BytesToContentNode(in []byte) (*yaml.Node, error) {
	node, err := g.BytesToNode(in)
	if err != nil {
		return nil, err
	}
	if node.Kind == yaml.DocumentNode {
		return node.Content[0], nil
	}
	return node, nil
}

func (g *goyamlYAMLLib) NodeToBytes(node *yaml.Node) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer func() {
		_ = enc.Close()
	}()
	enc.SetIndent(g.indentSpaces)
	if err := enc.Encode(node); err != nil {
		return nil, errors.Wrapf(err, "marshal patched node")
	}
	return buf.Bytes(), nil
}

func (g *goyamlYAMLLib) NodeToValue(node *yaml.Node) (any, error) {
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

func (g *goyamlYAMLLib) ValueToNode(value any, comment string) (*yaml.Node, error) {
	yamlBytes, err := yaml.Marshal(value)
	if err != nil {
		return nil, err
	}
	node, err := g.BytesToContentNode(yamlBytes)
	if err != nil {
		return nil, err
	}
	g.clearYAMLStyle(node)
	node.HeadComment = comment
	return node, nil
}

func (g *goyamlYAMLLib) SetDocumentNodeContent(documentNode *yaml.Node, valueNode *yaml.Node) error {
	switch {
	case documentNode.Kind == yaml.DocumentNode && documentNode.Content[0].Kind == yaml.ScalarNode && documentNode.Content[0].Tag == "!!null":
		// Empty document
		documentNode.Content[0] = valueNode
		return nil
	case documentNode.Kind == 0 && len(documentNode.Content) == 0:
		// Create new document
		documentNode.Kind = yaml.DocumentNode
		documentNode.Content = []*yaml.Node{valueNode}
		return nil
	}
	return nil
}

func (g *goyamlYAMLLib) NewContainer(node *yaml.Node) (YAMLContainer[*yaml.Node], error) {
	return newGoyamlContainer(node)
}

func (g *goyamlYAMLLib) CopyNode(node *yaml.Node) (*yaml.Node, error) {
	// Create a deep copy of the value
	nodeYAMLBytes, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}
	nodeCopy, err := g.BytesToContentNode(nodeYAMLBytes)
	if err != nil {
		return nil, err
	}
	return nodeCopy, nil
}

func (g *goyamlYAMLLib) clearYAMLStyle(node *yaml.Node) {
	node.Style = 0
	node.Line = 0
	node.Column = 0
	for _, n := range node.Content {
		g.clearYAMLStyle(n)
	}
}

func (g *goyamlYAMLLib) NodeIsNil(node *yaml.Node) bool {
	return node == nil
}
