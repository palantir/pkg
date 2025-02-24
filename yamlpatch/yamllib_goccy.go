// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
	"github.com/pkg/errors"
)

type GoccyYAMLLibraryOption interface {
	apply(opt *goccyYAMLLib)
}

type goccyYAMLLibOptionFunc func(opt *goccyYAMLLib)

func (f goccyYAMLLibOptionFunc) apply(opt *goccyYAMLLib) {
	f(opt)
}

func GoccyYAMLEncodeOption(encodeOption yaml.EncodeOption) GoccyYAMLLibraryOption {
	return goccyYAMLLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.encodeOptions = append(opt.encodeOptions, encodeOption)
	})
}

func GoccyUseNonFlowWhenModifyingEmptyContainer(useNonFlowWhenModifyingEmptyContainer bool) GoccyYAMLLibraryOption {
	return goccyYAMLLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.useNonFlowWhenModifyingEmptyContainer = useNonFlowWhenModifyingEmptyContainer
	})
}

func GoccyDisableAdjustIndentLevelWorkaround(disableAdjustIndentLevelWorkaround bool) GoccyYAMLLibraryOption {
	return goccyYAMLLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.disableAdjustIndentLevelWorkaround = disableAdjustIndentLevelWorkaround
	})
}

func NewGoccyYAMLLibrary(opts ...GoccyYAMLLibraryOption) YAMLLibrary[ast.Node] {
	yamllib := &goccyYAMLLib{}

	defaultOptions := []GoccyYAMLLibraryOption{
		GoccyUseNonFlowWhenModifyingEmptyContainer(true),
		GoccyYAMLEncodeOption(yaml.Indent(defaultIndentSpaces)),
		GoccyYAMLEncodeOption(yaml.IndentSequence(true)),
	}
	allOptions := append(defaultOptions, opts...)
	for _, opt := range allOptions {
		opt.apply(yamllib)
	}

	return yamllib
}

var _ YAMLLibrary[ast.Node] = (*goccyYAMLLib)(nil)

type goccyYAMLLib struct {
	useNonFlowWhenModifyingEmptyContainer bool
	disableAdjustIndentLevelWorkaround    bool
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
	documentNode := parsedYAML.Docs[0]
	if !g.disableAdjustIndentLevelWorkaround {
		g.adjustIndentLevels(documentNode, 0)
	}
	return documentNode, nil
}

// Exists as a workaround for https://github.com/goccy/go-yaml/issues/672.
// Currently, goccy/go-yaml appears to have a bug where, when parsing YAML, the indent level is not properly adjusted
// down. For example, for the input:
/*
level-0-key-0:
  level-1-key-1:
    level-2-key-1:
      level-3-key-1:
        level-4-key-1: value-1
level-0-key-1:
  level-1-key-2: value-2
*/
// The indent level for "level-0-key-1" should be "0". However, due to the bug referenced above, the level is actually
// "3" because the level of the previous token was "4" and the indent level is only decremented by 1.
//
// This function is a somewhat crude workaround: it traverses the nodes of the parsed YAML, tracks the "expected" indent
// level, and if the indent level of a node is greater/larger than the "expected" level, it sets the indent level to the
// "expected" one. It does not modify any other values. This works to fix the common case, but may not work properly for
// certain edge cases or node types. This behavior can be toggled off to avoid performing any adjustments.
func (g *goccyYAMLLib) adjustIndentLevels(node ast.Node, currExpectedLevel int) {
	switch typedNode := node.(type) {
	case ast.ScalarNode:
		typedNode.GetToken().Position.IndentLevel = min(currExpectedLevel, typedNode.GetToken().Position.IndentLevel)
	case *ast.DocumentNode:
		g.adjustIndentLevels(typedNode.Body, currExpectedLevel)
	case *ast.SequenceNode:
		// heuristic: if flow style is true and open brace is on same line as definition, expected level for key
		// and all values is the current level
		if typedNode.IsFlowStyle {
			if typedNode.Start != nil && typedNode.Start.Position != nil {
				typedNode.Start.Position.IndentLevel = currExpectedLevel
			}

			// start token is '['
			var startTokenPos *token.Position
			if typedNode.GetToken() != nil && typedNode.GetToken().Position != nil {
				startTokenPos = typedNode.GetToken().Position
			}

			// previous token is ':'
			var prevTokenPos *token.Position
			if typedNode.GetToken() != nil && typedNode.GetToken().Prev != nil && typedNode.GetToken().Prev.Position != nil {
				prevTokenPos = typedNode.GetToken().Prev.Position
			}

			// flow style is true.
			// Start with assumption that values are on same line: in that case, indent level is the same.
			// If not, then expect values indented by 1 more level if "isIndentSequence" is set.
			childIndentLevel := currExpectedLevel
			if startTokenPos != nil && prevTokenPos != nil && startTokenPos.Line != prevTokenPos.Line && getIsIndentSequence(g.encodeOptions) {
				childIndentLevel++
			}
			for _, entry := range typedNode.Entries {
				g.adjustIndentLevels(entry, childIndentLevel)
			}

			return
		}

		if typedNode.Start != nil && typedNode.Start.Position != nil {
			typedNode.Start.Position.IndentLevel = min(typedNode.Start.Position.IndentLevel, currExpectedLevel)
		}
		for _, entry := range typedNode.Entries {
			nextIndentLevel := currExpectedLevel
			if getIsIndentSequence(g.encodeOptions) {
				nextIndentLevel++
			}
			g.adjustIndentLevels(entry, nextIndentLevel)
		}
	case *ast.SequenceEntryNode:
		if typedNode.Start != nil && typedNode.Start.Position != nil {
			typedNode.Start.Position.IndentLevel = min(typedNode.Start.Position.IndentLevel, currExpectedLevel)
		}
		g.adjustIndentLevels(typedNode.Value, currExpectedLevel)
	case ast.MapNode:
		mapIter := typedNode.MapRange()
		for mapIter.Next() {
			// key of map is "current" level
			mapEntryKeyToken := mapIter.Key().GetToken()
			mapKeyTokenPos := mapEntryKeyToken.Position
			if mapKeyTokenPos != nil {
				mapEntryKeyToken.Position.IndentLevel = min(mapEntryKeyToken.Position.IndentLevel, currExpectedLevel)
			}

			// heuristic: if value of map is on same line, then it is the same level; otherwise, it is indented another level
			mapEntryValueNode := mapIter.Value()
			mapEntryValueToken := mapEntryValueNode.GetToken()
			mapEntryValueTokenPos := mapEntryValueToken.Position
			if mapEntryValueTokenPos != nil && mapKeyTokenPos != nil && mapEntryValueTokenPos.Line == mapKeyTokenPos.Line {
				g.adjustIndentLevels(mapEntryValueNode, currExpectedLevel)
			} else {
				g.adjustIndentLevels(mapEntryValueNode, currExpectedLevel+1)
			}
		}
	}
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

func (g *goccyYAMLLib) NewContainer(node ast.Node) (YAMLContainer[ast.Node], error) {
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
