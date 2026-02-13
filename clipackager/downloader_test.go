// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureFileWithSHA256ChecksumExists verifies the behavior of EnsureFileWithSHA256ChecksumExists.
// Test code written by Claude with guided prompts and iteration.
func TestEnsureFileWithSHA256ChecksumExists(t *testing.T) {
	const (
		testContentString   = "test file content for download"
		testContentChecksum = "10590227a96a18687944335ea90b4315597bbe943cf280454330d55cad933107"
	)

	// Test data
	testContent := []byte(testContentString)

	// Calculate expected checksum
	hasher := sha256.New()
	hasher.Write(testContent)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))
	if expectedChecksum != testContentChecksum {
		t.Fatalf("SHA256 checksum does not match. Expected %s, got %s", expectedChecksum, testContentChecksum)
	}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(testContent)
	}))
	defer server.Close()

	t.Run("downloads and checksums file successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-download.bin")

		err := EnsureFileWithSHA256ChecksumExists(testFile, server.URL, expectedChecksum)
		if err != nil {
			t.Fatalf("failed to download file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testFile); err != nil {
			t.Fatalf("file does not exist after download: %v", err)
		}

		// Verify file content
		actualContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read downloaded file: %v", err)
		}

		if string(actualContent) != string(testContent) {
			t.Errorf("file content mismatch:\nwant: %q\ngot:  %q", string(testContent), string(actualContent))
		}
	})

	t.Run("does not re-download when file with correct checksum exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-existing.bin")

		// Create file with correct checksum
		if err := os.WriteFile(testFile, testContent, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Track if server is called
		serverCalled := false
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(testContent)
		}))
		defer testServer.Close()

		err := EnsureFileWithSHA256ChecksumExists(testFile, testServer.URL, expectedChecksum)
		if err != nil {
			t.Fatalf("failed when file with correct checksum exists: %v", err)
		}

		if serverCalled {
			t.Error("server was called when file with correct checksum already exists")
		}

		// Verify file content is unchanged
		actualContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(actualContent) != string(testContent) {
			t.Errorf("file content changed:\nwant: %q\ngot:  %q", string(testContent), string(actualContent))
		}
	})

	t.Run("overwrites file with incorrect checksum", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-incorrect.bin")

		// Create file with incorrect content
		incorrectContent := []byte("incorrect content")
		if err := os.WriteFile(testFile, incorrectContent, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		err := EnsureFileWithSHA256ChecksumExists(testFile, server.URL, expectedChecksum)
		if err != nil {
			t.Fatalf("failed to overwrite file with incorrect checksum: %v", err)
		}

		// Verify file was overwritten with correct content
		actualContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(actualContent) != string(testContent) {
			t.Errorf("file was not overwritten correctly:\nwant: %q\ngot:  %q", string(testContent), string(actualContent))
		}
	})

	t.Run("returns error when directory exists at file path", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-dir")

		// Create a directory at the file path
		if err := os.Mkdir(testFile, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		err := EnsureFileWithSHA256ChecksumExists(testFile, server.URL, expectedChecksum)
		if err == nil {
			t.Fatal("expected error when directory exists at file path, got nil")
		}

		// Verify exact error message
		expectedErr := fmt.Sprintf("%s is not a regular file: has mode %s", testFile, os.ModeDir|0755)
		if err.Error() != expectedErr {
			t.Errorf("error message mismatch:\nwant: %q\ngot:  %q", expectedErr, err.Error())
		}
	})

	t.Run("returns error when downloaded content has incorrect checksum", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-bad-checksum.bin")

		// Create server that returns different content than expected
		badContent := []byte("different content from server")
		badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(badContent)
		}))
		defer badServer.Close()

		err := EnsureFileWithSHA256ChecksumExists(testFile, badServer.URL, expectedChecksum)
		if err == nil {
			t.Fatal("expected error when downloaded content has incorrect checksum, got nil")
		}

		// Calculate actual checksum of bad content
		badHasher := sha256.New()
		badHasher.Write(badContent)
		badChecksum := badHasher.Sum(nil)

		// Decode expected checksum bytes
		wantChecksumBytes, _ := hex.DecodeString(expectedChecksum)

		// Verify exact error message
		expectedErr := fmt.Sprintf("file downloaded from %s had checksum %x, wanted %x", badServer.URL, badChecksum, wantChecksumBytes)
		if err.Error() != expectedErr {
			t.Errorf("error message mismatch:\nwant: %q\ngot:  %q", expectedErr, err.Error())
		}
	})

	t.Run("returns error when HTTP server responds with non-200 status", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-http-error.bin")

		// Create server that returns 404 Not Found
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer errorServer.Close()

		err := EnsureFileWithSHA256ChecksumExists(testFile, errorServer.URL, expectedChecksum)
		if err == nil {
			t.Fatal("expected error when HTTP server responds with non-200 status, got nil")
		}

		// Verify exact error message
		expectedErr := fmt.Sprintf("expected status %d from GET to URL %s, got %d", http.StatusOK, errorServer.URL, http.StatusNotFound)
		if err.Error() != expectedErr {
			t.Errorf("error message mismatch:\nwant: %q\ngot:  %q", expectedErr, err.Error())
		}

		// Verify file was not created
		if _, err := os.Stat(testFile); err == nil {
			t.Error("file should not exist after failed download")
		}
	})
}
