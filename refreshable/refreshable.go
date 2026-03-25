// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"sync"
)

// A Refreshable is a generic container type for a volatile underlying value.
// It supports atomic access and user-provided callback "subscriptions" on updates.
type Refreshable[T any] interface {
	// Current returns the most recent value of this Refreshable.
	// If the value has not been initialized, returns T's zero value.
	Current() T

	// Subscribe calls the consumer function when Value updates until stop is closed.
	// The consumer must be relatively fast: Updatable.Set blocks until all subscribers have returned.
	// Expensive or error-prone responses to refreshed values should be asynchronous.
	// Updates considered no-ops by reflect.DeepEqual may be skipped.
	// When called, consumer is executed with the Current value.
	Subscribe(consumer func(T)) UnsubscribeFunc
}

// A Updatable is a Refreshable which supports setting the value with a user-provided value.
// When a utility returns a (non-Updatable) Refreshable, it implies that value updates are handled internally.
type Updatable[T any] interface {
	Refreshable[T]
	// Update updates the Refreshable with a new T.
	// It blocks until all subscribers have completed.
	Update(T)
}

// A Validated is capable of rejecting updates according to validation logic.
// Its Unvalidated method returns the most recent value to pass validation.
type Validated[T any] interface {
	// SubscribeValidated calls the consumer function when the validated value updates until stop is closed.
	// The consumer receives the latest value and its validation error (nil if valid).
	SubscribeValidated(consumer func(Validated[T])) UnsubscribeFunc
	// Unvalidated returns the most recent value to pass validation.
	Unvalidated() T
	// Validation returns the result of the most recent validation.
	// If the last value was valid, Validation returns the same value as Unvalidated and a nil error.
	// If the last value was invalid, Validation returns T's zero value and the error. Unvalidated returns the most recent valid value.
	Validation() (T, error)
}

// Ready extends Refreshable for asynchronous implementations which may not have a value when they are constructed.
// Callers should check that the Ready channel is closed before using the Current value.
type Ready[T any] interface {
	Refreshable[T]
	// ReadyC returns a channel which is closed after a value is successfully populated.
	ReadyC() <-chan struct{}
}

// UnsubscribeFunc removes a subscription from a refreshable's internal tracking and/or stops its update routine.
// It is safe to call multiple times.
type UnsubscribeFunc func()

// New returns a new Updatable that begins with the given value.
func New[T any](val T) Updatable[T] {
	return newDefault(val)
}

// Cached returns a new Refreshable that subscribes to the original Refreshable and caches its value.
// This is useful in combination with View to avoid recomputing an expensive mapped value
// each time it is retrieved. The returned refreshable is read-only (does not implement Update).
func Cached[T any](original Refreshable[T]) (Refreshable[T], UnsubscribeFunc) {
	out := newZero[T]()
	stop := original.Subscribe(out.Update)
	return newDerivedRefreshable(out, stop), stop
}

// CachedAuto is like Cached with automatic GC-based cleanup of the upstream subscription.
func CachedAuto[T any](original Refreshable[T]) Refreshable[T] {
	out, _ := Cached(original)
	return out
}

// View returns a Refreshable implementation that converts the original Refreshable value to a new value using mapFn.
// Current() and Subscribe() invoke mapFn as needed on the current value of the original Refreshable.
// Subscription callbacks are invoked with the mapped value each time the original value changes
// and the result is not cached nor compared for equality with the previous value, so functions
// subscribing to View refreshables are more likely to receive duplicate updates.
func View[T any, M any](original Refreshable[T], mapFn func(T) M) Refreshable[M] {
	return mapperRefreshable[T, M]{
		base:   original,
		mapper: mapFn,
	}
}

// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
// See Cached and View for more information.
func Map[T any, M any](original Refreshable[T], mapFn func(T) M) (Refreshable[M], UnsubscribeFunc) {
	return Cached(View(original, mapFn))
}

// MapAuto is like Map with automatic GC-based cleanup of the upstream subscription.
func MapAuto[T any, M any](original Refreshable[T], mapFn func(T) M) Refreshable[M] {
	out, _ := Map(original, mapFn)
	return out
}

// MapContext is like Map but unsubscribes when the context is cancelled.
func MapContext[T any, M any](ctx context.Context, original Refreshable[T], mapFn func(T) M) Refreshable[M] {
	out, stop := Map(original, mapFn)
	go func() {
		<-ctx.Done()
		stop()
	}()
	return out
}

// MapWithError is similar to Validate but allows for the function to return a mapping/mutation
// of the input object in addition to returning an error. The returned validRefreshable will contain the mapped value.
// The context is passed to the mapFn but is not considered in the subscription lifecycle.
// An error is returned if the current original value fails to map.
// The subscription and mapping continue until the UnsubscribeFunc is called or the Validated is garbage-collected.
func MapWithError[T any, M any](ctx context.Context, original Refreshable[T], mapFn func(context.Context, T) (M, error)) (Validated[M], UnsubscribeFunc, error) {
	v := newValidRefreshable[M]()
	intermediate := validatedFromRefreshable(original)
	stop := subscribeValidRefreshable(ctx, v, intermediate, mapFn)
	_, err := v.Validation()
	return newDerivedValidated(v, stop), stop, err
}

// MapWithErrorAuto is like MapWithError with automatic GC-based cleanup of the upstream subscription.
func MapWithErrorAuto[T any, M any](ctx context.Context, original Refreshable[T], mapFn func(context.Context, T) (M, error)) (Validated[M], error) {
	out, _, err := MapWithError(ctx, original, mapFn)
	return out, err
}

// Validate returns a new Refreshable that returns the latest original value accepted by the validatingFn.
// The context is passed to the validatingFn but is not considered in the subscription lifecycle.
// If the upstream value results in an error, it is reported by Validation().
// An error is returned if the current original value is invalid.
// The subscription and mapping continue until the UnsubscribeFunc is called or the Validated is garbage-collected.
func Validate[T any](ctx context.Context, original Refreshable[T], validatingFn func(context.Context, T) error) (Validated[T], UnsubscribeFunc, error) {
	return MapWithError(ctx, original, identity(validatingFn))
}

// ValidateAuto is like Validate with automatic GC-based cleanup of the upstream subscription.
func ValidateAuto[T any](ctx context.Context, original Refreshable[T], validatingFn func(context.Context, T) error) (Validated[T], error) {
	out, _, err := Validate(ctx, original, validatingFn)
	return out, err
}

// Merge returns a new Refreshable that combines the latest values of two Refreshables of different types using the mergeFn.
// The returned Refreshable is updated whenever either of the original Refreshables updates.
func Merge[T1 any, T2 any, R any](original1 Refreshable[T1], original2 Refreshable[T2], mergeFn func(T1, T2) R) (Refreshable[R], UnsubscribeFunc) {
	out := newZero[R]()
	// Subscriber callbacks push latest values into shared state so doUpdate
	// does not capture any Refreshable interface values. This avoids reference
	// cycles between derivedRefreshable wrappers and upstream subscriber lists
	// that would prevent runtime.AddCleanup from firing.
	var mu sync.Mutex
	val1 := original1.Current()
	val2 := original2.Current()
	doUpdate := func() {
		// mu must be held by caller.
		out.Update(mergeFn(val1, val2))
	}
	stop1 := original1.Subscribe(func(v T1) {
		mu.Lock()
		defer mu.Unlock()
		val1 = v
		doUpdate()
	})
	stop2 := original2.Subscribe(func(v T2) {
		mu.Lock()
		defer mu.Unlock()
		val2 = v
		doUpdate()
	})
	combined := func() {
		stop1()
		stop2()
	}
	return newDerivedRefreshable(out, combined), combined
}

// MergeAuto is like Merge with automatic GC-based cleanup of the upstream subscriptions.
func MergeAuto[T1 any, T2 any, R any](original1 Refreshable[T1], original2 Refreshable[T2], mergeFn func(T1, T2) R) Refreshable[R] {
	out, _ := Merge(original1, original2, mergeFn)
	return out
}

// Collect returns a new Refreshable that combines the latest values of multiple Refreshables into a slice.
// The returned Refreshable is updated whenever any of the original Refreshables updates.
func Collect[T any](list ...Refreshable[T]) (Refreshable[[]T], UnsubscribeFunc) {
	out, _, unsub := CollectMutable(list...)
	return out, unsub
}

// CollectAuto is like Collect with automatic GC-based cleanup of the upstream subscriptions.
func CollectAuto[T any](list ...Refreshable[T]) Refreshable[[]T] {
	out, _ := Collect(list...)
	return out
}

// AddFunc is a function that adds a new Refreshable to a collection.
type AddFunc[T any] func(Refreshable[T])

// CollectMutable returns a new Refreshable that combines the latest values of multiple Refreshables into a slice.
// The returned Refreshable is updated whenever any of the Refreshables updates.
// The add function allows adding new Refreshables to the collection after creation.
// The unsubscribe function removes subscriptions from all Refreshables in the collection.
func CollectMutable[T any](list ...Refreshable[T]) (Refreshable[[]T], AddFunc[T], UnsubscribeFunc) {
	out := newZero[[]T]()
	// Subscriber callbacks push latest values into a parallel slice so doUpdate
	// does not capture any Refreshable interface values. This avoids reference
	// cycles between derivedRefreshable wrappers and upstream subscriber lists
	// that would prevent runtime.AddCleanup from firing.
	var mu sync.Mutex
	vals := make([]T, len(list))
	for i, r := range list {
		vals[i] = r.Current()
	}
	stops := make([]UnsubscribeFunc, 0, len(list))
	doUpdate := func() {
		// mu must be held by caller.
		current := make([]T, len(vals))
		copy(current, vals)
		out.Update(current)
	}
	for i, r := range list {
		idx := i
		stops = append(stops, r.Subscribe(func(v T) {
			mu.Lock()
			defer mu.Unlock()
			vals[idx] = v
			doUpdate()
		}))
	}
	add := func(r Refreshable[T]) {
		mu.Lock()
		vals = append(vals, r.Current())
		idx := len(vals) - 1
		mu.Unlock()
		// Subscribe outside of lock since it immediately invokes the callback.
		stop := r.Subscribe(func(v T) {
			mu.Lock()
			defer mu.Unlock()
			vals[idx] = v
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
	return newDerivedRefreshable(out, combined), add, combined
}
