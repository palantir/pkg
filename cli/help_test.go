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

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/cli/flag"
)

var stringCases = []struct {
	flag flag.Flag
	str  string
}{{
	flag: flag.StringFlag{Name: "path", Alias: "p", Usage: "where it is"},
	str:  "   --path, -p\twhere it is",
}, {
	flag: flag.StringFlag{Name: "path", Usage: "where it is", Required: true},
	str:  " * --path\twhere it is",
}, {
	flag: flag.StringFlag{Name: "path", Value: "/", Usage: "where it is"},
	str:  "   --path\twhere it is (default=\"/\")",
}, {
	flag: flag.StringFlag{Name: "path", EnvVar: "THE_PATH", Usage: "where it is"},
	str:  "   --path\twhere it is (default=$THE_PATH)",
}, {
	flag: flag.StringFlag{Name: "path", Value: "/", EnvVar: "THE_PATH", Usage: "where it is"},
	str:  "   --path\twhere it is (default=$THE_PATH,\"/\")",
}, {
	flag: flag.StringParam{Name: "path", Usage: "where it is"},
	str:  " * <path>\twhere it is",
}, {
	flag: flag.BoolFlag{Name: "force", Alias: "f", Usage: "forcefully"},
	str:  "   --force, -f\tforcefully",
}}

func TestString(t *testing.T) {
	for _, c := range stringCases {
		actual := flagHelp(c.flag)
		assert.Equal(t, c.str, actual)
	}
}
