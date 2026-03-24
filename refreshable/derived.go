// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// cleanupState tracks the lifecycle of a derivedRefreshable and is responsible
// for calling unsubscribe functions when the derived is no longer needed.
// Cleanup fires when the derivedRefreshable is garbage collected AND all
// user subscriptions on it have been removed.
type cleanupState struct {
	subCount atomic.Int32
	gcDone   atomic.Bool
	unsubs   []func()
	once     sync.Once
}

func (s *cleanupState) tryCleanup() {
	if s.subCount.Load() > 0 || !s.gcDone.Load() {
		return
	}
	s.once.Do(func() {
		for _, unsub := range s.unsubs {
			unsub()
		}
	})
}

// derivedRefreshable wraps an inner Refreshable and manages cleanup of parent
// subscriptions when the derived is garbage collected. The refs field holds
// references to upstream objects (e.g. parent derivedRefreshable wrappers) to
// prevent them from being GC'd prematurely in chained Map() scenarios.
type derivedRefreshable struct {
	inner Refreshable
	state *cleanupState
	refs  []any
}

func newDerivedRefreshable(inner Refreshable, unsubs ...func()) *derivedRefreshable {
	state := &cleanupState{unsubs: unsubs}
	d := &derivedRefreshable{
		inner: inner,
		state: state,
	}
	runtime.AddCleanup(d, func(s *cleanupState) {
		s.gcDone.Store(true)
		s.tryCleanup()
	}, state)
	return d
}

func (d *derivedRefreshable) Current() interface{} {
	return d.inner.Current()
}

func (d *derivedRefreshable) Subscribe(consumer func(interface{})) (unsubscribe func()) {
	d.state.subCount.Add(1)
	innerUnsub := d.inner.Subscribe(consumer)
	state := d.state
	var once sync.Once
	return func() {
		once.Do(func() {
			innerUnsub()
			state.subCount.Add(-1)
			state.tryCleanup()
		})
	}
}

func (d *derivedRefreshable) Map(mapFn func(interface{}) interface{}) Refreshable {
	result := d.inner.Map(mapFn)
	if dr, ok := result.(*derivedRefreshable); ok {
		dr.refs = append(dr.refs, d)
	}
	return result
}
