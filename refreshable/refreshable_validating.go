// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"errors"
	"sync/atomic"
)

type ValidatingRefreshable struct {
	Refreshable
	lastValidateErr *atomic.Value
}

// this is needed to be able to store the absence of an error in an atomic.Value
type errorWrapper struct {
	err error
}

func (v *ValidatingRefreshable) LastValidateErr() error {
	return v.lastValidateErr.Load().(errorWrapper).err
}

// NewValidatingRefreshable returns a new Refreshable whose current value is the latest value that passes the provided
// validatingFn successfully. This returns an error if the current value of the passed in Refreshable does not pass the
// validatingFn or if the validatingFn or Refreshable are nil.
func NewValidatingRefreshable(origRefreshable Refreshable, validatingFn func(interface{}) error) (*ValidatingRefreshable, error) {
	mappingFn := func(i interface{}) (interface{}, error) {
		if err := validatingFn(i); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return newValidatingRefreshable(origRefreshable, mappingFn, false)
}

// NewMapValidatingRefreshable is similar to NewValidatingRefreshable but allows for the function to return a mapping/mutation
// of the input object in addition to returning an error. The returned ValidatingRefreshable will contain the mapped value.
// The mapped value must always be of the same type (but not necessarily that of the input type).
func NewMapValidatingRefreshable(origRefreshable Refreshable, mappingFn func(interface{}) (interface{}, error)) (*ValidatingRefreshable, error) {
	return newValidatingRefreshable(origRefreshable, mappingFn, true)
}

func newValidatingRefreshable(origRefreshable Refreshable, validatingFn func(interface{}) (interface{}, error), storeMappedVal bool) (*ValidatingRefreshable, error) {
	if validatingFn == nil {
		return nil, errors.New("failed to create validating Refreshable because the validating function was nil")
	}

	if origRefreshable == nil {
		return nil, errors.New("failed to create validating Refreshable because the passed in Refreshable was nil")
	}

	var validatedRefreshable *DefaultRefreshable
	currentVal := origRefreshable.Current()
	mappedVal, err := validatingFn(currentVal)
	if err != nil {
		return nil, err
	}
	if storeMappedVal {
		validatedRefreshable = NewDefaultRefreshable(mappedVal)
	} else {
		validatedRefreshable = NewDefaultRefreshable(currentVal)
	}

	var lastValidateErr atomic.Value
	lastValidateErr.Store(errorWrapper{})
	v := ValidatingRefreshable{
		Refreshable:     validatedRefreshable,
		lastValidateErr: &lastValidateErr,
	}

	updateValueFn := func(i interface{}) {
		mappedVal, err := validatingFn(i)
		if err != nil {
			v.lastValidateErr.Store(errorWrapper{err})
			return
		}
		if storeMappedVal {
			err = validatedRefreshable.Update(mappedVal)
		} else {
			err = validatedRefreshable.Update(i)
		}
		v.lastValidateErr.Store(errorWrapper{err: err})
	}

	origRefreshable.Subscribe(updateValueFn)

	// manually update value after performing subscription. This ensures that, if the current value changed between when
	// it was fetched earlier in the function and when the subscription was performed, it is properly captured.
	updateValueFn(origRefreshable.Current())

	return &v, nil
}
