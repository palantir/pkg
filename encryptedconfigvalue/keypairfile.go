// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	KeyPathProperty      = "palantir.config.key_path"
	DefaultPublicKeyPath = "var/conf/encrypted-config-value.key"
)

// LoadKeyPairFromDefaultPath loads an encrypted-config-value key pair from the default locations as specified by the
// encrypted-config-value project.
func LoadKeyPairFromDefaultPath() (KeyPair, error) {
	pubKeyPath := os.Getenv(KeyPathProperty)
	if pubKeyPath == "" {
		pubKeyPath = DefaultPublicKeyPath
	}
	pubKey, err := readKeyFromPath(pubKeyPath, pubKeyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %v", err)
	}

	privKey, err := readKeyFromPath(toPrivKeyPath(pubKeyPath), privKeyReader)
	if err != nil {
		// if private key file could not be read, assume that this is a public-key only pair
		return &keyPair{
			algorithm: pubKey.Algorithm(),
			pubKey:    pubKey,
		}, nil
	}

	if pubKey.Algorithm() != privKey.Algorithm() {
		return nil, fmt.Errorf("algorithm for public and private key must match, but %q != %q", pubKey.Algorithm(), privKey.Algorithm())
	}

	return &keyPair{
		algorithm: pubKey.Algorithm(),
		pubKey:    pubKey,
		privKey:   privKey,
	}, nil
}

func LoadSymmetricKeyFromPath(keyPath string) (KeyPair, error) {
	pubKey, err := readKeyFromPath(keyPath, pubKeyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load symmetric key: %v", err)
	}

	return &keyPair{
		algorithm: pubKey.Algorithm(),
		pubKey:    pubKey,
	}, nil
}

func LoadKeyPairFromPaths(pubKeyPath, privKeyPath string) (KeyPair, error) {
	pubKey, err := readKeyFromPath(pubKeyPath, pubKeyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %v", err)
	}
	privKey, err := readKeyFromPath(privKeyPath, privKeyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}
	return &keyPair{
		algorithm: pubKey.Algorithm(),
		pubKey:    pubKey,
		privKey:   privKey,
	}, nil
}

var pubKeyReader = func(alg Algorithm) func(SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error) {
	return alg.ToPublicKey
}

var privKeyReader = func(alg Algorithm) func(SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error) {
	return alg.ToPrivateKey
}

func readKeyFromPath(keyPath string, reader func(alg Algorithm) func(SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error)) (KeyWithAlgorithm, error) {
	keyContent, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contents of key file: %v", err)
	}
	key, err := NewSerializableKeyWithAlgorithm(string(keyContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create key from contents of file: %v", err)
	}
	kwa, err := reader(key.algorithm.Definition())(key)
	if err != nil {
		return nil, fmt.Errorf("failed to transform SerializableKeyWithAlgorithm to KeyWithAlgorithm: %v", err)
	}
	return kwa, nil
}

func toPrivKeyPath(pubKeyPath string) string {
	return strings.TrimSuffix(pubKeyPath, path.Ext(pubKeyPath)) + ".private"
}
