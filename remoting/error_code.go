// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode is an enum describing error category.
//
// Each error code has associated HTTP status codes.
type ErrorCode int16

const (
	_                              ErrorCode = iota // there is no good candidate for zero value
	ErrorCodePermissionDenied                       // ErrorCodePermissionDenied has status code 403 Forbidden.
	ErrorCodeInvalidArgument                        // ErrorCodeInvalidArgument has status code 400 BadRequest.
	ErrorCodeNotFound                               // ErrorCodeNotFound  has status code 404 NotFound.
	ErrorCodeConflict                               // ErrorCodeConflict has status code 409 Conflict.
	ErrorCodeRequestEntityTooLarge                  // ErrorCodeRequestEntityTooLarge has status code 413 RequestEntityTooLarge.
	ErrorCodeFailedPrecondition                     // ErrorCodeFailedPrecondition has status code 500 InternalServerError.
	ErrorCodeInternal                               // ErrorCodeInternal has status code 500 InternalServerError.
	ErrorCodeTimeout                                // ErrorCodeTimeout has status code 500 InternalServerError.
	ErrorCodeCustomClient                           // ErrorCodeCustomClient has status code 400 BadRequest.
	ErrorCodeCustomServer                           // ErrorCodeCustomServer has status code 500 InternalServerError.
)

// StatusCode returns HTTP status code associated with this error code.
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

// String representation of this error code.
//
// For example "PERMISSION_DENIED" or "TIMEOUT".
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
	return fmt.Sprintf("<invalid error code: %d>", ec)
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
