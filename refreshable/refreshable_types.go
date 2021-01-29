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
}

type StringPtr interface {
	Refreshable
	CurrentStringPtr() *string
}

type StringSlice interface {
	Refreshable
	CurrentStringSlice() []string
}

type Int interface {
	Refreshable
	CurrentInt() int
}

type IntPtr interface {
	Refreshable
	CurrentIntPtr() *int
}

type Bool interface {
	Refreshable
	CurrentBool() bool
}

type BoolPtr interface {
	Refreshable
	CurrentBoolPtr() *bool
}

// Duration is a Refreshable that can return the current time.Duration.
type Duration interface {
	Refreshable
	CurrentDuration() time.Duration
}

type DurationPtr interface {
	Refreshable
	CurrentDurationPtr() *time.Duration
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

var (
	_ Bool        = (*refreshableTyped)(nil)
	_ BoolPtr     = (*refreshableTyped)(nil)
	_ Duration    = (*refreshableTyped)(nil)
	_ Int         = (*refreshableTyped)(nil)
	_ IntPtr      = (*refreshableTyped)(nil)
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

func (rt refreshableTyped) CurrentStringPtr() *string {
	return rt.Current().(*string)
}

func (rt refreshableTyped) CurrentStringSlice() []string {
	return rt.Current().([]string)
}

func (rt refreshableTyped) CurrentInt() int {
	return rt.Current().(int)
}

func (rt refreshableTyped) CurrentIntPtr() *int {
	return rt.Current().(*int)
}

func (rt refreshableTyped) CurrentBool() bool {
	return rt.Current().(bool)
}

func (rt refreshableTyped) CurrentBoolPtr() *bool {
	return rt.Current().(*bool)
}

func (rt refreshableTyped) CurrentDuration() time.Duration {
	return rt.Current().(time.Duration)
}

func (rt refreshableTyped) CurrentDurationPtr() *time.Duration {
	return rt.Current().(*time.Duration)
}
