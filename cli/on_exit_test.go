// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnExit(t *testing.T) {
	ran := false
	f := func() {
		ran = true
	}

	onExit := newOnExit()
	onExit.Register(f)
	onExit.run()

	assert.True(t, ran)
}

func TestOnExitUnregister(t *testing.T) {
	ran := false
	f := func() {
		ran = true
	}

	onExit := newOnExit()
	id := onExit.Register(f)
	priorityID := onExit.register(f, highPriority)
	onExit.Unregister(id)
	onExit.Unregister(priorityID)
	onExit.run()

	assert.False(t, ran)
}

func TestOnExitWithPanic(t *testing.T) {
	ran := false
	f1 := func() {
		panic("f1 panic")
	}
	f2 := func() {
		ran = true
	}

	onExit := newOnExit()
	onExit.Register(f1)
	onExit.Register(f2)
	onExit.run()

	assert.True(t, ran)
}

func TestRunInOrder(t *testing.T) {
	var got []int

	addToGot := func(n int) func() {
		return func() {
			got = append(got, n)
		}
	}

	onExit := newOnExit()
	onExit.Register(addToGot(0))
	oneID := onExit.Register(addToGot(1))
	onExit.Register(addToGot(2))
	onExit.Register(addToGot(3))
	onExit.Register(addToGot(4))
	onExit.Register(addToGot(5))
	onExit.register(addToGot(10), highPriority)
	onExit.Unregister(oneID)
	onExit.register(addToGot(11), highPriority)

	onExit.run()

	assert.Equal(t, []int{11, 10, 5, 4, 3, 2, 0}, got)
}
