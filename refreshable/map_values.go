// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
	"sync"
)

// MapValues creates a Validated[map[K]R] by applying mapperFn to each entry of a map Refreshable.
// When keys are added, mapperFn is called with a per-key context that is cancelled on removal.
// When any per-key refreshable updates, the output map is rebuilt with aggregated validation errors.
//
// Use this instead of Map on a map[K]V when you need per-key refreshables that update independently.
func MapValues[K comparable, V, R any](
	ctx context.Context,
	refreshableMap Refreshable[map[K]V],
	mapperFn func(context.Context, K, V) Validated[R],
) Validated[map[K]R] {
	type perKey struct {
		val    Validated[R]
		unsub  UnsubscribeFunc
		cancel context.CancelFunc
	}

	out := newValidRefreshable[map[K]R]()
	var mu sync.Mutex
	keys := make(map[K]*perKey)

	updateOutput := func() {
		mu.Lock()
		result := make(map[K]R)
		var errs []error
		for k, pk := range keys {
			result[k] = pk.val.Unvalidated()
			if _, err := pk.val.Validation(); err != nil {
				errs = append(errs, err)
			}
		}
		mu.Unlock()
		joined := errors.Join(errs...)
		if joined == nil {
			out.r.Update(validRefreshableContainer[map[K]R]{unvalidated: result, validated: result})
		} else {
			out.r.Update(validRefreshableContainer[map[K]R]{unvalidated: result, lastErr: joined})
		}
	}

	unsub := refreshableMap.Subscribe(func(currentMap map[K]V) {
		mu.Lock()
		var removed []*perKey
		for k := range keys {
			if _, ok := currentMap[k]; !ok {
				removed = append(removed, keys[k])
				delete(keys, k)
			}
		}
		type newEntry struct {
			key K
			pk  *perKey
		}
		var added []newEntry
		for k, v := range currentMap {
			if _, ok := keys[k]; !ok {
				keyCtx, keyCancel := context.WithCancel(ctx)
				pk := &perKey{val: mapperFn(keyCtx, k, v), cancel: keyCancel}
				keys[k] = pk
				added = append(added, newEntry{key: k, pk: pk})
			}
		}
		mu.Unlock()

		// Run outside lock: unsub/cancel may trigger callbacks that call updateOutput,
		// and SubscribeValidated immediately invokes its callback.
		for _, pk := range removed {
			pk.unsub()
			pk.cancel()
		}
		for _, entry := range added {
			stop := entry.pk.val.SubscribeValidated(func(Validated[R]) { updateOutput() })
			mu.Lock()
			entry.pk.unsub = stop
			mu.Unlock()
		}
		updateOutput()
	})

	combinedUnsub := func() {
		unsub()
		mu.Lock()
		all := make([]*perKey, 0, len(keys))
		for _, pk := range keys {
			all = append(all, pk)
		}
		clear(keys)
		mu.Unlock()
		for _, pk := range all {
			pk.unsub()
			pk.cancel()
		}
	}

	return newDerivedValidated(out, combinedUnsub)
}
