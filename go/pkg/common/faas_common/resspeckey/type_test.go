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

// Package resspeckey -
package resspeckey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceSpecification_String(t *testing.T) {
	tests := []struct {
		name     string
		input    ResourceSpecification
		expected string
	}{
		{
			name: "only_cpu_and_memory",
			input: ResourceSpecification{
				CPU:    4,
				Memory: 16,
			},
			expected: "cpu-4-mem-16-ephemeral-storage-0",
		},
		{
			name: "with_custom_resources_and_spec",
			input: ResourceSpecification{
				CPU:    8,
				Memory: 32,
				CustomResources: map[string]int64{
					"gpu":  2,
					"nvme": 5,
				},
				CustomResourcesSpec: map[string]interface{}{
					"feature-flag": "beta",
					"version":      1.2,
				},
			},
			// 注意：CustomResourcesSpec 的键会被排序，所以 version 在前
			expected: "cpu-8-mem-32-gpu-2-nvme-5-feature-flag-beta-version-1.2-ephemeral-storage-0",
		},
		{
			name: "with_invoke_label",
			input: ResourceSpecification{
				CPU:         2,
				Memory:      8,
				InvokeLabel: "production-canary",
			},
			expected: "cpu-2-mem-8-invoke-label-production-canary-ephemeral-storage-0",
		},
		{
			name: "all_fields_populated",
			input: ResourceSpecification{
				CPU:    16,
				Memory: 64,
				CustomResources: map[string]int64{
					"gpu": 4,
				},
				CustomResourcesSpec: map[string]interface{}{
					"network-mode": "enhanced",
				},
				InvokeLabel:      "staging",
				EphemeralStorage: 1234,
			},
			expected: "cpu-16-mem-64-gpu-4-network-mode-enhanced-invoke-label-staging-ephemeral-storage-1234",
		},
		{
			name: "empty_optional_fields",
			input: ResourceSpecification{
				CPU:    1,
				Memory: 2,
				// CustomResources 为空
				// CustomResourcesSpec 为空
				// InvokeLabel 为空 ""
				// EphemeralStorage 为 false (零值)
			},
			expected: "cpu-1-mem-2-ephemeral-storage-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 testify/assert 进行更清晰的断言
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}
