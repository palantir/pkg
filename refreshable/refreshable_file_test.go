// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileRefreshable(t *testing.T) {
	ctx := t.Context()

	dir := t.TempDir()
	filename := filepath.Join(dir, "file.txt")
	ticker := make(chan time.Time, 1)
	refreshableFile := NewFileRefreshableWithTicker(ctx, filename, ticker)
	// Assert we start with IsNotExist error.
	curr, err := refreshableFile.Validation()
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Empty(t, curr)
	require.Empty(t, refreshableFile.Current())

	// Create file.
	require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test", string(curr))
		require.Equal(t, "test", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)

	// Update file.
	require.NoError(t, os.WriteFile(filename, []byte("test2"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test2", string(curr))
		require.Equal(t, "test2", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)

	// Delete file.
	require.NoError(t, os.Remove(filename))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Empty(t, curr)
		require.Equal(t, "test2", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)

	// Create file.
	require.NoError(t, os.WriteFile(filename, []byte("test3"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test3", string(curr))
		require.Equal(t, "test3", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)
}

func TestNewMultiFileRefreshable(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()

	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	file3 := filepath.Join(dir, "file3.txt")

	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("content3"), 0644))

	paths := New(map[string]struct{}{
		file1: {},
		file2: {},
	})

	multiFile := NewMultiFileRefreshable(ctx, paths)

	// Verify initial contents
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Equal(t, []byte("content2"), current[file2])
		assert.Len(t, current, 2)
	}, 2*time.Second, 10*time.Millisecond)

	// Add a new file to the set
	paths.Update(map[string]struct{}{
		file1: {},
		file2: {},
		file3: {},
	})

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("content3"), current[file3])
		assert.Len(t, current, 3)
	}, 2*time.Second, 10*time.Millisecond)

	// Remove a file from the set
	paths.Update(map[string]struct{}{
		file1: {},
	})

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Len(t, current, 1)
	}, 2*time.Second, 10*time.Millisecond)

	// Update file contents and verify refresh picks it up
	require.NoError(t, os.WriteFile(file1, []byte("updated1"), 0644))

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("updated1"), current[file1])
	}, 2*time.Second, 10*time.Millisecond)
}

func TestNewMultiFileRefreshableValidationError(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()

	nonExistentFile := filepath.Join(dir, "does_not_exist.txt")

	paths := New(map[string]struct{}{
		nonExistentFile: {},
	})

	multiFile := NewMultiFileRefreshable(ctx, paths)

	// Validation should return an error for non-existent file
	_, err := multiFile.Validation()
	require.Error(t, err)
	require.True(t, errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT))

	// Current should return empty map (no successfully read files)
	current := multiFile.Current()
	require.Empty(t, current)
}

func TestNewMultiFileRefreshableCanMap(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
	paths := New(map[string]struct{}{
		file1: {},
		file2: {},
	})
	aggregateToList, _ := Map(NewMultiFileRefreshable(ctx, paths), func(t map[string][]byte) [][]byte {
		var byteSlices [][]byte
		for _, v := range t {
			byteSlices = append(byteSlices, v)
		}
		return byteSlices
	})
	additionalByteSlice := New([]byte("additional"))
	merged, _ := Merge(aggregateToList, additionalByteSlice, func(t1 [][]byte, t2 []byte) [][]byte {
		t1 = append(t1, t2)
		return t1
	})
	current := merged.Current()
	require.Equal(t, 3, len(current))
	assert.Equal(t, "content1", string(current[0]))
	assert.Equal(t, "content2", string(current[1]))
	assert.Equal(t, "additional", string(current[2]))
}
