// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"testing"

	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/palantir/pkg/rid"
)

func TestRID(t *testing.T) {
	type ridAlias rid.ResourceIdentifier
	tests := []struct {
		Name string
		Test typeTest
	}{
		{
			Name: "empty",
			Test: typeTestCase[rid.ResourceIdentifier]{Codec: cj.RID[rid.ResourceIdentifier](), Value: rid.ResourceIdentifier{}, JSON: "\"ri....\"", ErrUnmarshalJSONFrom: "json: cannot unmarshal JSON string into Go rid.ResourceIdentifier after offset 8: rid first segment (service) does not match ^[a-z][a-z0-9\\-]*$ pattern: rid third segment (type) does not match ^[a-z][a-z0-9\\-]*$ pattern: rid fourth segment (locator) does not match ^[a-zA-Z0-9\\-\\._]+$ pattern"},
		},
		{
			Name: "resource",
			Test: typeTestCase[rid.ResourceIdentifier]{Codec: cj.RID[rid.ResourceIdentifier](), Value: must(rid.ParseRID("ri.example.main.foo.bar")), JSON: "\"ri.example.main.foo.bar\""},
		},
		{
			Name: "alias",
			Test: typeTestCase[ridAlias]{Codec: cj.RID[ridAlias](), Value: ridAlias(must(rid.ParseRID("ri.example.main.foo.bar"))), JSON: "\"ri.example.main.foo.bar\""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("Marshal", tc.Test.TestMarshal)
			t.Run("Unmarshal", tc.Test.TestUnmarshal)
		})
	}
}
