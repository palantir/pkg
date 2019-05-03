// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag

import (
	"strconv"
)

type IntParam struct {
	Name  string
	Usage string
}

func (f IntParam) MainName() string {
	return f.Name
}

func (f IntParam) FullNames() []string {
	return []string{f.Name}
}

func (f IntParam) IsRequired() bool {
	return true
}

func (f IntParam) DeprecationStr() string {
	return ""
}

func (f IntParam) HasLeader() bool {
	return false
}

func (f IntParam) Default() interface{} {
	panic("always required")
}

func (f IntParam) Parse(str string) (interface{}, error) {
	i, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return nil, err
	}
	return int(i), nil
}

func (f IntParam) PlaceholderStr() string {
	return defaultPlaceholder(f.Name)
}

func (f IntParam) DefaultStr() string {
	panic("always required")
}

func (f IntParam) EnvVarStr() string {
	panic("always required")
}

func (f IntParam) UsageStr() string {
	return f.Usage
}
