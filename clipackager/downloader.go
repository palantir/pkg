// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
)

// EnsureFileWithSHA256ChecksumExists is a convenience function that wraps ensureFileWithChecksumExists. If the provided
// wantChecksum is non-empty, uses SHA256 as the hasher and the bytes obtained by hex-decoding the provided wantChecksum
// as the expected checksum. If the provided wantChecksum is the empty string, does not perform checksum validation.
func EnsureFileWithSHA256ChecksumExists(filepath, url, wantChecksum string) error {
	var verifier *ChecksumVerifier
	if wantChecksum != "" {
		var err error
		wantChecksumBytes, err := hex.DecodeString(wantChecksum)
		if err != nil {
			return fmt.Errorf("could not decode checksum %s: %w", wantChecksum, err)
		}
		verifier = &ChecksumVerifier{
			Hasher:       sha256.New(),
			WantChecksum: wantChecksumBytes,
		}
	}
	return ensureFileWithChecksumExists(filepath, url, verifier)
}

// ChecksumVerifier provides the information necessary to compute and verify a checksum.
type ChecksumVerifier struct {
	Hasher       hash.Hash
	WantChecksum []byte
}

// ensureFileWithChecksumExists ensures that a file exists at the provided filepath and, if the provided
// checksumVerifier is non-nil, verifies that the file's computed checksum matches the expected one. If a file with a
// matching checksum does not exist at the path, downloads the file from the provided url to filepath and, if the
// provided checksumVerifier is non-nil, verifies that the checksum matches. If checksumVerifier is nil, no checksum
// computation or verification is performed.
func ensureFileWithChecksumExists(filepath, url string, checksumVerifier *ChecksumVerifier) error {
	if exists, err := fileWithChecksumExists(filepath, checksumVerifier); err != nil {
		return err
	} else if exists {
		// file exists and checksum is valid
		return nil
	}

	// download file from URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status %d from GET to URL %s, got %d", http.StatusOK, url, resp.StatusCode)
	}

	reader := io.Reader(resp.Body)
	if checksumVerifier != nil {
		checksumVerifier.Hasher.Reset()
		reader = io.TeeReader(resp.Body, checksumVerifier.Hasher)
	}

	// create output file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, reader); err != nil {
		return err
	}

	if checksumVerifier != nil {
		if gotChecksum := checksumVerifier.Hasher.Sum(nil); !bytes.Equal(checksumVerifier.WantChecksum, gotChecksum) {
			return fmt.Errorf("file downloaded from %s had checksum %x, wanted %x", url, gotChecksum, checksumVerifier.WantChecksum)
		}
	}
	return nil
}

func fileWithChecksumExists(filepath string, checksumVerifier *ChecksumVerifier) (bool, error) {
	if fi, err := os.Stat(filepath); err != nil {
		return false, nil
	} else if mode := fi.Mode(); !mode.IsRegular() {
		return false, fmt.Errorf("%s is not a regular file: has mode %s", filepath, mode)
	}

	// if no verifier was provided, assume that existence of file is sufficient
	if checksumVerifier == nil {
		return true, nil
	}

	existing, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = existing.Close()
	}()

	checksumVerifier.Hasher.Reset()
	if _, err := io.Copy(checksumVerifier.Hasher, existing); err != nil {
		return false, err
	}
	if bytes.Equal(checksumVerifier.WantChecksum, checksumVerifier.Hasher.Sum(nil)) {
		// file exists and checksum matches
		return true, nil
	}
	// file exists, but checksum does not match
	return false, nil
}
