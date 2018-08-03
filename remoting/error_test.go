// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/remoting"
)

func TestError_Error(t *testing.T) {
	err := remoting.NewError(
		remoting.MustErrorType(
			remoting.ErrorCodeTimeout,
			"MyApplication:DatabaseTimeout",
		),
		map[string]string{
			"ttl": "10s",
		},
	)
	assert.EqualError(t, err, fmt.Sprintf("TIMEOUT MyApplication:DatabaseTimeout (%s)", err.InstanceID()))
}

func TestWriteErrorResponse_ValidateJSON(t *testing.T) {
	remotingError := remoting.NewError(
		remoting.MustErrorType(
			remoting.ErrorCodeTimeout,
			"MyApplication:Timeout",
		),
		map[string]string{
			"ttl": "10s",
		},
	)

	expectedJSON := fmt.Sprintf(`{
  "errorCode": "TIMEOUT",
  "errorName": "MyApplication:Timeout",
  "errorInstanceId": "%s",
  "parameters": {
    "ttl": "10s"
  }
}`, remotingError.InstanceID())

	recorder := httptest.NewRecorder()
	remoting.WriteErrorResponse(recorder, remotingError)
	response := recorder.Result()

	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	var buffer bytes.Buffer
	require.NoError(t, json.Indent(&buffer, body, "", "  "))
	assert.Equal(t, expectedJSON, buffer.String())
}

func TestWriteErrorResponse_Then_ErrorFromResponse(t *testing.T) {
	tests := map[string]remoting.Error{
		"default timeout": remoting.NewTimeoutError(
			map[string]string{
				"ttl": "10s",
			},
		),
		"custom timeout": remoting.NewError(
			remoting.MustErrorType(
				remoting.ErrorCodeTimeout,
				"MyApplication:Timeout",
			),
			map[string]string{
				"ttl": "10s",
			},
		),
		"custom not found": remoting.NewError(
			remoting.MustErrorType(
				remoting.ErrorCodeNotFound,
				"MyApplication:MissingData",
			),
			map[string]string{},
		),
		"custom client": remoting.NewError(
			remoting.MustErrorType(
				remoting.ErrorCodeCustomClient,
				"MyApplication:CustomClientError",
			),
			map[string]string{},
		),
		"custom server": remoting.NewError(
			remoting.MustErrorType(
				remoting.ErrorCodeCustomServer,
				"MyApplication:CustomServerError",
			),
			map[string]string{},
		),
		"custom server with nil params": remoting.NewError(
			remoting.MustErrorType(
				remoting.ErrorCodeCustomServer,
				"MyApplication:CustomServerError",
			),
			nil,
		),
	}

	for name, expected := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			remoting.WriteErrorResponse(recorder, expected)

			response := recorder.Result()

			assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
			assert.Equal(t, expected.Code().StatusCode(), response.StatusCode)

			actual, err := remoting.ErrorFromResponse(response)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
