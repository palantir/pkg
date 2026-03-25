// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
	"sync"
)

type validRefreshable[T any] struct {
	r Updatable[validRefreshableContainer[T]]
}

type validRefreshableContainer[T any] struct {
	unvalidated T
	validated   T
	lastErr     error
}

func (v *validRefreshable[T]) Unvalidated() T { return v.r.Current().unvalidated }

func (v *validRefreshable[T]) SubscribeValidated(consumer func(Validated[T])) UnsubscribeFunc {
	return v.r.Subscribe(func(_ validRefreshableContainer[T]) {
		consumer(v)
	})
}

// Validation returns the most recent upstream Refreshable and its validation result.
// If the error is nil, the validRefreshable is up-to-date with its original and the value
// is equal to that returned by Unvalidated().
func (v *validRefreshable[T]) Validation() (T, error) {
	c := v.r.Current()
	return c.validated, c.lastErr
}

func newValidRefreshable[M any]() *validRefreshable[M] {
	valid := &validRefreshable[M]{
		r: newDefault(validRefreshableContainer[M]{}),
	}
	return valid
}

func subscribeValidRefreshable[T, M any](ctx context.Context, v *validRefreshable[M], original Validated[T], mapFn func(context.Context, T) (M, error)) UnsubscribeFunc {
	return original.SubscribeValidated(func(val Validated[T]) {
		_, lastErr := val.Validation()
		valueT := val.Unvalidated()
		updateValidRefreshableWithParents(ctx, v, lastErr, func(ctx context.Context) (M, error) {
			return mapFn(ctx, valueT)
		})
	})
}

func updateValidRefreshable[M any](ctx context.Context, valid *validRefreshable[M], mapFn func(context.Context) (M, error)) {
	updateValidRefreshableWithParents(ctx, valid, nil, mapFn)
}

func updateValidRefreshableWithParents[M any](ctx context.Context, valid *validRefreshable[M], validatedParentError error, mapFn func(context.Context) (M, error)) {
	unvalidated := valid.r.Current().unvalidated
	validated, mapperErr := mapFn(ctx)
	err := getError(mapperErr, validatedParentError)
	if err == nil {
		unvalidated = validated
	} else {
		var zero M
		validated = zero
	}
	valid.r.Update(validRefreshableContainer[M]{
		unvalidated: unvalidated,
		validated:   validated,
		lastErr:     err,
	})
}

func getError(mapperErr, validatedParentError error) error {
	if mapperErr != nil && validatedParentError != nil {
		return errors.Join(mapperErr, validatedParentError)
	}
	if mapperErr != nil {
		return mapperErr
	}
	if validatedParentError != nil {
		return validatedParentError
	}
	return nil
}

// identity is a validating map function that returns its input argument type.
func identity[T any](validatingFn func(context.Context, T) error) func(ctx context.Context, i T) (T, error) {
	return func(ctx context.Context, i T) (T, error) { return i, validatingFn(ctx, i) }
}

func validatedFromRefreshable[M any](original Refreshable[M]) Validated[M] {
	valid := &validRefreshable[M]{
		r: newDefault(validRefreshableContainer[M]{}),
	}
	unsub := original.Subscribe(func(m M) {
		valid.r.Update(validRefreshableContainer[M]{
			unvalidated: m,
			validated:   m,
			lastErr:     nil,
		})
	})
	return newDerivedValidated(valid, unsub)
}

// MapFromValidated returns a new Refreshable by applying mapFn to the most recent
// value to pass validation from the original Validated. Invalid updates are ignored.
func MapFromValidated[T any, M any](original Validated[T], mapFn func(T) M) (Refreshable[M], UnsubscribeFunc) {
	out := newZero[M]()
	stop := original.SubscribeValidated(func(v Validated[T]) {
		out.Update(mapFn(v.Unvalidated()))
	})
	return newDerivedRefreshable(out, stop), stop
}

// MapFromValidatedAuto is like MapFromValidated with automatic GC-based cleanup of the upstream subscription.
func MapFromValidatedAuto[T any, M any](original Validated[T], mapFn func(T) M) Refreshable[M] {
	out, _ := MapFromValidated(original, mapFn)
	return out
}

// MapFromValidatedChecked is identical to MapFromValidated but first checks if the
// original Validated currently has a validation error and returns it if so.
func MapFromValidatedChecked[T any, M any](original Validated[T], mapFn func(T) M) (Refreshable[M], UnsubscribeFunc, error) {
	if _, err := original.Validation(); err != nil {
		return nil, nil, err
	}
	out, stop := MapFromValidated(original, mapFn)
	return out, stop, nil
}

// MapFromValidatedCheckedAuto is like MapFromValidatedChecked with automatic GC-based cleanup of the upstream subscription.
func MapFromValidatedCheckedAuto[T any, M any](original Validated[T], mapFn func(T) M) (Refreshable[M], error) {
	if _, err := original.Validation(); err != nil {
		return nil, err
	}
	return MapFromValidatedAuto(original, mapFn), nil
}

// MapValidated returns a new Validated based on the current one that handles updates based on the current Validated.
// The context is passed to the mapFn but is not considered in the subscription lifecycle.
// An error is returned if the current original value fails to map.
// The subscription and mapping continue until the UnsubscribeFunc is called or the Validated is garbage-collected.
func MapValidated[T any, M any](ctx context.Context, original Validated[T], mapFn func(context.Context, T) (M, error)) (Validated[M], UnsubscribeFunc, error) {
	v := newValidRefreshable[M]()
	stop := subscribeValidRefreshable(ctx, v, original, mapFn)
	_, err := v.Validation()
	return newDerivedValidated(v, stop), stop, err
}

// MapValidatedAuto is like MapValidated with automatic GC-based cleanup of the upstream subscription.
func MapValidatedAuto[T any, M any](ctx context.Context, original Validated[T], mapFn func(context.Context, T) (M, error)) (Validated[M], error) {
	out, _, err := MapValidated(ctx, original, mapFn)
	return out, err
}

// ValidatedAddFunc is a function that adds a new Validated to a collection.
type ValidatedAddFunc[T any] func(Validated[T])

// CollectValidated returns a new Validated that combines the latest values of multiple Validated refreshables into a slice.
// The returned Validated is updated whenever any of the original Validated refreshables updates.
func CollectValidated[T any](list ...Validated[T]) (Validated[[]T], UnsubscribeFunc) {
	out, _, unsub := CollectValidatedMutable(list...)
	return out, unsub
}

// CollectValidatedAuto is like CollectValidated with automatic GC-based cleanup of the upstream subscriptions.
func CollectValidatedAuto[T any](list ...Validated[T]) Validated[[]T] {
	out, _ := CollectValidated(list...)
	return out
}

// CollectValidatedMutable returns a new Validated that combines the latest values of multiple Validated refreshables into a slice.
// The returned Validated is updated whenever any of the Validated refreshables updates.
// The add function allows adding new Validated refreshables to the collection after creation.
// The unsubscribe function removes subscriptions from all Validated refreshables in the collection.
func CollectValidatedMutable[T any](list ...Validated[T]) (Validated[[]T], ValidatedAddFunc[T], UnsubscribeFunc) {
	out := newValidRefreshable[[]T]()
	// Subscriber callbacks push latest values into parallel slices so doUpdate
	// does not capture any Validated interface values. This avoids reference
	// cycles between derivedValidated wrappers and upstream subscriber lists
	// that would prevent runtime.AddCleanup from firing.
	var mu sync.Mutex
	vals := make([]T, len(list))
	valErrs := make([]error, len(list))
	for i, r := range list {
		vals[i] = r.Unvalidated()
		_, valErrs[i] = r.Validation()
	}
	stops := make([]UnsubscribeFunc, 0, len(list))
	doUpdate := func() {
		// mu must be held by caller.
		current := make([]T, len(vals))
		copy(current, vals)
		var errs []error
		for _, err := range valErrs {
			if err != nil {
				errs = append(errs, err)
			}
		}
		joined := errors.Join(errs...)
		if joined == nil {
			out.r.Update(validRefreshableContainer[[]T]{unvalidated: current, validated: current, lastErr: nil})
		} else {
			out.r.Update(validRefreshableContainer[[]T]{unvalidated: current, validated: nil, lastErr: joined})
		}
	}
	for i, r := range list {
		idx := i
		stops = append(stops, r.SubscribeValidated(func(v Validated[T]) {
			mu.Lock()
			defer mu.Unlock()
			vals[idx] = v.Unvalidated()
			_, valErrs[idx] = v.Validation()
			doUpdate()
		}))
	}
	add := func(r Validated[T]) {
		mu.Lock()
		_, validErr := r.Validation()
		vals = append(vals, r.Unvalidated())
		valErrs = append(valErrs, validErr)
		idx := len(vals) - 1
		mu.Unlock()
		// Subscribe outside of lock since it immediately invokes the callback.
		stop := r.SubscribeValidated(func(v Validated[T]) {
			mu.Lock()
			defer mu.Unlock()
			vals[idx] = v.Unvalidated()
			_, valErrs[idx] = v.Validation()
			doUpdate()
		})
		mu.Lock()
		stops = append(stops, stop)
		mu.Unlock()
	}
	combined := func() {
		mu.Lock()
		defer mu.Unlock()
		for _, stop := range stops {
			stop()
		}
	}
	return newDerivedValidated(out, combined), add, combined
}

// MergeValidated returns a new Validated that combines the latest values of two Validated refreshables using the mergeFn.
func MergeValidated[T1 any, T2 any, R any](original1 Validated[T1], original2 Validated[T2], mergeFn func(T1, T2) R) (Validated[R], UnsubscribeFunc) {
	out := newValidRefreshable[R]()
	// Subscriber callbacks push latest values into shared state so doUpdate
	// does not capture any Validated interface values. This avoids reference
	// cycles between derivedValidated wrappers and upstream subscriber lists
	// that would prevent runtime.AddCleanup from firing.
	var mu sync.Mutex
	val1 := original1.Unvalidated()
	val2 := original2.Unvalidated()
	_, err1 := original1.Validation()
	_, err2 := original2.Validation()
	doUpdate := func() {
		// mu must be held by caller.
		merged := mergeFn(val1, val2)
		err := getError(err1, err2)
		if err == nil {
			out.r.Update(validRefreshableContainer[R]{unvalidated: merged, validated: merged, lastErr: nil})
		} else {
			var zero R
			out.r.Update(validRefreshableContainer[R]{unvalidated: merged, validated: zero, lastErr: err})
		}
	}
	stop1 := original1.SubscribeValidated(func(v Validated[T1]) {
		mu.Lock()
		defer mu.Unlock()
		val1 = v.Unvalidated()
		_, err1 = v.Validation()
		doUpdate()
	})
	stop2 := original2.SubscribeValidated(func(v Validated[T2]) {
		mu.Lock()
		defer mu.Unlock()
		val2 = v.Unvalidated()
		_, err2 = v.Validation()
		doUpdate()
	})
	combined := func() {
		stop1()
		stop2()
	}
	return newDerivedValidated(out, combined), combined
}

// MergeValidatedAuto is like MergeValidated with automatic GC-based cleanup of the upstream subscriptions.
func MergeValidatedAuto[T1 any, T2 any, R any](original1 Validated[T1], original2 Validated[T2], mergeFn func(T1, T2) R) Validated[R] {
	out, _ := MergeValidated(original1, original2, mergeFn)
	return out
}

// MergeValidatedAndRefreshable returns a new Validated that combines the latest values of a Validated
// and a plain Refreshable using the mergeFn. The Refreshable is wrapped with an always-valid Validate
// so that only errors from the Validated source propagate. The returned Validated is updated whenever
// either source updates.
// The context is used internally to wrap the Refreshable as a Validated but does not affect the subscription lifecycle.
// The subscription and mapping continue until the UnsubscribeFunc is called or the Validated is garbage-collected.
func MergeValidatedAndRefreshable[T1 any, T2 any, R any](
	ctx context.Context,
	original1 Validated[T1],
	refreshable1 Refreshable[T2],
	mergeFn func(T1, T2) R,
) (Validated[R], UnsubscribeFunc) {
	original2, _, _ := Validate(ctx, refreshable1, func(ctx context.Context, i T2) error {
		return nil
	})
	return MergeValidated(original1, original2, mergeFn)
}

// MergeValidatedAndRefreshableAuto is like MergeValidatedAndRefreshable with automatic GC-based cleanup of the upstream subscriptions.
func MergeValidatedAndRefreshableAuto[T1 any, T2 any, R any](
	ctx context.Context,
	original1 Validated[T1],
	refreshable1 Refreshable[T2],
	mergeFn func(T1, T2) R,
) Validated[R] {
	out, _ := MergeValidatedAndRefreshable(ctx, original1, refreshable1, mergeFn)
	return out
}
