// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safelong

import (
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"strconv"
)

const (
	safeIntVal = int64(1) << 53
	minVal     = -safeIntVal + 1
	maxVal     = safeIntVal - 1
)

type SafeLong int64

func NewSafeLong(val int64) (SafeLong, error) {
	if err := validate(val); err != nil {
		return 0, err
	}
	return SafeLong(val), nil
}

func ParseSafeLong(s string) (SafeLong, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return NewSafeLong(i)
}

func (s *SafeLong) UnmarshalJSON(b []byte) error {
	var val int64
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}

	newVal, err := NewSafeLong(val)
	if err != nil {
		return err
	}
	*s = newVal

	return nil
}

func (s SafeLong) MarshalJSON() ([]byte, error) {
	if err := validate(int64(s)); err != nil {
		return nil, err
	}
	return strconv.AppendInt(nil, int64(s), 10), nil
}

// MarshalJSONTo writes s as a JSON number, or as a JSON string when s is an object name,
// so that map[SafeLong]T works without encoding/json's legacy rule of ignoring MarshalJSON for map keys.
func (s SafeLong) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := validate(int64(s)); err != nil {
		return err
	}
	if isObjectName(enc.StackIndex(enc.StackDepth())) {
		return enc.WriteToken(jsontext.String(strconv.FormatInt(int64(s), 10)))
	}
	return enc.WriteToken(jsontext.Int(int64(s)))
}

// UnmarshalJSONFrom reads s from a JSON number, or from a JSON string when s is an object name.
func (s *SafeLong) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var val int64
	if isObjectName(dec.StackIndex(dec.StackDepth())) {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		if val, err = strconv.ParseInt(tok.String(), 10, 64); err != nil {
			return err
		}
	} else if err := jsonv2.UnmarshalDecode(dec, &val); err != nil {
		return err
	}
	newVal, err := NewSafeLong(val)
	if err != nil {
		return err
	}
	*s = newVal
	return nil
}

// isObjectName reports whether the next token at the given stack position is an object name.
// An object's length counts its names and values, so the next token is a name when the length is even.
func isObjectName(kind jsontext.Kind, length int64) bool {
	return kind == '{' && length%2 == 0
}

func validate(val int64) error {
	if val < minVal || val > maxVal {
		return fmt.Errorf("%d is not a valid value for a SafeLong as it is not safely representable in Javascript: must be between %d and %d", val, minVal, maxVal)
	}
	return nil
}
