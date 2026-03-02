// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

type validRefreshable[T any] struct {
	r Updatable[ValidRefreshableContainer[T]]
}

type ValidRefreshableContainer[T any] struct {
	Validated   T
	Unvalidated T
	LastErr     error
}

func (v *validRefreshable[T]) Current() T { return v.r.Current().Validated }

func (v *validRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	return v.r.Subscribe(func(val ValidRefreshableContainer[T]) {
		consumer(val.Validated)
	})
}

// Validation returns the most recent upstream Refreshable and its validation result.
// If the error is nil, the validRefreshable is up-to-date with its original and the value
// is equal to that returned by Current().
func (v *validRefreshable[T]) Validation() (T, error) {
	c := v.r.Current()
	return c.Unvalidated, c.LastErr
}

func newValidRefreshable[M any]() *validRefreshable[M] {
	valid := &validRefreshable[M]{
		r: newDefault(ValidRefreshableContainer[M]{}),
	}
	return valid
}

func subscribeValidRefreshable[T, M any](v *validRefreshable[M], original Refreshable[T], mapFn func(T) (M, error)) UnsubscribeFunc {
	return original.Subscribe(func(valueT T) {
		updateValidRefreshable(v, func() (M, error) {
			return mapFn(valueT)
		})
	})
}

func updateValidRefreshable[M any](valid *validRefreshable[M], mapFn func() (M, error)) {
	validated := valid.r.Current().Validated
	unvalidated, err := mapFn()
	if err == nil {
		validated = unvalidated
	}
	valid.r.Update(ValidRefreshableContainer[M]{
		Validated:   validated,
		Unvalidated: unvalidated,
		LastErr:     err,
	})
}

// identity is a validating map function that returns its input argument type.
func identity[T any](validatingFn func(T) error) func(i T) (T, error) {
	return func(i T) (T, error) { return i, validatingFn(i) }
}
