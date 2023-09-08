// Copyright (c) 2023 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.21

package metrics

import (
	"slices"
)

// sortStrings is the default slices.Sort function which does not force allocation like sort.Strings.
var sortStrings = slices.Sort[[]string]

// sortTags uses slices.SortFunc which does not force allocation like sort.Sort.
func sortTags(tags Tags) {
	slices.SortFunc(tags, compareTags)
}

func compareTags(a, b Tag) int {
	switch {
	case a.key < b.key:
		return -1
	case a.key > b.key:
		return 1
	case a.value < b.value:
		return -1
	case a.value > b.value:
		return 1
	default:
		return 0
	}
}
