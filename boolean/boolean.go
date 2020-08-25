// Copyright (c) 2020 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package boolean

import (
	"strconv"
)

// Boolean is an alias for bool which implements serialization matching the conjure wire specification.
// ref: https://github.com/palantir/conjure/blob/master/docs/spec/wire.md
type Boolean bool

// UnmarshalText implements encoding.TextUnmarshaler (used by encoding/json and others).
func (b *Boolean) UnmarshalText(data []byte) error {
	bo, err := strconv.ParseBool(string(data))
	if err != nil {
		return err
	}
	*b = Boolean(bo)
	return nil
}

// MarshalText implements encoding.TextMarshaler (used by encoding/json and others).
func (b Boolean) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatBool(bool(b))), nil
}
