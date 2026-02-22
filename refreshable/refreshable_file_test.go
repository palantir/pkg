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
	require.Empty(t, refreshableFile.LastCurrent())

	// Create file.
	require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test", string(curr))
		require.Equal(t, "test", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)

	// Update file.
	require.NoError(t, os.WriteFile(filename, []byte("test2"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test2", string(curr))
		require.Equal(t, "test2", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)

	// Delete file.
	require.NoError(t, os.Remove(filename))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Empty(t, curr)
		require.Equal(t, "test2", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)

	// Create file.
	require.NoError(t, os.WriteFile(filename, []byte("test3"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "test3", string(curr))
		require.Equal(t, "test3", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)
}

func TestNewFileRefreshableWithReaderFunc(t *testing.T) {
	ctx := t.Context()

	dir := t.TempDir()
	filename := filepath.Join(dir, "file.txt")
	ticker := make(chan time.Time, 1)

	// Custom reader that uppercases the content
	uppercaseReader := func(path string) ([]byte, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		// Convert to uppercase
		for i := range content {
			if content[i] >= 'a' && content[i] <= 'z' {
				content[i] = content[i] - 'a' + 'A'
			}
		}
		return content, nil
	}

	refreshableFile := NewFileRefreshableWithReaderFunc(ctx, filename, ticker, uppercaseReader)

	// Assert we start with IsNotExist error.
	curr, err := refreshableFile.Validation()
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Empty(t, curr)
	require.Empty(t, refreshableFile.LastCurrent())

	// Create file with lowercase content.
	require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "TEST", string(curr)) // Should be uppercased
		require.Equal(t, "TEST", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)

	// Update file.
	require.NoError(t, os.WriteFile(filename, []byte("hello world"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "HELLO WORLD", string(curr)) // Should be uppercased
		require.Equal(t, "HELLO WORLD", string(refreshableFile.LastCurrent()))
	}, time.Second, 10*time.Millisecond)

	// Delete file - Current() should retain last valid value.
	require.NoError(t, os.Remove(filename))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Empty(t, curr)
		require.Equal(t, "HELLO WORLD", string(refreshableFile.LastCurrent()))
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
	paths := New(map[string]struct{}{file1: {}, file2: {}})
	multiFile := NewMultiFileRefreshable(ctx, paths)
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.LastCurrent()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Equal(t, []byte("content2"), current[file2])
		assert.Len(t, current, 2)
	}, 2*time.Second, 10*time.Millisecond)
	paths.Update(map[string]struct{}{file1: {}, file2: {}, file3: {}})
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.LastCurrent()
		assert.Equal(t, []byte("content3"), current[file3])
		assert.Len(t, current, 3)
	}, 2*time.Second, 10*time.Millisecond)
	paths.Update(map[string]struct{}{file1: {}})
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.LastCurrent()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Len(t, current, 1)
	}, 2*time.Second, 10*time.Millisecond)
	require.NoError(t, os.WriteFile(file1, []byte("updated1"), 0644))
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.LastCurrent()
		assert.Equal(t, []byte("updated1"), current[file1])
	}, 2*time.Second, 10*time.Millisecond)
}

func TestNewMultiFileRefreshableValidationError(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	nonExistentFile := filepath.Join(dir, "does_not_exist.txt")
	paths := New(map[string]struct{}{nonExistentFile: {}})
	multiFile := NewMultiFileRefreshable(ctx, paths)
	_, err := multiFile.Validation()
	require.Error(t, err)
	require.True(t, errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT))
	current := multiFile.LastCurrent()
	require.Empty(t, current)
}

func TestNewMultiFileRefreshableCanMap(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
	paths := New(map[string]struct{}{file1: {}, file2: {}})
	aggregateToList, _, err := MapValidated(NewMultiFileRefreshable(ctx, paths), func(t map[string][]byte) ([][]byte, error) {
		var byteSlices [][]byte
		for _, v := range t {
			byteSlices = append(byteSlices, v)
		}
		return byteSlices, nil
	})
	assert.NoError(t, err)
	additionalByteSlice, _, _ := MapWithError(New([]byte("additional")), func(a []byte) ([]byte, error) {
		return a, nil
	})
	merged, _ := MergeValidated(aggregateToList, additionalByteSlice, func(t1 [][]byte, t2 []byte) [][]byte {
		return append(t1, t2)
	})
	current := merged.Current()
	require.Equal(t, 3, len(current))
	// Map iteration order is non-deterministic, so check contents without assuming order
	contents := make(map[string]bool)
	for _, v := range current {
		contents[string(v)] = true
	}
	assert.True(t, contents["content1"])
	assert.True(t, contents["content2"])
	assert.True(t, contents["additional"])
}
