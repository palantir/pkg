package encryption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type SymmetricKey struct {
	k []byte
}
type EncodedSymmetricKey string

const (
	NoSymmetricKey   = EncodedSymmetricKey("")
	symmetricKeyBits = 256
)

var (
	aesKeyPrefix   = []byte("AES:")
	aes32KeyPrefix = []byte("AES32:")
)

func NewSymmetricKey() (*SymmetricKey, error) {
	key, err := RandomBytes(symmetricKeyBits / 8)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate symmetric key: %v", err)
	}
	return &SymmetricKey{k: key}, nil
}

// Encode returns a base64-encoded version of the key using an "AES64" header.
// Compatible with https://github.com/palantir/encrypted-config-value.
func (sym *SymmetricKey) Encode() EncodedSymmetricKey {
	return sym.encodeWithFunc(aesKeyPrefix, Base64Encode)
}

// Encode32 returns a base32-encoded version of the key using an "AES32" header.
// Compatible with https://github.com/palantir/encrypted-config-value.
func (sym *SymmetricKey) Encode32() EncodedSymmetricKey {
	return sym.encodeWithFunc(aes32KeyPrefix, Base32Encode)
}

func (sym *SymmetricKey) encodeWithFunc(prefix []byte, encodingFunc func([]byte) []byte) EncodedSymmetricKey {
	if sym == nil {
		return NoSymmetricKey
	}
	encoded := encodingFunc(sym.k)
	return EncodedSymmetricKey(append(prefix, encoded...))
}

func (enc EncodedSymmetricKey) Decode() (*SymmetricKey, error) {
	encBytes := []byte(enc)
	switch {
	case bytes.HasPrefix(encBytes, aesKeyPrefix):
		withoutPrefix := encBytes[len(aesKeyPrefix):]
		raw, err := Base64Decode(withoutPrefix)
		if err != nil {
			return nil, fmt.Errorf("Encoded symmetric key is not valid base64: %v", err)
		}
		return &SymmetricKey{k: raw}, nil
	case bytes.HasPrefix(encBytes, aes32KeyPrefix):
		withoutPrefix := encBytes[len(aes32KeyPrefix):]
		raw, err := Base32Decode(withoutPrefix)
		if err != nil {
			return nil, fmt.Errorf("Encoded symmetric key is not valid base32: %v", err)
		}
		return &SymmetricKey{k: raw}, nil
	default:
		return nil, fmt.Errorf("Encoded symmetric key must start with %q or %q", aesKeyPrefix, aes32KeyPrefix)
	}
}

func (sym *SymmetricKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(sym.Encode())
}

func (sym *SymmetricKey) UnmarshalJSON(data []byte) error {
	var encoded EncodedSymmetricKey
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	result, err := encoded.Decode()
	if err != nil {
		return err
	}
	*sym = *result
	return nil
}

func (sym *SymmetricKey) MarshalYAML() (interface{}, error) {
	return sym.Encode(), nil
}

func (sym *SymmetricKey) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var encoded EncodedSymmetricKey
	if err := unmarshal(&encoded); err != nil {
		return err
	}
	result, err := encoded.Decode()
	if err != nil {
		return err
	}
	*sym = *result
	return nil
}

func KeyFromFile(keyPath string) (*SymmetricKey, error) {
	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read key file %s: %v", keyPath, err)
	}
	key, err := EncodedSymmetricKey(keyData).Decode()
	if err != nil {
		return nil, fmt.Errorf("Failed to decode key file %s: %v", keyPath, err)
	}
	return key, nil
}
