// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"time"
)

// NewFromChannel populates an Updatable with the values channel.
// If an element is already available, the returned Value is guaranteed to be populated.
// The channel should be closed when no longer used to avoid leaking resources.
func NewFromChannel[T any](values <-chan T) Ready[T] {
	out := newReady[T]()
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

// NewFromTickerFunc returns a Ready Refreshable populated by the result of the provider called each interval.
// If the providers bool return is false, the value is ignored.
// The result's ReadyC channel is closed when a new value is populated.
// The refreshable will stop updating when the provided context is cancelled or the returned UnsubscribeFunc func is called.
func NewFromTickerFunc[T any](ctx context.Context, interval time.Duration, provider func(ctx context.Context) (T, bool)) (Ready[T], UnsubscribeFunc) {
	out := newReady[T]()
	ctx, cancel := context.WithCancel(ctx)
	values := make(chan T)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(values)
		for {
			if value, ok := provider(ctx); ok {
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

// ready is an Updatable which exposes a channel that is closed when a value is first available.
// Current returns the zero value before Update is called, marking the value ready.
type ready[T any] struct {
	in     Updatable[T]
	readyC <-chan struct{}
	cancel context.CancelFunc
}

func newReady[T any]() *ready[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &ready[T]{
		in:     newZero[T](),
		readyC: ctx.Done(),
		cancel: cancel,
	}
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
	r.in.Update(val)
	r.cancel()
}
