// Copyright (c) 2020 Palantir Technologies. All rights reserved.
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

// MustNewUnhealthyIfAtLeastOneErrorSource returns the result of calling NewUnhealthyIfAtLeastOneErrorSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize))
}

// NewUnhealthyIfAtLeastOneErrorSource creates an unhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// NewUnhealthyIfAtLeastOneErrorSource creates an unhealthyIfAtLeastOneErrorSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewUnhealthyIfAtLeastOneErrorSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, UnhealthyIfAtLeastOneError,
		WithWindowSize(windowSize))
}

// MustNewHealthyIfNotAllErrorsSource returns the result of calling NewHealthyIfNotAllErrorsSource, but panics if it returns an error.
// Should only be used in instances where the inputs are statically defined and known to be valid.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRequireFullWindow())
}

// NewHealthyIfNotAllErrorsSource creates an healthyIfNotAllErrorsSource
// with a sliding window of size windowSize and uses the checkType.
// windowSize must be a positive value, otherwise returns error.
// Errors submitted in the first time window cause the health check to go to REPAIRING instead of ERROR.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRequireFullWindow())
}

// MustNewAnchoredHealthyIfNotAllErrorsSource returns the result of calling
// NewAnchoredHealthyIfNotAllErrorsSource but panics if that call returns an error
// Should only be used in instances where the inputs are statically defined and known to be valid.
// Care should be taken in considering health submission rate and window size when using anchored
// windows. Windows too close to service emission frequency may cause errors to not surface.
// DEPRECATED: please use MustNewErrorHealthCheckSource.
func MustNewAnchoredHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) ErrorHealthCheckSource {
	return MustNewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}

// NewAnchoredHealthyIfNotAllErrorsSource creates an healthyIfNotAllErrorsSource
// with supplied checkType, using sliding window of size windowSize, which will
// anchor (force the window to be at least the grace period) by defining a repairing deadline
// at the end of the initial window or one window size after the end of a gap.
// If all errors happen before the repairing deadline, the health check returns REPAIRING instead of ERROR.
// windowSize must be a positive value, otherwise returns error.
// Care should be taken in considering health submission rate and window size when using anchored
// windows. Windows too close to service emission frequency may cause errors to not surface.
// DEPRECATED: please use NewErrorHealthCheckSource.
func NewAnchoredHealthyIfNotAllErrorsSource(checkType health.CheckType, windowSize time.Duration) (ErrorHealthCheckSource, error) {
	return NewErrorHealthCheckSource(checkType, HealthyIfNotAllErrors,
		WithWindowSize(windowSize),
		WithRepairingGracePeriod(windowSize),
		WithRequireFullWindow())
}
