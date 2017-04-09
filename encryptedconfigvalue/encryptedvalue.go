// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
)

// EncryptedValue is a value that is encrypted by the encrypted-config-value library.
// The serialized value is a string of the form "${enc:...}"
type EncryptedValue struct {
	// ciphertext is the content of encrypted value. In the format "${enc:<base64-encoded-value>}", it is the
	// "<base64-encoded-value>".
	ciphertext string
}

func (v *EncryptedValue) Ciphertext() string {
	return v.ciphertext
}

func (v *EncryptedValue) Decrypt(key KeyWithAlgorithm) ([]byte, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(v.ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64-decode ciphertext: %v", err)
	}
	decrypted, err := key.Decrypt(decodedCiphertext)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func (v EncryptedValue) String() string {
	return v.SerializedStringForm()
}

func (v *EncryptedValue) SerializedStringForm() string {
	return fmt.Sprintf("${enc:%v}", v.ciphertext)
}

func (v *EncryptedValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.SerializedStringForm())
}

func (v *EncryptedValue) UnmarshalJSON(data []byte) error {
	var stringVal string
	if err := json.Unmarshal(data, &stringVal); err != nil {
		return err
	}
	encryptedVal, err := NewEncryptedValue(stringVal)
	if err != nil {
		return err
	}
	*v = encryptedVal
	return nil
}

func (v *EncryptedValue) MarshalYAML() (interface{}, error) {
	return v.SerializedStringForm(), nil
}

func (v *EncryptedValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringVal string
	if err := unmarshal(&stringVal); err != nil {
		return err
	}
	encryptedVal, err := NewEncryptedValue(stringVal)
	if err != nil {
		return err
	}
	*v = encryptedVal
	return nil
}

// MustNewEncryptedValue is a version of NewEncryptedValue that panics if any error is encountered. This can be used as
// a shorthand when the input value is known to be of the correct format.
func MustNewEncryptedValue(val string) EncryptedValue {
	ev, err := NewEncryptedValue(val)
	if err != nil {
		panic(err)
	}
	return ev
}

// NewEncryptedValue returns a new EncryptedValue created from the provided serialized form of the string. The provided
// string should be of the form "${enc:<base64-encoded-value>}".
func NewEncryptedValue(val string) (EncryptedValue, error) {
	matches := encryptedValueRegexp.FindStringSubmatch(val)
	if len(matches) != 2 {
		return EncryptedValue{}, fmt.Errorf("encrypted value must be of the form ${enc:<base64-encoded-value>}, but was %q", val)
	}
	return EncryptedValue{
		ciphertext: matches[1],
	}, nil
}

// Encrypt returns a new EncryptedValue created by encrypting the provided value using the provided key.
func Encrypt(val []byte, key KeyWithAlgorithm) (EncryptedValue, error) {
	encryptedBytes, err := key.Encrypt(val)
	if err != nil {
		return EncryptedValue{}, err
	}
	return EncryptedValue{
		ciphertext: base64.StdEncoding.EncodeToString(encryptedBytes),
	}, nil
}

var encryptedValueRegexp = regexp.MustCompile(`\${enc:([^}]+)}`)

// ContainsEncryptedConfigValue returns true if the provided input has any occurrences of an encrypted value in it.
func ContainsEncryptedConfigValue(input []byte) bool {
	return encryptedValueRegexp.Match(input)
}

// DecryptAllEncryptedValues returns a new byte slice that is based on the input but where all occurrences of encrypted
// values are replaced with the value obtained when decrypting the encrypted value using the provided key. If an error
// is encountered while decrypting any of the encountered values, returns the unmodified original input.
func DecryptAllEncryptedValues(input []byte, key KeyWithAlgorithm) []byte {
	decrypt := func(raw []byte) []byte {
		encryptedVal, err := NewEncryptedValue(string(raw))
		if err != nil {
			return raw
		}
		decrypted, err := encryptedVal.Decrypt(key)
		if err != nil {
			return raw
		}
		return decrypted
	}
	return encryptedValueRegexp.ReplaceAllFunc(input, decrypt)
}

// NormalizeEncryptedValues returns a new byte slice in which all of the encrypted values in the input that have the
// same decrypted plaintext representation when decrypted using the provided key will be normalized such that their
// encrypted values are the same. If the decrypted plaintext exists as a key in the "normalized" map, then it is
// substituted with the value in that map. If the map does not contain an entry for the plaintext, the first time it is
// encountered it is added to the map with its corresponding EncryptedValue, and every subsequent occurrence in the
// input will use the normalized value. On completion of the function, the "normalized" map will contain a key for every
// plaintext value in the input where the value will be the EncryptedValue that was used for all occurrences.
//
// WARNING: after this function has been executed, the keys of the "normalized" map will contain all of the decrypted
// values in the input -- its use should be tracked carefully. The "normalized" version of the input is also less
// cryptographically secure because it makes the output more predictable -- for example, it makes it possible to
// determine that multiple different encrypted values have the same underlying plaintext value.
//
// The intended usage of this function is limited to very specific cases in which there is a requirement that the same
// plaintext must render to the same encrypted value for a specific key. Ensure that you fully understand the
// ramifications of this and only use this function if it is absolutely necessary.
func NormalizeEncryptedValues(input []byte, key KeyWithAlgorithm, normalized map[string]EncryptedValue) []byte {
	normalizeEncryptedValue := func(raw []byte) []byte {
		encryptedVal, err := NewEncryptedValue(string(raw))
		if err != nil {
			return raw
		}
		decrypted, err := encryptedVal.Decrypt(key)
		if err != nil {
			return raw
		}
		plaintext := string(decrypted)
		// if an entry for the plaintext of the current encrypted value exists in the normalized map, replace
		// the encrypted value with the normalized one.
		if sub, present := normalized[plaintext]; present {
			return []byte(sub.SerializedStringForm())
		}
		// this is the first time that this plaintext has been encountered for an encrypted value. Store the
		// current encrypted value as the value for the plaintext in the map so that all subsequent occurrences
		// will use this encrypted value.
		normalized[plaintext] = encryptedVal
		return raw
	}
	return encryptedValueRegexp.ReplaceAllFunc(input, normalizeEncryptedValue)
}
