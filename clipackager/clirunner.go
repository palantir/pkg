// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mholt/archives"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// NewDefaultPackagedCLIRunner returns a PackagedCLIRunner that uses opinionated defaults. cliName and cliVersion are
// the name and version of the CLI. The archiveBytes are bytes for the archive and archiveExtension is the file
// extension that represents the format of the archive.
//
// Assumes that the archive contains a top-level directory named "{cliName}-{cliVersion}", and that the CLI executable
// is at {cliName}-{cliVersion}/bin/{cliName}, with ".bat" appended if the GOOS is windows. The directory
// filepath.Join(os.TempDir(), "_"+cliName) is used as the package working directory.
func NewDefaultPackagedCLIRunner(
	cliName,
	cliVersion string,
	archiveBytes []byte,
	archiveExtension string,
) PackagedCLIRunner {
	cliDirName := fmt.Sprintf("%s-%s", cliName, cliVersion)
	return NewPackagedCLIRunner(
		cliDirName,
		filepath.Join(os.TempDir(), "_"+cliName),
		NewArchivePackagedCLIProviderFromBytes(
			archiveBytes,
			archiveExtension,
			AddExtensionForWindowsPathProvider(
				filepath.Join(cliDirName, "bin", cliName),
				".bat",
			),
		),
	)
}

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
func NewPackagedCLIRunner(name, pkgWorkDir string, cliProvider PackagedCLIProvider) PackagedCLIRunner {
	return &packgedCLIRunner{
		cliPkgName:  name,
		pkgWorkDir:  pkgWorkDir,
		cliProvider: cliProvider,
	}
}

// NewArchivePackagedCLIProviderFromBytes returns a PackagedCLIProvider that uses the provided bytes as the archive that contains the CLI.
func NewArchivePackagedCLIProviderFromBytes(archiveBytes []byte, archiveExtension string, pathToExecutableInArchive PathProviderFn) PackagedCLIProvider {
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

	// pkgWorkDir is the directory in which all package-related operations should occur. This directory may be used for
	// things like lock files and as a parent directory within which archives may be written or extracted, so it should
	// generally be assumed that the runner has full control over the directory. However, it may also make sense to have
	// this be somewhat stable (rather than fully random) so that CLIs that have been set up can be reused across runs.
	// For example, a value such as filepath.Join(os.TempDir(), "_conjureircli") is an example of such usage.
	pkgWorkDir string

	// cliProvider is the provider that extracts and writes the CLI if it does not exist.
	cliProvider PackagedCLIProvider
}

var _ PackagedCLIRunner = (*packgedCLIRunner)(nil)

func (r *packgedCLIRunner) EnsureCLIExistsAndReturnPath() (string, error) {
	cliPath, err := r.cliPath()
	if err != nil {
		return "", err
	}
	if err := r.ensureCLIExists(); err != nil {
		return "", err
	}
	return cliPath, nil
}

func (r *packgedCLIRunner) cliExtractDirPath() string {
	return filepath.Join(r.pkgWorkDir, r.cliPkgName+"-extract-dir")
}

func (r *packgedCLIRunner) cliPath() (string, error) {
	pathInExtractDir, err := r.cliProvider.PathInExtractDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(r.cliExtractDirPath(), pathInExtractDir), nil
}

// ensureCLIExists ensures that the CLI exists at the expected location. If the CLI exists at the expected location,
// returns nil without doing any work. If the CLI does not exist at the expected location, unarchives the CLI from the
// provider to ensure that it does exist at the expected location. Obtains and holds a global file-based lock based on
// the CLI package name that locks across different processes/executables. Returns an error if the CLI does not exist at
// the expected location and it was not possible to extract it.
func (r *packgedCLIRunner) ensureCLIExists() error {
	if err := os.MkdirAll(r.pkgWorkDir, 0755); err != nil {
		return fmt.Errorf("os.MkdirAll failed for package work directory %s: %w", r.pkgWorkDir, err)
	}
	installPkgLockFilePath := filepath.Join(r.pkgWorkDir, fmt.Sprintf("install-%s.lock", r.cliPkgName))
	installMutex := lockedfile.MutexAt(installPkgLockFilePath)
	unlockFn, err := installMutex.Lock()
	if err != nil {
		return fmt.Errorf("failed to lock file for installing CLI: %w", err)
	}
	defer unlockFn()

	// path to extracted CLI
	cliPath, err := r.cliPath()
	if err != nil {
		return err
	}

	// check if executable already exists: if so, return
	if checkNonEmptyFileExists(cliPath) == nil {
		return nil
	}

	// path to the directory into which CLI should be extracted. This is the base directory: extracting the CLI may
	// create further nested directories.
	cliExtractDir := r.cliExtractDirPath()

	// remove the extraction directory in case of a previous bad install
	if err := os.RemoveAll(cliExtractDir); err != nil {
		return fmt.Errorf("failed to remove destination dir %s: %w", cliExtractDir, err)
	}

	// extract the CLI into the destination directory
	if err := r.cliProvider.ExtractCLI(cliExtractDir); err != nil {
		return fmt.Errorf("failed to extract CLI into %s: %w", cliExtractDir, err)
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
	// ExtractCLI extracts the CLI into the provided directory. No file or directory should exist at the destination
	// path when this function is called, the path returned by filepath.Dir(destDir) must be valid and exist. The
	// operation should be implemented such that it either fully succeeds or does not create the directory at all.
	ExtractCLI(destDir string) error

	// PathInExtractDir returns the path to the CLI in the extraction directory. The CLI executable should exist at the
	// path returned by this function within destDir if ExtractCLI was called successfully with destDir. May return an
	// error if it is not possible to provide a path in the extracted directory for the CLI (for example, if running on
	// an OS/Architecture for which an executable does not exist).
	PathInExtractDir() (string, error)
}

var _ PackagedCLIProvider = (*archiveCLIProvider)(nil)

type archiveCLIProvider struct {
	// function that returns an io.ReadCloser that provides the bytes for the content of this runner.
	archiveByteProvider func() (io.ReadCloser, error)

	// the extension for the archive, including the dot. For example, ".tar.gz", ".tgz", ".zip", etc. Used if the
	// archive is written to disk or to help identify the archive format. Can be blank (in which case type of archive
	// is attempted to be identified based on content).
	archiveExtension string

	// returns the path to the CLI in the expanded archive.
	pathInExpandedArchive PathProviderFn
}

type PathProviderFn func() (string, error)

// StaticPathProvider returns a PathProviderFn that always returns the provided path.
func StaticPathProvider(fpath string) PathProviderFn {
	return func() (string, error) {
		return fpath, nil
	}
}

// AddExtensionForWindowsPathProvider returns a PathProviderFn that returns the provided path if the GOOS is not
// Windows; otherwise, returns the path with the provided extension appended. Example values for extension include
// ".exe" and ".bat".
func AddExtensionForWindowsPathProvider(fpath, extension string) PathProviderFn {
	return func() (string, error) {
		if runtime.GOOS == "windows" {
			return fpath + extension, nil
		}
		return fpath, nil
	}
}

func (p *archiveCLIProvider) PathInExtractDir() (string, error) {
	if p.pathInExpandedArchive == nil {
		return "", fmt.Errorf("pathInExpandedArchive function is nil")
	}
	return p.pathInExpandedArchive()
}

func (p *archiveCLIProvider) ExtractCLI(destDir string) error {
	// precondition of function is that destDir must not already exist
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("%s already exists", destDir)
	}

	// precondition of function is that parent of destDir must exist
	parentDir := filepath.Dir(destDir)
	if _, err := os.Stat(parentDir); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("parent directory %s does not exist", parentDir)
	}

	// Create a temporary directory into which archive is extracted.
	// Create the directory in parentDir to ensure that it is in the same volume as the destination.
	tmpExtractDir, err := os.MkdirTemp(parentDir, fmt.Sprintf("%s-tmp-*", filepath.Base(destDir)))
	if err != nil {
		return fmt.Errorf("failed to create temporary directory %s: %w", tmpExtractDir, err)
	}

	// extract archive into temporary extraction directory
	archiveReader, err := p.archiveByteProvider()
	if err != nil {
		return fmt.Errorf("failed to create archive byte reader: %w", err)
	}
	defer func() {
		_ = archiveReader.Close()
	}()
	if err := p.extractArchive(tmpExtractDir, archiveReader); err != nil {
		return fmt.Errorf("failed to extract CLI archive: %w", err)
	}

	// rename temporary extraction directory to actual destination directory.
	// Done so that this part of the operation is atomic and destination directory is either created with full archive
	// content or not created at all.
	if err := os.Rename(tmpExtractDir, destDir); err != nil {
		return fmt.Errorf("failed to move temporary extraction directory %s to %s: %w", tmpExtractDir, destDir, err)
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
		filenameForID = "archive-cli-provider" + p.archiveExtension
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
