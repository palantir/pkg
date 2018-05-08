// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Error interface {
	error

	Code() ErrorCode
	Name() ErrorName
	InstanceID() ErrorInstanceID
	Parameters() map[string]string
	WriteResponse(w http.ResponseWriter)
}

func NewError(code ErrorCode, name ErrorName, parameters map[string]string) Error {
	return internalError{
		code:       code,
		name:       name,
		instanceID: ErrorInstanceID{}, // TODO Generate UUID
		parameters: parameters,
	}
}

func NewCustomServerError(name ErrorName, parameters map[string]string) Error {
	return NewError(ErrorCodeCustomServer, name, parameters)
}

func NewCustomClientError(name ErrorName, parameters map[string]string) Error {
	return NewError(ErrorCodeCustomClient, name, parameters)
}

// internalError implements the Error interface. It can only be created with exported constructors,
// which guarantee correctness of the data.
type internalError struct {
	code       ErrorCode
	name       ErrorName
	instanceID ErrorInstanceID
	parameters map[string]string
}

func (e internalError) Error() string {
	// e.g. "NOT_FOUND MyApplication:MissingData (00010203-0405-0607-0809-0a0b0c0d0e0f)"
	return fmt.Sprintf("%s %s (%s)", e.code, e.name, e.instanceID)
}

func (e internalError) Code() ErrorCode {
	return e.code
}

func (e internalError) Name() ErrorName {
	return e.name
}

func (e internalError) InstanceID() ErrorInstanceID {
	return e.instanceID
}

func (e internalError) Parameters() map[string]string {
	return e.parameters
}

func (e internalError) WriteResponse(w http.ResponseWriter) {
	body, err := json.Marshal(serializableError{
		Code:       e.code,
		Name:       e.name,
		InstanceID: e.instanceID,
		Parameters: e.parameters,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.code.StatusCode())
	_, err = w.Write(body)
	if err != nil {
		// There is not much else we can do on failure.
		log.Printf("failed to write error: %v", err)
	}
}

func ErrorFromResponse(response *http.Response) (Error, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var unmarshalled serializableError
	if err := json.Unmarshal(body, &unmarshalled); err != nil {
		return nil, err
	}

	return internalError{
		code:       unmarshalled.Code,
		name:       unmarshalled.Name,
		instanceID: unmarshalled.InstanceID,
		parameters: unmarshalled.Parameters,
	}, nil
}

// serializableError is serializable version of the internalError with exported fields.
type serializableError struct {
	Code       ErrorCode         `json:"errorCode" yaml:"errorCode"`
	Name       ErrorName         `json:"errorName" yaml:"errorName"`
	InstanceID ErrorInstanceID   `json:"errorInstanceId" yaml:"errorInstanceId"`
	Parameters map[string]string `json:"parameters" yaml:"parameters"`
}
