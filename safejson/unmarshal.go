// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

import (
	"bytes"
	"encoding/json"
	"io"
)

func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalFrom(bytes.NewReader(data), v)
}

func UnmarshalFrom(reader io.Reader, v interface{}) error {
	dec := json.NewDecoder(reader)
	dec.UseNumber()
	return dec.Decode(v)
}
