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

// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
func Map[T any, M any](original Refreshable[T], mapFn func(T) M) (Refreshable[M], UnsubscribeFunc) {
	out := newDefault(mapFn(original.Current()))
	stop := original.Subscribe(func(v T) {
		out.Update(mapFn(v))
	})
	return (*readOnlyRefreshable[M])(out), stop
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
	v, stop := newValidRefreshable(original, mapFn)
	_, err := v.Validation()
	return v, stop, err
}

// Validate returns a new Refreshable that returns the latest original value accepted by the validatingFn.
// If the upstream value results in an error, it is reported by Validation().
// An error is returned if the current original value is invalid.
func Validate[T any](original Refreshable[T], validatingFn func(T) error) (Validated[T], UnsubscribeFunc, error) {
	return MapWithError(original, identity(validatingFn))
}
