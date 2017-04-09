// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/palantir/pkg/encryption"
)

// NewKeyWithAlgorithm returns a KeyWithAlgorithm based on the provided serialized string input. The input must be of
// the form "<algorithm>:<base64-encoded-key>".
func NewKeyWithAlgorithm(input string) (KeyWithAlgorithm, error) {
	skwa, err := NewSerializableKeyWithAlgorithm(input)
	if err != nil {
		return nil, err
	}

	// Implementation note: Because of constraints imposed by the upstream encrypted-config-value library, there is
	// not a 1:1 mapping between algorithms and key type (for example, "RSA:<base64>" could represent either the
	// public key or private key for the RSA algorithm). Because of this, the code below works by trying to parse
	// the key as the "public" key for the algorithm first, and then falls back to parsing as the "private" key for
	// the algorithm if that fails.

	// try parsing as public key
	alg := algorithms[skwa.algorithm]
	kwa, pubErr := alg.ToPublicKey(skwa)
	if pubErr == nil {
		return kwa, nil
	}
	// try parsing as private key
	kwa, privErr := alg.ToPrivateKey(skwa)
	if privErr == nil {
		return kwa, nil
	}
	errParts := []string{
		"could not parse key as public or private key",
		fmt.Sprintf("Pubic key parse error: %v", pubErr),
		fmt.Sprintf("Private key parse error: %v", privErr),
	}
	return nil, fmt.Errorf(strings.Join(errParts, "\n"))
}

type KeyWithAlgorithm interface {
	Algorithm() AlgorithmType

	encryption.Key

	// ToSerializable returns a representation of this KeyWithAlgorithm that is suitable for marshaling or
	// unmarshaling in a format that is compatible with encrypted-config-value.
	ToSerializable() (SerializableKeyWithAlgorithm, error)
}

type keyWithAlgorithm struct {
	encryption.Key
	algorithm  AlgorithmType
	serializer KeySerializer
}

func (kwa *keyWithAlgorithm) Algorithm() AlgorithmType {
	return kwa.algorithm
}

func (kwa *keyWithAlgorithm) ToSerializable() (SerializableKeyWithAlgorithm, error) {
	bytes, err := kwa.serializer.Marshal(kwa.Key)
	if err != nil {
		return SerializableKeyWithAlgorithm{}, err
	}
	// create SerializableKeyWithAlgorithm with "<key-algorithm>:<base64-encoded-key>"
	return NewSerializableKeyWithAlgorithm(fmt.Sprintf("%s:%s", kwa.algorithm, base64.StdEncoding.EncodeToString(bytes)))
}
