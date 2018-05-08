// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/remoting"
)

func TestError_Error(t *testing.T) {
	err := remoting.NewError(
		remoting.ErrorCodeTimeout,
		"MyApplication:Timeout",
		map[string]string{
			"ttl": "10s",
		},
	)

	assert.EqualError(t, err, fmt.Sprintf("TIMEOUT MyApplication:Timeout (%s)", err.InstanceID()))
}

func TestError_WriteResponse_Then_ErrorFromResponse(t *testing.T) {
	tests := map[string]remoting.Error{
		"timeout": remoting.NewError(
			remoting.ErrorCodeTimeout,
			"MyApplication:Timeout",
			map[string]string{
				"ttl": "10s",
			},
		),
		"not found": remoting.NewError(
			remoting.ErrorCodeNotFound,
			"MyApplication:MissingData",
			map[string]string{},
		),
		"custom server": remoting.NewCustomServerError(
			"MyApplication:CustomServerError",
			map[string]string{},
		),
		"custom client": remoting.NewCustomClientError(
			"MyApplication:CustomClientError",
			map[string]string{},
		),
	}

	for name, expected := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			expected.WriteResponse(recorder)

			response := recorder.Result()

			assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
			assert.Equal(t, expected.Code().StatusCode(), response.StatusCode)

			actual, err := remoting.ErrorFromResponse(response)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
