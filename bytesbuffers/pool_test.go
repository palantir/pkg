// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytesbuffers_test

import (
	"strings"
	"testing"

	"github.com/palantir/pkg/bytesbuffers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_ProvidesResetBuffer(t *testing.T) {
	for name, poolProvider := range map[string]func() bytesbuffers.Pool{
		"SyncPool": func() bytesbuffers.Pool {
			return bytesbuffers.NewSyncPool(1)
		},
		"SizedPool": func() bytesbuffers.Pool {
			return bytesbuffers.NewSizedPool(1, 16)
		},
	} {
		t.Run(name, func(t *testing.T) {
			pool := poolProvider()

			buf := pool.Get()
			assert.Equal(t, 0, buf.Len())
			require.NoError(t, buf.WriteByte('a'))
			require.Equal(t, 1, buf.Len())

			pool.Put(buf)

			buf = pool.Get()
			assert.Equal(t, 0, buf.Len())
		})
	}
}

func TestPool_DiscardsLargeBuffer(t *testing.T) {
	const allocSize = 64
	for name, poolProvider := range map[string]func() bytesbuffers.Pool{
		"SyncPool": func() bytesbuffers.Pool {
			return bytesbuffers.NewSyncPool(allocSize)
		},
		"SizedPool": func() bytesbuffers.Pool {
			return bytesbuffers.NewSizedPool(1, allocSize)
		},
	} {
		t.Run(name, func(t *testing.T) {
			pool := poolProvider()
			buf := pool.Get()
			require.Equal(t, allocSize, buf.Cap())
			_, _ = buf.WriteString(strings.Repeat("0", 2*allocSize))
			require.GreaterOrEqual(t, buf.Cap(), 2*allocSize)
			pool.Put(buf)

			newBuf := pool.Get()
			require.False(t, buf == newBuf, "grown buffer should not have been reused")
			require.Equal(t, allocSize, newBuf.Cap())
		})
	}
}

func TestSizedPool_HasFixedSize(t *testing.T) {
	pool := bytesbuffers.NewSizedPool(2, 16)

	buf1 := pool.Get()
	buf2 := pool.Get()
	buf3 := pool.Get()

	pool.Put(buf1)
	assert.True(t, buf1 == pool.Get(), "expected buffer 1 to get reused")

	pool.Put(buf2)
	assert.True(t, buf2 == pool.Get(), "expected buffer 2 to get reused")

	pool.Put(buf1)
	pool.Put(buf2)
	pool.Put(buf3)

	assert.True(t, buf1 == pool.Get(), "expected buffer 1 to get reused")
	assert.True(t, buf2 == pool.Get(), "expected buffer 2 to get reused")
	assert.False(t, buf3 == pool.Get(), "expected buffer 3 to not be reused")
}
