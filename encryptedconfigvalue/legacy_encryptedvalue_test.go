// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

func TestLegacyEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		name             string
		keyPairGenerator func() (encryptedconfigvalue.KeyPair, error)
		encrypter        encryptedconfigvalue.Encrypter
	}{
		{"AES", encryptedconfigvalue.NewAESKeyPair, encryptedconfigvalue.LegacyAESGCMEncrypter()},
		{"RSA", encryptedconfigvalue.NewRSAKeyPair, encryptedconfigvalue.LegacyRSAOAEPEncrypter()},
	} {
		keyPair, err := currCase.keyPairGenerator()
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		const plaintext = "test input plaintext"
		ev, err := currCase.encrypter.Encrypt(plaintext, keyPair.EncryptionKey)
		require.NoError(t, err)

		decrypted, err := ev.Decrypt(keyPair.DecryptionKey)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted)
	}
}
