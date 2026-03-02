/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package instancepool

import (
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"

	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestMakeOtelInitContainer(t *testing.T) {
	funcSpec := &types.FunctionSpecification{
		ExtendedMetaData: commontypes.ExtendedMetaData{
			UserOtelConfig: commontypes.UserOtelConfig{
				InitContainer: commontypes.OtelInitContainer{
					Image:    "otel-image",
					Command:  []string{"otel-command"},
					ShardDir: "/shard/dir",
					ResourceRequest: commontypes.ResourceRequire{
						Cpu:    500,
						Memory: 1024,
					},
					ResourceLimit: commontypes.ResourceRequire{
						Cpu:    1000,
						Memory: 2048,
					},
				},
			},
		},
	}

	result := makeOtelInitContainer(funcSpec)

	if result.Name != "otel-init" {
		t.Errorf("Expected container name 'otel-init', got '%s'", result.Name)
	}
	if result.Image != "otel-image" {
		t.Errorf("Expected image 'otel-image', got '%s'", result.Image)
	}
	if len(result.Command) != 1 || result.Command[0] != "otel-command" {
		t.Errorf("Expected command ['otel-command'], got %v", result.Command)
	}
	if len(result.VolumeMounts) != 1 || result.VolumeMounts[0].Name != "otel-shard-dir" {
		t.Errorf("Expected volume mount with name 'otel-shard-dir', got %v", result.VolumeMounts)
	}
	requestCpu, _ := result.ResourceRequirements.Requests[v1.ResourceCPU]
	if requestCpu.String() != "500m" {
		t.Errorf("Expected CPU request '500m', got '%s'", requestCpu.String())
	}
	requestMemory, _ := result.ResourceRequirements.Requests[v1.ResourceMemory]
	if requestMemory.String() != "1Gi" {
		t.Errorf("Expected memory request '1Gi', got '%s'", requestMemory.String())
	}
	limitCpu, _ := result.ResourceRequirements.Limits[v1.ResourceCPU]
	if limitCpu.String() != "1" {
		t.Errorf("Expected CPU limit '1', got '%s'", limitCpu.String())
	}
	limitMemory, _ := result.ResourceRequirements.Limits[v1.ResourceMemory]
	if limitMemory.String() != "2Gi" {
		t.Errorf("Expected memory limit '2Gi', got '%s'", limitMemory.String())
	}
}

func TestAddOtelSystemEnv(t *testing.T) {
	envs := []v1.EnvVar{
		{Name: "EXISTING_ENV_VAR", Value: "value"},
	}

	result := addOtelSystemEnv(envs)

	if len(result) != 5 {
		t.Errorf("Expected 5 env vars, got %d", len(result))
	}
	if result[0].Name != "EXISTING_ENV_VAR" {
		t.Errorf("Expected first env var to be 'EXISTING_ENV_VAR', got '%s'", result[0].Name)
	}
	if result[1].Name != "OTEL_NODE_IP" {
		t.Errorf("Expected second env var to be 'OTEL_NODE_IP', got '%s'", result[1].Name)
	}
	if result[2].Name != "OTEL_POD_IP" {
		t.Errorf("Expected third env var to be 'OTEL_POD_IP', got '%s'", result[2].Name)
	}
	if result[3].Name != "OTEL_RESOURCE_ATTRIBUTES_POD_NAME" {
		t.Errorf("Expected fourth env var to be 'OTEL_RESOURCE_ATTRIBUTES_POD_NAME', got '%s'", result[3].Name)
	}
	if result[4].Name != "OTEL_RESOURCE_ATTRIBUTES_NODE_NAME" {
		t.Errorf("Expected fifth env var to be 'OTEL_RESOURCE_ATTRIBUTES_NODE_NAME', got '%s'", result[4].Name)
	}
}

func TestGetFunctionAgentInitOtelEnv(t *testing.T) {
	tests := []struct {
		name          string
		funcSpec      *types.FunctionSpecification
		expectedEnv   []v1.EnvVar
		expectedError string
	}{
		{
			name: "Valid OTEL_EXPORTER_OTLP_ENDPOINT",
			funcSpec: &types.FunctionSpecification{
				ExtendedMetaData: commontypes.ExtendedMetaData{
					UserOtelConfig: commontypes.UserOtelConfig{
						OtelEnv: map[string]string{
							otelEndPointEnvKey: "http://example.com:4318",
						},
					},
				},
			},
			expectedEnv: []v1.EnvVar{
				{
					Name:  otelWhiteList,
					Value: "example.com,None,4318/TCP,;",
				},
			},
			expectedError: "",
		},
		{
			name: "Missing OTEL_EXPORTER_OTLP_ENDPOINT",
			funcSpec: &types.FunctionSpecification{
				ExtendedMetaData: commontypes.ExtendedMetaData{
					UserOtelConfig: commontypes.UserOtelConfig{
						OtelEnv: map[string]string{},
					},
				},
			},
			expectedEnv:   nil,
			expectedError: "",
		},
		{
			name: "Invalid OTEL_EXPORTER_OTLP_ENDPOINT",
			funcSpec: &types.FunctionSpecification{
				ExtendedMetaData: commontypes.ExtendedMetaData{
					UserOtelConfig: commontypes.UserOtelConfig{
						OtelEnv: map[string]string{
							otelEndPointEnvKey: "invalid-url",
						},
					},
				},
			},
			expectedEnv:   nil,
			expectedError: "failed to parse OTEL_EXPORTER_OTLP_ENDPOINT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := getFunctionAgentInitOtelEnv(tt.funcSpec)
			if len(env) != len(tt.expectedEnv) {
				t.Errorf("Expected %d env vars, got %d", len(tt.expectedEnv), len(env))
			}
			for i := range env {
				if env[i].Name != tt.expectedEnv[i].Name || env[i].Value != tt.expectedEnv[i].Value {
					t.Errorf("Expected env var %v, got %v", tt.expectedEnv[i], env[i])
				}
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
