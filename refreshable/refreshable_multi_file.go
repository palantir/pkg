// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
)

// NewMultiFileRefreshable creates a Refreshable that tracks the contents of multiple files.
// The input is a Refreshable of a set of file paths (map keys). The output is a Refreshable
// of a map from file path to file contents. When files are added to or removed from the input set,
// the corresponding file watchers are created or destroyed. Each file is read periodically
// using NewFileRefreshable.
func NewMultiFileRefreshable(ctx context.Context, paths Refreshable[map[string]struct{}]) Refreshable[map[string][]byte] {
	out := New(make(map[string][]byte))

	// Track active file refreshables and their unsubscribe functions
	fileRefreshables := make(map[string]Validated[[]byte])
	unsubscribers := make(map[string]UnsubscribeFunc)

	// Helper to rebuild output map from all file refreshables
	updateOutput := func() {
		result := make(map[string][]byte)
		for path, fr := range fileRefreshables {
			result[path] = fr.Current()
		}
		out.Update(result)
	}

	// Subscribe to paths changes to add/remove file refreshables
	paths.Subscribe(func(currentPaths map[string]struct{}) {
		// Remove files no longer in the set
		for path, unsub := range unsubscribers {
			if _, exists := currentPaths[path]; !exists {
				unsub()
				delete(unsubscribers, path)
				delete(fileRefreshables, path)
			}
		}

		// Add new files
		for path := range currentPaths {
			if _, exists := fileRefreshables[path]; !exists {
				fr := NewFileRefreshable(ctx, path)
				fileRefreshables[path] = fr
				unsubscribers[path] = fr.Subscribe(func([]byte) {
					updateOutput()
				})
			}
		}

		updateOutput()
	})

	return out
}
