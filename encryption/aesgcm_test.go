// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryption_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryption"
)

func TestAESEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		keyBits    int
		nonceBytes int
		input      []byte
	}{
		{256, 12, []byte("secret message")},
		{256, 32, []byte("secret message")},
		{128, 12, []byte{byte(1), byte(3)}},
	} {
		aesGCMKey, err := encryption.NewAESGCM(currCase.keyBits, currCase.nonceBytes)
		require.NoError(t, err, "Case %d", i)

		encrypted, err := aesGCMKey.Encrypt(currCase.input)
		require.NoError(t, err, "Case %d", i)

		decrypted, err := aesGCMKey.Decrypt(encrypted)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.input, decrypted, "Case %d", i)
	}
}

func TestAESDecryptUsingStoredKey(t *testing.T) {
	for i, currCase := range []struct {
		keyBase64        string
		nonceSize        int
		ciphertextBase64 string
		plaintext        []byte
	}{
		{
			keyBase64:        "d410u8X4spKYnSgsZ2dqibuCg8LOh1HU2bAIqRdDWNQ=",
			nonceSize:        12,
			ciphertextBase64: "oJ0dBgHS8XV9ABNxU6OV8ZPYteSOojYrVJc8NWP+Ava8OBWjvbEHCVBZ",
			plaintext:        []byte("secret message"),
		},
	} {
		keyBytes, err := base64.StdEncoding.DecodeString(currCase.keyBase64)
		require.NoError(t, err, "Case %d", i)
		key := encryption.AESGCMFromKey(keyBytes, currCase.nonceSize)

		encrypted, err := base64.StdEncoding.DecodeString(currCase.ciphertextBase64)
		require.NoError(t, err, "Case %d", i)

		gotPlaintext, err := key.Decrypt(encrypted)
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, currCase.plaintext, gotPlaintext, "Case %d", i)
	}
}
