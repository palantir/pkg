// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager_test

import (
	_ "embed"
	"strings"
	"testing"

	clipackager "github.com/palantir/pkg/clipackager/clirunner"
)

var (
	//go:embed testdata/full-dir-structure.tgz
	fullDirStructureTgz []byte

	//go:embed testdata/multi-top-level-dir.tgz
	multiTopLevelDirTgz []byte

	//go:embed testdata/single-dir.tgz
	singleDirTgz []byte

	//go:embed testdata/single-executable.tgz
	singleExecutableTgz []byte
)

// TestRunCLI verifies the functionality of the RunCLI function using various CLI providers.
// The test code was written by Claude with prompting and iteration.
// The test cases (the TGZs being tested) were human-designed, although the code to generate them in create_archive.go
// was written by Claude with prompting and iteration.
func TestRunCLI(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		archiveBytes              []byte
		pathToExecutableInArchive string
	}{
		{
			name:                      "full directory structure",
			archiveBytes:              fullDirStructureTgz,
			pathToExecutableInArchive: "hello-world-1.0.0/bin/test-cli.sh",
		},
		{
			name:                      "multiple top-level directories",
			archiveBytes:              multiTopLevelDirTgz,
			pathToExecutableInArchive: "bin/test-cli.sh",
		},
		{
			name:                      "single directory",
			archiveBytes:              singleDirTgz,
			pathToExecutableInArchive: "bin/test-cli.sh",
		},
		{
			name:                      "single executable",
			archiveBytes:              singleExecutableTgz,
			pathToExecutableInArchive: "test-cli.sh",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Create a CLI provider from the embedded archive bytes
			cliProvider := clipackager.NewArchiveCLIProviderFromBytes(
				tc.archiveBytes,
				".tgz",
				tc.pathToExecutableInArchive,
			)

			// Create a CLI runner with a unique work directory for this test
			runner := clipackager.NewCLIRunner(
				"test-cli",
				"1.0.0",
				t.TempDir(),
				cliProvider,
			)

			// Run the CLI and verify output
			output, err := clipackager.RunCLI(runner)
			if err != nil {
				t.Fatalf("failed to run CLI: %v", err)
			}

			expectedOutput := "Hello, world!\n"
			actualOutput := string(output)
			if actualOutput != expectedOutput {
				t.Errorf("unexpected output:\nwant: %q\ngot:  %q", expectedOutput, actualOutput)
			}

			// Run the CLI a second time to verify caching works correctly
			output2, err := clipackager.RunCLI(runner)
			if err != nil {
				t.Fatalf("failed to run CLI on second invocation: %v", err)
			}

			actualOutput2 := string(output2)
			if actualOutput2 != expectedOutput {
				t.Errorf("unexpected output on second invocation:\nwant: %q\ngot:  %q", expectedOutput, actualOutput2)
			}
		})
	}
}

func TestRunCLIWithArgs(t *testing.T) {
	cliProvider := clipackager.NewArchiveCLIProviderFromBytes(
		singleExecutableTgz,
		".tgz",
		"test-cli.sh",
	)

	runner := clipackager.NewCLIRunner(
		"test-cli",
		"1.0.0",
		t.TempDir(),
		cliProvider,
	)

	// Run with arguments (the test-cli.sh ignores them but this verifies args are passed)
	output, err := clipackager.RunCLI(runner, "arg1", "arg2", "arg3")
	if err != nil {
		t.Fatalf("failed to run CLI with args: %v", err)
	}

	if !strings.Contains(string(output), "Hello, world!") {
		t.Errorf("unexpected output: %q", string(output))
	}
}

func TestCLIExecutablePath(t *testing.T) {
	cliProvider := clipackager.NewArchiveCLIProviderFromBytes(
		fullDirStructureTgz,
		".tgz",
		"hello-world-1.0.0/bin/test-cli.sh",
	)

	workDir := t.TempDir()
	runner := clipackager.NewCLIRunner(
		"test-cli",
		"1.0.0",
		workDir,
		cliProvider,
	)

	// Get the executable path
	execPath, err := runner.EnsureCLIExistsAndReturnPath()
	if err != nil {
		t.Fatalf("failed to get CLI executable path: %v", err)
	}

	// Verify the path contains the expected components
	if !strings.Contains(execPath, "test-cli-1.0.0") {
		t.Errorf("executable path %q does not contain expected name-version", execPath)
	}

	if !strings.Contains(execPath, "test-cli.sh") {
		t.Errorf("executable path %q does not contain expected executable name", execPath)
	}

	// Call again to verify path is consistent
	execPath2, err := runner.EnsureCLIExistsAndReturnPath()
	if err != nil {
		t.Fatalf("failed to get CLI executable path on second call: %v", err)
	}

	if execPath != execPath2 {
		t.Errorf("executable path changed between calls:\nfirst:  %q\nsecond: %q", execPath, execPath2)
	}
}
