// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	OperationAdd     = "add"
	OperationReplace = "replace"
	OperationRemove  = "remove"
	OperationMove    = "move"
	OperationCopy    = "copy"
	OperationTest    = "test"
)

type Patch []Operation

// Operation represents a RFC6902 JSON Patch operation.
type Operation struct {
	Type  string      `json:"op" yaml:"op"`
	Path  Path        `json:"path" yaml:"path"`
	From  Path        `json:"from,omitempty" yaml:"from,omitempty"`
	Value interface{} `json:"value,omitempty" yaml:"value,omitempty"`

	// If not empty, will be inserted as a head comment above this node
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
}

func (op Operation) String() string {
	return fmt.Sprintf("op: %q, path: %q, value: %v, from: %q", op.Type, op.Path.String(), op.Value, op.From.String())
}

// Path represents a decoded JSON patch targeting a location within a document.
// Use ParsePath or UnmarshalText to construct a Path.
type Path []string

// MustParsePath is like ParsePath but panics in case of invalid input.
func MustParsePath(str string) Path {
	path, err := ParsePath(str)
	if err != nil {
		panic(err)
	}
	return path
}

func ParsePath(str string) (Path, error) {
	if !strings.HasPrefix(str, "/") {
		return nil, errors.Errorf("path must begin with leading slash")
	}
	// Special case for root node of document
	if str == "/" {
		return Path{""}, nil
	}
	// General case: leading slash results in empty leading element.
	// This represents indexing into the document node.
	split := strings.Split(str, "/")
	for i := range split {
		split[i] = rfc6901Decoder.Replace(split[i])
	}
	return split, nil
}

func (p Path) String() string {
	switch len(p) {
	case 0:
		return ""
	case 1:
		return "/"
	default:
		parts := make([]string, len(p))
		for i, part := range p {
			parts[i] = rfc6901Encoder.Replace(part)
		}
		return strings.Join(parts, "/")
	}
}

func (p Path) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Path) UnmarshalText(text []byte) error {
	path, err := ParsePath(string(text))
	if err != nil {
		return err
	}
	*p = path
	return nil
}

func (p Path) Key() string {
	if len(p) == 0 {
		return ""
	}
	return p[len(p)-1]
}

var (
	// rfc6901Decoder and rfc6901Encoder implement http://tools.ietf.org/html/rfc6901#section-4 :
	//
	// Evaluation of each reference token begins by decoding any escaped
	// character sequence.  This is performed by first transforming any
	// occurrence of the sequence '~1' to '/', and then transforming any
	// occurrence of the sequence '~0' to '~'.

	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
	rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")
)
