// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"regexp"
)

const (
	ErrorNameDefaultPermissionDenied      ErrorName = "Default:PermissionDenied"
	ErrorNameDefaultInvalidArgument       ErrorName = "Default:InvalidArgument"
	ErrorNameDefaultNotFound              ErrorName = "Default:NotFound"
	ErrorNameDefaultConflict              ErrorName = "Default:Conflict"
	ErrorNameDefaultRequestEntityTooLarge ErrorName = "Default:RequestEntityTooLarge"
	ErrorNameDefaultFailedPrecondition    ErrorName = "Default:FailedPrecondition"
	ErrorNameDefaultInternal              ErrorName = "Default:Internal"
	ErrorNameDefaultTimeout               ErrorName = "Default:Timeout"
)

// ErrorName identifies the type of error.
//
// Error name should be a compile-time constant and is considered part of the API
// of a service that produces error with such name.
//
// It should be in the "PascalCase:PascalCase" format, for example "Default:PermissionDenied",
// "Facebook:LikeAlreadyGiven" or "MyApplication:ErrorSpecificToMyBusinessDomain".
//
// Use MustErrorName() to ensure valid ErrorName is created.
type ErrorName string

func MustErrorName(errorName string) ErrorName {
	en, err := parseErrorName(errorName)
	if err != nil {
		panic(err)
	}
	return en
}

var errorNameRegexp = regexp.MustCompile(`^(([A-Z][a-z0-9]+)+):(([A-Z][a-z0-9]+)+)$`)

func parseErrorName(errorName string) (ErrorName, error) {
	if !errorNameRegexp.MatchString(errorName) {
		return "", fmt.Errorf("remoting: error name does not match regexp `%s`",
			errorNameRegexp.String())
	}
	return ErrorName(errorName), nil
}

func (en ErrorName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(en))
}

func (en *ErrorName) UnmarshalJSON(data []byte) error {
	var enString string

	err := json.Unmarshal(data, &enString)
	if err != nil {
		return err
	}

	*en, err = parseErrorName(enString)
	return err
}
