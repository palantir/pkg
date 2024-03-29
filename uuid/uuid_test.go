// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid_test

import (
	"encoding/json"
	"testing"

	"github.com/palantir/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUUID = uuid.UUID{
	0x0, 0x1, 0x2, 0x3,
	0x4, 0x5, 0x6, 0x7,
	0x8, 0x9, 0xA, 0xB,
	0xC, 0xD, 0xE, 0xF,
}

func TestUUID_MarshalJSON(t *testing.T) {
	marshalledUUID, err := json.Marshal(testUUID)
	assert.NoError(t, err)
	assert.Equal(t, `"00010203-0405-0607-0809-0a0b0c0d0e0f"`, string(marshalledUUID))
}

func TestUUID_MarshalBinary(t *testing.T) {
	marshalledUUID, err := testUUID.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, marshalledUUID)
}

func TestUUID_UnmarshalJSON(t *testing.T) {
	t.Run("correct lower case", func(t *testing.T) {
		var actual uuid.UUID
		err := json.Unmarshal([]byte(`"00010203-0405-0607-0809-0a0b0c0d0e0f"`), &actual)
		assert.NoError(t, err)
		assert.Equal(t, testUUID, actual)
	})

	t.Run("correct upper case", func(t *testing.T) {
		var actual uuid.UUID
		err := json.Unmarshal([]byte(`"00010203-0405-0607-0809-0A0B0C0D0E0F"`), &actual)
		assert.NoError(t, err)
		assert.Equal(t, testUUID, actual)
	})

	t.Run("incorrect group", func(t *testing.T) {
		var actual uuid.UUID
		err := json.Unmarshal([]byte(`"00010203-04Z5-0607-0809-0A0B0C0D0E0F"`), &actual)
		assert.EqualError(t, err, "invalid UUID format")
	})
}

func TestUUID_FromBytes(t *testing.T) {
	actual, err := uuid.FromBytes([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	require.NoError(t, err)
	require.Equal(t, testUUID, actual)

	t.Run("short []byte", func(t *testing.T) {
		actual, err := uuid.FromBytes([]byte{0, 1, 2, 3, 4, 5, 6, 7})
		require.EqualError(t, err, "invalid UUID (got 8 bytes)")
		require.Zero(t, actual)
	})

	t.Run("long []byte", func(t *testing.T) {
		actual, err := uuid.FromBytes([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		require.EqualError(t, err, "invalid UUID (got 17 bytes)")
		require.Zero(t, actual)
	})
}

func TestNewUUID(t *testing.T) {
	u1 := uuid.NewUUID()
	u2 := uuid.NewUUID()
	require.NotEqual(t, u1.String(), u2.String(), "Two UUIDs should not be equal.")
}
