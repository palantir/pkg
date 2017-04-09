// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
)

type RSAOAEPWithSHA256 interface {
	Key
	PublicKey() *rsa.PublicKey
	PrivateKey() *rsa.PrivateKey
}

type rsaOAEPWithSHA256 struct {
	// the public key. Is nil if the private key is non-nil (because in that case, priv.PublicKey already contains
	// the key).
	pub *rsa.PublicKey

	// the private key. If this value is nil, it indicates that this is an encryption-only key with only the public
	// key defined.
	priv *rsa.PrivateKey
}

func NewRSAOAEPWithSHA256(asymmetricBits int) (RSAOAEPWithSHA256, error) {
	priv, err := rsa.GenerateKey(rand.Reader, asymmetricBits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %v", err)
	}
	return &rsaOAEPWithSHA256{
		priv: priv,
	}, nil
}

func RSAOAEPWithSHA256FromPublicKey(key []byte) (RSAOAEPWithSHA256, error) {
	var pub *rsa.PublicKey
	pub, err := loadRSAPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %v", err)
	}
	return &rsaOAEPWithSHA256{
		pub: pub,
	}, nil
}

func RSAOAEPWithSHA256FromPrivateKey(key []byte) (RSAOAEPWithSHA256, error) {
	pkcsPrivKey, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid PKCS8 private key: %v", err)
	}
	rsaPrivKey, ok := pkcsPrivKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid PKCS8 private key: %v", err)
	}
	return &rsaOAEPWithSHA256{
		priv: rsaPrivKey,
	}, nil
}

func (a *rsaOAEPWithSHA256) PublicKey() *rsa.PublicKey {
	if a.pub != nil {
		return a.pub
	}
	return &a.priv.PublicKey
}

func (a *rsaOAEPWithSHA256) PrivateKey() *rsa.PrivateKey {
	return a.priv
}

func (a *rsaOAEPWithSHA256) Encrypt(data []byte) ([]byte, error) {
	pubKey := a.pub
	if a.pub == nil {
		pubKey = &a.priv.PublicKey
	}
	if pubKey == nil {
		return nil, fmt.Errorf("public key required for encryption")
	}
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, []byte{})
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

func (a *rsaOAEPWithSHA256) Decrypt(data []byte) ([]byte, error) {
	if a.priv == nil {
		return nil, fmt.Errorf("private key required for decryption")
	}
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, a.priv, data, []byte{})
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func loadRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	var errInvalidRSAPublicKeyError = fmt.Errorf("key is not a valid PEM-encoded RSA public key")

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errInvalidRSAPublicKeyError
	}
	pkixPubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errInvalidRSAPublicKeyError
	}
	rsaPubKey, ok := pkixPubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errInvalidRSAPublicKeyError
	}

	return rsaPubKey, nil
}

func RSAPublicKeyX509PEMBytes(key *rsa.PublicKey) ([]byte, error) {
	asn1PubKey, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1PubKey,
	}), nil
}

func RSAPrivateKeyPKCS8Bytes(key *rsa.PrivateKey) ([]byte, error) {
	pkey := struct {
		Version             int
		PrivateKeyAlgorithm []asn1.ObjectIdentifier
		PrivateKey          []byte
	}{
		Version:             0,
		PrivateKeyAlgorithm: make([]asn1.ObjectIdentifier, 1),
	}
	// https://tls.mbed.org/kb/cryptography/asn1-key-structures-in-der-and-pem, see bottom
	pkey.PrivateKeyAlgorithm[0] = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	pkey.PrivateKey = x509.MarshalPKCS1PrivateKey(key)
	return asn1.Marshal(pkey)
}
