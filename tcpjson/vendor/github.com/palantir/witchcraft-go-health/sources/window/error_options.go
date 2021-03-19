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

// ErrorMode is an enum for the available behaviors for error based window health check sources.
type ErrorMode string

const (
	// UnhealthyIfAtLeastOneError makes the error submitter based health
	// check source return unhealthy if there are any errors in the window.
	// Returns healthy if there no submissions in the window.
	UnhealthyIfAtLeastOneError ErrorMode = "UnhealthyIfAtLeastOneError"
	// HealthyIfNotAllErrors makes the error submitter based health
	// check source return unhealthy if there are only errors in the window.
	// Returns healthy if there no submissions in the window.
	HealthyIfNotAllErrors ErrorMode = "HealthyIfNotAllErrors"
	// HealthyIfNoRecentErrors makes the error submitter based health
	// check source return unhealthy if the most recent submission is an error.
	// Returns healthy if there no submissions in the window.
	HealthyIfNoRecentErrors ErrorMode = "HealthyIfNoRecentErrors"
)

// ErrorOption is an option for an error submitter based window health check source.
type ErrorOption func(conf *errorSourceConfig)

const (
	defaultWindowSize                         = 10 * time.Minute
	defaultRepairingGracePeriod time.Duration = 0
)

type errorSourceConfig struct {
	checkType              health.CheckType
	windowSize             time.Duration
	checkMessage           string
	repairingGracePeriod   time.Duration
	requireFirstFullWindow bool
	maxErrorAge            time.Duration
	timeProvider           TimeProvider
	healthState            health.HealthState_Value
}

func defaultErrorSourceConfig(checkType health.CheckType) errorSourceConfig {
	return errorSourceConfig{
		checkType:              checkType,
		windowSize:             defaultWindowSize,
		checkMessage:           "",
		repairingGracePeriod:   defaultRepairingGracePeriod,
		requireFirstFullWindow: false,
		maxErrorAge:            0,
		timeProvider:           NewOrdinaryTimeProvider(),
		healthState:            health.HealthState_ERROR,
	}
}

func (e *errorSourceConfig) apply(options ...ErrorOption) {
	for _, option := range options {
		option(e)
	}
}

// WithWindowSize modifies the window size.
// If not set, the default window size of 10 min is used.
func WithWindowSize(windowSize time.Duration) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.windowSize = windowSize
	}
}

// WithCheckMessage adds a message to the health check source.
// If not set, an empty message is used.
func WithCheckMessage(checkMessage string) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.checkMessage = checkMessage
	}
}

// WithRepairingGracePeriod adds a grace period for when the health check is coming from a long period with no events.
// When an error is submitted, if there have been no errors in the past window,
// a repairing deadline is set repairingGracePeriod into the future.
// All errors before that deadline are "downgraded" to "repairing errors".
// If a window only contains repairing errors, error health checks are converted to repairing health checks.
// This always happens when the health check is first set up.
// If not set, no grace period is used.
func WithRepairingGracePeriod(repairingGracePeriod time.Duration) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.repairingGracePeriod = repairingGracePeriod
	}
}

// WithRequireFullWindow adds a grace period for when the health check has just been initialized.
// A repairing deadline is set one window into the future.
// All errors before that deadline are "downgraded" to "repairing errors".
// If a window only contains repairing errors, error health checks are converted to repairing health checks.
// If not set, early errors might cause the health check to become unhealthy.
func WithRequireFullWindow() ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.requireFirstFullWindow = true
	}
}

// WithMaximumErrorAge sets the maximum age an error can have to make the health check become unhealthy.
// If the latest error happened more than the maximum age ago, error checks are converted to repairing checks.
// A maximum error age of zero is treated as infinity (no maximum error age).
// If not set, a non-nil error that has just been submitted can cause the health check to become unhealthy.
func WithMaximumErrorAge(maxErrorAge time.Duration) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.maxErrorAge = maxErrorAge
	}
}

// WithTimeProvider overrides the function used for fetching the current time.
// It is useful for writing time sensitive tests without having to actually wait.
// If not set, the default provider that returns time.Now() is used.
func WithTimeProvider(timeProvider TimeProvider) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.timeProvider = timeProvider
	}
}

// WithFailingHealthStateValue overrides the default health state value used when computing the health status in failure cases.
// All options that reduce errors to a REPAIRING health state will continue to work as expected, as this option
// strictly configures the base health state value before computing the full health status.
// If not set, the default health state value will be an ERROR.
func WithFailingHealthStateValue(healthState health.HealthState_Value) ErrorOption {
	return func(conf *errorSourceConfig) {
		conf.healthState = healthState
	}
}
