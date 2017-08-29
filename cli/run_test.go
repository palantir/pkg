// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/cli"
)

func TestRunErrorOutput(t *testing.T) {
	cases := []struct {
		err            error
		expectedOutput string
	}{
		// empty error should not produce any output
		{
			err:            fmt.Errorf(""),
			expectedOutput: "",
		},
		// non-empty error is printed with newline appended
		{
			err:            fmt.Errorf("foo"),
			expectedOutput: "foo\n",
		},
	}

	for i, currCase := range cases {
		app := cli.NewApp()
		app.Action = func(ctx cli.Context) error {
			return currCase.err
		}

		stderr := &bytes.Buffer{}
		app.Stderr = stderr

		exitStatus := app.Run([]string{"testApp"})
		if exitStatus == 0 {
			t.Errorf("Case %d: expected exitStatus to be non-0, was: %d", i, exitStatus)
		}

		if stderr.String() != currCase.expectedOutput {
			t.Errorf("Case %d:\nExpected: %q\nActual:   %q", i, currCase.expectedOutput, stderr.String())
		}
	}
}

func TestRunErrorHandler(t *testing.T) {
	cases := []struct {
		err              error
		handler          func(ctx cli.Context, err error) int
		expectedExitCode int
		expectedOutput   string
	}{
		// custom error handler is invoked if provided
		{
			err: fmt.Errorf(""),
			handler: func(ctx cli.Context, err error) int {
				ctx.Errorf("Error: %v\n", err)
				return 13
			},
			expectedExitCode: 13,
			expectedOutput:   "Error: \n",
		},
		// default behavior is used if custom error hander is nil
		{
			err:              fmt.Errorf("foo"),
			expectedExitCode: 1,
			expectedOutput:   "foo\n",
		},
	}

	for i, currCase := range cases {
		app := cli.NewApp()
		app.ErrorHandler = currCase.handler
		app.Action = func(ctx cli.Context) error {
			return currCase.err
		}

		stderr := &bytes.Buffer{}
		app.Stderr = stderr

		exitCode := app.Run([]string{"testApp"})
		assert.Equal(t, currCase.expectedExitCode, exitCode, "Case %d", i)
		assert.Equal(t, currCase.expectedOutput, stderr.String(), "Case %d", i)
	}
}

func TestRunContext(t *testing.T) {

	var customContext = context.WithValue(context.Background(), "message", "hello")

	cases := []struct {
		name  string
		check func(*testing.T)
	}{

		{
			name: "check that default app context is not nil",
			check: func(t *testing.T) {
				app := cli.NewApp()

				assert.NotNil(t, app.Context)
			},
		},
		{
			name: "check that an app can be configured with a custom option",
			check: func(t *testing.T) {
				app := cli.NewApp(func(theApp *cli.App) {
					theApp.Context = customContext
				})

				assert.Equal(t, "hello", app.Context.Value("message"))
			},
		},
		{
			name: "check that context is propagated to app action",
			check: func(t *testing.T) {
				app := cli.NewApp()

				app.Context = customContext

				app.Action = func(ctx cli.Context) error {
					assert.Equal(t, "hello", ctx.Context.Value("message"))
					return nil
				}

				app.Run([]string{"testApp"})
			},
		},
		{
			name: "check that context is propagated to app error handler",
			check: func(t *testing.T) {
				app := cli.NewApp()

				app.Context = customContext

				app.ErrorHandler = func(ctx cli.Context, err error) int {
					assert.Equal(t, "hello", ctx.Context.Value("message"))
					return 0
				}

				app.Action = func(ctx cli.Context) error {
					return fmt.Errorf("an error occured")
				}

				app.Run([]string{"testApp"})
			},
		},
		{
			name: "check that context is propagated to app subcommand",
			check: func(t *testing.T) {
				app := cli.NewApp()

				app.Context = customContext

				app.Subcommands = []cli.Command{
					{
						Name: "subCommand",

						Action: func(ctx cli.Context) error {
							assert.Equal(t, "hello", ctx.Context.Value("message"))

							return nil
						},
					},
				}

				app.Run([]string{"testApp", "subCommand"})
			},
		},
	}

	for i, currCase := range cases {

		name := fmt.Sprintf("Case %d - %s", i, currCase.name)

		t.Run(name, currCase.check)
	}
}
