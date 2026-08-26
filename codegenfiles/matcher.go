// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegenfiles

import (
	"regexp"
)

// ContentsMatcher reports whether a file's contents identify it as one that a
// generator owns. It is used to distinguish generated files from hand-written
// files that have been placed in a directory otherwise managed by a generator,
// so that only generated files are considered for deletion.
type ContentsMatcher interface {
	MatchFileContents(contents []byte) bool
}

// ContentsMatcherFunc adapts an ordinary function to a ContentsMatcher.
type ContentsMatcherFunc func(contents []byte) bool

func (f ContentsMatcherFunc) MatchFileContents(contents []byte) bool {
	return f(contents)
}

// generatedCodeLine matches the standard generated-code marker, which appears on a
// line by itself: "// Code generated <generator> DO NOT EDIT.". See
// https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source.
//
// The marker is recognized behind any of three comment syntaxes, because a generator
// commonly emits more than one kind of file into a directory and a project has a
// single ContentsMatcher for all of them: "//" for Go and other C-family languages,
// "#" for YAML, shell and the like, and "<!--  -->" for Markdown and HTML.
//
// The (?m) flag anchors ^ and $ to line boundaries so the marker is recognized on its
// own line wherever it appears in the file. The optional \r accepts CRLF line endings:
// RE2's $ matches only before \n, so without it a checkout with CRLF endings would
// contain no recognizably generated files at all.
var generatedCodeLine = regexp.MustCompile(`(?m)^(?://|#|<!--) Code generated .* DO NOT EDIT\.[ \t]*(?:-->)?[ \t]*\r?$`)

// GeneratedCodeMatcher returns a ContentsMatcher that matches files containing the
// standard generated-code marker line, in any of the comment syntaxes described on
// generatedCodeLine. Any generator that emits the marker is recognized without
// additional configuration.
//
// A file format that cannot carry a comment at all, such as JSON, can never be
// identified this way; a project generating those must rely on FileMatcher to
// establish ownership by path instead.
func GeneratedCodeMatcher() ContentsMatcher {
	return ContentsMatcherFunc(generatedCodeLine.Match)
}
