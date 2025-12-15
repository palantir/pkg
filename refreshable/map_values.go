// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
)

// MapValues creates a Validated Refreshable by applying a mapper function to each entry in a map.
// For each key-value pair in the input map, the mapper function creates a Validated[R] refreshable.
// The output is a Validated[map[K]R] that aggregates all mapped values.
//
// When keys are added to the input map, new refreshables are created via the mapper function.
// When keys are removed, their corresponding refreshables are unsubscribed.
// When any individual mapped refreshable updates, the output map is rebuilt.
//
// Current() returns a map containing only keys whose mapped refreshables are valid.
// Validation() returns the map and a joined error of all validation failures.
func MapValues[K comparable, V, R any](
	ctx context.Context,
	refreshableMap Refreshable[map[K]V],
	mapperFn func(context.Context, K, V) Validated[R],
) Validated[map[K]R] {
	out := newValidRefreshable[map[K]R]()
	mappedRefreshables := make(map[K]Validated[R])
	unsubscribers := make(map[K]UnsubscribeFunc)

	updateOutput := func() {
		result := make(map[K]R)
		var errs []error
		for key, refreshable := range mappedRefreshables {
			value, err := refreshable.Validation()
			if err != nil {
				errs = append(errs, err)
			} else {
				result[key] = value
			}
		}
		var joinedErr error
		if len(errs) > 0 {
			joinedErr = errors.Join(errs...)
		}
		out.r.Update(validRefreshableContainer[map[K]R]{
			validated:   result,
			unvalidated: result,
			lastErr:     joinedErr,
		})
	}

	refreshableMap.Subscribe(func(currentMap map[K]V) {
		// Remove keys no longer in the map
		for key, unsub := range unsubscribers {
			if _, exists := currentMap[key]; !exists {
				unsub()
				delete(unsubscribers, key)
				delete(mappedRefreshables, key)
			}
		}
		// Add new keys
		for key, value := range currentMap {
			if _, exists := mappedRefreshables[key]; !exists {
				mapped := mapperFn(ctx, key, value)
				mappedRefreshables[key] = mapped
				unsubscribers[key] = mapped.Subscribe(func(R) {
					updateOutput()
				})
			}
		}
		updateOutput()
	})

	return out
}
