// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	//go:embed testdata/single-executable.tgz
	internalSingleExecutableTGZ []byte
)

func TestDefaultPkgWorkDirUsesStableUserCachePath(t *testing.T) {
	cacheRoot := filepath.Join(t.TempDir(), "user-cache")
	got, err := defaultPkgWorkDir("conjure", func() (string, error) {
		return cacheRoot, nil
	})
	if err != nil {
		t.Fatalf("defaultPkgWorkDir returned error: %v", err)
	}

	want := filepath.Join(cacheRoot, "clipackager", "conjure")
	if got != want {
		t.Fatalf("unexpected default package work directory:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestNewDefaultPackagedCLIRunnerFailsClosedWhenUserCacheDirUnavailable(t *testing.T) {
	runner := newDefaultPackagedCLIRunner(
		"conjure",
		"4.51.0",
		nil,
		".tgz",
		func() (string, error) {
			return "", errors.New("cache root unavailable")
		},
	)

	_, err := runner.EnsureCLIExistsAndReturnPath()
	if err == nil {
		t.Fatal("expected error when user cache directory cannot be resolved")
	}
	if !strings.Contains(err.Error(), "failed to determine user cache directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewPackagedCLIRunnerNormalizesRelativeWorkDirOnce(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	runner := NewPackagedCLIRunner("test-cli-1.0.0", filepath.Join("cache", "work"), nil)
	got := runner.(*packgedCLIRunner).pkgWorkDir
	want := filepath.Join(root, "cache", "work")
	if got != want {
		t.Fatalf("unexpected normalized package work directory:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestResolvedUserCacheWorkDirReusesCachedCLI(t *testing.T) {
	cacheRoot := trustedTempDir(t)
	workDir, err := defaultPkgWorkDir("test-cli", func() (string, error) {
		return cacheRoot, nil
	})
	if err != nil {
		t.Fatalf("defaultPkgWorkDir returned error: %v", err)
	}

	newRunner := func() PackagedCLIRunner {
		return NewPackagedCLIRunner(
			"test-cli-1.0.0",
			workDir,
			NewArchivePackagedCLIProviderFromBytes(
				internalSingleExecutableTGZ,
				".tgz",
				StaticPathProvider("test-cli.sh"),
			),
		)
	}

	path1, output1, err := RunPackagedCLI(newRunner())
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	path2, output2, err := RunPackagedCLI(newRunner())
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if path1 != path2 {
		t.Fatalf("expected stable cached executable path:\nfirst:  %q\nsecond: %q", path1, path2)
	}
	if string(output1) != "Hello, world!\n" || string(output2) != "Hello, world!\n" {
		t.Fatalf("unexpected output:\nfirst:  %q\nsecond: %q", string(output1), string(output2))
	}
}

func trustedTempDir(t *testing.T) string {
	t.Helper()

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("failed to determine user cache directory for test: %v", err)
	}
	if err := os.MkdirAll(cacheRoot, 0700); err != nil {
		t.Fatalf("failed to create user cache directory for test: %v", err)
	}
	tempDir, err := os.MkdirTemp(cacheRoot, "clipackager-test-*")
	if err != nil {
		t.Fatalf("failed to create trusted temporary directory for test: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("failed to remove trusted temporary directory %s: %v", tempDir, err)
		}
	})
	return tempDir
}
