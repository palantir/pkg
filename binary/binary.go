// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"encoding/base64"
)

var encoding = base64.StdEncoding

// Binary wraps binary data and provides encoding helpers using base64.StdEncoding.
// Use this type for binary fields serialized/deserialized as base64 text.
// We store then encoded string instead of an aliased []byte so this type can be used as a map key.
// Use Bytes() to access the raw bytes.
type Binary string

func New(data []byte) Binary {
	return Binary(encoding.EncodeToString(data))
}

func (b Binary) Bytes() ([]byte, error) {
	return encoding.DecodeString(string(b))
}

func (b Binary) MarshalText() (text []byte, err error) {
	// Test that we can decode data before returning invalid base64
	if _, err := b.Bytes(); err != nil {
		return nil, err
	}

	return []byte(b), nil
}

func (b *Binary) UnmarshalText(data []byte) error {
	// Test that we can decode data before storing invalid base64
	if _, err := b.Bytes(); err != nil {
		return err
	}

	*b = Binary(data)
	return nil
}
