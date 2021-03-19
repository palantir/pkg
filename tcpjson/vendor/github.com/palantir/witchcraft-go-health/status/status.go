// Copyright (c) 2018 Palantir Technologies. All rights reserved.
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

package status

import (
	"context"
	"net/http"

	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
)

var (
	healthStateStatusCodes = map[health.HealthState_Value]int{
		health.HealthState_HEALTHY:   http.StatusOK,
		health.HealthState_DEFERRING: 518,
		health.HealthState_SUSPENDED: 519,
		health.HealthState_REPAIRING: 520,
		health.HealthState_WARNING:   521,
		health.HealthState_ERROR:     522,
		health.HealthState_TERMINAL:  523,
	}
)

// Source provides status that should be sent as a response.
type Source interface {
	Status() (respStatus int, metadata interface{})
}

// HealthCheckSource provides the SLS health status that should be sent as a response.
// Refer to the SLS specification for more information.
type HealthCheckSource interface {
	HealthStatus(ctx context.Context) health.HealthStatus
}

type combinedHealthCheckSource struct {
	healthCheckSources []HealthCheckSource
}

func NewCombinedHealthCheckSource(healthCheckSources ...HealthCheckSource) HealthCheckSource {
	return &combinedHealthCheckSource{
		healthCheckSources: healthCheckSources,
	}
}

func (c *combinedHealthCheckSource) HealthStatus(ctx context.Context) health.HealthStatus {
	result := health.HealthStatus{
		Checks: map[health.CheckType]health.HealthCheckResult{},
	}
	for _, healthCheckSource := range c.healthCheckSources {
		for k, v := range healthCheckSource.HealthStatus(ctx).Checks {
			result.Checks[k] = v
		}
	}
	return result
}

// HealthStateStatusCode returns the http status code for the provided health.HealthState_Value or
// http.StatusInternalServerError if the health state value is not recognized.
func HealthStateStatusCode(state health.HealthState_Value) int {
	code, ok := healthStateStatusCodes[state]
	if !ok {
		code = http.StatusInternalServerError
	}
	return code
}

func HealthStatusCode(metadata health.HealthStatus) int {
	worst := http.StatusOK
	for _, result := range metadata.Checks {
		code := HealthStateStatusCode(result.State.Value())
		if worst < code {
			worst = code
		}
	}
	return worst
}
