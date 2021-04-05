// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"time"
)

type String interface {
	Refreshable
	CurrentString() string
	MapString(func(string) interface{}) Refreshable
	SubscribeToString(func(string)) (unsubscribe func())
}

type StringPtr interface {
	Refreshable
	CurrentStringPtr() *string
	MapStringPtr(func(*string) interface{}) Refreshable
	SubscribeToStringPtr(func(*string)) (unsubscribe func())
}

type StringSlice interface {
	Refreshable
	CurrentStringSlice() []string
	MapStringSlice(func([]string) interface{}) Refreshable
	SubscribeToStringSlice(func([]string)) (unsubscribe func())
}

type Int interface {
	Refreshable
	CurrentInt() int
	MapInt(func(int) interface{}) Refreshable
	SubscribeToInt(func(int)) (unsubscribe func())
}

type IntPtr interface {
	Refreshable
	CurrentIntPtr() *int
	MapIntPtr(func(*int) interface{}) Refreshable
	SubscribeToIntPtr(func(*int)) (unsubscribe func())
}

type Int64 interface {
	Refreshable
	CurrentInt64() int64
	MapInt64(func(int64) interface{}) Refreshable
	SubscribeToInt64(func(int64)) (unsubscribe func())
}

type Int64Ptr interface {
	Refreshable
	CurrentInt64Ptr() *int64
	MapInt64Ptr(func(*int64) interface{}) Refreshable
	SubscribeToInt64Ptr(func(*int64)) (unsubscribe func())
}

type Float64 interface {
	Refreshable
	CurrentFloat64() float64
	MapFloat64(func(float64) interface{}) Refreshable
	SubscribeToFloat64(func(float64)) (unsubscribe func())
}

type Float64Ptr interface {
	Refreshable
	CurrentFloat64Ptr() *float64
	MapFloat64Ptr(func(*float64) interface{}) Refreshable
	SubscribeToFloat64Ptr(func(*float64)) (unsubscribe func())
}

type Bool interface {
	Refreshable
	CurrentBool() bool
	MapBool(func(bool) interface{}) Refreshable
	SubscribeToBool(func(bool)) (unsubscribe func())
}

type BoolPtr interface {
	Refreshable
	CurrentBoolPtr() *bool
	MapBoolPtr(func(*bool) interface{}) Refreshable
	SubscribeToBoolPtr(func(*bool)) (unsubscribe func())
}

// Duration is a Refreshable that can return the current time.Duration.
type Duration interface {
	Refreshable
	CurrentDuration() time.Duration
	MapDuration(func(time.Duration) interface{}) Refreshable
	SubscribeToDuration(func(time.Duration)) (unsubscribe func())
}

type DurationPtr interface {
	Refreshable
	CurrentDurationPtr() *time.Duration
	MapDurationPtr(func(*time.Duration) interface{}) Refreshable
	SubscribeToDurationPtr(func(*time.Duration)) (unsubscribe func())
}

func NewBool(in Refreshable) Bool {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewBoolPtr(in Refreshable) BoolPtr {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewDuration(in Refreshable) Duration {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewDurationPtr(in Refreshable) DurationPtr {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewString(in Refreshable) String {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewStringPtr(in Refreshable) StringPtr {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewStringSlice(in Refreshable) StringSlice {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewInt(in Refreshable) Int {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewIntPtr(in Refreshable) IntPtr {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewInt64(in Refreshable) Int64 {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewInt64Ptr(in Refreshable) Int64Ptr {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewFloat64(in Refreshable) Float64 {
	return refreshableTyped{
		Refreshable: in,
	}
}

func NewFloat64Ptr(in Refreshable) Float64Ptr {
	return refreshableTyped{
		Refreshable: in,
	}
}

var (
	_ Bool        = (*refreshableTyped)(nil)
	_ BoolPtr     = (*refreshableTyped)(nil)
	_ Duration    = (*refreshableTyped)(nil)
	_ Int         = (*refreshableTyped)(nil)
	_ IntPtr      = (*refreshableTyped)(nil)
	_ Int64       = (*refreshableTyped)(nil)
	_ Int64Ptr    = (*refreshableTyped)(nil)
	_ Float64     = (*refreshableTyped)(nil)
	_ Float64Ptr  = (*refreshableTyped)(nil)
	_ String      = (*refreshableTyped)(nil)
	_ StringPtr   = (*refreshableTyped)(nil)
	_ StringSlice = (*refreshableTyped)(nil)
)

type refreshableTyped struct {
	Refreshable
}

func (rt refreshableTyped) CurrentString() string {
	return rt.Current().(string)
}

func (rt refreshableTyped) MapString(mapFn func(string) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(string))
	})
}

func (rt refreshableTyped) SubscribeToString(subFn func(string)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(string))
	})
}

func (rt refreshableTyped) CurrentStringPtr() *string {
	return rt.Current().(*string)
}

func (rt refreshableTyped) MapStringPtr(mapFn func(*string) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*string))
	})
}

func (rt refreshableTyped) SubscribeToStringPtr(subFn func(*string)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*string))
	})
}

func (rt refreshableTyped) CurrentStringSlice() []string {
	return rt.Current().([]string)
}

func (rt refreshableTyped) MapStringSlice(mapFn func([]string) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.([]string))
	})
}

func (rt refreshableTyped) SubscribeToStringSlice(subFn func([]string)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.([]string))
	})
}

func (rt refreshableTyped) CurrentInt() int {
	return rt.Current().(int)
}

func (rt refreshableTyped) MapInt(mapFn func(int) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(int))
	})
}

func (rt refreshableTyped) SubscribeToInt(subFn func(int)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(int))
	})
}

func (rt refreshableTyped) CurrentIntPtr() *int {
	return rt.Current().(*int)
}

func (rt refreshableTyped) MapIntPtr(mapFn func(*int) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*int))
	})
}

func (rt refreshableTyped) SubscribeToIntPtr(subFn func(*int)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*int))
	})
}

func (rt refreshableTyped) CurrentInt64() int64 {
	return rt.Current().(int64)
}

func (rt refreshableTyped) MapInt64(mapFn func(int64) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(int64))
	})
}

func (rt refreshableTyped) SubscribeToInt64(subFn func(int64)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(int64))
	})
}

func (rt refreshableTyped) CurrentInt64Ptr() *int64 {
	return rt.Current().(*int64)
}

func (rt refreshableTyped) MapInt64Ptr(mapFn func(*int64) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*int64))
	})
}

func (rt refreshableTyped) SubscribeToInt64Ptr(subFn func(*int64)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*int64))
	})
}

func (rt refreshableTyped) CurrentFloat64() float64 {
	return rt.Current().(float64)
}

func (rt refreshableTyped) MapFloat64(mapFn func(float64) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(float64))
	})
}

func (rt refreshableTyped) SubscribeToFloat64(subFn func(float64)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(float64))
	})
}

func (rt refreshableTyped) CurrentFloat64Ptr() *float64 {
	return rt.Current().(*float64)
}

func (rt refreshableTyped) MapFloat64Ptr(mapFn func(*float64) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*float64))
	})
}

func (rt refreshableTyped) SubscribeToFloat64Ptr(subFn func(*float64)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*float64))
	})
}

func (rt refreshableTyped) CurrentBool() bool {
	return rt.Current().(bool)
}

func (rt refreshableTyped) MapBool(mapFn func(bool) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(bool))
	})
}

func (rt refreshableTyped) SubscribeToBool(subFn func(bool)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(bool))
	})
}

func (rt refreshableTyped) CurrentBoolPtr() *bool {
	return rt.Current().(*bool)
}

func (rt refreshableTyped) MapBoolPtr(mapFn func(*bool) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*bool))
	})
}

func (rt refreshableTyped) SubscribeToBoolPtr(subFn func(*bool)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*bool))
	})
}

func (rt refreshableTyped) CurrentDuration() time.Duration {
	return rt.Current().(time.Duration)
}

func (rt refreshableTyped) MapDuration(mapFn func(time.Duration) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(time.Duration))
	})
}

func (rt refreshableTyped) SubscribeToDuration(subFn func(time.Duration)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(time.Duration))
	})
}

func (rt refreshableTyped) CurrentDurationPtr() *time.Duration {
	return rt.Current().(*time.Duration)
}

func (rt refreshableTyped) MapDurationPtr(mapFn func(*time.Duration) interface{}) Refreshable {
	return rt.Map(func(i interface{}) interface{} {
		return mapFn(i.(*time.Duration))
	})
}

func (rt refreshableTyped) SubscribeToDurationPtr(subFn func(*time.Duration)) (unsubscribe func()) {
	return rt.Subscribe(func(i interface{}) {
		subFn(i.(*time.Duration))
	})
}
