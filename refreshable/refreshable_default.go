// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
)

type defaultRefreshable[T any] struct {
	mux         sync.Mutex
	current     atomic.Pointer[T]
	subscribers []*func(T)
}

func newDefault[T any](val T) *defaultRefreshable[T] {
	d := new(defaultRefreshable[T])
	d.current.Store(&val)
	return d
}

func newZero[T any]() *defaultRefreshable[T] {
	return newDefault(*new(T))
}

// Update changes the value of the Refreshable, then blocks while subscribers are executed.
func (d *defaultRefreshable[T]) Update(val T) {
	d.mux.Lock()
	defer d.mux.Unlock()
	old := d.current.Swap(&val)
	if reflect.DeepEqual(*old, val) {
		return
	}
	for _, sub := range d.subscribers {
		(*sub)(val)
	}
}

func (d *defaultRefreshable[T]) Current() T {
	return *d.current.Load()
}

func (d *defaultRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	d.mux.Lock()
	defer d.mux.Unlock()

	consumerFnPtr := &consumer
	d.subscribers = append(d.subscribers, consumerFnPtr)
	consumer(d.Current())
	return d.unsubscribe(consumerFnPtr)
}

func (d *defaultRefreshable[T]) unsubscribe(consumerFnPtr *func(T)) UnsubscribeFunc {
	return func() {
		d.mux.Lock()
		defer d.mux.Unlock()

		matchIdx := -1
		for idx, currSub := range d.subscribers {
			if currSub == consumerFnPtr {
				matchIdx = idx
				break
			}
		}
		if matchIdx != -1 {
			d.subscribers = slices.Delete(d.subscribers, matchIdx, matchIdx+1)
		}
	}
}

// mapperRefreshable wraps an existing Refreshable and applies a mapping function to its values.
// Subscribe may be called repeatedly with the same value when the underlying value changes but the mapped value does not.
// mapperRefreshable does not implement Updatable because the mapped value may not be able to be converted back to the original type.
type mapperRefreshable[S, T any] struct {
	base   Refreshable[S]
	mapper func(S) T
}

func (d mapperRefreshable[S, T]) Current() T {
	return d.mapper(d.base.Current())
}

func (d mapperRefreshable[S, T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	// Extract mapper to avoid capturing d.base in the closure, which would
	// prevent GC cleanup of upstream derived wrappers in Map chains.
	mapper := d.mapper
	return d.base.Subscribe(func(value S) { consumer(mapper(value)) })
}
