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

package sources

import (
	werror "github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-health/conjure/witchcraft/api/health"
)

// UnhealthyHealthCheckResult returns an unhealthy health check result with type checkType and message message.
func UnhealthyHealthCheckResult(checkType health.CheckType, message string, params map[string]interface{}) health.HealthCheckResult {
	return health.HealthCheckResult{
		Type:    checkType,
		State:   health.New_HealthState(health.HealthState_ERROR),
		Message: &message,
		Params:  params,
	}
}

// RepairingHealthCheckResult returns an repairing health check result with type checkType and message message.
func RepairingHealthCheckResult(checkType health.CheckType, message string, params map[string]interface{}) health.HealthCheckResult {
	return health.HealthCheckResult{
		Type:    checkType,
		State:   health.New_HealthState(health.HealthState_REPAIRING),
		Message: &message,
		Params:  params,
	}
}

// HealthyHealthCheckResult returns healthy health check result with type checkType.
func HealthyHealthCheckResult(checkType health.CheckType) health.HealthCheckResult {
	return health.HealthCheckResult{
		Type:  checkType,
		State: health.New_HealthState(health.HealthState_HEALTHY),
	}
}

// SafeParamsFromError returns the safeParam map from the given error
func SafeParamsFromError(err error) map[string]interface{} {
	safeParams, _ := werror.ParamsFromError(err)
	return safeParams
}
