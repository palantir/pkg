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

func TestRSAKeySerDe(t *testing.T) {
	pubKey, privKey, err := encryption.NewRSAKeyPair(2048)
	require.NoError(t, err)

	pubKeyBytes := pubKey.Bytes()
	pubKey, err = encryption.RSAPublicKeyFromPEMBytes(pubKeyBytes)
	require.NoError(t, err)

	privKeyBytes := privKey.Bytes()
	privKey, err = encryption.RSAPrivateKeyFromPKCS8Bytes(privKeyBytes)
	require.NoError(t, err)

	cipher := encryption.NewRSAOAEPCipher()
	plaintext := "input plaintext"
	encrypted, err := cipher.Encrypt([]byte(plaintext), pubKey)
	require.NoError(t, err)

	decrypted, err := cipher.Decrypt(encrypted, privKey)
	require.NoError(t, err)

	assert.Equal(t, plaintext, string(decrypted))
}
