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
	"context"
	"sync"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources"
	"github.com/palantir/witchcraft-go-health/status"
)

// ErrorSubmitter allows components whose functionality dictates a portion of health status to only consume this interface.
type ErrorSubmitter interface {
	Submit(error)
}

// ErrorHealthCheckSource is a health check source with statuses determined by submitted errors.
type ErrorHealthCheckSource interface {
	ErrorSubmitter
	status.HealthCheckSource
}

type errorHealthCheckSource struct {
	errorMode            ErrorMode
	timeProvider         TimeProvider
	windowSize           time.Duration
	checkMessage         string
	lastErrorTime        time.Time
	lastError            error
	lastSuccessTime      time.Time
	sourceMutex          sync.RWMutex
	checkType            health.CheckType
	repairingGracePeriod time.Duration
	repairingDeadline    time.Time
	maxErrorAge          time.Duration
	healthState          health.HealthState_Value
}

// MustNewErrorHealthCheckSource creates a new ErrorHealthCheckSource which will panic if any error is encountered.
// Should only be used in instances where the inputs are statically defined and known to be valid.
func MustNewErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) ErrorHealthCheckSource {
	source, err := NewErrorHealthCheckSource(checkType, errorMode, options...)
	if err != nil {
		panic(err)
	}
	return source
}

// NewErrorHealthCheckSource creates a new ErrorHealthCheckSource.
func NewErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) (ErrorHealthCheckSource, error) {
	conf := defaultErrorSourceConfig(checkType)
	conf.apply(options...)

	switch errorMode {
	case UnhealthyIfAtLeastOneError,
		HealthyIfNotAllErrors,
		HealthyIfNoRecentErrors:
	default:
		return nil, werror.Error("unknown or unsupported error mode",
			werror.SafeParam("errorMode", errorMode))
	}

	if conf.windowSize <= 0 {
		return nil, werror.Error("windowSize must be positive",
			werror.SafeParam("windowSize", conf.windowSize.String()))
	}
	if conf.repairingGracePeriod < 0 {
		return nil, werror.Error("repairingGracePeriod must be non negative",
			werror.SafeParam("repairingGracePeriod", conf.repairingGracePeriod.String()))
	}

	source := &errorHealthCheckSource{
		errorMode:            errorMode,
		timeProvider:         conf.timeProvider,
		windowSize:           conf.windowSize,
		checkMessage:         conf.checkMessage,
		checkType:            conf.checkType,
		repairingGracePeriod: conf.repairingGracePeriod,
		repairingDeadline:    conf.timeProvider.Now(),
		maxErrorAge:          conf.maxErrorAge,
		healthState:          conf.healthState,
	}

	// If requireFirstFullWindow, extend the repairing deadline to one windowSize from now.
	if conf.requireFirstFullWindow {
		source.repairingDeadline = conf.timeProvider.Now().Add(conf.windowSize)
	}

	return source, nil
}

// Submit submits an error.
func (e *errorHealthCheckSource) Submit(err error) {
	e.sourceMutex.Lock()
	defer e.sourceMutex.Unlock()

	// If using anchored windows when last submit is greater than the window
	// it will re-anchor the next window with a new repairing deadline.
	if !e.hasSuccessInWindow() && !e.hasErrorInWindow() {
		newRepairingDeadline := e.timeProvider.Now().Add(e.repairingGracePeriod)
		if newRepairingDeadline.After(e.repairingDeadline) {
			e.repairingDeadline = newRepairingDeadline
		}
	}

	if err != nil {
		e.lastError = err
		e.lastErrorTime = e.timeProvider.Now()
	} else {
		e.lastSuccessTime = e.timeProvider.Now()
	}
}

// HealthStatus polls the items inside the window and creates the HealthStatus.
func (e *errorHealthCheckSource) HealthStatus(ctx context.Context) health.HealthStatus {
	e.sourceMutex.RLock()
	defer e.sourceMutex.RUnlock()

	var healthCheckResult health.HealthCheckResult
	switch e.errorMode {
	case HealthyIfNotAllErrors:
		if e.hasSuccessInWindow() || !e.hasErrorInWindow() {
			healthCheckResult = sources.HealthyHealthCheckResult(e.checkType)
		} else {
			healthCheckResult = e.getFailureResult()
		}
	case UnhealthyIfAtLeastOneError:
		if e.hasErrorInWindow() {
			healthCheckResult = e.getFailureResult()
		} else {
			healthCheckResult = sources.HealthyHealthCheckResult(e.checkType)
		}
	case HealthyIfNoRecentErrors:
		if e.lastErrorTime.After(e.lastSuccessTime) {
			healthCheckResult = e.getFailureResult()
		} else {
			healthCheckResult = sources.HealthyHealthCheckResult(e.checkType)
		}
	}

	return health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			e.checkType: healthCheckResult,
		},
	}
}

func (e *errorHealthCheckSource) getFailureResult() health.HealthCheckResult {
	params := map[string]interface{}{
		"error": e.lastError.Error(),
	}
	healthCheckResult := health.HealthCheckResult{
		Type:    e.checkType,
		State:   health.New_HealthState(e.healthState),
		Message: &e.checkMessage,
		Params:  params,
	}
	if e.lastErrorTime.Before(e.repairingDeadline) {
		healthCheckResult.State = health.New_HealthState(health.HealthState_REPAIRING)
	}
	if e.maxErrorAge > 0 && e.timeProvider.Now().Sub(e.lastErrorTime) > e.maxErrorAge {
		healthCheckResult.State = health.New_HealthState(health.HealthState_REPAIRING)
	}
	return healthCheckResult
}

func (e *errorHealthCheckSource) hasSuccessInWindow() bool {
	return !e.lastSuccessTime.IsZero() && e.timeProvider.Now().Sub(e.lastSuccessTime) <= e.windowSize
}

func (e *errorHealthCheckSource) hasErrorInWindow() bool {
	return !e.lastErrorTime.IsZero() && e.timeProvider.Now().Sub(e.lastErrorTime) <= e.windowSize
}
