// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegenfiles

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/palantir/pkg/matcher"
)

// Project describes a directory that is reconciled against generated output. Callers produce the output
// themselves and hand it over as an Output; this package decides what has to change on disk and, on
// request, makes it so.
//
// All output written into a managed directory must be reconciled by a single Project. Reconciling part
// of it separately would let one pass delete files that another pass is responsible for producing.
type Project struct {
	// Dir is the project directory. Generated paths are resolved relative to it.
	Dir string

	// FileMatcher matches the files in Dir that the project is responsible for. A file matching
	// FileMatcher that the generated output does not contain is a candidate for deletion. It must match
	// every generated path; a path outside it is an error, because such a path could never be
	// reconciled. It is required whenever DeleteStale is set, since a nil matcher matches every file
	// in Dir.
	FileMatcher matcher.Matcher

	// ContentsMatcher identifies generated files by their content so that only generated files are
	// deleted and hand-written files placed in a generated directory are preserved. If nil, every file
	// matching FileMatcher that the output does not contain is eligible for deletion.
	ContentsMatcher ContentsMatcher

	// DeleteStale controls whether files matching FileMatcher that the output does not contain are
	// removed by Apply (and reported by Err). If false, such files are left alone.
	DeleteStale bool
}

// Plan computes the changes required to make Dir match out. It reads the filesystem but does not modify
// it, so the result can be reported (Changes.Err) or applied (Changes.Apply).
//
// It returns an error if a generated path escapes Dir, is not matched by FileMatcher, or is produced
// more than once.
func (p *Project) Plan(out *Output) (*Changes, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	files, err := p.resolve(out)
	if err != nil {
		return nil, err
	}

	changes := &Changes{dir: p.Dir}
	for _, relPath := range slices.Sorted(maps.Keys(files)) {
		write, err := p.planWrite(relPath, files[relPath])
		if err != nil {
			return nil, err
		}
		if write != nil {
			changes.writes = append(changes.writes, *write)
		}
	}

	if p.DeleteStale {
		onDisk, err := p.matchedFilesOnDisk()
		if err != nil {
			return nil, err
		}
		for _, relPath := range onDisk {
			if _, ok := files[relPath]; ok {
				// produced by the generated output: keep
				continue
			}
			generated, err := p.isGeneratedFile(filepath.Join(p.Dir, relPath))
			if err != nil {
				return nil, fmt.Errorf("failed to determine whether %s is generated: %w", relPath, err)
			}
			if !generated {
				// hand-written file in a generated directory: leave it untouched
				continue
			}
			changes.removals = append(changes.removals, relPath)
		}
	}
	return changes, nil
}

// planWrite reports the write needed to give relPath the provided content, or nil if the file on disk
// already matches.
func (p *Project) planWrite(relPath string, content []byte) (*Write, error) {
	fullPath := filepath.Join(p.Dir, relPath)
	info, err := os.Lstat(fullPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return &Write{Path: relPath, Content: content, Created: true}, nil
	case err != nil:
		return nil, fmt.Errorf("failed to stat %s: %w", relPath, err)
	case info.IsDir():
		// a directory stands where the file must go; Apply clears it first if its contents are stale
		// generated files, and reports the conflict precisely if they are not
		return &Write{Path: relPath, Content: content, Created: true}, nil
	}

	existing, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", relPath, err)
	}
	if bytes.Equal(existing, content) {
		return nil, nil
	}
	return &Write{Path: relPath, Content: content}, nil
}

// validate rejects configurations whose behavior would be destructive or impossible to satisfy.
func (p *Project) validate() error {
	if p.Dir == "" {
		return errors.New("Dir must be set")
	}
	// A nil matcher matches every path, so deleting against one would treat the entire project
	// directory as generated output and remove all of it.
	if p.DeleteStale && p.FileMatcher == nil {
		return errors.New("FileMatcher must be set when DeleteStale is true: a nil FileMatcher matches every file in Dir")
	}
	return nil
}

// resolve normalizes the paths in out to clean paths relative to Dir, keyed identically to the on-disk
// paths they are compared against, and rejects any that the project cannot reconcile.
func (p *Project) resolve(out *Output) (map[string][]byte, error) {
	absDir, err := filepath.Abs(p.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project directory %q: %w", p.Dir, err)
	}

	files := make(map[string][]byte, out.Len())
	for _, entry := range out.entries {
		relPath, err := relativizePath(absDir, entry.path)
		if err != nil {
			return nil, err
		}
		// FileMatcher bounds both the on-disk set the project reconciles against and the set it may
		// delete from, so a generated path outside it could never be reconciled.
		if p.FileMatcher != nil && !p.FileMatcher.Match(relPath) {
			return nil, fmt.Errorf("generated path %q is not matched by FileMatcher; FileMatcher must match every generated path", relPath)
		}
		if _, ok := files[relPath]; ok {
			return nil, fmt.Errorf("generated path %q was produced more than once", relPath)
		}
		files[relPath] = entry.content
	}
	return files, nil
}

// relativizePath normalizes a generated path to a clean path relative to the project directory.
// Absolute paths are made relative to it; paths that resolve outside it are rejected.
func relativizePath(absDir, filePath string) (string, error) {
	rel := filePath
	if filepath.IsAbs(filePath) {
		r, err := filepath.Rel(absDir, filePath)
		if err != nil {
			return "", fmt.Errorf("generated path %q is outside project directory %q: %w", filePath, absDir, err)
		}
		rel = r
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("generated path %q escapes the project directory", filePath)
	}
	return rel, nil
}

// matchedFilesOnDisk returns the paths, relative to Dir and sorted, of the files in Dir that
// FileMatcher matches. Directories are not returned: the project reconciles files, and empty
// directories are cleaned up as a consequence of removing the files in them.
func (p *Project) matchedFilesOnDisk() ([]string, error) {
	var paths []string
	err := filepath.WalkDir(p.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(p.Dir, path)
		if err != nil {
			return err
		}
		if p.FileMatcher == nil || p.FileMatcher.Match(relPath) {
			paths = append(paths, relPath)
		}
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		// the project directory does not exist yet, so nothing on disk can be stale
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", p.Dir, err)
	}
	// WalkDir descends depth-first, which does not order the relative paths it produces
	slices.Sort(paths)
	return paths, nil
}

// isGeneratedFile reports whether the file at fullPath is one the project owns, according to
// ContentsMatcher. A nil matcher treats every file as generated.
func (p *Project) isGeneratedFile(fullPath string) (bool, error) {
	if p.ContentsMatcher == nil {
		return true, nil
	}
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return false, err
	}
	return p.ContentsMatcher.MatchFileContents(contents), nil
}
