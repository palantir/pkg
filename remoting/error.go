// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"
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
//  var ErrorLikeAlreadyGiven = remoting.MustErrorType(
//    remoting.ErrorCodeConflict,
//    "Facebook:LikeAlreadyGiven",
//  )
//  ...
//  func HandleLikePost(w ResponseWriter, r *Request) {
//    ...
//    post := getPost(postID)
//    if post == nil {
//      remoteErr := remoting.NewNotFoundError(
//        map[string]string{
//          "postId": postID,
//        },
//      )
//      remoting.WriteErrorResponse(w, remoteErr)
//      return
//    }
//    if post.hasLike(userID) {
//      remoteErr := remoting.NewError(
//        ErrorLikeAlreadyGiven,
//        map[string]string{
//          "postId": postID,
//          "userId": userID,
//        },
//      )
//      remoting.WriteErrorResponse(w, remoteErr)
//      return
//    }
//    ...
//  }
type Error interface {
	error
	// Code returns an enum describing error category.
	Code() ErrorCode
	// Name returns an error name identifying error type.
	Name() string
	// InstanceID returns unique identifier of this particular error instance.
	InstanceID() ErrorInstanceID
	// Parameters returns a set of named parameters detailing this particular error instance,
	// for example error message.
	Parameters() map[string]string
}

// NewError returns new instance of an error of the specified type with provided parameters.
func NewError(errorType ErrorType, parameters map[string]string) Error {
	instanceID := NewErrorInstanceID()
	return internalError{
		errorType:  errorType,
		instanceID: instanceID,
		parameters: parameters,
	}
}

func NewPermissionDeniedError(parameters map[string]string) Error {
	return NewError(DefaultPermissionDenied, parameters)
}

func NewInvalidArgumentError(parameters map[string]string) Error {
	return NewError(DefaultInvalidArgument, parameters)
}

func NewNotFoundError(parameters map[string]string) Error {
	return NewError(DefaultNotFound, parameters)
}

func NewConflictError(parameters map[string]string) Error {
	return NewError(DefaultConflict, parameters)
}

func NewRequestEntityTooLargeError(parameters map[string]string) Error {
	return NewError(DefaultRequestEntityTooLarge, parameters)
}

func NewFailedPreconditionError(parameters map[string]string) Error {
	return NewError(DefaultFailedPrecondition, parameters)
}

func NewInternalError(parameters map[string]string) Error {
	return NewError(DefaultInternal, parameters)
}

func NewTimeoutError(parameters map[string]string) Error {
	return NewError(DefaultTimeout, parameters)
}

// internalError implements the Error interface. It can only be created with exported constructors,
// which guarantee correctness of the data.
type internalError struct {
	errorType  ErrorType
	instanceID ErrorInstanceID
	parameters map[string]string
}

func (e internalError) Error() string {
	return e.String()
}

// String representation of an error,
// for example "NOT_FOUND MyApplication:MissingData (00010203-0405-0607-0809-0a0b0c0d0e0f)".
func (e internalError) String() string {
	return fmt.Sprintf("%s (%s)", e.errorType, e.instanceID)
}

func (e internalError) Code() ErrorCode {
	return e.errorType.code
}

func (e internalError) Name() string {
	return e.errorType.name
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
	var unmarshalled serializableError
	if err := json.NewDecoder(response.Body).Decode(&unmarshalled); err != nil {
		return nil, err
	}

	errorType, err := NewErrorType(
		unmarshalled.Code,
		unmarshalled.Name,
	)
	if err != nil {
		return nil, err
	}

	return internalError{
		errorType:  errorType,
		instanceID: unmarshalled.InstanceID,
		parameters: unmarshalled.Parameters,
	}, nil
}

// serializableError is serializable version of the internalError with exported fields.
type serializableError struct {
	Code       ErrorCode         `json:"errorCode"`
	Name       string            `json:"errorName"`
	InstanceID ErrorInstanceID   `json:"errorInstanceId"`
	Parameters map[string]string `json:"parameters"`
}
