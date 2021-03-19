// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson

import (
	"encoding/json"
	"os"

	werror "github.com/palantir/witchcraft-go-error"
)

type LogEnvelopeV1 struct {
	LogEnvelopeMetadata
	Payload json.RawMessage `json:"payload"`
}

type LogEnvelopeMetadata struct {
	Type           string `json:"type"`
	Deployment     string `json:"deployment"`
	Environment    string `json:"environment"`
	EnvironmentID  string `json:"environmentId"`
	Host           string `json:"host"`
	NodeID         string `json:"nodeId"`
	Service        string `json:"service"`
	ServiceID      string `json:"serviceId"`
	Stack          string `json:"stack"`
	StackID        string `json:"stackId"`
	Product        string `json:"product"`
	ProductVersion string `json:"productVersion"`
}

const (
	envVarDeployment     = "LOG_ENVELOPE_DEPLOYMENT_NAME"
	envVarEnvironment    = "LOG_ENVELOPE_ENVIRONMENT_NAME"
	envVarEnvironmentID  = "LOG_ENVELOPE_ENVIRONMENT_ID"
	envVarHost           = "LOG_ENVELOPE_HOST"
	envVarNodeID         = "LOG_ENVELOPE_NODE_ID"
	envVarProduct        = "LOG_ENVELOPE_PRODUCT_NAME"
	envVarProductVersion = "LOG_ENVELOPE_PRODUCT_VERSION"
	envVarService        = "LOG_ENVELOPE_SERVICE_NAME"
	envVarServiceID      = "LOG_ENVELOPE_SERVICE_ID"
	envVarStack          = "LOG_ENVELOPE_STACK_NAME"
	envVarStackID        = "LOG_ENVELOPE_STACK_ID"
)

// GetEnvelopeMetadata retrieves all log envelope environment variables
// and returns the fully populated LogEnvelopeMetadata and a nil error.
// If any expected environment variables are not found or empty, then an empty LogEnvelopeMetadata
// will be returned along with an error that contains the missing environment variables.
func GetEnvelopeMetadata() (LogEnvelopeMetadata, error) {
	metadata := LogEnvelopeMetadata{
		Deployment:     os.Getenv(envVarDeployment),
		Environment:    os.Getenv(envVarEnvironment),
		EnvironmentID:  os.Getenv(envVarEnvironmentID),
		Host:           os.Getenv(envVarHost),
		NodeID:         os.Getenv(envVarNodeID),
		Product:        os.Getenv(envVarProduct),
		ProductVersion: os.Getenv(envVarProductVersion),
		Service:        os.Getenv(envVarService),
		ServiceID:      os.Getenv(envVarServiceID),
		Stack:          os.Getenv(envVarStack),
		StackID:        os.Getenv(envVarStackID),
	}
	var missingEnvVars []string
	if metadata.Deployment == "" {
		missingEnvVars = append(missingEnvVars, envVarDeployment)
	}
	if metadata.Environment == "" {
		missingEnvVars = append(missingEnvVars, envVarEnvironment)
	}
	if metadata.EnvironmentID == "" {
		missingEnvVars = append(missingEnvVars, envVarEnvironmentID)
	}
	if metadata.Host == "" {
		missingEnvVars = append(missingEnvVars, envVarHost)
	}
	if metadata.NodeID == "" {
		missingEnvVars = append(missingEnvVars, envVarNodeID)
	}
	if metadata.Product == "" {
		missingEnvVars = append(missingEnvVars, envVarProduct)
	}
	if metadata.ProductVersion == "" {
		missingEnvVars = append(missingEnvVars, envVarProductVersion)
	}
	if metadata.Service == "" {
		missingEnvVars = append(missingEnvVars, envVarService)
	}
	if metadata.ServiceID == "" {
		missingEnvVars = append(missingEnvVars, envVarServiceID)
	}
	if metadata.Stack == "" {
		missingEnvVars = append(missingEnvVars, envVarStack)
	}
	if metadata.StackID == "" {
		missingEnvVars = append(missingEnvVars, envVarStackID)
	}
	if len(missingEnvVars) > 0 {
		return LogEnvelopeMetadata{}, werror.Error("all log envelope environment variables are not set",
			werror.SafeParam("missingEnvVars", missingEnvVars))
	}
	return metadata, nil
}
