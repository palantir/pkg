// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
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
		requireErrorMatches(t, err, tc.ErrMarshalJSONTo, "expected error from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
		return
	}
	require.NoErrorf(t, err, "unexpected error from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
	got := strings.TrimSpace(buf.String())
	if assert.JSONEqf(t, tc.JSON, got, "unexpected JSON from (%T)(%v).MarshalJSON", tc.Value, tc.Value) {
		// If values are json-equivalent, assert JSON formatting/spacing
		assert.EqualValuesf(t, tc.JSON, got, "unexpected JSON formatting/spacing from (%T)(%v).MarshalJSON", tc.Value, tc.Value)
	}
}

// codecUnmarshaler adapts a Codec[T] to json.UnmarshalerFrom and mirrors the
// GoType handling of the package entry points (see fillGoType in codec.go).
// Decoding through it lets the harness assert the same SemanticError envelope
// that production code produces: json/v2 fills in the "unmarshal" action and the
// adapter records the Go type that a bare codec call would leave unset.
type codecUnmarshaler[T any] struct {
	codec    cj.Codec[T]
	receiver *T
}

func (c codecUnmarshaler[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := c.codec.UnmarshalJSONFrom(dec, c.receiver)
	if serr, ok := errors.AsType[*json.SemanticError](err); ok && serr.GoType == nil {
		serr.GoType = reflect.TypeFor[T]()
	}
	return err
}

func (tc typeTestCase[T]) TestUnmarshal(t *testing.T) {
	t.Helper()
	if tc.SkipTestUnmarshal {
		t.SkipNow()
	}
	dec := jsontext.NewDecoder(strings.NewReader(tc.JSON), json.JoinOptions(cj.DefaultOptions, tc.Options))
	var got T
	err := json.UnmarshalDecode(dec, &codecUnmarshaler[T]{tc.Codec, &got})
	if tc.ErrUnmarshalJSONFrom != "" {
		requireErrorMatches(t, err, tc.ErrUnmarshalJSONFrom, "expected error from (%T).UnmarshalJSON(%q)", tc.Value, tc.JSON)
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
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &t.Name); err != nil {
				return err
			}
			seenName = true
		case "age":
			if seenAge {
				return cj.NewDuplicateFieldKeyError(dec)
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
		return cj.NewMissingFieldsError(dec, missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := json.GetOption(dec.Options(), json.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError(dec, unknownMembers)
		}
	}
	return nil
}

// requireErrorMatches asserts that err's message matches want apart from
// json/v2's deliberately-randomized modal verb. The json package switches
// between "json: cannot " and "json: unable to " once per process (tied to Go's
// map-iteration randomization; see errorModalVerb) to discourage depending on
// the exact wording. That verb is the only varying part and it is a prefix, so
// we match the stable suffix. Everything after the verb -- action, Go type,
// byte offset, JSON pointer, and cause -- is asserted verbatim.
func requireErrorMatches(t *testing.T, err error, want, format string, args ...any) {
	t.Helper()
	require.Errorf(t, err, format, args...)
	suffix := strings.TrimPrefix(want, "json: cannot ")
	suffix = strings.TrimPrefix(suffix, "json: unable to ")
	require.Truef(t, strings.HasSuffix(err.Error(), suffix),
		"error %q does not end with %q; "+format, append([]any{err.Error(), suffix}, args...)...)
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
