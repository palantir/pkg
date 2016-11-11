/*
Copyright 2016 Palantir Technologies, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clitest

import (
	"github.com/palantir/pkg/cli/flag"
)

type dummyFlag struct {
	Name  string
	Value interface{}
}

func (f dummyFlag) MainName() string {
	return f.Name
}

func (f dummyFlag) FullNames() []string {
	return []string{flag.WithPrefix(f.Name)}
}

func (f dummyFlag) IsRequired() bool {
	return false
}

func (f dummyFlag) DeprecationStr() string {
	return ""
}

func (f dummyFlag) HasLeader() bool {
	return true
}

func (f dummyFlag) Default() interface{} {
	return f.Value
}

func (f dummyFlag) Parse(string) (interface{}, error) {
	return f.Value, nil
}

func (f dummyFlag) PlaceholderStr() string {
	panic("not implemented")
}

func (f dummyFlag) DefaultStr() string {
	panic("not implemented")
}

func (f dummyFlag) EnvVarStr() string {
	panic("not implemented")
}

func (f dummyFlag) UsageStr() string {
	panic("not implemented")
}
