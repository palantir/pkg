// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package clipackager

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func ensurePkgWorkDirPlatform(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute package work directory %s: %w", path, err)
	}
	absPath = filepath.Clean(absPath)

	current := string(filepath.Separator)
	fi, err := os.Lstat(current)
	if err != nil {
		return fmt.Errorf("failed to inspect package work directory component %s: %w", current, err)
	}
	if err := validatePkgWorkDirUnix(current, fi); err != nil {
		return err
	}

	relative := strings.TrimPrefix(absPath, current)
	if relative == "" {
		return nil
	}
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" {
			continue
		}
		current = filepath.Join(current, component)
		fi, err = os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.Mkdir(current, 0700); err != nil && !errors.Is(err, fs.ErrExist) {
				return fmt.Errorf("failed to create package work directory component %s: %w", current, err)
			}
			fi, err = os.Lstat(current)
		}
		if err != nil {
			return fmt.Errorf("failed to inspect package work directory component %s: %w", current, err)
		}
		if err := validatePkgWorkDirUnix(current, fi); err != nil {
			return err
		}
	}
	return nil
}

func validatePkgWorkDirUnix(path string, fi os.FileInfo) error {
	if err := validatePkgWorkDirEntry(path, fi); err != nil {
		return err
	}
	if fi.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("package work directory component %s is group or world writable: mode %o", path, fi.Mode().Perm())
	}

	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("could not determine owner of package work directory component %s", path)
	}
	if got, want := stat.Uid, uint32(os.Geteuid()); got != want && got != 0 {
		return fmt.Errorf("package work directory component %s is owned by uid %d, expected uid %d or root", path, got, want)
	}
	return nil
}
