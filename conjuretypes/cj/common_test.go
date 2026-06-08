// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typeTest interface {
	TestMarshal(t *testing.T)
	TestUnmarshal(t *testing.T)
}

type typeTestCase[T any] struct {
	Codec cj.Codec[T]

	// Value is the value to encode/decode.
	Value T
	// JSON is the JSON representation of the value.
	JSON string
	// Options (optional)
	json.Options

	SkipTestMarshal      bool
	ErrMarshalJSONTo     string // if nonempty, expect MarshalJSONTo to fail
	SkipTestUnmarshal    bool
	ErrUnmarshalJSONFrom string // if nonempty, expect UnmarshalJSONFrom to fail
}

func (tc typeTestCase[T]) TestMarshal(t *testing.T) {
	t.Helper()
	if tc.SkipTestMarshal {
		t.SkipNow()
	}
	buf := bytes.NewBuffer(nil)
	enc := jsontext.NewEncoder(buf, json.JoinOptions(cj.DefaultOptions, tc.Options))
	err := tc.Codec.MarshalJSONTo(enc, tc.Value)
	if tc.ErrMarshalJSONTo != "" {
		require.EqualErrorf(t, err, tc.ErrMarshalJSONTo, "expected error from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
		return
	}
	require.NoErrorf(t, err, "unexpected error from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
	got := strings.TrimSpace(buf.String())
	if assert.JSONEqf(t, tc.JSON, got, "unexpected JSON from (%T)(%v).MarshalJSON", tc.Value, tc.Value) {
		// If values are json-equivalent, assert JSON formatting/spacing
		assert.EqualValuesf(t, tc.JSON, got, "unexpected JSON formatting/spacing from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
	}
}

func (tc typeTestCase[T]) TestUnmarshal(t *testing.T) {
	t.Helper()
	if tc.SkipTestUnmarshal {
		t.SkipNow()
	}
	dec := jsontext.NewDecoder(strings.NewReader(tc.JSON), json.JoinOptions(cj.DefaultOptions, tc.Options))
	var got T
	err := tc.Codec.UnmarshalJSONFrom(dec, &got)
	if tc.ErrUnmarshalJSONFrom != "" {
		require.ErrorContainsf(t, err, expectedErrorFragment(tc.ErrUnmarshalJSONFrom), "expected error from (%T).UnmarshalJSON(%q)", tc.Value, tc.JSON)
		return
	}
	require.NoErrorf(t, err, "unexpected error from (%T).UnmarshalJSON(%q)", tc.Value, tc.JSON)
	_, err = dec.ReadToken()
	require.ErrorIsf(t, err, io.EOF, "unmarshal from (%T).UnmarshalJSON(%q) did not consume exactly one JSON value", tc.Value, tc.JSON)
	if isNaN(tc.Value) {
		assert.Truef(t, isNaN(got), "unexpected value from (%T).UnmarshalJSON(%q)", tc.Value, tc.JSON)
	} else {
		assert.EqualValuesf(t, tc.Value, got, "unexpected value from (%T).UnmarshalJSON(%q)", tc.Value, tc.JSON)
	}
}

func expectedErrorFragment(expected string) string {
	if idx := strings.Index(expected, ": "); idx >= 0 {
		return expected[idx+2:]
	}
	return expected
}

// Create a simple struct-like type that visits object fields
type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (t testStruct) MarshalJSON() ([]byte, error) {
	return json.Marshal(t, jsontext.AllowDuplicateNames(true))
}

func (t testStruct) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	{
		if err := enc.WriteToken(jsontext.String("name")); err != nil {
			return err
		}
		if err := cj.String[string]().MarshalJSONTo(enc, t.Name); err != nil {
			return err
		}
	}
	{
		if err := enc.WriteToken(jsontext.String("age")); err != nil {
			return err
		}
		if err := cj.Int32[int]().MarshalJSONTo(enc, t.Age); err != nil {
			return err
		}
	}
	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return err
	}
	return nil
}

func (t *testStruct) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *testStruct) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return cj.WrapSyntaxError(dec, err)
	}
	if kind := tok.Kind(); kind != jsontext.KindBeginObject {
		return cj.NewKindMismatchError(dec, kind, "testStruct opening brace")
	}
	var seenName bool
	var seenAge bool
	var unknownMembers []string
	for {
		key, err := dec.ReadToken()
		if err != nil {
			return cj.WrapSyntaxError(dec, err)
		}
		kind := key.Kind()
		if kind == jsontext.KindEndObject {
			break
		}
		if kind != jsontext.KindString {
			return cj.NewKindMismatchError(dec, kind, "ConjureDefinition closing brace or next key")
		}
		switch key.String() {
		case "name":
			if seenName {
				return cj.NewDuplicateFieldKeyError(dec, "testStruct[\"name\"]")
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &t.Name); err != nil {
				return err
			}
			seenName = true
		case "age":
			if seenAge {
				return cj.NewDuplicateFieldKeyError(dec, "testStruct[\"age\"]")
			}
			if err := cj.Int32[int]().UnmarshalJSONFrom(dec, &t.Age); err != nil {
				return err
			}
			seenAge = true
		default:
			unknownMembers = append(unknownMembers, key.String())
			if err := dec.SkipValue(); err != nil {
				return err
			}
		}
	}
	var missingFields []string
	if !seenName {
		missingFields = append(missingFields, "name")
	}
	if !seenAge {
		missingFields = append(missingFields, "age")
	}
	if len(missingFields) > 0 {
		return cj.NewMissingFieldsError(dec, "testStruct", missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := json.GetOption(dec.Options(), json.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError(dec, "testStruct", unknownMembers)
		}
	}
	return nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func isNaN(v any) bool {
	f, ok := v.(float64)
	return ok && math.IsNaN(f)
}
