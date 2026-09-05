// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package clipackager

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsurePkgWorkDirRejectsSymlinkComponent(t *testing.T) {
	parent := trustedTempDir(t)
	target := filepath.Join(parent, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	link := filepath.Join(parent, "cache-root")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	err := ensurePkgWorkDir(filepath.Join(link, "clipackager", "conjure"))
	if err == nil {
		t.Fatal("expected symlink path component to be rejected")
	}
	if !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsurePkgWorkDirRejectsStickyWorldWritableParentBeforeCreatingChildren(t *testing.T) {
	parent := filepath.Join(trustedTempDir(t), "shared-cache")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatalf("failed to create parent directory: %v", err)
	}
	if err := os.Chmod(parent, 01777); err != nil {
		t.Fatalf("failed to chmod parent directory: %v", err)
	}

	workDir := filepath.Join(parent, "clipackager", "conjure")
	err := ensurePkgWorkDir(workDir)
	if err == nil {
		t.Fatal("expected world-writable ancestor to be rejected")
	}
	if !strings.Contains(err.Error(), "group or world writable") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Lstat(filepath.Join(parent, "clipackager")); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("expected no child directories to be created below unsafe parent, got: %v", statErr)
	}
}

func TestEnsurePkgWorkDirCreatesPrivateChildrenUnderTrustedAncestors(t *testing.T) {
	workDir := filepath.Join(trustedTempDir(t), "cache-root", "clipackager", "conjure")
	if err := ensurePkgWorkDir(workDir); err != nil {
		t.Fatalf("expected trusted work directory to be accepted: %v", err)
	}
	fi, err := os.Stat(workDir)
	if err != nil {
		t.Fatalf("failed to stat created work directory: %v", err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0700); got != want {
		t.Fatalf("unexpected created work directory mode: want %o, got %o", want, got)
	}
}

func TestPackagedCLIRunnerRejectsCacheHitUnderStickyWorldWritableParent(t *testing.T) {
	sharedParent := filepath.Join(trustedTempDir(t), "shared-cache")
	if err := os.Mkdir(sharedParent, 0700); err != nil {
		t.Fatalf("failed to create shared parent directory: %v", err)
	}
	if err := os.Chmod(sharedParent, 01777); err != nil {
		t.Fatalf("failed to chmod shared parent directory: %v", err)
	}

	workDir := filepath.Join(sharedParent, "clipackager", "conjure")
	attackerCLI := filepath.Join(workDir, "conjure-1.0.0-extract-dir", "conjure")
	if err := os.MkdirAll(filepath.Dir(attackerCLI), 0700); err != nil {
		t.Fatalf("failed to create attacker-controlled cache tree: %v", err)
	}
	if err := os.WriteFile(attackerCLI, []byte("#!/bin/sh\necho attacker-controlled-cli\n"), 0700); err != nil {
		t.Fatalf("failed to create attacker-controlled CLI: %v", err)
	}

	runner := NewPackagedCLIRunner(
		"conjure-1.0.0",
		workDir,
		NewArchivePackagedCLIProviderFromBytes(nil, ".tgz", StaticPathProvider("conjure")),
	)
	_, output, err := RunPackagedCLI(runner)
	if err == nil {
		t.Fatalf("expected unsafe cache hit to be rejected before execution, got output %q", output)
	}
	if !strings.Contains(err.Error(), "group or world writable") {
		t.Fatalf("unexpected error: %v", err)
	}
}
