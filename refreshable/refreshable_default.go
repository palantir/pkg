// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

type DefaultRefreshable struct {
	typ     reflect.Type
	current *atomic.Value

	sync.Mutex  // protects subscribers
	subscribers []*func(interface{})
}

func NewDefaultRefreshable(val interface{}) *DefaultRefreshable {
	current := atomic.Value{}
	current.Store(val)

	return &DefaultRefreshable{
		current: &current,
		typ:     reflect.TypeOf(val),
	}
}

func (d *DefaultRefreshable) Update(val interface{}) error {
	d.Lock()
	defer d.Unlock()

	if valType := reflect.TypeOf(val); valType != d.typ {
		return fmt.Errorf("new refreshable value must be type %s: got %s", d.typ, valType)
	}

	if reflect.DeepEqual(d.current.Load(), val) {
		return nil
	}
	d.current.Store(val)

	for _, sub := range d.subscribers {
		(*sub)(val)
	}
	return nil
}

func (d *DefaultRefreshable) Current() interface{} {
	return d.current.Load()
}

func (d *DefaultRefreshable) Subscribe(consumer func(interface{})) (unsubscribe func()) {
	d.Lock()
	defer d.Unlock()

	consumerFnPtr := &consumer
	d.subscribers = append(d.subscribers, consumerFnPtr)
	return func() {
		d.unsubscribe(consumerFnPtr)
	}
}

func (d *DefaultRefreshable) unsubscribe(consumerFnPtr *func(interface{})) {
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

func (d *DefaultRefreshable) Map(mapFn func(interface{}) interface{}) Refreshable {
	newRefreshable := NewDefaultRefreshable(mapFn(d.Current()))
	d.Subscribe(func(updatedVal interface{}) {
		_ = newRefreshable.Update(mapFn(updatedVal))
	})
	return newRefreshable
}
