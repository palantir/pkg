// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testStr1              = "renderConf1"
	testStr2              = "renderConf2"
	refreshableSyncPeriod = time.Millisecond * 100
	sleepPeriod           = time.Millisecond * 175
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
	require.Empty(t, refreshableFile.Current())

	// Create file with lowercase content.
	require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "TEST", string(curr)) // Should be uppercased
		require.Equal(t, "TEST", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)

	// Update file.
	require.NoError(t, os.WriteFile(filename, []byte("hello world"), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.NoError(t, err)
		require.Equal(t, "HELLO WORLD", string(curr)) // Should be uppercased
		require.Equal(t, "HELLO WORLD", string(refreshableFile.Current()))
	}, time.Second, 10*time.Millisecond)

	// Delete file - Current() should retain last valid value.
	require.NoError(t, os.Remove(filename))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		curr, err := refreshableFile.Validation()
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Empty(t, curr)
		require.Equal(t, "HELLO WORLD", string(refreshableFile.Current()))
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
		current := multiFile.Current()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Equal(t, []byte("content2"), current[file2])
		assert.Len(t, current, 2)
	}, 2*time.Second, 10*time.Millisecond)
	paths.Update(map[string]struct{}{file1: {}, file2: {}, file3: {}})
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("content3"), current[file3])
		assert.Len(t, current, 3)
	}, 2*time.Second, 10*time.Millisecond)
	paths.Update(map[string]struct{}{file1: {}})
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		current := multiFile.Current()
		assert.Equal(t, []byte("content1"), current[file1])
		assert.Len(t, current, 1)
	}, 2*time.Second, 10*time.Millisecond)
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
	paths := New(map[string]struct{}{nonExistentFile: {}})
	multiFile := NewMultiFileRefreshable(ctx, paths)
	_, err := multiFile.Validation()
	require.Error(t, err)
	require.True(t, errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT))
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
	paths := New(map[string]struct{}{file1: {}, file2: {}})
	aggregateToList, _ := Map(NewMultiFileRefreshable(ctx, paths), func(t map[string][]byte) [][]byte {
		var byteSlices [][]byte
		for _, v := range t {
			byteSlices = append(byteSlices, v)
		}
		return byteSlices
	})
	additionalByteSlice := New([]byte("additional"))
	merged, _ := Merge(aggregateToList, additionalByteSlice, func(t1 [][]byte, t2 []byte) [][]byte {
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

// Verifies that a RefreshableFile can follow a symlink when the original file updates
// We ensure we wait long enough (80ms) such that the refreshable duration (50ms) will read the change
func TestRefreshableFileCanFollowSymLink(t *testing.T) {
	tempDir := t.TempDir()
	// We will have fileToWritePointingAtActual point at fileToWriteActual
	fileToWriteActual := filepath.Join(tempDir, "fileToWriteActual")
	fileToWritePointingAtActual := filepath.Join(tempDir, "fileToWritePointingAtActual")
	// Write the old file
	require.NoError(t, os.WriteFile(fileToWriteActual, []byte(testStr1), 0644))
	// Symlink the old file to point at the new file
	err := os.Symlink(fileToWriteActual, fileToWritePointingAtActual)
	assert.NoError(t, err)
	// Point the refreshable towards the new file
	r := NewFileRefreshableWithTicker(context.Background(), fileToWritePointingAtActual, time.Tick(refreshableSyncPeriod))
	str := getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf1")
	// Update the actual file
	require.NoError(t, os.WriteFile(fileToWriteActual, []byte(testStr2), 0644))
	time.Sleep(sleepPeriod)
	str = getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf2")
}

// Verifies that a RefreshableFile can follow a symlink to a symlink when the original file updates
// We ensure we wait long enough (80ms) such that the refreshable duration (50ms) will read the change
func TestRefreshableFileCanFollowMultipleSymLinks(t *testing.T) {
	tempDir := t.TempDir()
	// We will have fileToWritePointingAtActual point at fileToWriteActual
	fileToWriteActual := filepath.Join(tempDir, "fileToWriteActual")
	fileToWritePointingAtActual := filepath.Join(tempDir, "fileToWritePointingAtActual")
	fileToWritePointingAtSymlink := filepath.Join(tempDir, "fileToWritePointingAtSymlink")
	// Write the old file
	require.NoError(t, os.WriteFile(fileToWriteActual, []byte(testStr1), 0644))
	// Symlink the old file to point at the new file
	err := os.Symlink(fileToWriteActual, fileToWritePointingAtActual)
	assert.NoError(t, err)
	// Symlink a to a symlink
	err = os.Symlink(fileToWritePointingAtActual, fileToWritePointingAtSymlink)
	assert.NoError(t, err)
	// Point the refreshable towards the new file
	r := NewFileRefreshableWithTicker(context.Background(), fileToWritePointingAtSymlink, time.Tick(refreshableSyncPeriod))
	str := getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf1")
	// Update the symlink file
	require.NoError(t, os.WriteFile(fileToWriteActual, []byte(testStr2), 0644))
	time.Sleep(sleepPeriod)
	str = getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf2")
}

// Verifies that a RefreshableFile can follow a symlink to a symlink when the symlink changes
// We ensure we wait long enough (80ms) such that the refreshable duration (50ms) will read the change
func TestRefreshableFileCanFollowMovingSymLink(t *testing.T) {
	tempDir := t.TempDir()
	// We will have fileToWritePointingAtActual point at fileToWriteActual
	fileToWriteActualOriginal := filepath.Join(tempDir, "fileToWriteActualOriginal")
	fileToWriteActualUpdated := filepath.Join(tempDir, "fileToWriteActualUpdated")
	fileToWritePointingAtActual := filepath.Join(tempDir, "fileToWritePointingAtActual")
	fileToWritePointingAtSymlink := filepath.Join(tempDir, "fileToWritePointingAtSymlink")
	// Write the old file
	require.NoError(t, os.WriteFile(fileToWriteActualOriginal, []byte(testStr1), 0644))
	// Write the old file
	require.NoError(t, os.WriteFile(fileToWriteActualUpdated, []byte(testStr2), 0644))
	// Symlink the old file to point at the new file
	err := os.Symlink(fileToWriteActualOriginal, fileToWritePointingAtActual)
	assert.NoError(t, err)
	// Symlink a to a symlink
	err = os.Symlink(fileToWritePointingAtActual, fileToWritePointingAtSymlink)
	assert.NoError(t, err)
	// Point the refreshable towards the new file
	r := NewFileRefreshableWithTicker(context.Background(), fileToWritePointingAtSymlink, time.Tick(refreshableSyncPeriod))
	str := getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf1")
	// Change where the symlink points
	err = os.Remove(fileToWritePointingAtActual)
	assert.NoError(t, err)
	err = os.Symlink(fileToWriteActualUpdated, fileToWritePointingAtActual)
	assert.NoError(t, err)

	// Update the symlink file
	time.Sleep(sleepPeriod)
	str = getStringFromRefreshable(t, r)
	assert.Equal(t, str, "renderConf2")
}

func getStringFromRefreshable(t *testing.T, r Refreshable[[]byte]) string {
	return string(r.Current())
}
