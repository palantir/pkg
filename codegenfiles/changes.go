// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegenfiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Write is a file that Apply will create or overwrite.
type Write struct {
	// Path is relative to the project directory.
	Path    string
	Content []byte
	// Created reports that no file exists at Path yet.
	Created bool
}

// Changes is the difference between a project's generated output and what is on disk, as computed by
// Project.Plan. A Changes value describes the filesystem but has not modified it: use Err to report the
// difference (verification) or Apply to make it (generation).
type Changes struct {
	dir      string
	writes   []Write  // sorted by Path
	removals []string // sorted
}

// Empty reports whether the project directory already matches the generated output.
func (c *Changes) Empty() bool {
	return len(c.writes) == 0 && len(c.removals) == 0
}

// Writes returns the files that Apply would create or overwrite, sorted by path.
func (c *Changes) Writes() []Write {
	return slices.Clone(c.writes)
}

// Removals returns the paths, relative to the project directory, that Apply would delete, sorted.
func (c *Changes) Removals() []string {
	return slices.Clone(c.removals)
}

// String describes every change on its own line, sorted by path.
func (c *Changes) String() string {
	lines := make([]string, 0, len(c.writes)+len(c.removals))
	for _, w := range c.writes {
		reason := "out of date"
		if w.Created {
			reason = "missing"
		}
		lines = append(lines, fmt.Sprintf("%s: %s", w.Path, reason))
	}
	for _, relPath := range c.removals {
		lines = append(lines, fmt.Sprintf("%s: no longer generated", relPath))
	}
	// each line begins with its path, so sorting the formatted lines orders them by path
	slices.Sort(lines)
	return strings.Join(lines, "\n")
}

// Err returns nil if there are no changes, and otherwise an error describing them. It is the
// verification counterpart of Apply.
func (c *Changes) Err() error {
	if c.Empty() {
		return nil
	}
	return fmt.Errorf("generated output is out of date:\n%s", c)
}

// Apply makes the changes on disk, reporting progress to os.Stdout.
//
// Removals are performed before writes so that a path whose shape changed since the last run -- a
// directory standing where a generated file must now be written -- is cleared first where its contents
// are themselves stale generated files.
func (c *Changes) Apply() error {
	return c.apply(os.Stdout)
}

func (c *Changes) apply(stdout io.Writer) error {
	for _, relPath := range c.removals {
		_, _ = fmt.Fprintf(stdout, "removing file %s\n", relPath)
		if err := os.Remove(filepath.Join(c.dir, relPath)); err != nil {
			return fmt.Errorf("failed to remove %s: %w", relPath, err)
		}
		if err := removeEmptyParentDirectories(c.dir, relPath, stdout); err != nil {
			return fmt.Errorf("failed to remove empty parent directories for %s: %w", relPath, err)
		}
	}

	for _, w := range c.writes {
		fullPath := filepath.Join(c.dir, w.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", w.Path, err)
		}
		if err := writeFileAtomic(fullPath, w.Content); err != nil {
			if info, statErr := os.Lstat(fullPath); statErr == nil && info.IsDir() {
				return fmt.Errorf("failed to write %s: a directory exists at that path", w.Path)
			}
			return fmt.Errorf("failed to write %s: %w", w.Path, err)
		}
	}
	return nil
}

// newFileMode is the permission given to a generated file that does not exist yet.
const newFileMode os.FileMode = 0644

// writeFileAtomic writes content to path by way of a temporary file in the same directory, which is
// then renamed into place. A reader therefore never observes a half-written file, and a run interrupted
// partway through leaves the previous contents intact rather than truncated.
//
// The permissions of an existing file are preserved. os.WriteFile applies its mode argument only when
// creating a file, so replacing a file through a fresh temporary would otherwise quietly reset the mode
// of every file it wrote.
func writeFileAtomic(path string, content []byte) error {
	mode := newFileMode
	if info, err := os.Lstat(path); err == nil && info.Mode().IsRegular() {
		mode = info.Mode().Perm()
	}

	// the temporary lives in the destination directory so that the rename stays within one filesystem,
	// and is named with a leading dot so a concurrent reader is unlikely to pick it up
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err != nil {
		return err
	}
	tmpPath := f.Name()
	// removes the temporary only if the rename below did not happen
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := f.Write(content); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// os.CreateTemp creates with 0600, so the mode must be set explicitly either way
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// removeEmptyParentDirectories removes the parent directories of deletedFile (a path relative to
// rootDir) that have become empty, walking upward until it reaches a non-empty directory or rootDir.
//
// The climb intentionally ignores FileMatcher, so it can remove a directory that FileMatcher does not
// itself match. Only directories that this removal emptied are ever removed, and a directory left
// holding nothing existed solely for the generated files that were just deleted. Consulting the matcher
// instead would disable cleanup altogether for the common case, since a matcher that selects files
// (matcher.Name(`.+\.go`)) matches no directory at all.
func removeEmptyParentDirectories(rootDir string, deletedFile string, stdout io.Writer) error {
	relParent := filepath.Dir(deletedFile)
	if relParent == "" || relParent == "." {
		// reached the project root: stop
		return nil
	}
	parentDir := filepath.Join(rootDir, relParent)
	f, err := os.Open(parentDir)
	if err != nil {
		return err
	}
	// read at most one entry to test emptiness rather than listing the whole directory
	_, readErr := f.Readdirnames(1)
	_ = f.Close()
	if readErr == io.EOF {
		_, _ = fmt.Fprintf(stdout, "removing empty directory %s\n", relParent)
		if err := os.Remove(parentDir); err != nil {
			return fmt.Errorf("failed to remove empty directory %s: %w", relParent, err)
		}
		return removeEmptyParentDirectories(rootDir, relParent, stdout)
	}
	if readErr != nil {
		return readErr
	}
	return nil
}
