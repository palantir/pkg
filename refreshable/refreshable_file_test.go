// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"os"
	"path/filepath"
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
