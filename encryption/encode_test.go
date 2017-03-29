// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryption_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryption"
)

func TestBase32(t *testing.T) {
	msg := []byte("☃ testing base32 ☃")

	b64 := encryption.Base32Encode(msg)
	expected := "4KMIGIDUMVZXI2LOM4QGEYLTMUZTEIHCTCBQ===="
	assert.Equal(t, expected, string(b64), "wrong base32")

	msg2, err := encryption.Base32Decode(b64)
	require.NoError(t, err, "base32 decode failed")
	assert.Equal(t, msg, msg2, "bad base32 result")
}

func TestBase32Error(t *testing.T) {
	bad := []byte("???")
	_, err := encryption.Base32Decode(bad)
	assert.Error(t, err, "expected base32 decode to fail")
}

func TestBase64(t *testing.T) {
	msg := []byte("☃ testing base64 ☃")

	b64 := encryption.Base64Encode(msg)
	expected := "4piDIHRlc3RpbmcgYmFzZTY0IOKYgw=="
	assert.Equal(t, expected, string(b64), "wrong base64")

	msg2, err := encryption.Base64Decode(b64)
	require.NoError(t, err, "base64 decode failed")
	assert.Equal(t, msg, msg2, "bad base64 result")
}

func TestBase64Error(t *testing.T) {
	bad := []byte("???")
	_, err := encryption.Base64Decode(bad)
	assert.Error(t, err, "expected base64 decode to fail")
}
