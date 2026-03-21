// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// derivedRefreshable wraps a Refreshable and ties upstream subscription lifetime
// to its own GC reachability and active subscriber count via cleanupState.
//
// Must be a separate allocation from the inner Refreshable: the inner is
// typically captured by upstream subscription callbacks, so combining them
// would create a reference cycle that prevents runtime.AddCleanup from firing.
type derivedRefreshable[T any] struct {
	inner Refreshable[T]
	state *cleanupState
	// refs keeps upstream objects alive to prevent premature GC cleanup of
	// upstream derived wrappers that this consumer depends on.
	refs []any
}

func newDerivedRefreshable[T any](inner Refreshable[T], unsubs ...UnsubscribeFunc) *derivedRefreshable[T] {
	state := &cleanupState{
		unsubs: unsubs,
	}
	d := &derivedRefreshable[T]{
		inner: inner,
		state: state,
	}
	runtime.AddCleanup(d, runCleanup, state)
	return d
}

func (d *derivedRefreshable[T]) Current() T {
	return d.inner.Current()
}

func (d *derivedRefreshable[T]) Subscribe(consumer func(T)) UnsubscribeFunc {
	d.state.addSubscriber()
	innerUnsub := d.inner.Subscribe(consumer)
	state := d.state // capture state not d, so d remains GC-eligible
	var once sync.Once
	return func() {
		once.Do(func() {
			innerUnsub()
			state.removeSubscriber()
		})
	}
}

// derivedValidated wraps a Validated with the same GC cleanup semantics as
// derivedRefreshable. See derivedRefreshable for design rationale.
type derivedValidated[T any] struct {
	inner Validated[T]
	state *cleanupState
	refs  []any // see derivedRefreshable.refs
}

func newDerivedValidated[T any](inner Validated[T], unsubs ...UnsubscribeFunc) *derivedValidated[T] {
	state := &cleanupState{
		unsubs: unsubs,
	}
	d := &derivedValidated[T]{
		inner: inner,
		state: state,
	}
	runtime.AddCleanup(d, runCleanup, state)
	return d
}

func (d *derivedValidated[T]) Unvalidated() T {
	return d.inner.Unvalidated()
}

func (d *derivedValidated[T]) Validation() (T, error) {
	return d.inner.Validation()
}

func (d *derivedValidated[T]) SubscribeValidated(consumer func(Validated[T])) UnsubscribeFunc {
	d.state.addSubscriber()
	innerUnsub := d.inner.SubscribeValidated(consumer)
	state := d.state // capture state not d, so d remains GC-eligible
	var once sync.Once
	return func() {
		once.Do(func() {
			innerUnsub()
			state.removeSubscriber()
		})
	}
}

// cleanupState coordinates between GC and active subscribers. Upstream
// unsubscribe functions fire only when the derived wrapper is GC'd AND
// the subscriber count reaches zero, allowing subscriptions on a derived
// refreshable to outlive the derived reference itself.
type cleanupState struct {
	subCount atomic.Int32      // incremented by Subscribe, decremented by its returned unsub
	gcDone   atomic.Bool       // set by runtime.AddCleanup callback
	unsubs   []UnsubscribeFunc // upstream unsubscribe functions
	once     sync.Once         // ensures cleanup runs exactly once
}

func (s *cleanupState) addSubscriber() {
	s.subCount.Add(1)
}

func (s *cleanupState) removeSubscriber() {
	s.subCount.Add(-1)
	s.tryCleanup()
}

func (s *cleanupState) markGCd() {
	s.gcDone.Store(true)
	s.tryCleanup()
}

func (s *cleanupState) tryCleanup() {
	if s.gcDone.Load() && s.subCount.Load() <= 0 {
		s.once.Do(func() {
			for _, unsub := range s.unsubs {
				unsub()
			}
		})
	}
}

// runCleanup is the runtime.AddCleanup callback. Must be a package-level
// function (not a closure) to avoid preventing GC of the tracked object.
func runCleanup(state *cleanupState) { state.markGCd() }
