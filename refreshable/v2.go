// Copyright (c) 2023 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	refreshablev2 "github.com/palantir/pkg/refreshable/v2"
)

// ToV2 converts from a v1 Refreshable created by this package to v2 supporting type safety via generics.
// If v1's value is not of type T, the function panics.
func ToV2[T any](v1 Refreshable) refreshablev2.Refreshable[T] {
	v2 := refreshablev2.New[T](v1.Current().(T))
	v1.Subscribe(func(i interface{}) {
		v2.Update(i.(T))
	})
	return v2
}

// FromV2 converts from a v1 Refreshable created by this package to v2 supporting type safety via generics.
func FromV2[T any](v2 refreshablev2.Refreshable[T]) Refreshable {
	v1 := NewDefaultRefreshable(v2.Current())
	v2.Subscribe(func(i T) {
		if err := v1.Update(i); err != nil {
			panic(err)
		}
	})
	return v1
}
