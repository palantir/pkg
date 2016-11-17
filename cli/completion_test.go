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

	"github.com/palantir/pkg/cli/completion"
	"github.com/palantir/pkg/cli/flag"
)

var cases = []struct {
	args     []string
	expected []string
	exitCode uint8
}{
	{
		args:     []string{""},
		expected: []string{"init", "commit", "checkout", "merge", "-c"},
	}, {
		args:     []string{"_m"},
		expected: []string{"_manpage"},
	}, {
		args:     []string{"-"},
		expected: []string{"-c"},
	}, {
		args:     []string{"-c", "a=b", "c"},
		expected: []string{"commit", "checkout"},
	}, {
		args:     []string{"c"},
		expected: []string{"commit", "checkout"},
	}, {
		args:     []string{"co"},
		expected: []string{"commit", "co"},
	}, {
		args:     []string{"co", ""},
		expected: []string{"develop", "master", "origin/"},
	}, {
		args:     []string{"co", "origin/"},
		expected: []string{"origin/develop", "origin/master"},
	}, {
		args:     []string{"commit", ""},
		expected: []string{"#-]", "--all"},
	}, {
		args:     []string{"init", ""},
		expected: []string{"#-]"},
	}, {
		args:     []string{"init", "-x", ""},
		expected: []string{"#-["},
	}, {
		args:     []string{"checkout", "tree", ""},
		expected: []string{"#-]", "--track"},
	}, {
		args:     []string{"checkout", "tree", "--"},
		expected: []string{"--track"},
	}, {
		args:     []string{"checkout", "tree", "--track", ""},
		expected: []string{"origin/develop", "origin/master"},
		exitCode: completion.CustomFilepathCode,
	}, {
		args:     []string{"merge", ""},
		expected: []string{"TREE_ISH", nbsp},
	}, {
		args:     []string{"-c", "panic", "merge", ""},
		expected: none, // because of provider panic
	},
}

func TestCompletion(t *testing.T) {
	for _, testcase := range cases {
		app := createApp()
		args := append(testcase.args, "--generate-bash-completion")
		completions, exitCode := doCompletion(app, args)
		assert.Equal(t, testcase.expected, completions)
		assert.Equal(t, testcase.exitCode, exitCode)
	}
}

func createApp() *App {
	app := NewApp()
	app.Name = "git"
	app.Subcommands = []Command{
		{
			Name: "init",
		},
		{
			Name: "commit",
			Flags: []flag.Flag{
				flag.BoolFlag{Name: "all", Alias: "a"},
				flag.BoolFlag{Name: "allow-empty", Deprecated: "allow-empty is now the default behavior"},
			},
		},
		{
			Name:  "checkout",
			Alias: "co",
			Flags: []flag.Flag{
				flag.StringParam{Name: "tree-ish"},
				flag.StringFlag{Name: "track", Alias: "t"},
			},
		},
		{
			Name: "merge",
			Flags: []flag.Flag{
				flag.StringParam{Name: "tree-ish"},
			},
		},
		{
			Name: "_manpage",
		},
	}
	app.Flags = []flag.Flag{
		flag.StringFlag{Name: "c"},
	}
	app.Completion = createProviders()
	return app
}

func createProviders() map[string]completion.Provider {
	return map[string]completion.Provider{
		"tree-ish": func(ctx *completion.ProviderCtx) []string {
			if ctx.Flags["c"] == "panic" {
				panic("provider failed")
			} else if ctx.CommandIs("merge") {
				return none
			} else if ctx.Partial == "origin/" {
				return []string{"origin/develop", "origin/master"}
			}
			return []string{"develop", "master", "origin/"}
		},
		"track": func(ctx *completion.ProviderCtx) []string {
			return completion.CustomFilepath(ctx, []string{"origin/develop", "origin/master"})
		},
	}
}
