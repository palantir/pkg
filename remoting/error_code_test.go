// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/remoting"
)

func TestErrorCode_String(t *testing.T) {
	for ec, expectedString := range map[remoting.ErrorCode]string{
		remoting.ErrorCodePermissionDenied:      "PERMISSION_DENIED",
		remoting.ErrorCodeInvalidArgument:       "INVALID_ARGUMENT",
		remoting.ErrorCodeNotFound:              "NOT_FOUND",
		remoting.ErrorCodeConflict:              "CONFLICT",
		remoting.ErrorCodeRequestEntityTooLarge: "REQUEST_ENTITY_TOO_LARGE",
		remoting.ErrorCodeFailedPrecondition:    "FAILED_PRECONDITION",
		remoting.ErrorCodeInternal:              "INTERNAL",
		remoting.ErrorCodeTimeout:               "TIMEOUT",
		remoting.ErrorCodeCustomClient:          "CUSTOM_CLIENT",
		remoting.ErrorCodeCustomServer:          "CUSTOM_SERVER",
		remoting.ErrorCode(0):                   "<invalid error code: 0>",
		remoting.ErrorCode(200):                 "<invalid error code: 200>",
	} {
		assert.Equal(t, expectedString, ec.String())
	}
}

var validErrorCodes = []remoting.ErrorCode{
	remoting.ErrorCodePermissionDenied,
	remoting.ErrorCodeInvalidArgument,
	remoting.ErrorCodeNotFound,
	remoting.ErrorCodeConflict,
	remoting.ErrorCodeRequestEntityTooLarge,
	remoting.ErrorCodeFailedPrecondition,
	remoting.ErrorCodeInternal,
	remoting.ErrorCodeTimeout,
	remoting.ErrorCodeCustomClient,
	remoting.ErrorCodeCustomServer,
}

func TestErrorCode_MarshalJSON(t *testing.T) {
	for _, ec := range validErrorCodes {
		t.Run(ec.String(), func(t *testing.T) {
			serialized, err := json.Marshal(ec)
			assert.NoError(t, err)
			assert.Equal(t, `"`+ec.String()+`"`, string(serialized))
		})
	}
}

func TestErrorCode_UnmarshalJSON(t *testing.T) {
	for _, ec := range validErrorCodes {
		t.Run(ec.String(), func(t *testing.T) {
			serialized := `"` + ec.String() + `"`
			var actual remoting.ErrorCode
			err := json.Unmarshal([]byte(serialized), &actual)
			assert.NoError(t, err)
			assert.Equal(t, ec, actual)
		})
	}

	for _, s := range []string{
		"INVALID_ERROR_CODE",
	} {
		t.Run(s, func(t *testing.T) {
			serialized := `"` + s + `"`
			var actual remoting.ErrorCode
			err := json.Unmarshal([]byte(serialized), &actual)
			assert.Error(t, err)
		})
	}
}
