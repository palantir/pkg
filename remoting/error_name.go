// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	ErrorNamePermissionDenied      ErrorName = "Default:PermissionDenied"
	ErrorNameInvalidArgument       ErrorName = "Default:InvalidArgument"
	ErrorNameNotFound              ErrorName = "Default:NotFound"
	ErrorNameConflict              ErrorName = "Default:Conflict"
	ErrorNameRequestEntityTooLarge ErrorName = "Default:RequestEntityTooLarge"
	ErrorNameFailedPrecondition    ErrorName = "Default:FailedPrecondition"
	ErrorNameInternal              ErrorName = "Default:Internal"
	ErrorNameTimeout               ErrorName = "Default:Timeout"
)

// ErrorName identifies error type.
//
// Error name should be a compile-time constant and is considered part of the API
// of a service that produces error with such name.
//
// It must be in the "PascalCase:PascalCase" format, for example "Default:PermissionDenied",
// "Facebook:LikeAlreadyGiven" or "MyApplication:ErrorSpecificToMyBusinessDomain".
// The first part of an error name is a namespace and the second part contains cause on an error.
// "Default" namespace is reserved for predefined error names.
//
// Use MustErrorName() to ensure valid custom ErrorName is created.
type ErrorName string

func (en ErrorName) IsDefault() bool {
	return strings.HasPrefix(string(en), "Default:")
}

// MustErrorName creates error name or panics.
func MustErrorName(s string) ErrorName {
	if err := verifyErrorNameString(s); err != nil {
		panic(err)
	}
	return ErrorName(s)
}

// Using regexp from https://github.com/palantir/http-remoting-api/blob/develop/errors/src/main/java/com/palantir/remoting/api/errors/ErrorType.java#L33.
//
// Note that this regexp does not accept single upper case letter, see
// https://github.com/palantir/http-remoting-api/issues/110.
var errorNameRegexp = regexp.MustCompile("^(([A-Z][a-z0-9]+)+):(([A-Z][a-z0-9]+)+)$")

func verifyErrorNameString(s string) error {
	if !errorNameRegexp.MatchString(s) {
		return fmt.Errorf("remoting: error name does not match regexp `%s`", errorNameRegexp.String())
	}
	if strings.HasPrefix(s, "Default:") {
		switch ErrorName(s) {
		case ErrorNamePermissionDenied:
		case ErrorNameInvalidArgument:
		case ErrorNameNotFound:
		case ErrorNameConflict:
		case ErrorNameRequestEntityTooLarge:
		case ErrorNameFailedPrecondition:
		case ErrorNameInternal:
		case ErrorNameTimeout:
		default:
			return fmt.Errorf("remoting: custom error name cannot use default namespace")
		}
	}
	return nil
}

func (en ErrorName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(en))
}

func (en *ErrorName) UnmarshalJSON(data []byte) error {
	var enString string
	if err := json.Unmarshal(data, &enString); err != nil {
		return err
	}
	if err := verifyErrorNameString(enString); err != nil {
		return err
	}
	*en = ErrorName(enString)
	return nil
}
