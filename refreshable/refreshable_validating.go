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

	validatedRefreshable Refreshable
	lastValidateErr      *atomic.Value
}

// this is needed to be able to store the absence of an error in an atomic.Value
type errorWrapper struct {
	err error
}

func (v *ValidatingRefreshable) Current() interface{} {
	return v.validatedRefreshable.Current()
}

func (v *ValidatingRefreshable) Subscribe(consumer func(interface{})) (unsubscribe func()) {
	return v.validatedRefreshable.Subscribe(consumer)
}

func (v *ValidatingRefreshable) Map(mapFn func(interface{}) interface{}) Refreshable {
	return v.validatedRefreshable.Map(mapFn)
}

func (v *ValidatingRefreshable) LastValidateErr() error {
	return v.lastValidateErr.Load().(errorWrapper).err
}

// NewValidatingRefreshable returns a new Refreshable whose current value is the latest value that passes the provided
// validatingFn successfully. This returns an error if the current value of the passed in Refreshable does not pass the
// validatingFn or if the validatingFn or Refreshable are nil.
func NewValidatingRefreshable(origRefreshable Refreshable, validatingFn func(interface{}) error) (*ValidatingRefreshable, error) {
	if validatingFn == nil {
		return nil, errors.New("failed to create validating Refreshable because the validating function was nil")
	}

	if origRefreshable == nil {
		return nil, errors.New("failed to create validating Refreshable because the passed in Refreshable was nil")
	}

	currentVal := origRefreshable.Current()
	if err := validatingFn(currentVal); err != nil {
		return nil, err
	}

	validatedRefreshable := NewDefaultRefreshable(currentVal)

	var lastValidateErr atomic.Value
	lastValidateErr.Store(errorWrapper{})
	v := ValidatingRefreshable{
		validatedRefreshable: validatedRefreshable,
		lastValidateErr:      &lastValidateErr,
	}

	_ = origRefreshable.Subscribe(func(i interface{}) {
		if err := validatingFn(i); err != nil {
			v.lastValidateErr.Store(errorWrapper{err})
			return
		}

		if err := validatedRefreshable.Update(i); err != nil {
			v.lastValidateErr.Store(errorWrapper{err})
			return
		}

		v.lastValidateErr.Store(errorWrapper{})
	})
	return &v, nil
}
