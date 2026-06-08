// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"fmt"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
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

type simpleStruct struct {
	Name string
	Num  int
}

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
