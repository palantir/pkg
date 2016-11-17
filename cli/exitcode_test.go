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

package cli_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exitCoderTemplate = `package main

import (
	"fmt"
	"os"

	"github.com/palantir/pkg/cli"
)

func main() {
	app := cli.NewApp()
	app.Action = func(ctx cli.Context) error {
		%v
	}
	os.Exit(app.Run(os.Args))
}
`

// Verify that a CLI app that returns an ExitCoder in its action exits with the specified exit code
func TestExitCoder(t *testing.T) {
	for i, currCase := range []struct {
		action           string
		expectedExitCode int
		expectedOutput   string
	}{
		{action: `return cli.WithExitCode(2, fmt.Errorf("action failed"))`, expectedExitCode: 2, expectedOutput: "action failed\n"},
		{action: `return fmt.Errorf("action failed")`, expectedExitCode: 1, expectedOutput: "action failed\n"},
	} {
		output, err := runGoFile(t, fmt.Sprintf(exitCoderTemplate, currCase.action))
		require.Error(t, err)
		assert.Equal(t, currCase.expectedOutput, string(output), "Case %d", i)

		exiterr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		status, ok := exiterr.Sys().(syscall.WaitStatus)
		require.True(t, ok)
		assert.Equal(t, currCase.expectedExitCode, status.ExitStatus(), "Case %d", i)
	}
}

func runGoFile(t *testing.T, src string) ([]byte, error) {
	tmpDir, err := ioutil.TempDir(".", "")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("Failed to remove directory %v: %v\n", tmpDir, err)
		}
	}()

	err = ioutil.WriteFile(path.Join(tmpDir, "test_cli.go"), []byte(src), 0644)
	require.NoError(t, err)

	buildCmd := exec.Command("go", "build", "-o", "test-cli", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "%v failed: %s", buildCmd.Args, string(output))

	testCLICmd := exec.Command(path.Join(tmpDir, "test-cli"))
	return testCLICmd.CombinedOutput()
}
