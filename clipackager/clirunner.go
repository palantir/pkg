// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mholt/archives"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// RunPackagedCLI is a convenience function that runs the CLI executable provided by packgedCLIRunner using the provided
// arguments. Returns the path to the executable and the combined output of stdout and stderr.
func RunPackagedCLI(cliRunner PackagedCLIRunner, args ...string) (string, []byte, error) {
	executablePath, err := cliRunner.EnsureCLIExistsAndReturnPath()
	if err != nil {
		return executablePath, nil, err
	}
	output, err := exec.Command(executablePath, args...).CombinedOutput()
	return executablePath, output, err
}

type PackagedCLIRunner interface {
	// EnsureCLIExistsAndReturnPath ensures the CLI exists at the expected path and returns the path to the executable
	// for the CLI. Returns an error if the CLI does not exist or there is an error creating the CLI.
	EnsureCLIExistsAndReturnPath() (string, error)
}

// NewPackagedCLIRunner returns a new PackagedCLIRunner that uses the provided parameters.
func NewPackagedCLIRunner(name, workdir string, cliProvider PackagedCLIProvider) PackagedCLIRunner {
	return &packgedCLIRunner{
		cliPkgName:  name,
		workDir:     workdir,
		cliProvider: cliProvider,
	}
}

// NewArchivePackagedCLIProviderFromBytes returns a PackagedCLIProvider that uses the provided bytes as the archive that contains the CLI.
func NewArchivePackagedCLIProviderFromBytes(archiveBytes []byte, archiveExtension string, pathToExecutableInArchive string) PackagedCLIProvider {
	return &archiveCLIProvider{
		archiveByteProvider:   func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(archiveBytes)), nil },
		archiveExtension:      archiveExtension,
		pathInExpandedArchive: pathToExecutableInArchive,
	}
}

// packgedCLIRunner is a CLI runner that runs a CLI that is provided by a PackagedCLIProvider. If the CLI exists at the
// expected location, the runner runs it; otherwise, it extracts the CLI to the expected destination using the
// PackagedCLIProvider and then runs it.
type packgedCLIRunner struct {
	// name of the packaged CLI. Used as part of the name for the directory into which the package is expanded, so it
	// should be unique per package. Typically, the value is something like "{name}-{version}" (for example, "conjure-4.35.0").
	cliPkgName string

	// workDir is the base directory in which all operations should occur. This directory may be used for things like
	// lock files and as a parent directory within which archives may be downloaded or unpacked, so it should generally
	// be assumed that the runner has full control over the directory. However, it may also make sense to have this be
	// somewhat stable (rather than fully random) so that CLIs that have been set up can be reused across runs. For
	// example, a value such as filepath.Join(os.TempDir(), "_conjureircli") is an example of such usage.
	workDir string

	// cliProvider is the provider that extracts and writes the CLI if it does not exist.
	cliProvider PackagedCLIProvider
}

var _ PackagedCLIRunner = (*packgedCLIRunner)(nil)

func (r *packgedCLIRunner) EnsureCLIExistsAndReturnPath() (string, error) {
	return r.cliPath(), r.ensureCLIExists()
}

func (r *packgedCLIRunner) cliExtractDirPath() string {
	return filepath.Join(r.workDir, r.cliPkgName+"-extract-dir")
}

func (r *packgedCLIRunner) cliPath() string {
	return filepath.Join(r.cliExtractDirPath(), r.cliProvider.PathInExtractDir())
}

// ensureCLIExists ensures that the CLI exists at the expected location. If the CLI exists at the expected location,
// returns nil without doing any work. If the CLI does not exist at the expected location, unarchives the CLI from the
// provider to ensure that it does exist at the expected location. Obtains and holds a global file-based lock based on
// the CLI package name that locks across different processes/executables. Returns an error if the CLI does not exist at
// the expected location and it was not possible to extract it.
func (r *packgedCLIRunner) ensureCLIExists() error {
	installPkgLockFilePath := filepath.Join(r.workDir, fmt.Sprintf("install-%s.lock", r.cliPkgName))
	installMutex := lockedfile.MutexAt(installPkgLockFilePath)
	unlockFn, err := installMutex.Lock()
	if err != nil {
		return fmt.Errorf("failed to lock mutex for installing CLI: %w", err)
	}
	defer unlockFn()

	// path to extracted CLI
	cliPath := r.cliPath()
	if checkNonEmptyFileExists(cliPath) == nil {
		// executable already exists
		return nil
	}

	// path to the directory into which CLI should be extracted. This is the base directory: extracting the CLI may
	// create further nested directories.
	cliExtractDir := r.cliExtractDirPath()

	// remove the CLI dir just in case of a previous bad install
	if err := os.RemoveAll(cliExtractDir); err != nil {
		return fmt.Errorf("failed to remove destination dir %s: %w", cliExtractDir, err)
	}

	// extract the CLI into the destination directory
	if err := r.cliProvider.ExtractCLI(cliExtractDir); err != nil {
		return fmt.Errorf("failed to extract CLI int %s: %w", cliExtractDir, err)
	}

	// check that we can now find the CLI
	if err := checkNonEmptyFileExists(cliPath); err != nil {
		return fmt.Errorf("CLI does not exist at %s after extracting: %w", cliPath, err)
	}

	return nil
}

// PackagedCLIProvider provides a CLI. Supports extracting a CLI into a destination directory and returning the path in that
// directory to the executable CLI.
type PackagedCLIProvider interface {
	// ExtractCLI extracts the CLI into the provided directory.
	ExtractCLI(destDir string) error

	// PathInExtractDir returns the path to the CLI in the extraction directory. The CLI executable should exist at the
	// path returned by this function within destDir if ExtractCLI was called successfully with destDir.
	PathInExtractDir() string
}

var _ PackagedCLIProvider = (*archiveCLIProvider)(nil)

type archiveCLIProvider struct {
	// function that returns an io.ReadCloser that provides the bytes for the content of this runner.
	archiveByteProvider func() (io.ReadCloser, error)

	// the extension for the archive, including the dot. For example, ".tar.gz", ".tgz", ".zip", etc. Used if the
	// archive is written to disk or to help identify the archive format. Can be blank (in which case type of archive
	// is attempted to be identified based on content).
	archiveExtension string

	// path in the expanded archive in which the CLI exists.
	pathInExpandedArchive string
}

func (p *archiveCLIProvider) PathInExtractDir() string {
	return p.pathInExpandedArchive
}

func (p *archiveCLIProvider) ExtractCLI(destDir string) error {
	archiveReader, err := p.archiveByteProvider()
	if err != nil {
		return fmt.Errorf("failed to create archive byte reader: %w", err)
	}
	defer func() {
		_ = archiveReader.Close()
	}()
	// extract archive to destination
	if err := p.extractArchive(destDir, archiveReader); err != nil {
		return errors.Wrap(err, "failed to extract CLI archive")
	}
	return nil
}

// extractArchive extracts the archive stored in the packgedCLIRunner to the provided dstDir.
func (p *archiveCLIProvider) extractArchive(dstDir string, archiveReader io.Reader) error {
	extractor, archiveReader, err := p.getArchiveExtractor(archiveReader)
	if err != nil {
		return err
	}

	if err := extractor.Extract(context.Background(), archiveReader, func(ctx context.Context, fi archives.FileInfo) error {
		currPath := filepath.Join(dstDir, fi.NameInArchive)

		// Handle directories
		if fi.IsDir() {
			if err := os.MkdirAll(currPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", currPath, err)
			}
			return nil
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(currPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", currPath, err)
		}

		// Create output file
		outFile, err := os.OpenFile(currPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %q: %w", currPath, err)
		}
		defer func() {
			_ = outFile.Close()
		}()

		// Copy file contents from archive to output file
		archiveFile, err := fi.Open()
		if err != nil {
			return fmt.Errorf("failed to open archive file %q: %w", currPath, err)
		}
		defer func() {
			_ = archiveFile.Close()
		}()

		if _, err := io.Copy(outFile, archiveFile); err != nil {
			return fmt.Errorf("failed to copy archive file %q: %w", currPath, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	return nil
}

func (p *archiveCLIProvider) getArchiveExtractor(archiveReader io.Reader) (archives.Extractor, io.Reader, error) {
	var filenameForID string
	if p.archiveExtension != "" {
		filenameForID = "archive-cli-provider." + p.archiveExtension
	}

	var (
		archiveFormat archives.Format
		err           error
	)
	archiveFormat, archiveReader, err = archives.Identify(context.Background(), filenameForID, archiveReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to identify archive format: %w", err)
	}

	extractor, ok := archiveFormat.(archives.Extractor)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported archive format (cannot be converted to an archiver.Extractor): %s", archiveFormat)
	}
	return extractor, archiveReader, nil
}

// checkNonEmptyFileExists returns an error if the file at fpath is not a regular file with nonzero size.
func checkNonEmptyFileExists(fpath string) error {
	fi, err := os.Stat(fpath)
	switch {
	case err != nil:
		return err
	case fi.Size() == 0:
		return fmt.Errorf("file was empty")
	case !fi.Mode().IsRegular():
		return fmt.Errorf("file mode %s was unexpected", fi.Mode().String())
	}
	return nil
}
