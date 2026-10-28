// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

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

func (r *bashRunner) run(t testing.TB, function string, args ...string) bashResult {
	cmd := exec.Command(r.bash, append([]string{"-c", `set -euo pipefail; source "$0"; "$@"`, r.script, function}, args...)...)
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

func TestBashNormalizeMatchesGo(t *testing.T) {
	r := newBashRunner(t)
	for _, tc := range normalizeCases {
		for _, mode := range []gotoolchain.NormalizeMode{gotoolchain.NormalizeExact, gotoolchain.NormalizeLanguage} {
			want, err := gotoolchain.Normalize(tc.version, mode)
			assertBashMatchesGo(t, r.run(t, "go_toolchain_version_normalize", tc.version, mode.String()), want, err)
		}
	}
}
