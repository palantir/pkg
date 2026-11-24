// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotoolchain_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/palantir/pkg/gotoolchain"
)

type bashResult struct {
	stdout, stderr string
	ok             bool
}
type bashRunner struct{ bash, script string }

func newBashRunner(t testing.TB) *bashRunner {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	script := filepath.Join(t.TempDir(), gotoolchain.BashScriptFileName)
	requireNoError(t, os.WriteFile(script, gotoolchain.BashScript(), 0644))
	return &bashRunner{bash, script}
}

// run sources the script and runs the provided function with the provided arguments.
func (r *bashRunner) run(t testing.TB, function string, args ...string) bashResult {
	return r.runScript(t, `source "$0"; "$@"`, append([]string{function}, args...)...)
}

// runScript runs the provided bash script with "set -euo pipefail". In the script, $0 is the path to the gotoolchain
// script and the positional parameters are the provided arguments.
func (r *bashRunner) runScript(t testing.TB, script string, args ...string) bashResult {
	cmd := exec.Command(r.bash, append([]string{"-c", "set -euo pipefail; " + script, r.script}, args...)...)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "LC_ALL=C"}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		requireNoError(t, err)
	}
	return bashResult{stdout.String(), stderr.String(), err == nil}
}

func assertBashMatchesGo(t *testing.T, got bashResult, want string, wantErr error) {
	t.Helper()
	if wantErr != nil {
		if got.ok {
			t.Error("bash command succeeded, want failure")
		}
		if got.stdout != "" {
			t.Errorf("stdout = %q, want empty", got.stdout)
		}
		if !strings.Contains(got.stderr, "gotoolchain: ") {
			t.Errorf("stderr = %q, want a gotoolchain error", got.stderr)
		}
		return
	}
	if !got.ok {
		t.Errorf("bash command failed: %s", got.stderr)
	}
	if got.stdout != want+"\n" {
		t.Errorf("stdout = %q, want %q", got.stdout, want+"\n")
	}
	if got.stderr != "" {
		t.Errorf("stderr = %q, want empty", got.stderr)
	}
}

func TestBashScriptIsEmbeddedFile(t *testing.T) {
	content, err := os.ReadFile(gotoolchain.BashScriptFileName)
	requireNoError(t, err)
	got := gotoolchain.BashScript()
	if !bytes.Equal(got, content) {
		t.Error("BashScript does not match the embedded file")
	}
	got[0] = 'X'
	if !bytes.Equal(gotoolchain.BashScript(), content) {
		t.Error("modifying the returned BashScript changed the embedded content")
	}
}

func TestBashFromGoModMatchesGo(t *testing.T) {
	r := newBashRunner(t)
	for _, tc := range loadSelectionCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			path := writeGoMod(t, tc.GoMod)
			want, wantErr := gotoolchain.FromGoMod(path)
			assertBashMatchesGo(t, r.run(t, "go_toolchain_version_from_go_mod", path), want, wantErr)
		})
	}
}

func TestBashDirectivesMatchesGo(t *testing.T) {
	r := newBashRunner(t)
	for _, tc := range loadSelectionCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			path := writeGoMod(t, tc.GoMod)
			d, err := gotoolchain.ReadDirectives(path)
			assertBashMatchesGo(t, r.run(t, "go_toolchain_version_directives", path), "go="+d.Go+"\ntoolchain="+d.Toolchain, err)
		})
	}
}

func TestBashSelectMatchesGo(t *testing.T) {
	r := newBashRunner(t)
	for _, tc := range selectCases {
		want, err := tc.d.SelectedToolchain()
		assertBashMatchesGo(t, r.run(t, "go_toolchain_version_select", tc.d.Go, tc.d.Toolchain), want, err)
		if tc.d.Toolchain == "" {
			// an omitted toolchain argument is equivalent to an empty one
			assertBashMatchesGo(t, r.run(t, "go_toolchain_version_select", tc.d.Go), want, err)
		}
	}
}

func TestBashNormalizeMatchesGo(t *testing.T) {
	r := newBashRunner(t)
	for _, tc := range normalizeCases {
		for _, mode := range []gotoolchain.NormalizeMode{gotoolchain.NormalizeExact, gotoolchain.NormalizeLanguage} {
			want, err := gotoolchain.Normalize(tc.version, mode)
			assertBashMatchesGo(t, r.run(t, "go_toolchain_version_normalize", tc.version, mode.String()), want, err)
		}
		// exact is the default if the mode is omitted
		want, err := gotoolchain.Normalize(tc.version, gotoolchain.NormalizeExact)
		assertBashMatchesGo(t, r.run(t, "go_toolchain_version_normalize", tc.version), want, err)
	}
}

// TestBashSourcingOnlyDefinesFunctions verifies that the script can be sourced by other scripts without side effects:
// sourcing must not print output, define or change variables, or change shell options, and must only define functions
// with the go_toolchain_version_ prefix.
func TestBashSourcingOnlyDefinesFunctions(t *testing.T) {
	r := newBashRunner(t)
	got := r.runScript(t, `vars() { compgen -v | grep -v -x -e _ -e BASH_ARGC -e BASH_ARGV -e BASH_LINENO -e BASH_SOURCE -e BASH_COMMAND -e FUNCNAME -e vars_before -e opts_before -e output | sort; }
vars_before='' opts_before='' output=''
vars_before=$(vars)
opts_before="$(set -o)$(shopt -p)"
output=$(source "$0" 2>&1)
source "$0"
[[ -z $output ]] || { echo "sourcing printed output: $output" >&2; exit 1; }
[[ $(vars) == "$vars_before" ]] || { echo "sourcing changed variables" >&2; exit 1; }
[[ "$(set -o)$(shopt -p)" == "$opts_before" ]] || { echo "sourcing changed shell options" >&2; exit 1; }
unset -f vars
compgen -A function`)
	if !got.ok {
		t.Fatalf("sourcing failed: %s", got.stderr)
	}
	for _, function := range strings.Fields(got.stdout) {
		if !strings.HasPrefix(function, "go_toolchain_version_") {
			t.Errorf("sourcing defined function %q without the go_toolchain_version_ prefix", function)
		}
	}
}
