// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Error is an error representation intended for transport through
// RPC channels such as HTTP responses.
//
// Error is represented by its error code, an error name identifying the type of error and
// an optional set of named parameters detailing the error.
//
// Example usage:
//
//  func ServeHTTP(w ResponseWriter, r *Request) {
//    ...
//    err := foo()
//    if err != nil {
//      remoting.WriteErrorResponse(w, remoting.NewCustomServerError(
//        remoting.MustErrorName("MyApplication:SomethingWrongWithFoo"),
//        map[string]string{
//          "message": err.Error(),
//        },
//      ))
//    }
//  }
type Error interface {
	error
	// Code returns an enum identifying the type of error.
	Code() ErrorCode
	// Name returns an error name identifying the type of error.
	Name() ErrorName
	// InstanceID returns unique identifier of this particular error instance.
	InstanceID() ErrorInstanceID
	// Parameters returns a set of named parameters detailing the error, for example error message.
	Parameters() map[string]string
}

// NewError return new error with generated error instance identifier and provided code, name and parameters.
func NewError(code ErrorCode, name ErrorName, parameters map[string]string) Error {
	return internalError{
		code:       code,
		name:       name,
		instanceID: NewErrorInstanceID(),
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

func WriteErrorResponse(w http.ResponseWriter, e Error) {
	body, err := json.Marshal(serializableError{
		Code:       e.Code(),
		Name:       e.Name(),
		InstanceID: e.InstanceID(),
		Parameters: e.Parameters(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code().StatusCode())
	_, _ = w.Write(body) // There is nothing we can do on write failure.
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
	Code       ErrorCode         `json:"errorCode"`
	Name       ErrorName         `json:"errorName"`
	InstanceID ErrorInstanceID   `json:"errorInstanceId"`
	Parameters map[string]string `json:"parameters"`
}
