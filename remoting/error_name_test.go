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

var invalidErrorNames = map[string]string{
	"no namespace":                             ":PascalCase",
	"no cause":                                 "PascalCase:",
	"no namespace nor cause":                   ":",
	"no colon":                                 "PascalCase",
	"namespace with invalid case":              "notPascalCase:PascalCase",
	"cause with invalid case":                  "PascalCase:notPascalCase",
	"name with three parts":                    "PascalCase:PascalCase:PascalCase",
	"custom error name with default namespace": "Default:CustomError",
}

var validErrorNames = []remoting.ErrorName{
	remoting.MustErrorName("MyApplication:MyCustomError"),
	remoting.MustErrorName("Aa:Bb"),
	remoting.MustErrorName("A1:B2"),
	remoting.MustErrorName("A1A1:B2B2"),
	remoting.MustErrorName("Aa1Aa1:Bb2Bb2"),
	remoting.MustErrorName("A1aA1a:B2bB2b"),
	remoting.MustErrorName("MyApplication:MyCustomError"),
	remoting.ErrorNamePermissionDenied,
	remoting.ErrorNameInvalidArgument,
	remoting.ErrorNameNotFound,
	remoting.ErrorNameConflict,
	remoting.ErrorNameRequestEntityTooLarge,
	remoting.ErrorNameFailedPrecondition,
	remoting.ErrorNameInternal,
	remoting.ErrorNameTimeout,
}

func TestMustErrorName_PanicForInvalidErrorName(t *testing.T) {
	for name, en := range invalidErrorNames {
		t.Run(name, func(t *testing.T) {
			panicked := false
			defer func() {
				assert.True(t, panicked)
			}()
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			remoting.MustErrorName(en)
		})
	}
}

func TestErrorName_MarshalJSON(t *testing.T) {
	for i, en := range validErrorNames {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			serialized, err := en.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, `"`+string(en)+`"`, string(serialized))
		})
	}
}

func TestErrorName_UnmarshalJSON(t *testing.T) {
	for i, en := range validErrorNames {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			serialized := `"` + string(en) + `"`
			var actual remoting.ErrorName
			err := json.Unmarshal([]byte(serialized), &actual)
			assert.NoError(t, err)
			assert.Equal(t, en, actual)
		})
	}
	for name, en := range invalidErrorNames {
		t.Run(name, func(t *testing.T) {
			serialized := `"` + string(en) + `"`
			var actual remoting.ErrorName
			err := json.Unmarshal([]byte(serialized), &actual)
			assert.Error(t, err)
		})
	}
}
