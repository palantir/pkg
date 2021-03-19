// Copyright (c) 2019 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package window

import (
	"time"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
)

// MustNewMultiKeyUnhealthyIfAtLeastOneErrorSource returns the result of calling NewMultiKeyUnhealthyIfAtLeastOneErrorSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewKeyedErrorHealthCheckSource.
func MustNewMultiKeyUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) KeyedErrorHealthCheckSource {
	return MustNewKeyedErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError))
}

// NewMultiKeyUnhealthyIfAtLeastOneErrorSource creates an multiKeyUnhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType and a message in case of errors.
// windowSize must be a positive value, otherwise returns error.
// DEPRECATED: please use NewKeyedErrorHealthCheckSource.
func NewMultiKeyUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) (KeyedErrorHealthCheckSource, error) {
	return NewKeyedErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError))
}

// MustNewMultiKeyHealthyIfNoRecentErrorsSource returns the result of calling NewMultiKeyHealthyIfNoRecentErrorsSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewKeyedErrorHealthCheckSource.
func MustNewMultiKeyHealthyIfNoRecentErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) KeyedErrorHealthCheckSource {
	return MustNewKeyedErrorHealthCheckSource(checkType, HealthyIfNoRecentErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError))
}

// NewMultiKeyHealthyIfNoRecentErrorsSource creates an multiKeyUnhealthyIfNoRecentErrorsSource
// with a sliding window of size windowSize and uses the checkType and a message in case of errors.
// windowSize must be a positive value, otherwise returns error.
// Once a non-nil error has been submitted, this will be unhealthy until a nil error is submitted or `windowSize` time
// has passed without a non-nil error. Submitting a non-nil error resets the timer and stays unhealthy
// DEPRECATED: please use NewKeyedErrorHealthCheckSource.
func NewMultiKeyHealthyIfNoRecentErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) (KeyedErrorHealthCheckSource, error) {
	return NewKeyedErrorHealthCheckSource(checkType, HealthyIfNoRecentErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError))
}

// MustNewMultiKeyHealthyIfNotAllErrorsSource returns the result of calling NewMultiKeyHealthyIfNotAllErrorsSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewKeyedErrorHealthCheckSource.
func MustNewMultiKeyHealthyIfNotAllErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) KeyedErrorHealthCheckSource {
	return MustNewKeyedErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError),
		WithRequireFullWindow())
}

// NewMultiKeyHealthyIfNotAllErrorsSource creates an multiKeyUnhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType and a message in case of errors.
// windowSize must be a positive value, otherwise returns error.
// Errors submitted in the first time window cause the health check to go to REPAIRING instead of ERROR.
// DEPRECATED: please use NewKeyedErrorHealthCheckSource.
func NewMultiKeyHealthyIfNotAllErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) (KeyedErrorHealthCheckSource, error) {
	return NewKeyedErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError),
		WithRequireFullWindow())
}

// MustNewAnchoredMultiKeyHealthyIfNotAllErrorsSource returns the result of calling NewAnchoredMultiKeyHealthyIfNotAllErrorsSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewKeyedErrorHealthCheckSource.
func MustNewAnchoredMultiKeyHealthyIfNotAllErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) KeyedErrorHealthCheckSource {
	return MustNewKeyedErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}

// NewAnchoredMultiKeyHealthyIfNotAllErrorsSource creates an multiKeyUnhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType and a message in case of errors.
// Each key has a repairing deadline that is one window size after a moment of idleness.
// If all errors happen before their respective repairing deadline, the health check returns REPAIRING instead of ERROR.
// windowSize must be a positive value, otherwise returns error.
// Errors submitted in the first time window cause the health check to go to REPAIRING instead of ERROR.
// DEPRECATED: please use NewKeyedErrorHealthCheckSource.
func NewAnchoredMultiKeyHealthyIfNotAllErrorsSource(checkType health.CheckType, messageInCaseOfError string, windowSize time.Duration) (KeyedErrorHealthCheckSource, error) {
	return NewKeyedErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithCheckMessage(messageInCaseOfError),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}
