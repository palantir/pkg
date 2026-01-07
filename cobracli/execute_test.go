// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cobracli_test

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/palantir/pkg/cobracli"
	werror "github.com/palantir/witchcraft-go-error"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteWithParams(t *testing.T) {
	// local variable that can be used as a debug variable
	debugVar := false

	for i, tc := range []struct {
		name       string
		runE       func(cmd *cobra.Command, args []string) error
		configure  func(cmd *cobra.Command)
		args       []string
		params     []cobracli.Param
		wantRV     int
		wantOutput interface{}
	}{
		{
			"standard output",
			func(cmd *cobra.Command, args []string) error {
				cmd.Println("hello, world!")
				return nil
			},
			nil,
			nil,
			cobracli.DefaultParams(nil),
			0,
			"hello, world!\n",
		},
		{
			"version command prints version",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("hello-error")
			},
			nil,
			[]string{"version"},
			append(cobracli.DefaultParams(nil), cobracli.VersionCmdParam("1.0.0")),
			0,
			"my-app version 1.0.0\n",
		},
		{
			"version command not present if not requested",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("hello-error")
			},
			nil,
			[]string{"version"},
			cobracli.DefaultParams(nil),
			1,
			`Error: hello-error
`,
		},
		{
			"version flag prints version",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("hello-error")
			},
			nil,
			[]string{"--version"},
			append(cobracli.DefaultParams(nil), cobracli.VersionFlagParam("1.0.0")),
			0,
			"my-app version 1.0.0\n",
		},
		{
			"version flag exists if version is set on root command",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("hello-error")
			},
			func(cmd *cobra.Command) {
				cmd.Version = "1.0.0"
			},
			[]string{"--version"},
			cobracli.DefaultParams(nil),
			0,
			"my-app version 1.0.0\n",
		},
		{
			"version flag not present if not requested",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("hello-error")
			},
			nil,
			[]string{"--version"},
			cobracli.DefaultParams(nil),
			1,
			`Error: unknown flag: --version
Usage:
  my-app [flags]

Flags:
  -h, --help   help for my-app
`,
		},
		{
			"standard fail",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("custom failure")
			},
			nil,
			nil,
			cobracli.DefaultParams(nil),
			1,
			"Error: custom failure\n",
		},
		{
			"standard fail with package debug variable prints stack trace",
			func(cmd *cobra.Command, args []string) error {
				return errors.New("custom failure")
			},
			nil,
			[]string{
				"--debug",
			},
			append(cobracli.DefaultParams(&debugVar), cobracli.AddDebugPersistentFlagParam(&debugVar)),
			1,
			regexp.MustCompile(`(?s)^Error: custom failure
	github.com/palantir/pkg/cobracli_test.TestExecuteWithParams.func9.+
$`),
		},
	} {
		func() {
			// reset value of the variable on each run
			debugVar = false

			outBuf := &bytes.Buffer{}
			rootCmd := &cobra.Command{
				Use:  "my-app",
				RunE: tc.runE,
			}
			rootCmd.SetOutput(outBuf)
			rootCmd.SetArgs(tc.args)
			if tc.configure != nil {
				tc.configure(rootCmd)
			}

			rv := cobracli.Execute(rootCmd, tc.params...)
			require.Equal(t, tc.wantRV, rv, "Case %d: %s", i, tc.name)

			switch val := tc.wantOutput.(type) {
			case *regexp.Regexp:
				assert.Regexp(t, val, outBuf.String(), "Case %d: %s\nOutput:\n%s", i, tc.name, outBuf.String())
			case string:
				assert.Equal(t, val, outBuf.String(), "Case %d: %s\nOutput:\n%s", i, tc.name, outBuf.String())
			default:
				require.Fail(t, "unsupported type: %s. Case %d, %s", val, i, tc.name)
			}
		}()
	}
}

func TestPrintInfoLevelErrorAndParamsWithDebugTransformer(t *testing.T) {
	t.Run("nil error produces no output", func(t *testing.T) {
		var cmd cobra.Command
		var out bytes.Buffer
		cmd.SetErr(&out)
		cobracli.PrintInfoLevelErrorAndParamsWithDebugTransformer(nil, nil)(&cmd, nil)
		require.Equal(t, "", out.String())
	})

	t.Run("error without params prints only error message", func(t *testing.T) {
		var cmd cobra.Command
		var out bytes.Buffer
		cmd.SetErr(&out)
		cobracli.PrintInfoLevelErrorAndParamsWithDebugTransformer(nil, nil)(&cmd, errors.New("🥳"))
		require.Equal(t, "Error: 🥳\n", out.String())
	})

	t.Run("error with params prints error and params", func(t *testing.T) {
		var cmd cobra.Command
		var out bytes.Buffer
		cmd.SetErr(&out)
		type jsonInvalidMapKey struct{}
		cobracli.PrintInfoLevelErrorAndParamsWithDebugTransformer(nil, nil)(&cmd, werror.ErrorWithContextParams(
			context.Background(),
			"🥳",
			werror.SafeParam("foo", "bar"),
			werror.UnsafeParam("baz", 7),
			werror.UnsafeParam("qux", map[jsonInvalidMapKey]string{{}: ""}),
		))
		require.Equal(t, `Error: 🥳
Error params:
  baz: 7
  foo: "bar"
  qux: error json marshalling parameter value: json: unsupported type: map[cobracli_test.jsonInvalidMapKey]string
`, out.String())
	})

	t.Run("debug mode uses debug transformer", func(t *testing.T) {
		var cmd cobra.Command
		var out bytes.Buffer
		cmd.SetErr(&out)
		cobracli.PrintInfoLevelErrorAndParamsWithDebugTransformer(ptr(true), func(err error) string {
			return "✅"
		})(&cmd, werror.ErrorWithContextParams(
			context.Background(),
			"🥳",
			werror.SafeParam("foo", "bar"),
			werror.UnsafeParam("baz", "qux"),
		))
		require.Equal(t, "Error: ✅\n", out.String())
	})
}

func ptr[T any](v T) *T {
	return &v
}
