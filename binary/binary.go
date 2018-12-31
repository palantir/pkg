// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"encoding/base64"
)

// Binary wraps a []byte and provides MarshalText/UnmarshalText encoding helpers using base64.StdEncoding.
type Binary []byte

var encoding = base64.StdEncoding

func (b Binary) MarshalText() ([]byte, error) {
	encoded := make([]byte, encoding.EncodedLen(len(b)))
	encoding.Encode(encoded, b)
	return encoded, nil
}

func (b *Binary) UnmarshalText(data []byte) error {
	decoded := make([]byte, encoding.DecodedLen(len(data)))
	n, err := encoding.Decode(decoded, data)
	if err != nil {
		return err
	}
	*b = Binary(decoded[:n])
	return nil
}
