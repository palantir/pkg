// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
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

// A Validated is a Refreshable capable of rejecting updates according to validation logic.
// Its Current method returns the most recent value to pass validation.
type Validated[T any] interface {
	Refreshable[T]
	// Validation returns the result of the most recent validation.
	// If the last value was valid, Validation returns the same value as Current and a nil error.
	// If the last value was invalid, it and the error are returned. Current returns the most recent valid value.
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
	return out.readOnly(), stop
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
// An error is returned if the current original value fails to map.
func MapWithError[T any, M any](original Refreshable[T], mapFn func(T) (M, error)) (Validated[M], UnsubscribeFunc, error) {
	v := newValidRefreshable[M]()
	stop := subscribeValidRefreshable(v, original, mapFn)
	_, err := v.Validation()
	return v, stop, err
}

// Validate returns a new Refreshable that returns the latest original value accepted by the validatingFn.
// If the upstream value results in an error, it is reported by Validation().
// An error is returned if the current original value is invalid.
func Validate[T any](original Refreshable[T], validatingFn func(T) error) (Validated[T], UnsubscribeFunc, error) {
	return MapWithError(original, identity(validatingFn))
}

// Merge returns a new Refreshable that combines the latest values of two Refreshables of different types using the mergeFn.
// The returned Refreshable is updated whenever either of the original Refreshables updates.
// The unsubscribe function removes subscriptions from both original Refreshables.
func Merge[T1 any, T2 any, R any](original1 Refreshable[T1], original2 Refreshable[T2], mergeFn func(T1, T2) R) (Refreshable[R], UnsubscribeFunc) {
	out := newZero[R]()
	doUpdate := func() {
		out.Update(mergeFn(original1.Current(), original2.Current()))
	}
	stop1 := original1.Subscribe(func(T1) { doUpdate() })
	stop2 := original2.Subscribe(func(T2) { doUpdate() })
	return out.readOnly(), func() {
		stop1()
		stop2()
	}
}

// Collect returns a new Refreshable that combines the latest values of multiple Refreshables into a slice.
// The returned Refreshable is updated whenever any of the original Refreshables updates.
// The unsubscribe function removes subscriptions from all original Refreshables.
func Collect[T any](list ...Refreshable[T]) (Refreshable[[]T], UnsubscribeFunc) {
	out := newZero[[]T]()
	doUpdate := func() {
		current := make([]T, len(list))
		for i := range list {
			current[i] = list[i].Current()
		}
		out.Update(current)
	}
	stops := make([]UnsubscribeFunc, len(list))
	for i := range list {
		stops[i] = list[i].Subscribe(func(T) { doUpdate() })
	}
	return out.readOnly(), func() {
		for _, stop := range stops {
			stop()
		}
	}
}
