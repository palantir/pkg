// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

type validRefreshable[T any] struct {
	r Updatable[validRefreshableContainer[T]]
}

type validRefreshableContainer[T any] struct {
	validated   T
	unvalidated T
	lastErr     error
}

func (v *validRefreshable[T]) Current() T { return v.r.Current().validated }

func (v *validRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	return v.r.Subscribe(func(val validRefreshableContainer[T]) {
		consumer(val.validated)
	})
}

// Validation returns the most recent upstream Refreshable and its validation result.
// If the error is nil, the validRefreshable is up-to-date with its original and the value
// is equal to that returned by Current().
func (v *validRefreshable[T]) Validation() (T, error) {
	c := v.r.Current()
	return c.unvalidated, c.lastErr
}

func newValidRefreshable[M any]() *validRefreshable[M] {
	valid := &validRefreshable[M]{
		r: newDefault(validRefreshableContainer[M]{}),
	}
	return valid
}

func subscribeValidRefreshable[T, M any](v *validRefreshable[M], original Refreshable[T], mapFn func(T) (M, error)) UnsubscribeFunc {
	return original.Subscribe(func(valueT T) {
		updateValidRefreshable(v, valueT, mapFn)
	})
}

func updateValidRefreshable[T any, M any](valid *validRefreshable[M], value T, mapFn func(T) (M, error)) {
	validated := valid.r.Current().validated
	unvalidated, err := mapFn(value)
	if err == nil {
		validated = unvalidated
	}
	valid.r.Update(validRefreshableContainer[M]{
		validated:   validated,
		unvalidated: unvalidated,
		lastErr:     err,
	})
}
