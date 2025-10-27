// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

import (
	"bytes"
)

// Unmarshal unmarshals the provided bytes (which should be valid JSON)
// into "v" using safejson.Decoder.
//
// Note: Unlike encoding/json.Unmarshal, this function does NOT return an error
// if there are trailing bytes after the first valid JSON value. The stdlib
// json.Unmarshal would return an error like "invalid character 'b' after top-level value"
// in such cases, but safejson.Unmarshal will successfully unmarshal the first value
// and ignore any trailing bytes.
func Unmarshal(data []byte, v interface{}) error {
	return Decoder(bytes.NewReader(data)).Decode(v)
}
