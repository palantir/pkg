// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package safejson implements encoding and decoding of JSON with the following
// special configurations:
//
// - json.Decoder.UseNumber
// - json.Encoder.SetEscapeHTML(false)
// - json.Encoder.Encode(big floats as json.Number)
package safejson

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/palantir/pkg/transform"
)

// Marshal returns the JSON encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}), nil
}

// MarshalIndent is like Marshal but indents the output to be human readable.
func MarshalIndent(v interface{}) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, b, "", "    "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// An Encoder writes JSON objects to an output stream.
//
// Use NewEncoder to make a new Encoder. NewEncoder is implemented differently
// from Go version 1.7 onward.
type Encoder struct {
	enc *json.Encoder
}

// Encode writes the JSON encoding of v to the stream, followed by a newline
// character.
//
// See the documentation for json.Marshal for details about the conversion of Go
// values to JSON.
func (e *Encoder) Encode(v interface{}) error {
	return e.enc.Encode(bigFloatToJSONNumber(v))
}

// bigFloatToJSONNumber returns json.Number in place of big.Float.
//
// This is necessary because big.Float's text marshaller turns them into normal
// strings. For usage, see Encoder.Encode.
func bigFloatToJSONNumber(obj interface{}) interface{} {
	rules := transform.Rules{
		func(bf *big.Float) json.Number {
			return json.Number(bf.Text('g', -1))
		},
	}
	return rules.Apply(obj)
}
