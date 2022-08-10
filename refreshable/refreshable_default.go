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
	mux         sync.Mutex
	current     atomic.Value
	subscribers []*func(T)
}

// Update changes the value of the Refreshable, then blocks while subscribers are executed.
func (d *DefaultRefreshable[T]) Update(val T) {
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

func (d *DefaultRefreshable[T]) Current() T {
	return *(d.current.Load().(*T))
}

func (d *DefaultRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	d.mux.Lock()
	defer d.mux.Unlock()

	consumerFnPtr := &consumer
	d.subscribers = append(d.subscribers, consumerFnPtr)
	return d.unsubscribe(consumerFnPtr)
}

func (d *DefaultRefreshable[T]) unsubscribe(consumerFnPtr *func(T)) UnsubscribeFunc {
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
