// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clipackager provides functions that help with downloading and packaging CLIs for Go programs.
//
// The EnsureFileWithSHA256ChecksumExists can be used in a file run by "go generate" to download a file or archive that
// contains a CLI that can be embedded in programs using the "embed" directive.
//
// The PackagedCLIProvider interface provides functions for extracting a CLI from an archive into a directory for use.
//
// The PackagedCLIRunner interface provides a function that ensures that a CLI exists using PackagedCLIProvider and
// returns its path. The functionality is implemented in such a way that the CLI should generally be cached across
// different runs and also uses file-based locking to ensure safety across processes.
package clipackager
