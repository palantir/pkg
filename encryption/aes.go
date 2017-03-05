package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type Plaintext string
type Ciphertext string

const nonceSize = 32

// EncryptSym encrypts the provided plaintext with the provided symmetric key. Compatible with
// https://github.com/palantir/encrypted-config-value.
func EncryptSym(plain Plaintext, sym *SymmetricKey) (Ciphertext, error) {
	gcm, err := newBlockCipher(sym)
	if err != nil {
		return "", err
	}

	// Generate random nonce (javax crypto calls this the initialization vector)
	nonce, err := RandomBytes(gcm.NonceSize())
	if err != nil {
		return "", fmt.Errorf("Failed to generate nonce: %v", err)
	}

	encrypted := gcm.Seal(nil, nonce, []byte(plain), nil)

	// The IV needs to be unique but not secure, so include it at the beginning of the ciphertext
	withNonce := append(nonce, encrypted...)

	b64 := Base64Encode(withNonce)
	return Ciphertext(b64), nil
}

func DecryptSym(value Ciphertext, sym *SymmetricKey) (Plaintext, error) {
	raw, err := Base64Decode([]byte(value))
	if err != nil {
		return "", fmt.Errorf("Encrypted value is not valid base64: %v", err)
	}

	if len(raw) <= nonceSize {
		return "", fmt.Errorf("Expected encrypted value to be at least %v bytes but was %v", nonceSize, len(raw))
	}

	gcm, err := newBlockCipher(sym)
	if err != nil {
		return "", err
	}

	plain, err := gcm.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("Failed to decrypt value: %v", err)
	}

	return Plaintext(plain), nil
}

func newBlockCipher(sym *SymmetricKey) (cipher.AEAD, error) {
	if sym == nil {
		return nil, fmt.Errorf("symmetric key must not be nil")
	}
	block, err := aes.NewCipher(sym.k)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct symmetric cipher: %v", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct block cipher: %v", err)
	}
	return gcm, nil
}
