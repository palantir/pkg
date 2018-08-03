// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	DefaultPermissionDenied      = ErrorType{ErrorCodePermissionDenied, "Default:PermissionDenied"}
	DefaultInvalidArgument       = ErrorType{ErrorCodeInvalidArgument, "Default:InvalidArgument"}
	DefaultNotFound              = ErrorType{ErrorCodeNotFound, "Default:NotFound"}
	DefaultConflict              = ErrorType{ErrorCodeConflict, "Default:Conflict"}
	DefaultRequestEntityTooLarge = ErrorType{ErrorCodeRequestEntityTooLarge, "Default:RequestEntityTooLarge"}
	DefaultFailedPrecondition    = ErrorType{ErrorCodeFailedPrecondition, "Default:FailedPrecondition"}
	DefaultInternal              = ErrorType{ErrorCodeInternal, "Default:Internal"}
	DefaultTimeout               = ErrorType{ErrorCodeTimeout, "Default:Timeout"}
)

// ErrorType represents certain class of errors. Each error type is uniquely identified by an error name
// and has assigned one of possible error codes.
//
// Error type should be a compile-time constant and is considered part of the API
// of a service that produces error of such type.
type ErrorType struct {
	code ErrorCode
	name string
}

func MustErrorType(code ErrorCode, name string) ErrorType {
	errorType, err := NewErrorType(code, name)
	if err != nil {
		panic(err)
	}
	return errorType
}

// NewErrorType returns error type with the provided error code and name,
// or returns an error if error of such type cannot be created.
//
// Error name must be in the "PascalCase:PascalCase" format,
// for example "Default:PermissionDenied", "Facebook:LikeAlreadyGiven" or
// "MyApplication:ErrorSpecificToMyBusinessDomain". The first part of an error name is a namespace
// and the second part contains cause on an error. "Default" namespace is reserved for
// error types defined in this package.
//
// Example:
//
//  var ErrorLikeAlreadyGiven = remoting.MustErrorType(
//    remoting.ErrorCodeConflict,
//    "Facebook:LikeAlreadyGiven",
//  )
//
func NewErrorType(code ErrorCode, name string) (ErrorType, error) {
	if err := verifyErrorNameString(name); err != nil {
		return ErrorType{}, err
	}
	if err := verifyErrorCodeErrorNameCombination(code, name); err != nil {
		return ErrorType{}, err
	}
	return ErrorType{
		code: code,
		name: name,
	}, nil
}

// String representation of an error type,
// for example "NOT_FOUND MyApplication:MissingData".
func (et ErrorType) String() string {
	return fmt.Sprintf("%s %s", et.code, et.name)
}

func (et ErrorType) Code() ErrorCode {
	return et.code
}

func (et ErrorType) Name() string {
	return et.name
}

// Using regexp from https://github.com/palantir/http-remoting-api/blob/develop/errors/src/main/java/com/palantir/remoting/api/errors/ErrorType.java#L33.
//
// Note that this regexp does not accept single upper case letter, see
// https://github.com/palantir/http-remoting-api/issues/110.
var errorNameRegexp = regexp.MustCompile("^(([A-Z][a-z0-9]+)+):(([A-Z][a-z0-9]+)+)$")

const (
	errorNamePermissionDenied      = "Default:PermissionDenied"
	errorNameInvalidArgument       = "Default:InvalidArgument"
	errorNameNotFound              = "Default:NotFound"
	errorNameConflict              = "Default:Conflict"
	errorNameRequestEntityTooLarge = "Default:RequestEntityTooLarge"
	errorNameFailedPrecondition    = "Default:FailedPrecondition"
	errorNameInternal              = "Default:Internal"
	errorNameTimeout               = "Default:Timeout"
)

func verifyErrorNameString(name string) error {
	if !errorNameRegexp.MatchString(name) {
		return fmt.Errorf("remoting: error name does not match regexp `%s`", errorNameRegexp.String())
	}
	if strings.HasPrefix(name, "Default:") {
		switch name {
		case errorNamePermissionDenied:
		case errorNameInvalidArgument:
		case errorNameNotFound:
		case errorNameConflict:
		case errorNameRequestEntityTooLarge:
		case errorNameFailedPrecondition:
		case errorNameInternal:
		case errorNameTimeout:
		default:
			return fmt.Errorf("remoting: error name with default namespace cannot use custom cause")
		}
	}
	return nil
}

func verifyErrorCodeErrorNameCombination(code ErrorCode, name string) error {
	if strings.HasPrefix(name, "Default:") {
		defaultErrorType := ErrorType{code, name}
		switch defaultErrorType {
		case DefaultPermissionDenied:
		case DefaultInvalidArgument:
		case DefaultNotFound:
		case DefaultConflict:
		case DefaultRequestEntityTooLarge:
		case DefaultFailedPrecondition:
		case DefaultInternal:
		case DefaultTimeout:
		default:
			return fmt.Errorf("remoting: invalid combination of default error name and error code")
		}
	}
	return nil
}
