// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"encoding/base64"
)

// Binary wraps binary data and provides MarshalText/UnmarshalText encoding helpers using base64.StdEncoding.
// We use a struct instead of an aliased []byte so this type can be used as a map key.
type Binary struct {
	Data []byte
}

func New(data []byte) Binary {
	return Binary{Data: data}
}

var encoding = base64.StdEncoding

func (b Binary) String() string {
	return encoding.EncodeToString(b.Data)
}

func (b Binary) MarshalText() ([]byte, error) {
	encoded := make([]byte, encoding.EncodedLen(len(b.Data)))
	encoding.Encode(encoded, b.Data)
	return encoded, nil
}

func (b *Binary) UnmarshalText(data []byte) error {
	decoded := make([]byte, encoding.DecodedLen(len(data)))
	n, err := encoding.Decode(decoded, data)
	if err != nil {
		return err
	}
	b.Data = decoded[:n]
	return nil
}
