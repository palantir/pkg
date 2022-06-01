// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type DefaultRefreshable[T any] struct {
	current *atomic.Value

	sync.Mutex  // protects subscribers
	subscribers []*func(T)
}

func New[T any](val T) *DefaultRefreshable[T] {
	current := atomic.Value{}
	current.Store(&val)

	return &DefaultRefreshable[T]{
		current: &current,
	}
}

// Update changes the value of the Refreshable, then blocks while subscribers are executed.
func (d *DefaultRefreshable[T]) Update(val T) {
	d.Lock()
	defer d.Unlock()

	if reflect.DeepEqual(d.Current(), val) {
		return
	}

	d.current.Store(&val)

	for _, sub := range d.subscribers {
		(*sub)(val)
	}
}

func (d *DefaultRefreshable[T]) Current() T {
	return *(d.current.Load().(*T))
}

func (d *DefaultRefreshable[T]) Subscribe(consumer func(T)) (unsubscribe func()) {
	d.Lock()
	defer d.Unlock()

	consumerFnPtr := &consumer
	d.subscribers = append(d.subscribers, consumerFnPtr)
	return func() {
		d.unsubscribe(consumerFnPtr)
	}
}

func (d *DefaultRefreshable[T]) unsubscribe(consumerFnPtr *func(T)) {
	d.Lock()
	defer d.Unlock()

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
