// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatchcommon

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// YAMLContainer is an interface to abstract away indexing into sequence (list) and mapping (object) nodes.
// Keys are strings for compatibility with JSONPatch and JSON map keys.
// Sequence containers parse integer indices from the string representation.
type YAMLContainer[NodeT any] interface {
	// Get returns the child node at the key index. If the key does not exist, it returns nil, nil.
	Get(key string) (NodeT, error)
	// Set overwrites the key with val. It returns an error if the key does not already exist (or index out of bounds, for sequences).
	Set(key string, val NodeT) error
	// Add adds a new node to the container. It returns an error if the key already exists.
	Add(key string, val NodeT) error
	// Remove removes a node from the container. It returns an error if the key does not exist.
	Remove(key string) error
}

func ParseSeqIndex(indexStr string) (int, error) {
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, errors.Wrapf(err, "index into SequenceNode with non-integer %q key", indexStr)
	}
	if idx < 0 {
		return 0, errors.Errorf("index into SequenceNode with negative %q key", indexStr)
	}
	return idx, nil
}

var ErrIllegalDocumentAccess = fmt.Errorf("*goyamlDocumentContainer does not allow non-empty key access")
