// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package safejson

import (
	"encoding/json"
	"io"
)

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	return &Encoder{enc: e}
}
