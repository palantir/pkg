// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"errors"
)

type validRefreshable[T any] struct {
	r Updatable[validRefreshableContainer[T]]
}

type validRefreshableContainer[T any] struct {
	validated   T
	unvalidated T
	lastErr     error
}

func (v *validRefreshable[T]) LastCurrent() T { return v.r.Current().validated }

func (v *validRefreshable[T]) SubscribeValidated(consumer func(T, error)) UnsubscribeFunc {
	return v.r.Subscribe(func(val validRefreshableContainer[T]) {
		consumer(val.validated, val.lastErr)
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

func subscribeValidRefreshable[T, M any](v *validRefreshable[M], original Validated[T], mapFn func(T) (M, error)) UnsubscribeFunc {
	return original.SubscribeValidated(func(valueT T, lastErr error) {
		updateValidRefreshableWithParents(v, valueT, lastErr, mapFn)
	})
}

func updateValidRefreshable[T any, M any](valid *validRefreshable[M], value T, mapFn func(T) (M, error)) {
	updateValidRefreshableWithParents(valid, value, nil, mapFn)
}

func updateValidRefreshableWithParents[T any, M any](valid *validRefreshable[M], value T, validatedParentError error, mapFn func(T) (M, error)) {
	validated := valid.r.Current().validated
	unvalidated, mapperErr := mapFn(value)
	err := getError(mapperErr, validatedParentError)
	if err == nil {
		validated = unvalidated
	}
	valid.r.Update(validRefreshableContainer[M]{
		validated:   validated,
		unvalidated: unvalidated,
		lastErr:     err,
	})
}

func getError(mapperErr, validatedParentError error) error {
	if mapperErr == nil && validatedParentError == nil {
		return nil
	}
	if mapperErr != nil && validatedParentError != nil {
		return errors.Join(mapperErr, validatedParentError)
	}
	if mapperErr != nil {
		return mapperErr
	}
	return validatedParentError
}

// identity is a validating map function that returns its input argument type.
func identity[T any](validatingFn func(T) error) func(i T) (T, error) {
	return func(i T) (T, error) { return i, validatingFn(i) }
}

func ValidatedFromRefreshable[M any](original Refreshable[M]) Validated[M] {
	valid := &validRefreshable[M]{
		r: newDefault(validRefreshableContainer[M]{}),
	}
	original.Subscribe(func(m M) {
		valid.r.Update(validRefreshableContainer[M]{
			validated:   m,
			unvalidated: m,
			lastErr:     nil,
		})
	})
	return valid
}

func MapValidated[T any, M any](original Validated[T], mapFn func(T) (M, error)) (Validated[M], UnsubscribeFunc, error) {
	v := newValidRefreshable[M]()
	stop := subscribeValidRefreshable(v, original, mapFn)
	_, err := v.Validation()
	return v, stop, err
}

func MergeValidated[T1 any, T2 any, R any](original1 Validated[T1], original2 Validated[T2], mergeFn func(T1, T2) R) (Validated[R], UnsubscribeFunc) {
	out := newValidRefreshable[R]()
	doUpdate := func() {
		merged := mergeFn(original1.LastCurrent(), original2.LastCurrent())
		_, err1 := original1.Validation()
		_, err2 := original2.Validation()
		err := getError(err1, err2)
		if err == nil {
			out.r.Update(validRefreshableContainer[R]{validated: merged, unvalidated: merged, lastErr: nil})
		} else {
			out.r.Update(validRefreshableContainer[R]{validated: out.r.Current().validated, unvalidated: merged, lastErr: err})
		}
	}
	stop1 := original1.SubscribeValidated(func(T1, error) { doUpdate() })
	stop2 := original2.SubscribeValidated(func(T2, error) { doUpdate() })
	return out, func() {
		stop1()
		stop2()
	}
}
