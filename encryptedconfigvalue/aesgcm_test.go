// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		name    string
		keyBits int
		input   string
	}{
		{"256-bit AES key", 256, "secret message"},
		{"128-bit AES key", 128, "secret message"},
	} {
		aesKey, err := NewAESKey(currCase.keyBits)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		encrypter := NewAESGCMEncrypter()
		ev, err := encrypter.Encrypt(currCase.input, aesKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		decrypted, err := ev.Decrypt(aesKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		assert.Equal(t, currCase.input, decrypted, "Case %d: %s", i, currCase.name)
	}
}

func TestAESJSONSerDe(t *testing.T) {
	for i, currCase := range []struct {
		name          string
		keyBase64     string
		json          string
		wantDecrypted string
	}{
		{
			"test decode JSON",
			"0JlMK+vn1T8+d43NRp49xi35lA/NQVSTeowTw4iLw5M=",
			`{"type":"AES","mode":"GCM","ciphertext":"hGYI+23l1vDMjQ==","iv":"DbEqWuhTvB9x1wkA","tag":"wmCU2C8xTtWc4+er22oXLA=="}`,
			"test input",
		},
	} {
		aesKeyBytes, err := base64.StdEncoding.DecodeString(currCase.keyBase64)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		aesKey, err := AESKeyFromBytes(aesKeyBytes)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		var ev aesGCMEncryptedValue
		err = json.Unmarshal([]byte(currCase.json), &ev)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		decrypted, err := ev.Decrypt(aesKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		assert.Equal(t, currCase.wantDecrypted, string(decrypted), "Case %d: %s", i, currCase.name)

		marshaledJSON, err := json.Marshal(ev)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		assert.Equal(t, currCase.json, string(marshaledJSON), "Case %d: %s", i, currCase.name)
	}
}

func TestAESDecryptUsingStoredKey(t *testing.T) {
	for i, currCase := range []struct {
		name            string
		keyBase64       string
		nonceSize       int
		encryptedBase64 string
		nonceBase64     string
		tagBase64       string
		plaintext       []byte
	}{
		{
			name:            "Standard AES value",
			keyBase64:       "jL+EV1ylmlj0TqW6ZlbjB/FOJDUSsc2YoofV7gd7Hms=",
			nonceSize:       12,
			encryptedBase64: "DRkNGsmzUw3J22MmgjA=",
			nonceBase64:     "twixk4mwt3YXfGXg",
			tagBase64:       "vWqoZlV+VKAEYmk9Pz7vpw==",
			plaintext:       []byte("secret message"),
		},
		{
			name:            "AES value with non-standard nonce size",
			keyBase64:       "JVXsw1LD7eB77/+s53As8JxPUyHSSSM1lH85NoR/fMs=",
			nonceSize:       32,
			encryptedBase64: "fI7IEXyyFCxH3uKv+Sw=",
			nonceBase64:     "isN/8Z7FhVLJ5v//CamggWaskyAUrEIiqknppwT6p3U=",
			tagBase64:       "VBs8LgommFWD52WAzNaQbw==",
			plaintext:       []byte("secret message"),
		},
	} {
		keyBytes, err := base64.StdEncoding.DecodeString(currCase.keyBase64)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)
		aesKey, err := AESKeyFromBytes(keyBytes)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		ciphertext, err := base64.StdEncoding.DecodeString(currCase.encryptedBase64)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		nonce, err := base64.StdEncoding.DecodeString(currCase.nonceBase64)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		tag, err := base64.StdEncoding.DecodeString(currCase.tagBase64)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		ev := &aesGCMEncryptedValue{
			encrypted: ciphertext,
			nonce:     nonce,
			tag:       tag,
		}
		gotPlaintext, err := ev.Decrypt(aesKey)
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, string(currCase.plaintext), gotPlaintext, "Case %d: %s", i, currCase.name)
	}
}
