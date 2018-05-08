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

var errorCodes = []remoting.ErrorCode{
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
	for _, ec := range errorCodes {
		serialized, err := json.Marshal(ec)
		assert.NoError(t, err)
		assert.Equal(t, `"`+ec.String()+`"`, string(serialized))
	}
}

func TestErrorCode_UnmarshalJSON(t *testing.T) {
	for _, ec := range errorCodes {
		serialized := `"` + ec.String() + `"`
		var actual remoting.ErrorCode
		err := json.Unmarshal([]byte(serialized), &actual)
		assert.NoError(t, err)
		assert.Equal(t, ec, actual)
	}
}
