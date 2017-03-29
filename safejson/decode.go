// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

func Decode(in, out interface{}) error {
	if out == nil {
		return nil // nothing to do
	}

	b, err := Marshal(in)
	if err != nil {
		return err
	}
	return Unmarshal(b, &out)
}
