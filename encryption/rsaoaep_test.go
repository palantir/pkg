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

func TestRSAEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		keyBits int
		input   []byte
	}{
		{2048, []byte("secret message")},
		{1024, []byte("secret message")},
		{1024, []byte{byte(1), byte(3)}},
	} {
		pubKey, privKey, err := encryption.NewRSAKeyPair(currCase.keyBits)
		require.NoError(t, err, "Case %d", i)

		cipher := encryption.NewRSAOAEPCipher()
		encrypted, err := cipher.Encrypt(currCase.input, pubKey)
		require.NoError(t, err, "Case %d", i)

		decrypted, err := cipher.Decrypt(encrypted, privKey)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.input, decrypted, "Case %d", i)
	}
}
