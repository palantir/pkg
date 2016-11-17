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

package clitest

import (
	"bytes"
	"fmt"

	"github.com/palantir/pkg/cli"
	"github.com/palantir/pkg/cli/flag"
)

func Context(flags map[string]interface{}) cli.Context {
	app := cli.NewApp()
	app.Stdout = new(bytes.Buffer)
	app.Stderr = new(bytes.Buffer)
	app.Flags = make([]flag.Flag, 0, len(flags))

	args := make([]string, 0, len(flags)*2+1)
	args = append(args, "dummyApp")
	for name, value := range flags {
		app.Flags = append(app.Flags, dummyFlag{Name: name, Value: value})
		args = append(args, fmt.Sprintf("--%v", name), "OVERWRITTEN_VALUE")
	}

	var theCtx cli.Context
	app.Action = func(ctx cli.Context) error {
		theCtx = ctx
		return nil
	}

	status := app.Run(args)
	if status != 0 {
		panic(status)
	}

	return theCtx
}

// Stdout returns the output printed to ctx.App.Stdout as a string. Assumes context was created by clitest.Context.
func Stdout(ctx cli.Context) string { return ctx.App.Stdout.(*bytes.Buffer).String() }

// Stderr returns the output printed to ctx.App.Stderr as a string. Assumes context was created by clitest.Context.
func Stderr(ctx cli.Context) string { return ctx.App.Stderr.(*bytes.Buffer).String() }
