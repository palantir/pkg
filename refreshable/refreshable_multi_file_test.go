// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	multiFile := NewMultiFileRefreshable(ctx, paths)
	vvv, _ := Map(multiFile, func(t map[string][]byte) [][]byte {
		var byteSlices [][]byte
		for _, v := range t {
			byteSlices = append(byteSlices, v)
		}
		return byteSlices
	})
	additionalByteSlice := New([]byte("additional"))
	done, _ := Merge(vvv, additionalByteSlice, func(t1 [][]byte, t2 []byte) [][]byte {
		all := [][]byte{}
		for _, v := range t1 {
			all = append(all, v)
		}
		all = append(all, t2)
		return all
	})
	for _, v := range done.Current() {
		fmt.Println(string(v))
	}

}
