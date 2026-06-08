// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/bearertoken"
	"github.com/palantir/pkg/conjuretypes/cj"
)

func TestBearerToken(t *testing.T) {
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "bearertoken",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), Value: "foo", JSON: "\"foo\""},
		},
		{
			Name: "null",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "null", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON null into Go bearertoken.Token after offset 4: want json string"},
		},
		{
			Name: "invalid",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "\" \"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \" \" into Go bearertoken.Token after offset 3: invalid character for bearer token"},
		},
		{
			Name: "non_ascii",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "\"☃\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"☃\" into Go bearertoken.Token after offset 5: invalid character for bearer token"},
		},
		{
			Name: "empty",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "\"\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"\" into Go bearertoken.Token after offset 2: empty bearer token"},
		},
		{
			Name: "trailing_padding",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), Value: "Zm9vYmFy==", JSON: "\"Zm9vYmFy==\""},
		},
		{
			Name: "interspersed_padding",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "\"fo=o\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"fo=o\" into Go bearertoken.Token after offset 6: invalid character for bearer token"},
		},
		{
			Name: "only_padding",
			Test: typeTestCase[bearertoken.Token]{Codec: cj.BearerToken[bearertoken.Token](), JSON: "\"==\"", SkipTestMarshal: true, ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string \"==\" into Go bearertoken.Token after offset 4: invalid character for bearer token"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
