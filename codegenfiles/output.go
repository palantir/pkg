// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegenfiles

import (
	"maps"
	"slices"
)

// Output accumulates the files that a caller wants a project directory to contain. Paths may be relative
// to the project directory or absolute; they are normalized by Project.Plan.
//
// Adding the same path twice is recorded rather than rejected, so that Plan can report the collision
// against the normalized path with the project directory in hand. Output is not safe for concurrent use.
type Output struct {
	entries []outputEntry
}

type outputEntry struct {
	path    string
	content []byte
}

func NewOutput() *Output {
	return &Output{}
}

// Add records that path should contain content.
func (o *Output) Add(path string, content []byte) {
	o.entries = append(o.entries, outputEntry{path: path, content: content})
}

// AddAll records every entry in files. Entries are added in sorted order so that the collision reported
// for a path present in more than one call does not depend on map iteration order.
func (o *Output) AddAll(files map[string][]byte) {
	for _, path := range slices.Sorted(maps.Keys(files)) {
		o.Add(path, files[path])
	}
}

// Len returns the number of entries added, counting any duplicate paths separately.
func (o *Output) Len() int {
	return len(o.entries)
}
