// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"encoding/base64"
)

var b64 = base64.StdEncoding

// Binary wraps binary data and provides encoding helpers using base64.StdEncoding.
// Use this type for binary fields serialized/deserialized as base64 text.
// Type is specified as string rather than []byte so that it can be used as a map key.
// Use Bytes() to access the raw bytes.
// Values of this type are only valid when the backing string is Base64-encoded using standard encoding.
type Binary string

func New(data []byte) Binary {
	return Binary(b64.EncodeToString(data))
}

func (b Binary) Bytes() ([]byte, error) {
	return b64.DecodeString(string(b))
}

func (b Binary) String() string {
	return string(b)
}

func (b Binary) MarshalText() (text []byte, err error) {
	// Verify that data is base64-encoded
	if _, err := b.Bytes(); err != nil {
		return nil, err
	}

	return []byte(b), nil
}

func (b *Binary) UnmarshalText(data []byte) error {
	// Verify that data is base64-encoded
	if _, err := Binary(data).Bytes(); err != nil {
		return err
	}

	*b = Binary(data)
	return nil
}
