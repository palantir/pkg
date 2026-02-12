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

// EnsureFileWithSHA256ChecksumExists is a convenience function that wraps ensureFileWithChecksumExists. It uses SHA256
// as the hasher and the bytes obtained by hex-decoding the provided wantChecksum as the expected checksum.
func EnsureFileWithSHA256ChecksumExists(filepath, url, wantChecksum string) error {
	wantChecksumBytes, err := hex.DecodeString(wantChecksum)
	if err != nil {
		return fmt.Errorf("could not decode checksum %s: %w", wantChecksum, err)
	}
	return ensureFileWithChecksumExists(filepath, url, sha256.New(), wantChecksumBytes)
}

// ensureFileWithChecksumExists ensures that a file exists at the provided filepath and that the hash computed for that
// file using the provided hasher matches the provided wantChecksum. If such a file does not exist, downloads the file
// from the provided url to filepath and verifies that the checksum matches.
func ensureFileWithChecksumExists(filepath, url string, hasher hash.Hash, wantChecksum []byte) error {
	if exists, err := fileWithChecksumExists(filepath, hasher, wantChecksum); err != nil {
		return err
	} else if exists {
		// file exists and checksum is valid
		return nil
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	hasher.Reset()
	reader := io.TeeReader(resp.Body, hasher)
	if _, err := io.Copy(out, reader); err != nil {
		return err
	}

	if gotChecksum := hasher.Sum(nil); !bytes.Equal(wantChecksum, gotChecksum) {
		return fmt.Errorf("file downloaded from %s had checksum %x, wanted %x", url, gotChecksum, wantChecksum)
	}
	return nil
}

func fileWithChecksumExists(filepath string, hasher hash.Hash, wantChecksum []byte) (bool, error) {
	if hasher == nil {
		return false, fmt.Errorf("hasher cannot be nil")
	}

	if _, err := os.Stat(filepath); err != nil {
		return false, nil
	}

	existing, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = existing.Close()
	}()

	hasher.Reset()
	if _, err := io.Copy(hasher, existing); err != nil {
		return false, err
	}
	if bytes.Equal(wantChecksum, hasher.Sum(nil)) {
		// file exists and checksum matches
		return true, nil
	}
	return false, nil
}
