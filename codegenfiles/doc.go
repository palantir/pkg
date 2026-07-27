// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package codegenfiles reconciles the output of a code generator against a directory.
//
// Generating code is the caller's business. This package deals with what comes after: writing the files
// whose content changed, leaving alone the ones that did not, deleting output that is no longer
// produced, preserving hand-written files that share the directory, and answering "would this change
// anything?" without touching the filesystem.
//
// A caller collects generated content into an [Output] and asks a [Project] what would have to change:
//
//	out := codegenfiles.NewOutput()
//	out.AddAll(generated)
//
//	p := &codegenfiles.Project{
//		Dir:             projectDir,
//		FileMatcher:     matcher.Name(`.+\.go`),
//		ContentsMatcher: codegenfiles.GeneratedCodeMatcher(),
//		DeleteStale:     true,
//	}
//	changes, err := p.Plan(out)
//	if err != nil {
//		return err
//	}
//	if verify {
//		return changes.Err()
//	}
//	return changes.Apply()
//
// [Project.Plan] reads the filesystem but never writes to it, so verifying and generating differ only
// in what is done with the result. A tool exposing the two as separate modes runs the same code either
// way, and its verify cannot corrupt the tree it is checking.
//
// # Ownership
//
// Deleting output nobody produced is the delicate part, and two independent questions decide it.
//
// FileMatcher answers "is this file mine by location?" and bounds everything: the on-disk files a
// project reconciles against, and the paths its generators may produce. A generated path it does not
// match is an error, because such a path could never be reconciled.
//
// ContentsMatcher answers "is this file mine by content?", normally via the generated-code marker
// recognized by [GeneratedCodeMatcher]. It exists so that a hand-written file placed in a generated
// directory survives. A format that cannot carry a comment, such as JSON, can never be identified this
// way, so a project generating those must establish ownership by path alone.
//
// DeleteStale turns deletion on. It requires FileMatcher, because a nil matcher matches every file in
// Dir.
//
// # One project per managed directory
//
// Everything written into a managed directory must go through a single [Project]. Reconciling part of
// it separately would let one pass delete files another pass is responsible for producing. When several
// generators, configurations or schemas contribute to one directory, they contribute to one [Output].
//
// The corollary is that two generators writing the same path is a collision rather than a silent
// last-writer-wins, and [Project.Plan] reports it.
//
// # Collecting output from a generator
//
// Generators expose their output in different ways, and the ones seen so far all reduce to filling an
// [Output]:
//
//   - A generator returning a map of path to content: hand it to [Output.AddAll].
//   - A generator returning file values that render on demand: render each and [Output.Add] it. Paths
//     may be absolute; Plan resolves them against Dir.
//   - A generator that writes through an interface of its own: implement that interface over an Output.
//     A generator library commonly has one, since writing to disk is rarely its only use.
//   - A generator with a pluggable assembler, such as gengo, whose FileType decides how an assembled
//     file is committed: register one that captures instead of writing.
//
// The last two are worth looking for before reaching for a scratch directory. A generator that can only
// write to disk has to be pointed at a temporary directory, and its output read back from there.
package codegenfiles
