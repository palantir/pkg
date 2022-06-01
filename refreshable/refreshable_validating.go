// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"errors"
	"sync/atomic"
)

type ValidatingRefreshable[T any] struct {
	Refreshable[T]
	lastValidateErr *atomic.Value
}

// this is needed to be able to store the absence of an error in an atomic.Value
type errorWrapper struct {
	err error
}

func (v *ValidatingRefreshable[T]) LastValidateErr() error {
	return v.lastValidateErr.Load().(errorWrapper).err
}

// NewValidatingRefreshable returns a new Refreshable whose current value is the latest value that passes the provided
// validatingFn successfully. This returns an error if the current value of the passed in Refreshable does not pass the
// validatingFn or if the validatingFn or Refreshable are nil.
func NewValidatingRefreshable[T any](origRefreshable Refreshable[T], validatingFn func(T) error) (*ValidatingRefreshable[T], error) {
	mappingFn := func(i T) (T, error) {
		if err := validatingFn(i); err != nil {
			var zero T
			return zero, err
		}
		return i, nil
	}
	return newValidatingRefreshable(origRefreshable, mappingFn)
}

// NewMapValidatingRefreshable is similar to NewValidatingRefreshable but allows for the function to return a mapping/mutation
// of the input object in addition to returning an error. The returned ValidatingRefreshable will contain the mapped value.
// The mapped value must always be of the same type (but not necessarily that of the input type).
func NewMapValidatingRefreshable[T any, M any](origRefreshable Refreshable[T], mappingFn func(T) (M, error)) (*ValidatingRefreshable[M], error) {
	return newValidatingRefreshable(origRefreshable, mappingFn)
}

func newValidatingRefreshable[T any, M any](origRefreshable Refreshable[T], mappingFn func(T) (M, error)) (*ValidatingRefreshable[M], error) {
	if mappingFn == nil {
		return nil, errors.New("failed to create validating Refreshable because the validating function was nil")
	}
	if origRefreshable == nil {
		return nil, errors.New("failed to create validating Refreshable because the passed in Refreshable was nil")
	}

	var validatedRefreshable *DefaultRefreshable[M]
	currentVal := origRefreshable.Current()
	mappedVal, err := mappingFn(currentVal)
	if err != nil {
		return nil, err
	}
	validatedRefreshable = New(mappedVal)

	var lastValidateErr atomic.Value
	lastValidateErr.Store(errorWrapper{})
	v := ValidatingRefreshable[M]{
		Refreshable:     validatedRefreshable,
		lastValidateErr: &lastValidateErr,
	}

	updateValueFn := func(i T) {
		mappedVal, err := mappingFn(i)
		v.lastValidateErr.Store(errorWrapper{err})
		if err == nil {
			validatedRefreshable.Update(mappedVal)
		}
	}

	origRefreshable.Subscribe(updateValueFn)

	// manually update value after performing subscription. This ensures that, if the current value changed between when
	// it was fetched earlier in the function and when the subscription was performed, it is properly captured.
	updateValueFn(origRefreshable.Current())

	return &v, nil
}
