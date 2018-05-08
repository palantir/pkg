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

// ErrorName is a human readable name of the error type.
//
// It should be in the "PascalCase:PascalCase" format, for example "Default:PermissionDenied" or
// "MyApplication:DatasetNotFound"
//
// TODO: Consider using struct{Namespace, Name string} instead of string?
type ErrorName string

func MustErrorName(errorName string) ErrorName {
	en, err := parseErrorName(errorName)
	if err != nil {
		panic(err)
	}
	return en
}

const errorNameFormat = `(([A-Z][a-z0-9]+)+):(([A-Z][a-z0-9]+)+)`

var errorNameRegexp = regexp.MustCompile(errorNameFormat)

func parseErrorName(errorName string) (ErrorName, error) {
	if !errorNameRegexp.MatchString(errorName) {
		return "", fmt.Errorf("remoting: error name `%s` is not in the valid format `%s`",
			errorName, errorNameFormat)
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
