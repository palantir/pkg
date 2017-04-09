// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"fmt"
	"reflect"

	"github.com/palantir/pkg/encryption"
)

// key size is hard-coded for compatibility with encrypted-config-value
const (
	rsaKeySizeBits = 2048
)

var rsaAlgorithm = &algorithmDefinition{
	name: RSA,
	generator: func() (pubKey encryption.Key, privKey encryption.Key, err error) {
		rsaKey, err := encryption.NewRSAOAEPWithSHA256(rsaKeySizeBits)
		if err != nil {
			return nil, nil, err
		}
		return rsaKey, rsaKey, nil
	},
	pubKeySerializer:  rsaOAEPWithSHA256PublicKeyX509Serializer,
	privKeySerializer: rsaOAEPWithSHA256PrivateKeyPKCS8Serializer,
}

var (
	rsaOAEPWithSHA256PublicKeyX509Serializer = &defaultKeySerializer{
		marshalFunc: func(key encryption.Key) ([]byte, error) {
			k, ok := key.(encryption.RSAOAEPWithSHA256)
			if !ok {
				return nil, fmt.Errorf("key must be an RSA OAEP with SHA-256 key, but is %q", reflect.TypeOf(key).String())
			}
			return encryption.RSAPublicKeyX509PEMBytes(k.PublicKey())
		},
		unmarshalFunc: func(key []byte) (encryption.Key, error) {
			return encryption.RSAOAEPWithSHA256FromPublicKey(key)
		},
	}
	rsaOAEPWithSHA256PrivateKeyPKCS8Serializer = &defaultKeySerializer{
		marshalFunc: func(key encryption.Key) ([]byte, error) {
			k, ok := key.(encryption.RSAOAEPWithSHA256)
			if !ok {
				return nil, fmt.Errorf("key must be an RSA OAEP with SHA-256 key, but is %q", reflect.TypeOf(key).String())
			}
			return encryption.RSAPrivateKeyPKCS8Bytes(k.PrivateKey())
		},
		unmarshalFunc: func(key []byte) (encryption.Key, error) {
			return encryption.RSAOAEPWithSHA256FromPrivateKey(key)
		},
	}
)
