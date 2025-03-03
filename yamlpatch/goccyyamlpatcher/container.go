// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"io"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
	"github.com/pkg/errors"
)

type flowStyler interface {
	SetIsFlowStyle(isFlow bool)
}

// newGoccyContainer returns the container impl matching node.Kind.
// If the node is not a Map or Sequence, an error is returned.
func newGoccyContainer(node ast.Node, useNonFlowWhenAddingToEmpty bool, encodeOptions ...yaml.EncodeOption) (yamlpatchcommon.YAMLContainer[ast.Node], error) {
	if node == nil {
		return nil, errors.Errorf("unexpected nil yaml node")
	}

	switch node.Type() {
	case ast.MappingType:
		return &goccyMappingContainer{
			node:                         node.(*ast.MappingNode),
			useNonFlowWhenModifyingEmpty: useNonFlowWhenAddingToEmpty,
			encodeOptions:                encodeOptions,
		}, nil
	case ast.SequenceType:
		return &goccySequenceContainer{
			node:                         node.(*ast.SequenceNode),
			useNonFlowWhenModifyingEmpty: useNonFlowWhenAddingToEmpty,
			encodeOptions:                encodeOptions,
		}, nil
	case ast.DocumentType:
		return &goccyDocumentContainer{
			node:          node.(*ast.DocumentNode),
			encodeOptions: encodeOptions,
		}, nil
	case ast.AliasType:
		// Recursive call to bypass alias wrapping
		// TODO(maybe): Block writes to nodes accessed via alias since they may have unintended side effects.
		//  When generating the JSONPatch for a diff, the values are fully dealiased so if two paths that share an alias
		//  begin to differ, a change will be produced that ends up changing the alias target. This will change the
		//  resolved value(s) for the path that was supposed to remain unchanged. In this case the "best" approach is
		//  probably to copy the alias target to the original path then edit the copy and remove the alias reference.
		return newGoccyContainer(node.(*ast.AliasNode).Value, useNonFlowWhenAddingToEmpty, encodeOptions...)
	default:
		return nil, errors.Errorf("unexpected yaml node: type %s", node.Type())
	}
}

var _ yamlpatchcommon.YAMLContainer[ast.Node] = (*goccyMappingContainer)(nil)

type goccyMappingContainer struct {
	node                         *ast.MappingNode
	useNonFlowWhenModifyingEmpty bool
	encodeOptions                []yaml.EncodeOption
}

func (g *goccyMappingContainer) Get(key string) (ast.Node, error) {
	if err := g.validate(); err != nil {
		return nil, err
	}
	_, _, valNode := g.find(key)
	return valNode, nil
}

func (g *goccyMappingContainer) Set(key string, val ast.Node) error {
	if err := g.validate(); err != nil {
		return err
	}

	keyIdx, keyNode, prevValNode := g.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be replaced", key)
	}

	if len(g.node.Values) == 0 && g.useNonFlowWhenModifyingEmpty {
		// special case: when setting on empty map ("{}") with useNonFlowWhenModifyingEmpty=true, use non-flow style
		g.node.IsFlowStyle = false
	}

	// match flow style of node
	if flowNode, ok := val.(flowStyler); ok {
		flowNode.SetIsFlowStyle(g.node.IsFlowStyle)
	}

	// if new and old values are of same type (scalar or non-scalar) match indent level
	_, newValueIsScalar := val.(ast.ScalarNode)
	_, prevValueIsScalar := prevValNode.(ast.ScalarNode)
	if newValueIsScalar == prevValueIsScalar {
		indentNum := prevValNode.GetToken().Position.IndentNum
		if _, ok := val.(*ast.SequenceNode); ok && getIsIndentSequence(g.encodeOptions) && !(prevValNode.GetToken().Position.Line == keyNode.GetToken().Position.Line) {
			indentNum -= getIndentSpaces(g.encodeOptions)
		}
		val.AddColumn(indentNum)
	} else if !newValueIsScalar {
		// new value is not a scalar, but old value was: set indent level of new value to be 1 more than key
		val.AddColumn(keyNode.GetToken().Position.IndentNum + getIndentSpaces(g.encodeOptions))
	}

	g.node.Values[keyIdx].Value = val
	return nil
}

// roundabout workaround to extract the indent level from the encoding options
func getIndentSpaces(encodeOptions []yaml.EncodeOption) int {
	encoder := yaml.NewEncoder(io.Discard, encodeOptions...)
	node, _ := encoder.EncodeToNode(true)
	return node.GetToken().Position.IndentNum
}

// roundabout workaround to extract "IsIndentSequence" value from the encoding options
func getIsIndentSequence(encodeOptions []yaml.EncodeOption) bool {
	encoder := yaml.NewEncoder(io.Discard, encodeOptions...)
	// write a single sequence node and verify whether it is indented
	node, _ := encoder.EncodeToNode([]bool{true})
	nodePos := node.GetToken().Position
	return nodePos.Column > nodePos.IndentNum
}

func (g *goccyMappingContainer) Add(key string, val ast.Node) error {
	if err := g.validate(); err != nil {
		return err
	}
	if _, _, existingValue := g.find(key); existingValue != nil {
		return errors.Errorf("key %s already exists and can not be added", key)
	}

	if len(g.node.Values) == 0 && g.useNonFlowWhenModifyingEmpty {
		// special case: when adding to empty map ("{}") with useNonFlowWhenModifyingEmpty=true, use non-flow style
		g.node.IsFlowStyle = false
	}

	// match flow style of node
	if flowNode, ok := val.(flowStyler); ok {
		flowNode.SetIsFlowStyle(g.node.IsFlowStyle)
	}

	mapEntryNode, err := yaml.ValueToNode(yaml.MapSlice{
		{
			Key:   key,
			Value: val,
		},
	}, yaml.Indent(getIndentSpaces(g.encodeOptions)), yaml.IndentSequence(true))
	if err != nil {
		return err
	}
	mapEntryNodeTyped := mapEntryNode.(*ast.MappingNode)
	mapEntryValue := mapEntryNodeTyped.Values[0]

	if len(g.node.Values) > 0 {
		mapEntryValue.AddColumn(g.node.Values[0].Key.GetToken().Position.IndentNum)
	} else {
		mapEntryValue.AddColumn(g.node.GetToken().Prev.Position.IndentNum + getIndentSpaces(g.encodeOptions))
	}

	g.node.Values = append(g.node.Values, mapEntryValue)
	return nil
}

func (g *goccyMappingContainer) Remove(key string) error {
	if err := g.validate(); err != nil {
		return err
	}

	keyIdx, _, _ := g.find(key)
	if keyIdx == -1 {
		return errors.Errorf("key %s does not exist and can not be removed", key)
	}
	// remove entry from content
	g.node.Values = append(g.node.Values[:keyIdx], g.node.Values[keyIdx+1:]...)
	return nil
}

func (g *goccyMappingContainer) find(key string) (keyIdx int, keyNode, valNode ast.Node) {
	for entryIdx, mapEntry := range g.node.Values {
		// Use mapEntry.Key.GetToken().Value instead of mapEntry.Key.String() because former should always be the
		// raw/unquoted string value (latter may return string value that includes quotes)
		if mapEntry.Key.GetToken().Value == key {
			return entryIdx, mapEntry.Key, mapEntry.Value
		}
	}
	return -1, nil, nil
}

func (g *goccyMappingContainer) validate() error {
	for _, mapEntry := range g.node.Values {
		if _, isScalar := mapEntry.Key.(ast.ScalarNode); !isScalar {
			return errors.Errorf("jsonpatch only supports scalar mapping keys, got %s", mapEntry.Key.Type())
		}
	}
	return nil
}

var _ yamlpatchcommon.YAMLContainer[ast.Node] = (*goccySequenceContainer)(nil)

type goccySequenceContainer struct {
	node                         *ast.SequenceNode
	useNonFlowWhenModifyingEmpty bool
	encodeOptions                []yaml.EncodeOption
}

func (g *goccySequenceContainer) Get(key string) (ast.Node, error) {
	if key == "-" {
		return nil, nil
	}
	// Parse key into integer and index into array
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return nil, err
	}
	if idx > len(g.node.Values)-1 {
		// key is out of bounds
		return nil, nil
	}
	return g.node.Values[idx], nil
}

func (g *goccySequenceContainer) Set(key string, val ast.Node) error {
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Values)-1 {
		return errors.Errorf("set index key out of bounds (idx %d, len %d)", idx, len(g.node.Values))
	}

	if len(g.node.Values) == 0 && g.useNonFlowWhenModifyingEmpty {
		// special case: when setting on empty sequence ("[]]") with useNonFlowWhenModifyingEmpty=true, use non-flow style
		g.node.IsFlowStyle = false
	}

	// match flow style of node
	if flowNode, ok := val.(flowStyler); ok {
		flowNode.SetIsFlowStyle(g.node.IsFlowStyle)
	}

	g.node.Values[idx] = val
	return nil
}

func (g *goccySequenceContainer) Add(key string, val ast.Node) error {
	if len(g.node.Values) == 0 && g.useNonFlowWhenModifyingEmpty {
		// special case: when adding to empty sequence ("[]") with useNonFlowWhenModifyingEmpty=true, use non-flow style
		g.node.IsFlowStyle = false

		// when sequences are formatted, formatting is performed based on the column.
		// If sequence was previously flow style AND the sequence content was on the same line as the previous token (:),
		// set column for node to be 1 indented from key
		if g.node.Start != nil && g.node.Start.Position != nil &&
			g.node.GetToken().Prev != nil && g.node.GetToken().Prev.Position != nil &&
			g.node.GetToken().Prev.Prev != nil && g.node.GetToken().Prev.Prev.Position != nil {
			seqStartBracketPos := g.node.Start.Position
			colonBeforeSeq := g.node.GetToken().Prev
			defBeforeSeq := colonBeforeSeq.Prev

			// previously, sequence was on same line as definition: since this is changing, update position
			if seqStartBracketPos.Line == colonBeforeSeq.Position.Line {
				g.node.Start.Position.Column = defBeforeSeq.Position.Column
				if getIsIndentSequence(g.encodeOptions) {
					g.node.Start.Position.Column += getIndentSpaces(g.encodeOptions)
				}
			}
		}
	}

	// match flow style of node
	if flowNode, ok := val.(flowStyler); ok {
		flowNode.SetIsFlowStyle(g.node.IsFlowStyle)
	}

	var commentGroup *ast.CommentGroupNode
	if !g.node.IsFlowStyle {
		// comments for sequence items should be "head" comments (occur on the line before the value):
		// if node has a comment set, clear it and add it to the appropriate head comment index.
		// Do so even if empty so that the head comments for other entries line up properly.
		commentGroup = val.GetComment()
		if commentGroup != nil {
			if err := val.SetComment(nil); err != nil {
				return errors.Wrapf(err, "failed to clear comment for node")
			}
		}
	}

	if key == "-" {
		g.node.Values = append(g.node.Values, val)
		if !g.node.IsFlowStyle {
			g.node.ValueHeadComments = append(g.node.ValueHeadComments, commentGroup)
		}
		return nil
	}
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Values) {
		return errors.Errorf("add index key out of bounds (idx %d, len %d)", idx, len(g.node.Values))
	}

	// update values with inserted element
	g.node.Values = slices.Insert(g.node.Values, idx, val)

	// if sequence is not "flow" style (single line), then also insert comment into value head comments slice
	if !g.node.IsFlowStyle {
		g.node.ValueHeadComments = slices.Insert(g.node.ValueHeadComments, idx, commentGroup)

		// special case: comments on the first element of a sequence need special handling. Conceptually, the comment for
		// the first element should be in g.node.ValueHeadsComments[0]. However, such a comment can also be interpreted as a
		// comment on the overall sequence node itself, and it appears that this is how the library interprets such a
		// comment when reading/loading from rendered YAML. In most cases this is fine, but if a new element is inserted at
		// position 0, this can cause an issue: if the comment is in g.node.ValueHeadComments then the comment would be
		// properly "shifted" to position 1, but if it's on the node, then the comment will remain rendered above position
		// 0. Add special-case logic to deal with this: if the insertion is for position 0, the overall node has a comment,
		// and there used to be another node at position 0, then move the comment from the overall node to position 1 and
		// clear out the original value.
		if idx == 0 && len(g.node.ValueHeadComments) > 1 && g.node.Comment != nil && g.node.ValueHeadComments[1] == nil {
			g.node.ValueHeadComments[1] = g.node.Comment
			g.node.Comment = nil
		}
	}

	return nil
}

func (g *goccySequenceContainer) Remove(key string) error {
	idx, err := yamlpatchcommon.ParseSeqIndex(key)
	if err != nil {
		return err
	}
	if idx > len(g.node.Values)-1 {
		return errors.Errorf("remove index key out of bounds (idx %d, len %d)", idx, len(g.node.Values))
	}
	// remove value from sequence
	g.node.Values = slices.Delete(g.node.Values, idx, idx+1)
	// if sequence is not flow style, remove comment node
	if !g.node.IsFlowStyle {
		g.node.ValueHeadComments = slices.Delete(g.node.ValueHeadComments, idx, idx+1)
	}
	return nil
}

var _ yamlpatchcommon.YAMLContainer[ast.Node] = (*goccyDocumentContainer)(nil)

// goyamlDocumentContainer is a special container that wraps a yaml.Document.
// Since documents have a single element, the 'key' argument in all methods must be the empty string "".
// An error is returned if any other key is provided, since the intention is likely not to be accessing a document node.
type goccyDocumentContainer struct {
	node          *ast.DocumentNode
	encodeOptions []yaml.EncodeOption
}

func (g *goccyDocumentContainer) Get(key string) (ast.Node, error) {
	if key != "" {
		return nil, yamlpatchcommon.ErrIllegalDocumentAccess
	}
	return g.node.Body, nil
}

func (g *goccyDocumentContainer) Set(key string, val ast.Node) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	g.node.Body = val
	return nil
}

func (g *goccyDocumentContainer) Add(key string, val ast.Node) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	g.node.Body = val
	return nil
}

func (g *goccyDocumentContainer) Remove(key string) error {
	if key != "" {
		return yamlpatchcommon.ErrIllegalDocumentAccess
	}
	return errors.Errorf("document does not implement Remove()")
}
