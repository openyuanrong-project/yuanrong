/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

// Package utils -
package utils

import (
	"fmt"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	schedulePolicyKey = "schedule_policy"
	scheduleCPU       = "CPU"
	scheduleMemory    = "Memory"
)

const (
	// NodeSelectorKey -
	NodeSelectorKey = "node_selector"
	// MonopolyPolicyValue -
	MonopolyPolicyValue = "monopoly"
	// SharedPolicyValue -
	SharedPolicyValue = "shared"
)

// CreateCustomExtensions create customExtensions
func CreateCustomExtensions(customExtensions map[string]string, schedulePolicy string) map[string]string {
	if customExtensions == nil {
		customExtensions = make(map[string]string, 1)
	}
	customExtensions[schedulePolicyKey] = schedulePolicy
	return customExtensions
}

// CreatePodAffinity - create pod affinity
func CreatePodAffinity(key, label string, affinityType api.AffinityType) []api.Affinity {
	var (
		operators []api.LabelOperator
		affinity  []api.Affinity
	)
	if label != "" {
		operators = append(operators, api.LabelOperator{
			Type:        api.LabelOpIn,
			LabelKey:    key,
			LabelValues: []string{label},
		})
	} else {
		operators = append(operators, api.LabelOperator{
			Type:        api.LabelOpExists,
			LabelKey:    key,
			LabelValues: []string{},
		})
	}
	affinity = append(affinity, api.Affinity{
		Kind:                     api.AffinityKindInstance,
		Affinity:                 affinityType,
		PreferredPriority:        false,
		PreferredAntiOtherLabels: false,
		LabelOps:                 operators,
	})
	return affinity
}

// CreateCreateOptions create CreateOptions
func CreateCreateOptions(createOptions map[string]string, key, value string) map[string]string {
	if createOptions == nil {
		return make(map[string]string)
	}
	createOptions[key] = value
	return createOptions
}

// GenerateResourcesMap -
func GenerateResourcesMap(cpu, memory float64) map[string]float64 {
	resourcesMap := make(map[string]float64)
	resourcesMap[scheduleCPU] = cpu
	resourcesMap[scheduleMemory] = memory
	return resourcesMap
}

// AddNodeSelector -
func AddNodeSelector(nodeSelectorMap map[string]string, extraParams *types.ExtraParams) {
	if extraParams.CustomExtensions == nil {
		extraParams.CustomExtensions = make(map[string]string, 1)
	}
	if nodeSelectorMap != nil && len(nodeSelectorMap) != 0 {
		for k, v := range nodeSelectorMap {
			extraParams.CustomExtensions[NodeSelectorKey] = fmt.Sprintf(`{"%s": "%s"}`, k, v)
		}
	}
}
