// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"fmt"
	"reflect"

	"github.com/palantir/pkg/encryption"
)

// key and nonce sizes are hard-coded for compatibility with encrypted-config-value
const (
	aesKeySizeBits = 256
	nonceSizeBytes = 32
)

var aesAlgorithm = &algorithmDefinition{
	name: AES,
	generator: func() (pubKey encryption.Key, privKey encryption.Key, err error) {
		aesKey, err := encryption.NewAESGCM(aesKeySizeBits, nonceSizeBytes)
		if err != nil {
			return nil, nil, err
		}
		return aesKey, nil, nil
	},
	pubKeySerializer: &aesKeyWithSetNonceSerializer{nonceSizeBytes: nonceSizeBytes},
}

// aesKeyWithSetNonceSerializer is a serializer that will marshal and unmarshal AES-GCM keys. This serializer uses the
// provided nonce size as the fixed nonce size for all keys that it deals with. Because of this, the serialized
// representation does not include the nonce size as part of its representation.
type aesKeyWithSetNonceSerializer struct {
	nonceSizeBytes int
}

func (a *aesKeyWithSetNonceSerializer) Unmarshal(key []byte) (encryption.Key, error) {
	return encryption.AESGCMFromKey(key, a.nonceSizeBytes), nil
}

func (a *aesKeyWithSetNonceSerializer) Marshal(key encryption.Key) ([]byte, error) {
	k, ok := key.(encryption.AESGCMKey)
	if !ok {
		return nil, fmt.Errorf("key must be an AES GCM key, but is %q", reflect.TypeOf(key).String())
	}
	return k.Key(), nil
}
