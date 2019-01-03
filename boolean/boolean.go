// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package boolean

import (
	"strconv"
)

// Boolean wraps a bool value and provides encoding methods allowing its use as a map key in JSON objects.
type Boolean bool

func (b Boolean) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatBool(bool(b))), nil
}

func (b *Boolean) UnmarshalText(data []byte) error {
	rawBoolean, err := strconv.ParseBool(string(data))
	if err != nil {
		return err
	}
	*b = Boolean(rawBoolean)
	return nil
}
