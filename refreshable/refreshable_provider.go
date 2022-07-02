// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"sync"
	"time"
)

type ProviderRefreshable[T any] interface {
	Refreshable[T]
	Start(ctx context.Context)
	ReadyC() <-chan struct{}
	ValuesC(buf int) (values <-chan T, cleanup func())
}

func NewTickerRefreshable[T any](interval time.Duration, getValue func(context.Context) (T, bool)) *TickerRefreshable[T] {
	var zeroValue T
	return &TickerRefreshable[T]{
		DefaultRefreshable: New[T](zeroValue),
		ready:              make(chan struct{}),
		interval:           interval,
		getValue:           getValue,
	}
}

type TickerRefreshable[T any] struct {
	*DefaultRefreshable[T]
	readyOnce sync.Once
	ready     chan struct{}
	interval  time.Duration
	getValue  func(ctx context.Context) (T, bool)
}

// compile time interface check
var _ Refreshable[any] = (*TickerRefreshable[any])(nil)

func (p *TickerRefreshable[T]) Ready() bool {
	select {
	case <-p.ReadyC():
		return true
	default:
		return false
	}
}

func (p *TickerRefreshable[T]) ReadyC() <-chan struct{} {
	return p.ready
}

func (p *TickerRefreshable[T]) Wait(ctx context.Context) (t T, ok bool) {
	select {
	case <-p.ReadyC():
		return p.Current(), true
	case <-ctx.Done():
		return
	}
}

func (p *TickerRefreshable[T]) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			value, ok := p.getValue(ctx)
			if ok {
				p.Update(value)
				p.readyOnce.Do(func() {
					close(p.ready)
				})
			}

			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *TickerRefreshable[T]) ValuesC(buf int) (values <-chan T, cleanup func()) {
	ch := make(chan T, buf)
	unsub := p.Subscribe(func(t T) {
		ch <- t
	})
	return ch, func() {
		unsub()
		close(ch)
	}
}
