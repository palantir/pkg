package encryption

import (
	"encoding/base32"
	"encoding/base64"
)

func Base32Encode(input []byte) []byte {
	return []byte(base32.StdEncoding.EncodeToString(input))
}

func Base32Decode(input []byte) ([]byte, error) {
	return base32.StdEncoding.DecodeString(string(input))
}

func Base64Encode(input []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(input))
}

func Base64Decode(input []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(input))
}
