// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
	"github.com/palantir/pkg/yamlpatch/yamlpatch"
	"github.com/pkg/errors"
)

func New(opts ...YAMLOption) yamlpatch.Patcher {
	return yamlpatchcommon.NewYAMLPatchApplier(newGoccyYAMLLibrary(opts...))
}

func newGoccyYAMLLibrary(opts ...YAMLOption) yamlpatchcommon.YAMLLibrary[ast.Node] {
	yamllib := &goccyYAMLLib{}

	defaultOptions := []YAMLOption{
		UseNonFlowWhenModifyingEmptyContainer(true),
		YAMLEncodeOption(yaml.Indent(yamlpatchcommon.DefaultIndentSpaces)),
		YAMLEncodeOption(yaml.IndentSequence(true)),
	}
	allOptions := append(defaultOptions, opts...)
	for _, opt := range allOptions {
		opt.apply(yamllib)
	}

	return yamllib
}

var _ yamlpatchcommon.YAMLLibrary[ast.Node] = (*goccyYAMLLib)(nil)

type goccyYAMLLib struct {
	useNonFlowWhenModifyingEmptyContainer bool
	encodeOptions                         []yaml.EncodeOption
}

func (g *goccyYAMLLib) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

func (g *goccyYAMLLib) BytesToNode(in []byte) (ast.Node, error) {
	parsedYAML, err := parser.ParseBytes(in, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal original yaml bytes")
	}
	if numParsedYAMLDocs := len(parsedYAML.Docs); numParsedYAMLDocs != 1 {
		return nil, errors.Errorf("parsed YAML must contain exactly one document, but got %d", numParsedYAMLDocs)
	}
	return parsedYAML.Docs[0], nil
}

func (g *goccyYAMLLib) BytesToContentNode(in []byte) (ast.Node, error) {
	node, err := g.BytesToNode(in)
	if err != nil {
		return nil, err
	}
	if docNode, ok := node.(*ast.DocumentNode); ok {
		return docNode.Body, nil
	}
	return node, nil
}

func (g *goccyYAMLLib) NodeToBytes(node ast.Node) ([]byte, error) {
	nodeBytes, err := node.MarshalYAML()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal node to YAML")
	}
	if len(nodeBytes) > 0 {
		nodeBytes = append(nodeBytes, '\n')
	}
	return nodeBytes, nil
}

func (g *goccyYAMLLib) NodeToValue(node ast.Node) (any, error) {
	var obj any
	if err := yaml.NodeToValue(node, &obj); err != nil {
		return nil, errors.Wrapf(err, "failed to convert YAML node to value")
	}
	return obj, nil
}

func (g *goccyYAMLLib) ValueToNode(value any, comment string) (ast.Node, error) {
	node, err := yaml.ValueToNode(value, g.encodeOptions...)
	if err != nil {
		return nil, err
	}

	if comment != "" {
		comment = " " + comment
		commentGroup := ast.CommentGroup([]*token.Token{
			token.Comment(comment, comment, nil),
		})

		if mappingNode, ok := node.(*ast.MappingNode); ok && len(mappingNode.Values) > 0 {
			// if value is a mapping node with at least one value, set comment on the first mapping node value, since
			// that is what will render the comment "above" the key
			mappingNode.Values[0].Comment = commentGroup
		} else {
			if err := node.SetComment(commentGroup); err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}

func (g *goccyYAMLLib) SetDocumentNodeContent(documentNode ast.Node, valueNode ast.Node) error {
	docNode, ok := documentNode.(*ast.DocumentNode)
	if !ok {
		return errors.Errorf("documentNode must be a DocumentNode, but was %s", documentNode.Type())
	}

	// nil out any start/end markers since there is now content
	docNode.Start = nil
	docNode.End = nil

	docNode.Body = valueNode
	return nil
}

func (g *goccyYAMLLib) NewContainer(node ast.Node) (yamlpatchcommon.YAMLContainer[ast.Node], error) {
	return newGoccyContainer(node, g.useNonFlowWhenModifyingEmptyContainer, g.encodeOptions...)
}

func (g *goccyYAMLLib) CopyNode(node ast.Node) (ast.Node, error) {
	// Create a deep copy of the node
	nodeBytes, err := g.NodeToBytes(node)
	if err != nil {
		return nil, err
	}
	nodeCopy, err := g.BytesToContentNode(nodeBytes)
	if err != nil {
		return nil, err
	}
	return nodeCopy, nil
}

func (g *goccyYAMLLib) NodeIsNil(node ast.Node) bool {
	return node == nil
}
