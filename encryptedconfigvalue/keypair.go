// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

type KeyPair interface {
	Algorithm() AlgorithmType

	// PublicKey returns the public key if this keypair is for an asymmetric encryption algorithm, or the shared
	// encryption key if this keypair is for a symmetric encryption algorithm.
	PublicKey() KeyWithAlgorithm

	// PrivateKey returns the private key if this keypair is for an asymmetric algorithm. Returns nil if this
	// keypair is for a symmetric encryption algorithm.
	PrivateKey() KeyWithAlgorithm
}

type keyPair struct {
	algorithm AlgorithmType
	pubKey    KeyWithAlgorithm
	privKey   KeyWithAlgorithm
}

func (k *keyPair) Algorithm() AlgorithmType {
	return k.algorithm
}

func (k *keyPair) PublicKey() KeyWithAlgorithm {
	return k.pubKey
}

func (k *keyPair) PrivateKey() KeyWithAlgorithm {
	return k.privKey
}
