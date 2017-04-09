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

func TestAESEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		name    string
		keyBits int
		input   []byte
	}{
		{"256-bit AES key", 256, []byte("secret message")},
		{"128-bit AES key", 128, []byte{byte(1), byte(3)}},
	} {
		aesKey, err := encryption.NewAESKey(currCase.keyBits)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		cipher := encryption.NewAESGCMCipher()
		encrypted, err := cipher.Encrypt(currCase.input, aesKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		decrypted, err := cipher.Decrypt(encrypted, aesKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		assert.Equal(t, currCase.input, decrypted, "Case %d: %s", i, currCase.name)
	}
}
