// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

func TestAESEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		input []byte
	}{
		{[]byte("secret message")},
		{[]byte{byte(1), byte(3)}},
	} {
		kp, err := encryptedconfigvalue.AES.Definition().NewKeyPair()
		require.NoError(t, err, "Case %d", i)

		encryptedBytes, err := kp.PublicKey().Encrypt(currCase.input)
		require.NoError(t, err, "Case %d", i)

		decryptedBytes, err := kp.PublicKey().Decrypt(encryptedBytes)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.input, decryptedBytes, "Case %d", i)
	}
}

func TestAESLoadStoredKeyAndWrite(t *testing.T) {
	for i, currCase := range []struct {
		serializedKey string
	}{
		{"AES:s0u/5zMvOz9bd2/7QSJ0yaRpav9kgAmLh6GyXkttwC4="},
	} {
		kwa, err := encryptedconfigvalue.NewKeyWithAlgorithm(currCase.serializedKey)
		require.NoError(t, err, "Case %d", i)

		serialized, err := kwa.ToSerializable()
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.serializedKey, serialized.SerializedStringForm(), "Case %d", i)
	}
}

func TestAESDecryptUsingStoredKey(t *testing.T) {
	for i, currCase := range []struct {
		serializedKey    string
		ciphertextBase64 string
		plaintext        []byte
	}{
		{
			"AES:s0u/5zMvOz9bd2/7QSJ0yaRpav9kgAmLh6GyXkttwC4=",
			"KDXbSDZnEOYfSF6sL28Eh60HP0Lo7GZBIKPufqveF/aso9hEw/2F80ox5s8kkK+7e/jQE1vUZ3f+n33x4OM=",
			[]byte("secret message"),
		},
	} {
		kwa, err := encryptedconfigvalue.NewKeyWithAlgorithm(currCase.serializedKey)
		require.NoError(t, err, "Case %d", i)

		wantCiphertext, err := base64.StdEncoding.DecodeString(currCase.ciphertextBase64)
		require.NoError(t, err, "Case %d", i)

		gotPlaintext, err := kwa.Decrypt(wantCiphertext)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.plaintext, gotPlaintext)
	}
}
