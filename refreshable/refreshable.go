// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

type Refreshable[T any] interface {
	// Current returns the most recent value of this Refreshable.
	// If the value has not been initialized, returns T's zero value.
	Current() T

	// Subscribe subscribes to changes of this Supplier.
	// The provided function is called with the value of Get() whenever the value changes.
	Subscribe(consumer func(T)) (unsubscribe func())
}

// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
func Map[T any, M any](t Refreshable[T], mapFn func(T) M) (m Refreshable[M], unsubscribe func()) {
	out := New(mapFn(t.Current()))
	unsubscribe = t.Subscribe(func(v T) {
		out.Update(mapFn(v))
	})
	return out, unsubscribe
}
