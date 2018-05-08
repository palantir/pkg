// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/remoting"
)

func TestMustErrorName_SucceedForValidErrorName(t *testing.T) {
	remoting.MustErrorName("MyApplication:MyCustomError")
}

func TestMustErrorName_PanicForInvalidErrorName(t *testing.T) {
	for _, tc := range []string{
		":PascalCase",
		"PascalCase:",
		":",
		"PascalCase",
		"notPascalCase:PascalCase",
		"PascalCase:notPascalCase",
		"PascalCase:PascalCase:PascalCase",
	} {
		t.Run(fmt.Sprintf(`"%s"`, tc), func(t *testing.T) {
			panicked := false
			defer func() {
				assert.True(t, panicked)
			}()
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			remoting.MustErrorName(tc)
		})
	}
}

var ens = []remoting.ErrorName{
	remoting.MustErrorName("MyApplication:MyCustomError"),
	remoting.ErrorNameDefaultPermissionDenied,
	remoting.ErrorNameDefaultInvalidArgument,
	remoting.ErrorNameDefaultNotFound,
	remoting.ErrorNameDefaultConflict,
	remoting.ErrorNameDefaultRequestEntityTooLarge,
	remoting.ErrorNameDefaultFailedPrecondition,
	remoting.ErrorNameDefaultInternal,
	remoting.ErrorNameDefaultTimeout,
}

func TestErrorName_MarshalJSON(t *testing.T) {
	for _, en := range ens {
		serialized, err := en.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, `"`+string(en)+`"`, string(serialized))
	}
}

func TestErrorName_UnmarshalJSON(t *testing.T) {
	for _, en := range ens {
		serialized := `"` + string(en) + `"`
		var actual remoting.ErrorName
		err := json.Unmarshal([]byte(serialized), &actual)
		assert.NoError(t, err)
		assert.Equal(t, en, actual)
	}
}
