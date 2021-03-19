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
	"context"
	"sync"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-health/sources"
	"github.com/palantir/witchcraft-go-health/status"
)

// KeyedErrorSubmitter allows components whose functionality dictates a portion of health status to only consume this interface.
type KeyedErrorSubmitter interface {
	Submit(key string, err error)
}

// KeyedErrorHealthCheckSource is a health check source with statuses determined by submitted key error pairs.
type KeyedErrorHealthCheckSource interface {
	KeyedErrorSubmitter
	status.HealthCheckSource
}

type keyedErrorHealthCheckSource struct {
	errorMode               ErrorMode
	windowSize              time.Duration
	errorStore              TimedKeyStore
	successStore            TimedKeyStore
	gapEndTimeStore         TimedKeyStore
	repairingGracePeriod    time.Duration
	maxErrorAge             time.Duration
	globalRepairingDeadline time.Time
	sourceMutex             sync.Mutex
	checkType               health.CheckType
	checkMessage            string
	timeProvider            TimeProvider
}

// MustNewKeyedErrorHealthCheckSource creates a new KeyedErrorHealthCheckSource which will panic if any error is encountered.
// Should only be used in instances where the inputs are statically defined and known to be valid.
func MustNewKeyedErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) KeyedErrorHealthCheckSource {
	source, err := NewKeyedErrorHealthCheckSource(checkType, errorMode, options...)
	if err != nil {
		panic(err)
	}
	return source
}

// NewKeyedErrorHealthCheckSource creates a new KeyedErrorHealthCheckSource.
func NewKeyedErrorHealthCheckSource(checkType health.CheckType, errorMode ErrorMode, options ...ErrorOption) (KeyedErrorHealthCheckSource, error) {
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

	source := &keyedErrorHealthCheckSource{
		errorMode:               errorMode,
		windowSize:              conf.windowSize,
		errorStore:              NewTimedKeyStore(conf.timeProvider),
		successStore:            NewTimedKeyStore(conf.timeProvider),
		gapEndTimeStore:         NewTimedKeyStore(conf.timeProvider),
		repairingGracePeriod:    conf.repairingGracePeriod,
		maxErrorAge:             conf.maxErrorAge,
		globalRepairingDeadline: conf.timeProvider.Now(),
		checkType:               conf.checkType,
		checkMessage:            conf.checkMessage,
		timeProvider:            conf.timeProvider,
	}

	if conf.requireFirstFullWindow {
		source.globalRepairingDeadline = conf.timeProvider.Now().Add(conf.windowSize)
	}

	return source, nil
}

// Submit submits an item as a key error pair.
func (k *keyedErrorHealthCheckSource) Submit(key string, err error) {
	k.sourceMutex.Lock()
	defer k.sourceMutex.Unlock()

	k.errorStore.PruneKeysAboveAge(k.windowSize)
	k.successStore.PruneKeysAboveAge(k.windowSize)
	k.gapEndTimeStore.PruneKeysAboveAge(k.repairingGracePeriod + k.windowSize)

	_, hasError := k.errorStore.Get(key)
	_, hasSuccess := k.successStore.Get(key)
	if !hasError && !hasSuccess {
		k.gapEndTimeStore.Put(key, nil)
	}

	if err == nil {
		k.successStore.Put(key, nil)
	} else {
		k.errorStore.Put(key, err)
	}
}

// HealthStatus polls the items inside the window and creates the HealthStatus.
func (k *keyedErrorHealthCheckSource) HealthStatus(ctx context.Context) health.HealthStatus {
	k.sourceMutex.Lock()
	defer k.sourceMutex.Unlock()

	var healthCheckResult health.HealthCheckResult

	k.errorStore.PruneKeysAboveAge(k.windowSize)
	k.successStore.PruneKeysAboveAge(k.windowSize)
	k.gapEndTimeStore.PruneKeysAboveAge(k.repairingGracePeriod + k.windowSize)

	params := make(map[string]interface{})
	shouldError := false
	for _, errItem := range k.errorStore.List() {
		switch k.errorMode {
		case HealthyIfNotAllErrors:
			if _, hasSuccess := k.successStore.Get(errItem.Key); hasSuccess {
				continue
			}
		case UnhealthyIfAtLeastOneError:
		case HealthyIfNoRecentErrors:
			if successItem, hasSuccess := k.successStore.Get(errItem.Key); hasSuccess {
				if !errItem.Time.After(successItem.Time) {
					continue
				}
			}
		}

		shouldError = shouldError || k.shouldError(errItem)
		params[errItem.Key] = errItem.Payload.(error).Error()
	}

	if len(params) > 0 {
		healthCheckResult = health.HealthCheckResult{
			Type:    k.checkType,
			State:   health.New_HealthState(health.HealthState_REPAIRING),
			Message: &k.checkMessage,
			Params:  params,
		}
		if shouldError {
			healthCheckResult.State = health.New_HealthState(health.HealthState_ERROR)
		}
	} else {
		healthCheckResult = sources.HealthyHealthCheckResult(k.checkType)
	}

	return health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{
			k.checkType: healthCheckResult,
		},
	}
}

func (k *keyedErrorHealthCheckSource) shouldError(item TimedKey) bool {
	if k.maxErrorAge > 0 && k.timeProvider.Now().Sub(item.Time) > k.maxErrorAge {
		return false
	}
	repairingDeadline := k.globalRepairingDeadline
	if gapEndTime, hasRepairingDeadline := k.gapEndTimeStore.Get(item.Key); hasRepairingDeadline {
		newRepairingDeadline := gapEndTime.Time.Add(k.repairingGracePeriod)
		if newRepairingDeadline.After(repairingDeadline) {
			repairingDeadline = newRepairingDeadline
		}
	}
	return item.Time.After(repairingDeadline) || item.Time.Equal(repairingDeadline)
}
