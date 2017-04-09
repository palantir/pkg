// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"encoding/base64"
	"fmt"

	"github.com/palantir/pkg/encryption"
)

type Algorithm interface {
	Name() AlgorithmType
	NewKeyPair() (KeyPair, error)

	// ToPublicKey returns a public key for this algorithm based on the provided input. If this is an asymmetric-key
	// algorithm, it returns the public key; if this algorithm is a symmetric-key algorithm, this function returns
	// the symmetric key.
	ToPublicKey(key SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error)

	// ToPrivateKey returns a private key for this algorithm based on the provided input. If this is an
	// asymmetric-key algorithm, it returns the private key; if this algorithm is a symmetric-key algorithm, this
	// function returns an error.
	ToPrivateKey(key SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error)
}

type KeySerializer interface {
	// Marshal returns the byte representation of the provided key. The output of this function should be able to be
	// processed by Unmarshal.
	Marshal(key encryption.Key) ([]byte, error)

	// Unmarshal returns a key given the byte representation produced by its Marshal function.
	Unmarshal(key []byte) (encryption.Key, error)
}

type defaultKeySerializer struct {
	marshalFunc   func(key encryption.Key) ([]byte, error)
	unmarshalFunc func(key []byte) (encryption.Key, error)
}

func (s *defaultKeySerializer) Marshal(key encryption.Key) ([]byte, error) {
	return s.marshalFunc(key)
}

func (s *defaultKeySerializer) Unmarshal(key []byte) (encryption.Key, error) {
	return s.unmarshalFunc(key)
}

type algorithmDefinition struct {
	name              AlgorithmType
	generator         func() (pubKey encryption.Key, privKey encryption.Key, err error)
	pubKeySerializer  KeySerializer
	privKeySerializer KeySerializer
}

func (a *algorithmDefinition) Name() AlgorithmType {
	return a.name
}

func (a *algorithmDefinition) NewKeyPair() (KeyPair, error) {
	pubKey, privKey, err := a.generator()
	if err != nil {
		return nil, err
	}
	kp := &keyPair{
		algorithm: a.name,
		pubKey: &keyWithAlgorithm{
			algorithm:  a.name,
			Key:        pubKey,
			serializer: a.pubKeySerializer,
		},
	}
	if privKey != nil {
		kp.privKey = &keyWithAlgorithm{
			algorithm:  a.name,
			Key:        privKey,
			serializer: a.privKeySerializer,
		}
	}
	return kp, nil
}

func (a *algorithmDefinition) ToPublicKey(key SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error) {
	return a.unmarshalKey(a.pubKeySerializer, key)
}

func (a *algorithmDefinition) ToPrivateKey(key SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error) {
	if a.privKeySerializer == nil {
		return nil, fmt.Errorf("private key does not exist for symmetric-key algorithm %s", a.name)
	}
	return a.unmarshalKey(a.privKeySerializer, key)
}

func (a *algorithmDefinition) unmarshalKey(serializer KeySerializer, key SerializableKeyWithAlgorithm) (KeyWithAlgorithm, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(key.encodedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to base64-decode key content: %v", err)
	}

	unmarshaledKey, err := serializer.Unmarshal(decodedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal key: %v", err)
	}
	return &keyWithAlgorithm{
		algorithm:  a.name,
		Key:        unmarshaledKey,
		serializer: serializer,
	}, nil
}
