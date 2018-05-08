// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode is an enum identifying the type of error.
type ErrorCode int16

const (
	_ ErrorCode = iota // there is no good candidate for zero value
	ErrorCodePermissionDenied
	ErrorCodeInvalidArgument
	ErrorCodeNotFound
	ErrorCodeConflict
	ErrorCodeRequestEntityTooLarge
	ErrorCodeFailedPrecondition
	ErrorCodeInternal
	ErrorCodeTimeout
	ErrorCodeCustomClient
	ErrorCodeCustomServer
)

func (ec ErrorCode) StatusCode() int {
	switch ec {
	case ErrorCodePermissionDenied:
		return http.StatusForbidden
	case ErrorCodeInvalidArgument:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeRequestEntityTooLarge:
		return http.StatusRequestEntityTooLarge
	case ErrorCodeFailedPrecondition:
		return http.StatusInternalServerError
	case ErrorCodeInternal:
		return http.StatusInternalServerError
	case ErrorCodeTimeout:
		return http.StatusInternalServerError
	case ErrorCodeCustomClient:
		return http.StatusBadRequest
	case ErrorCodeCustomServer:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

func (ec ErrorCode) String() string {
	switch ec {
	case ErrorCodePermissionDenied:
		return "PERMISSION_DENIED"
	case ErrorCodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case ErrorCodeNotFound:
		return "NOT_FOUND"
	case ErrorCodeConflict:
		return "CONFLICT"
	case ErrorCodeRequestEntityTooLarge:
		return "REQUEST_ENTITY_TOO_LARGE"
	case ErrorCodeFailedPrecondition:
		return "FAILED_PRECONDITION"
	case ErrorCodeInternal:
		return "INTERNAL"
	case ErrorCodeTimeout:
		return "TIMEOUT"
	case ErrorCodeCustomClient:
		return "CUSTOM_CLIENT"
	case ErrorCodeCustomServer:
		return "CUSTOM_SERVER"
	}
	return "<invalid error code>"
}

func (ec ErrorCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(ec.String())
}

func (ec *ErrorCode) UnmarshalJSON(data []byte) error {
	var ecString string

	err := json.Unmarshal(data, &ecString)
	if err != nil {
		return err
	}

	switch ecString {
	case "PERMISSION_DENIED":
		*ec = ErrorCodePermissionDenied
	case "INVALID_ARGUMENT":
		*ec = ErrorCodeInvalidArgument
	case "NOT_FOUND":
		*ec = ErrorCodeNotFound
	case "CONFLICT":
		*ec = ErrorCodeConflict
	case "REQUEST_ENTITY_TOO_LARGE":
		*ec = ErrorCodeRequestEntityTooLarge
	case "FAILED_PRECONDITION":
		*ec = ErrorCodeFailedPrecondition
	case "INTERNAL":
		*ec = ErrorCodeInternal
	case "TIMEOUT":
		*ec = ErrorCodeTimeout
	case "CUSTOM_CLIENT":
		*ec = ErrorCodeCustomClient
	case "CUSTOM_SERVER":
		*ec = ErrorCodeCustomServer
	default:
		return fmt.Errorf(`remoting: unknown error code string`)
	}
	return nil
}
