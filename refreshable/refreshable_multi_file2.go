// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
)

func TransitiveMapper[K comparable, V, R any](
	ctx context.Context,
	refreshableMap Refreshable[map[K]V],
	mapperFn func(context.Context, K, V) Validated[R],
) Validated[map[K]R] {
	out := newValidRefreshable[map[K]R]()
	// Track active file refreshables and their unsubscribe functions
	fileRefreshables := make(map[K]Validated[R])
	unsubscribers := make(map[K]UnsubscribeFunc)

	// Helper to rebuild output map from all file refreshables
	updateOutput := func() {
		result2 := make(map[K]R)
		var errs []error
		for path, fr := range fileRefreshables {
			content, err := fr.Validation()
			if err != nil {
				errs = append(errs, err)
			} else {
				result2[path] = content
			}
		}
		var joinedErr error
		if len(errs) > 0 {
			joinedErr = errors.Join(errs...)
		}
		out.r.Update(validRefreshableContainer[map[K]R]{
			validated:   result2,
			unvalidated: result2,
			lastErr:     joinedErr,
		})

	}

	refreshableMap.Subscribe(func(currentPaths map[K]V) {
		// Remove files no longer in the set
		for path, unsub := range unsubscribers {
			if _, exists := currentPaths[path]; !exists {
				unsub()
				delete(unsubscribers, path)
				delete(fileRefreshables, path)
			}
		}
		// Add new files
		for path, value := range currentPaths {
			if _, exists := fileRefreshables[path]; !exists {
				validatedR := mapperFn(ctx, path, value)
				fileRefreshables[path] = validatedR
				unsubscribers[path] = validatedR.Subscribe(func(R) {
					updateOutput()
				})
			}
		}
		updateOutput()
	})

	return out
}
