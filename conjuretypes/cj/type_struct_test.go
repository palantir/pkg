// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"fmt"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/assert"
)

func TestStructs(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "simpleStruct",
			Test: typeTestCase[simpleStruct]{Codec: cj.Struct[simpleStruct](), Value: simpleStruct{Name: "foo", Num: 42}, JSON: `{"name":"foo","num":42}`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}

// TestStructEqual verifies the codec delegates to the type's own Equal method. The
// ptrStruct case is the proof: its Equal compares by pointee, so two distinct
// pointers to equal values report equal -- the opposite of Go's == on the struct.
func TestStructEqual(t *testing.T) {
	codec := cj.Struct[simpleStruct]()
	assert.True(t, codec.Equal(simpleStruct{Name: "a", Num: 1}, simpleStruct{Name: "a", Num: 1}))
	assert.False(t, codec.Equal(simpleStruct{Name: "a", Num: 1}, simpleStruct{Name: "b", Num: 1}))

	ptrCodec := cj.Struct[ptrStruct]()
	x, y, z := 7, 7, 8
	assert.True(t, ptrCodec.Equal(ptrStruct{Opt: &x}, ptrStruct{Opt: &y}), "delegates to Equal, which compares by pointee, not identity")
	assert.False(t, ptrCodec.Equal(ptrStruct{Opt: &x}, ptrStruct{Opt: &z}))
}

type simpleStruct struct {
	Name string
	Num  int
}

func (s simpleStruct) Equal(other simpleStruct) bool {
	return s.Name == other.Name && s.Num == other.Num
}

// ptrStruct's Equal compares by pointee, demonstrating that the codec adopts the
// type's own equality semantics. The methods exist only to satisfy the Struct
// constructor's constraint.
type ptrStruct struct {
	Opt *int
}

func (s ptrStruct) Equal(other ptrStruct) bool {
	if s.Opt == nil || other.Opt == nil {
		return s.Opt == other.Opt
	}
	return *s.Opt == *other.Opt
}

func (ptrStruct) MarshalJSONTo(*jsontext.Encoder) error      { return nil }
func (*ptrStruct) UnmarshalJSONFrom(*jsontext.Decoder) error { return nil }

func (s simpleStruct) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(jsontext.Value(fmt.Sprintf(`{"name":"%s","num":%d}`, s.Name, s.Num)))
}

func (s *simpleStruct) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	if kind := tok.Kind(); kind != jsontext.KindBeginObject {
		return cj.NewKindMismatchError(dec, kind, "json object")
	}
	for {
		key, err := dec.ReadToken()
		if err != nil {
			return err
		}
		switch kind := key.Kind(); kind {
		case jsontext.KindString:
			switch key.String() {
			case "name":
				if err := cj.String[string]().UnmarshalJSONFrom(dec, &s.Name); err != nil {
					return err
				}
			case "num":
				if err := cj.Int32[int]().UnmarshalJSONFrom(dec, &s.Num); err != nil {
					return err
				}
			}
		case jsontext.KindEndObject:
			return nil // end of object
		default:
			return cj.NewKindMismatchError(dec, kind, "next key or closing brace for simpleStruct")
		}
	}
}
