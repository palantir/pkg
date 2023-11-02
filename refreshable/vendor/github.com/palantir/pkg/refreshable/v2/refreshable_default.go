// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type defaultRefreshable[T any] struct {
	mux         sync.Mutex
	current     atomic.Value
	subscribers []*func(T)
}

func newDefault[T any](val T) *defaultRefreshable[T] {
	d := new(defaultRefreshable[T])
	d.current.Store(&val)
	return d
}

func newZero[T any]() *defaultRefreshable[T] {
	d := new(defaultRefreshable[T])
	var zero T
	d.current.Store(&zero)
	return d
}

// Update changes the value of the Refreshable, then blocks while subscribers are executed.
func (d *defaultRefreshable[T]) Update(val T) {
	d.mux.Lock()
	defer d.mux.Unlock()
	old := d.current.Swap(&val)
	if reflect.DeepEqual(*(old.(*T)), val) {
		return
	}
	for _, sub := range d.subscribers {
		(*sub)(val)
	}
}

func (d *defaultRefreshable[T]) Current() T {
	return *(d.current.Load().(*T))
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
			d.subscribers = append(d.subscribers[:matchIdx], d.subscribers[matchIdx+1:]...)
		}
	}

}

// readOnlyRefreshable aliases defaultRefreshable but hides the Update method so the type
// does not implement Updatable.
type readOnlyRefreshable[T any] defaultRefreshable[T]

func (d *readOnlyRefreshable[T]) Current() T {
	return (*defaultRefreshable[T])(d).Current()
}

func (d *readOnlyRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	return (*defaultRefreshable[T])(d).Subscribe(consumer)
}
