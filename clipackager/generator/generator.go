// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const VersionVar = "{{VERSION}}"

// DownloadFile downloads the file at the provided urlTemplate rendered with the provided version to a file at the
// provided destinationPath and verifies that the file matches the provided expectedChecksum. If a file with the
// provided checksum already exists at the provided path, this function does not do anything. The URL template rendering
// consists of replacing any occurrences of the string value of VersionVar with the provided version.
func DownloadFile(destinationFilePath, urlTemplate, version, expectedChecksum string) error {
	renderedURL := strings.ReplaceAll(urlTemplate, "{{VERSION}}", version)
	if err := downloadFile(destinationFilePath, renderedURL, expectedChecksum); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

func downloadFile(filepath, url, expectedChecksum string) error {
	// if file exists and checksum matches, return nil without doing anything;
	// otherwise, download file from URL
	if _, err := os.Stat(filepath); err == nil {
		computedChecksum, err := computeSHA256ChecksumForFile(filepath)
		if err != nil {
			return err
		}
		if computedChecksum == expectedChecksum {
			// existing file up to date
			return nil
		}
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer func() {
		_ = out.Close()
	}()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// teeReader will write whatever bytes are read to the output file
	teeReader := io.TeeReader(resp.Body, out)
	// computing the checksum from the reader reads all file bytes into the hasher, and since the reader is a TeeReader,
	// the bytes that are read for computing the checksum are written to the output file as well.
	computedChecksum, err := computeSHA256ChecksumForReader(teeReader)
	if err != nil {
		return fmt.Errorf("error computing checksum: %w", err)
	}

	if computedChecksum != expectedChecksum {
		return fmt.Errorf(
			"computed checksum for downloaded file (%s) does not match expected checksum (%s)",
			computedChecksum,
			expectedChecksum,
		)
	}
	return nil
}

func computeSHA256ChecksumForFile(filepath string) (string, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer func() {
		_ = file.Close()
	}()
	return computeSHA256ChecksumForReader(file)
}

func computeSHA256ChecksumForReader(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("failed to copy bytes to compute hash: %w", err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
