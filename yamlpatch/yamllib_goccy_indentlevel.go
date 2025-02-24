// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// AddIndentLevel is analogous to AddColumn. Content is based on implementations of AddColumn and types that implement
// them.
func AddIndentLevel(input ast.Node, indentLevel int) {
	switch typed := input.(type) {
	case *ast.NullNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.FloatNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.NanNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.LiteralNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	case *ast.InfinityNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.AnchorNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Name != nil {
			AddIndentLevel(typed.Name, indentLevel)
		}
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	case *ast.AliasNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	case *ast.StringNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.MappingNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		AddIndentLevelToToken(typed.End, indentLevel)
		for _, value := range typed.Values {
			AddIndentLevel(value, indentLevel)
		}
	case *ast.CommentNode:
		if typed.Token == nil {
			return
		}
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.SequenceNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		AddIndentLevelToToken(typed.End, indentLevel)
		for _, value := range typed.Values {
			AddIndentLevel(value, indentLevel)
		}
	case *ast.BoolNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.IntegerNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.MergeKeyNode:
		AddIndentLevelToToken(typed.Token, indentLevel)
	case *ast.MappingKeyNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	case *ast.SequenceEntryNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
	case *ast.TagNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	case *ast.MappingValueNode:
		AddIndentLevelToToken(typed.Start, indentLevel)
		if typed.Key != nil {
			AddIndentLevel(typed.Key, indentLevel)
		}
		if typed.Value != nil {
			AddIndentLevel(typed.Value, indentLevel)
		}
	}
}

func AddIndentLevelToToken(input *token.Token, indentLevel int) {
	if input == nil {
		return
	}
	input.Position.IndentLevel += indentLevel
}
