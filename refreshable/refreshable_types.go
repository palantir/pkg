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

func NewString(in Refreshable) String {
	return refreshableTyped{
		Refreshable: in,
	}
}

type Int interface {
	Refreshable
	CurrentInt() int
}

func NewInt(in Refreshable) Int {
	return refreshableTyped{
		Refreshable: in,
	}
}

type Bool interface {
	Refreshable
	CurrentBool() bool
}

func NewBool(in Refreshable) Bool {
	return refreshableTyped{
		Refreshable: in,
	}
}

type refreshableTyped struct {
	Refreshable
}

func (rt refreshableTyped) CurrentString() string {
	return rt.Current().(string)
}

func (rt refreshableTyped) CurrentInt() int {
	return rt.Current().(int)
}

func (rt refreshableTyped) CurrentBool() bool {
	return rt.Current().(bool)
}

// Duration is a Refreshable that can return the current time.Duration.
type Duration interface {
	Refreshable
	CurrentDuration() time.Duration
}

func NewDuration(in Refreshable) Duration {
	return refreshableTyped{
		Refreshable: in,
	}
}

func (rt refreshableTyped) CurrentDuration() time.Duration {
	return rt.Current().(time.Duration)
}
