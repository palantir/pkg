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

package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirRaw(t *testing.T) {
	assert.Equal(t, "", dirRaw(""))
	assert.Equal(t, "./", dirRaw("./"))
	assert.Equal(t, "test/", dirRaw("test/"))
	assert.Equal(t, "test/", dirRaw("test/ing"))
	assert.Equal(t, "test/ing/", dirRaw("test/ing/"))
	assert.Equal(t, "test/ing/", dirRaw("test/ing/still"))
	assert.Equal(t, "/", dirRaw("/"))
	assert.Equal(t, "/", dirRaw("/test"))
	assert.Equal(t, "/test/", dirRaw("/test/"))
	assert.Equal(t, "/test/", dirRaw("/test/ing"))
	assert.Equal(t, "/test/ing/", dirRaw("/test/ing/"))
	assert.Equal(t, "/test/ing/", dirRaw("/test/ing/still"))
}
