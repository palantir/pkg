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
	return newDefault(*new(T))
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

func (d *defaultRefreshable[T]) readOnly() *readOnlyRefreshable[T] {
	return (*readOnlyRefreshable[T])(d)
}

type container2[T1, T2 any] struct {
	V1 T1
	V2 T2
}

type defaultRefreshable2[T1, T2 any] struct {
	mux         sync.Mutex
	current     atomic.Value
	subscribers []*func(T1, T2)
}

func newDefault2[T1, T2 any](val1 T1, val2 T2) *defaultRefreshable2[T1, T2] {
	d := new(defaultRefreshable2[T1, T2])
	d.current.Store(&container2[T1, T2]{val1, val2})
	return d
}

func newZero2[T1, T2 any]() *defaultRefreshable2[T1, T2] {
	return newDefault2(*new(T1), *new(T2))
}

// Update changes the value of the Refreshable, then blocks while subscribers are executed.
func (d *defaultRefreshable2[T1, T2]) Update(val1 T1, val2 T2) {
	d.mux.Lock()
	defer d.mux.Unlock()
	val := container2[T1, T2]{val1, val2}
	old := d.current.Swap(&val)
	if reflect.DeepEqual(*(old.(*container2[T1, T2])), val) {
		return
	}
	for _, sub := range d.subscribers {
		(*sub)(val1, val2)
	}
}

func (d *defaultRefreshable2[T1, T2]) Current() (T1, T2) {
	val := *(d.current.Load().(*container2[T1, T2]))
	return val.V1, val.V2
}

func (d *defaultRefreshable2[T1, T2]) Subscribe(consumer func(T1, T2)) UnsubscribeFunc {
	d.mux.Lock()
	defer d.mux.Unlock()

	consumerFnPtr := &consumer
	d.subscribers = append(d.subscribers, consumerFnPtr)
	consumer(d.Current())
	return d.unsubscribe(consumerFnPtr)
}

func (d *defaultRefreshable2[T1, T2]) unsubscribe(consumerFnPtr *func(T1, T2)) UnsubscribeFunc {
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

func (d *defaultRefreshable2[T1, T2]) readOnly() *readOnlyRefreshable2[T1, T2] {
	return (*readOnlyRefreshable2[T1, T2])(d)
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

// readOnlyRefreshable aliases defaultRefreshable but hides the Update method so the type
// does not implement Updatable.
type readOnlyRefreshable2[T1, T2 any] defaultRefreshable2[T1, T2]

func (d *readOnlyRefreshable2[T1, T2]) Current() (T1, T2) {
	return (*defaultRefreshable2[T1, T2])(d).Current()
}

func (d *readOnlyRefreshable2[T1, T2]) Subscribe(consumer func(T1, T2)) UnsubscribeFunc {
	return (*defaultRefreshable2[T1, T2])(d).Subscribe(consumer)
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
	return d.base.Subscribe(func(value S) { consumer(d.mapper(value)) })
}

// mapperRefreshable wraps an existing Refreshable and applies a mapping function to its values.
// Subscribe may be called repeatedly with the same value when the underlying value changes but the mapped value does not.
// mapperRefreshable does not implement Updatable because the mapped value may not be able to be converted back to the original type.
type mapperRefreshable2[S1, S2, T1, T2 any] struct {
	base   Refreshable2[S1, S2]
	mapper func(S1, S2) (T1, T2)
}

func (d mapperRefreshable2[S1, S2, T1, T2]) Current() (T1, T2) {
	return d.mapper(d.base.Current())
}

func (d mapperRefreshable2[S1, S2, T1, T2]) Subscribe(consumer func(T1, T2)) UnsubscribeFunc {
	return d.base.Subscribe(func(val1 S1, val2 S2) { consumer(d.mapper(val1, val2)) })
}

type mapperRefreshableFrom2[S1, S2, T any] struct {
	base   Refreshable2[S1, S2]
	mapper func(S1, S2) T
}

func (d mapperRefreshableFrom2[S1, S2, T]) Current() T {
	return d.mapper(d.base.Current())
}

func (d mapperRefreshableFrom2[S1, S2, T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	return d.base.Subscribe(func(val1 S1, val2 S2) { consumer(d.mapper(val1, val2)) })
}

type mapperRefreshableTo2[S, T1, T2 any] struct {
	base   Refreshable[S]
	mapper func(S) (T1, T2)
}

func (d mapperRefreshableTo2[S, T1, T2]) Current() (T1, T2) {
	return d.mapper(d.base.Current())
}

func (d mapperRefreshableTo2[S, T1, T2]) Subscribe(consumer func(T1, T2)) UnsubscribeFunc {
	return d.base.Subscribe(func(val S) { consumer(d.mapper(val)) })
}
