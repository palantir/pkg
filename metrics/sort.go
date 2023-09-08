// Copyright (c) 2023 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !go1.21

package metrics

import (
	"sort"
)

// sortStrings is the default sort.Strings function.
// Unfortunately this forces the slice to escape to the heap.
// See https://github.com/golang/go/issues/17332
// Go 1.21's slices package does not have this issue.
var sortStrings = sort.Strings

// sortTags is the default sort.Sort function.
// Unfortunately this forces the slice to escape to the heap.
// See https://github.com/golang/go/issues/17332
// Go 1.21's slices package does not have this issue.
var sortTags = sort.Sort
