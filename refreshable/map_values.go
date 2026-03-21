// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
	"sync"
)

// MapValues creates a Validated Refreshable by applying a mapper function to each entry in a map.
// For each key-value pair in the input map, the mapper function creates a Validated[R] refreshable.
// The output is a Validated[map[K]R] that aggregates all mapped values.
//
// When keys are added to the input map, new refreshables are created via the mapper function.
// A per-key derived context is passed to mapperFn; this context is cancelled when the key is removed,
// allowing mapperFn implementations that start goroutines (e.g. NewFileRefreshable) to clean up.
// When keys are removed, their corresponding refreshables are unsubscribed and contexts cancelled.
// When any individual mapped refreshable updates, the output map is rebuilt.
//
// Unvalidated() returns a map containing the last valid value for each key.
// Validation() returns the map and a joined error of all validation failures.
//
// This should be used instead of just calling Map on a map[K]V when you need to interject an additional refreshable that can be updated independently
func MapValues[K comparable, V, R any](
	ctx context.Context,
	refreshableMap Refreshable[map[K]V],
	mapperFn func(context.Context, K, V) Validated[R],
) Validated[map[K]R] {
	out := newValidRefreshable[map[K]R]()
	var mu sync.Mutex
	mappedRefreshables := make(map[K]Validated[R])
	unsubscribers := make(map[K]UnsubscribeFunc)
	cancels := make(map[K]context.CancelFunc)

	updateOutput := func() {
		mu.Lock()
		result := make(map[K]R)
		var errs []error
		for key, refreshable := range mappedRefreshables {
			result[key] = refreshable.Unvalidated()
			if _, err := refreshable.Validation(); err != nil {
				errs = append(errs, err)
			}
		}
		mu.Unlock()
		joined := errors.Join(errs...)
		if joined == nil {
			out.r.Update(validRefreshableContainer[map[K]R]{unvalidated: result, validated: result, lastErr: nil})
		} else {
			out.r.Update(validRefreshableContainer[map[K]R]{unvalidated: result, validated: nil, lastErr: joined})
		}
	}

	unsub := refreshableMap.Subscribe(func(currentMap map[K]V) {
		mu.Lock()
		// Collect keys to remove
		var removedUnsubs []UnsubscribeFunc
		var removedCancels []context.CancelFunc
		for key := range unsubscribers {
			if _, exists := currentMap[key]; !exists {
				removedUnsubs = append(removedUnsubs, unsubscribers[key])
				removedCancels = append(removedCancels, cancels[key])
				delete(unsubscribers, key)
				delete(mappedRefreshables, key)
				delete(cancels, key)
			}
		}
		// Create refreshables for new keys
		type newEntry struct {
			key    K
			mapped Validated[R]
		}
		var newEntries []newEntry
		for key, value := range currentMap {
			if _, exists := mappedRefreshables[key]; !exists {
				keyCtx, keyCancel := context.WithCancel(ctx)
				mapped := mapperFn(keyCtx, key, value)
				mappedRefreshables[key] = mapped
				cancels[key] = keyCancel
				newEntries = append(newEntries, newEntry{key: key, mapped: mapped})
			}
		}
		mu.Unlock()

		// Unsubscribe and cancel outside lock to avoid deadlock with
		// per-key subscription callbacks that call updateOutput.
		for _, unsub := range removedUnsubs {
			unsub()
		}
		for _, cancel := range removedCancels {
			cancel()
		}

		// Subscribe outside lock since SubscribeValidated immediately invokes the callback.
		for _, entry := range newEntries {
			stop := entry.mapped.SubscribeValidated(func(Validated[R]) {
				updateOutput()
			})
			mu.Lock()
			unsubscribers[entry.key] = stop
			mu.Unlock()
		}
		updateOutput()
	})

	combinedUnsub := func() {
		unsub()
		mu.Lock()
		stops := make([]UnsubscribeFunc, 0, len(unsubscribers))
		cancelFns := make([]context.CancelFunc, 0, len(cancels))
		for _, stop := range unsubscribers {
			stops = append(stops, stop)
		}
		for _, cancel := range cancels {
			cancelFns = append(cancelFns, cancel)
		}
		clear(unsubscribers)
		clear(mappedRefreshables)
		clear(cancels)
		mu.Unlock()
		for _, stop := range stops {
			stop()
		}
		for _, cancel := range cancelFns {
			cancel()
		}
	}

	d := newDerivedValidated(out, combinedUnsub)
	d.refs = append(d.refs, refreshableMap)
	return d
}
