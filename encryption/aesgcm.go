// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type AESGCMKey interface {
	Key
	Key() []byte
	NonceSizeBytes() int
}

type aesGCMKey struct {
	k              []byte
	nonceSizeBytes int
}

func NewAESGCM(keySizeBits, nonceSizeBytes int) (AESGCMKey, error) {
	k, err := randomBytes(keySizeBits / 8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %v", err)
	}
	return AESGCMFromKey(k, nonceSizeBytes), err
}

func AESGCMFromKey(key []byte, nonceSizeBytes int) AESGCMKey {
	return &aesGCMKey{
		k:              key,
		nonceSizeBytes: nonceSizeBytes,
	}
}

func (k *aesGCMKey) Key() []byte {
	return k.k
}

func (k *aesGCMKey) NonceSizeBytes() int {
	return k.nonceSizeBytes
}

func (k *aesGCMKey) Encrypt(data []byte) ([]byte, error) {
	gcm, err := k.newBlockCipher()
	if err != nil {
		return nil, err
	}

	// generate random nonce (javax crypto calls this the initialization vector)
	nonce, err := randomBytes(gcm.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	encrypted := gcm.Seal(nil, nonce, data, nil)

	// IV needs to be unique but not secure, so include it at the beginning of the ciphertext
	withNonce := append(nonce, encrypted...)
	return withNonce, nil
}

func (k *aesGCMKey) Decrypt(data []byte) ([]byte, error) {
	if len(data) <= k.nonceSizeBytes {
		return nil, fmt.Errorf("expected encrypted value to be at least %d bytes but was %d", k.nonceSizeBytes, len(data))
	}
	gcm, err := k.newBlockCipher()
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, data[:k.nonceSizeBytes], data[k.nonceSizeBytes:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %v", err)
	}
	return plain, nil
}

func (k *aesGCMKey) newBlockCipher() (cipher.AEAD, error) {
	if len(k.k) == 0 {
		return nil, fmt.Errorf("symmetric key cannot be empty")
	}
	block, err := aes.NewCipher(k.k)
	if err != nil {
		return nil, fmt.Errorf("failed to construct symmetric cipher: %v", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, k.nonceSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to construct block cipher: %v", err)
	}
	return gcm, nil
}

func randomBytes(n int) ([]byte, error) {
	out := make([]byte, n)
	if _, err := rand.Read(out); err != nil {
		return nil, fmt.Errorf("failed to generate %d cryptographically strong pseudo-random bytes: %v", n, err)
	}
	return out, nil
}
