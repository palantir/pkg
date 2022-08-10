// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"time"
)

// A Refreshable is a generic container type for a volatile underlying value.
// It supports atomic access and user-provided callback "subscriptions" on updates.
type Refreshable[T any] interface {
	// Current returns the most recent value of this Refreshable.
	// If the value has not been initialized, returns T's zero value.
	Current() T

	// Subscribe calls the consumer function when Value updates until stop is closed.
	// The consumer should be relatively fast: Updatable.Set blocks until all subscribers have returned.
	// Updates considered no-ops by reflect.DeepEqual may be skipped.
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
	// If the last value was invalid, it and the validation error are returned. Current returns the newest valid value.
	Validation() (T, error)
}

// Ready extends Refreshable for asynchronous implementations which may not have a value when they are constructed.
// Callers should check that the Ready channel is closed before using the Current value.
type Ready[T any] interface {
	Refreshable[T]
	// ReadyC returns a channel which is closed after a value is successfully populated.
	ReadyC() <-chan struct{}
}

type UnsubscribeFunc func()

func New[T any](val T) *DefaultRefreshable[T] {
	d := new(DefaultRefreshable[T])
	d.current.Store(&val)
	return d
}

// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
func Map[T any, M any](t Refreshable[T], mapFn func(T) M) (Refreshable[M], UnsubscribeFunc) {
	out := New(mapFn(t.Current()))
	unsubscribe := t.Subscribe(func(v T) {
		out.Update(mapFn(v))
	})
	return out, unsubscribe
}

func SubscribeWithCurrent[T any](r Refreshable[T], consumer func(T)) UnsubscribeFunc {
	unsubscribe := r.Subscribe(consumer)
	consumer(r.Current())
	return unsubscribe
}

// UpdateFromChannel populates an Updatable with the values channel.
// If an element is already available, the returned Value is guaranteed to be populated.
// The channel should be closed when no longer used to avoid leaking resources.
func UpdateFromChannel[T any](in Updatable[T], values <-chan T) Ready[T] {
	out := newReady(in)
	select {
	case initial, ok := <-values:
		if !ok {
			return out // channel already closed
		}
		out.Update(initial)
	default:
	}

	go func() {
		for value := range values {
			out.Update(value)
		}
	}()

	return out
}

// UpdateFromTickerFunc returns a Refreshable populated by the result of the provider called each interval.
// If the providers bool return is false, the value is ignored.
func UpdateFromTickerFunc[T any](in Updatable[T], interval time.Duration, provider func() (T, bool)) (Ready[T], UnsubscribeFunc) {
	out := newReady(in)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if value, ok := provider(); ok {
				out.Update(value)
			}
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, UnsubscribeFunc(cancel)
}

// Wait waits until the Ready has a current value or the context expires.
func Wait[T any](ctx context.Context, ready Ready[T]) (T, bool) {
	select {
	case <-ready.ReadyC():
		return ready.Current(), true
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

type ready[T any] struct {
	in     Updatable[T]
	readyC <-chan struct{}
	cancel context.CancelFunc
}

func newReady[T any](in Updatable[T]) *ready[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &ready[T]{in: in, readyC: ctx.Done(), cancel: cancel}
}

func (r *ready[T]) Current() T {
	return r.in.Current()
}

func (r *ready[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	return r.in.Subscribe(consumer)
}

func (r *ready[T]) ReadyC() <-chan struct{} {
	return r.readyC
}

func (r *ready[T]) Update(val T) {
	r.cancel()
	r.in.Update(val)
}
