// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package codegenfiles reconciles the output of a code generator against a directory.
//
// Generating code is the caller's business. This package deals with what comes after: writing the files
// whose content changed, leaving alone the ones that did not, deleting output no longer produced,
// preserving hand-written files that share the directory, and answering "would this change anything?"
// without touching the filesystem.
//
// Collect generated content into an [Output], ask a [Project] what would have to change, then either
// report the result with [Changes.Err] or carry it out with [Changes.Apply]. [Project.Plan] reads the
// filesystem but never writes to it, so verifying and generating differ only in what is done with its
// result: a tool offering both runs the same code either way, and its verify cannot corrupt the tree it
// is checking.
//
// Everything written into a managed directory must go through one [Project]. Reconciling part of it
// separately would let one pass delete what another is responsible for producing, so several generators,
// configurations or schemas contributing to one directory contribute to one [Output] — which is also why
// two of them writing the same path is a collision that Plan reports rather than a silent
// last-writer-wins.
//
// # Ownership
//
// Deleting output nobody produced is the delicate part, and three fields decide it. FileMatcher bounds
// ownership by location, covering both the on-disk files a project reconciles against and the paths its
// generators may produce; a generated path it does not match is an error, since such a path could never
// be reconciled. ContentsMatcher bounds ownership by content, normally via the marker recognized by
// [GeneratedCodeMatcher], so a hand-written file in a generated directory survives. DeleteStale turns
// deletion on, and requires FileMatcher, since a nil matcher matches everything in Dir.
//
// Leaving ContentsMatcher nil is common: a format that cannot carry a comment cannot be identified this
// way, a marker does not match output generated before the marker last changed, and where a generator
// owns a distinctive filename it rules out nothing FileMatcher has not. FileMatcher is then all that
// stands between DeleteStale and a hand-written file, so match the paths the current configuration could
// produce — not a bare filename, which may also appear in vendored copies of a dependency's output, and
// not a directory shared with hand-written files or another generator. Bounded tightly enough it matches
// only the generated set, leaving DeleteStale unable to fire at all: safe, but not working stale-output
// handling.
//
// # Collecting output
//
// However a generator exposes what it produces, it reduces to filling an [Output]. A map of path to
// content goes to [Output.AddAll]; file values that render on demand render into [Output.Add]; a
// generator writing through an interface of its own, such as gengo's FileType, gets one implemented over
// an Output. Paths may be absolute, and Plan resolves them against Dir. Look for that interface before
// reaching for a scratch directory, which only a generator that can write nowhere else needs, and which
// works only if its output does not depend on where it was written.
package codegenfiles
