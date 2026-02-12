// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build generate
// +build generate

// Written by Claude with prompts and iteration.
// Run using the command "go run create_archive.go".
// Creates the archives used for testing clirunner.

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	// Create single-executable.tgz with test-cli.sh at root
	if err := createArchive("single-executable.tgz", "test-cli.sh", "test-cli.sh"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating single-executable.tgz: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created single-executable.tgz")

	// Create single-dir.tgz with bin/test-cli.sh
	if err := createArchive("single-dir.tgz", "test-cli.sh", "bin/test-cli.sh"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating single-dir.tgz: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created single-dir.tgz")

	// Create full-dir-structure.tgz with hello-world-1.0.0 as top-level directory
	if err := createMultiFileArchive("full-dir-structure.tgz", "hello-world-1.0.0"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating full-dir-structure.tgz: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created full-dir-structure.tgz")

	// Create multi-top-level-dir.tgz without a top-level directory
	if err := createMultiFileArchive("multi-top-level-dir.tgz", ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating multi-top-level-dir.tgz: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created multi-top-level-dir.tgz")
}

func createArchive(outputFile, sourceFile, archivePath string) error {
	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add source file to the archive with the specified path
	if err := addFileToArchive(tarWriter, sourceFile, archivePath); err != nil {
		return fmt.Errorf("failed to add %s to archive: %w", sourceFile, err)
	}

	return nil
}

func createMultiFileArchive(outputFile, topLevelDir string) error {
	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Build paths with optional top-level directory
	binPath := "bin/test-cli.sh"
	jarPath := "jars/test-jar.txt"
	if topLevelDir != "" {
		binPath = topLevelDir + "/" + binPath
		jarPath = topLevelDir + "/" + jarPath
	}

	// Add test-cli.sh from file
	if err := addFileToArchive(tarWriter, "test-cli.sh", binPath); err != nil {
		return fmt.Errorf("failed to add test-cli.sh to archive: %w", err)
	}

	// Add test-jar.txt with content
	if err := addContentToArchive(tarWriter, []byte("test-jar-content"), jarPath, 0644); err != nil {
		return fmt.Errorf("failed to add test-jar.txt to archive: %w", err)
	}

	return nil
}

func addFileToArchive(tarWriter *tar.Writer, filename, archivePath string) error {
	// Open the file to be archived
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for the header
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Create tar header
	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}

	// Use the specified archive path for the header name
	header.Name = archivePath

	// Write header to tar archive
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Copy file content to tar archive
	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

func addContentToArchive(tarWriter *tar.Writer, content []byte, archivePath string, mode int64) error {
	// Create tar header
	header := &tar.Header{
		Name: archivePath,
		Mode: mode,
		Size: int64(len(content)),
	}

	// Write header to tar archive
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Write content to tar archive
	if _, err := tarWriter.Write(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}
