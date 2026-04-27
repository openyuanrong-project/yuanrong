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

// Package resspeckey -
package resspeckey

import (
	"encoding/json"
	"fmt"
	"sort"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// ResourceSpecification contains resource specification of a requested instance
type ResourceSpecification struct {
	CPU                 int64                  `json:"cpu"`
	Memory              int64                  `json:"memory"`
	InvokeLabel         string                 `json:"invokeLabels"`
	CustomResources     map[string]int64       `json:"customResources"`
	CustomResourcesSpec map[string]interface{} `json:"customResourcesSpec"`
	EphemeralStorage    int                    `json:"ephemeral_storage"`
}

// DeepCopy return a ResourceSpecification Copy
func (rs *ResourceSpecification) DeepCopy() *ResourceSpecification {
	customResource := map[string]int64{}
	for k, v := range rs.CustomResources {
		customResource[k] = v
	}
	customResourcesSpec := map[string]interface{}{}
	for k, v := range rs.CustomResourcesSpec {
		customResourcesSpec[k] = v
	}
	return &ResourceSpecification{
		CPU:                 rs.CPU,
		Memory:              rs.Memory,
		CustomResources:     customResource,
		InvokeLabel:         rs.InvokeLabel,
		CustomResourcesSpec: customResourcesSpec,
		EphemeralStorage:    rs.EphemeralStorage,
	}
}

// String returns ResourceSpecification as string
func (rs *ResourceSpecification) String() string {
	resourceExpression := fmt.Sprintf("cpu-%d-mem-%d", rs.CPU, rs.Memory)
	keys := make([]string, 0)
	for key, value := range rs.CustomResources {
		if value <= constant.MinCustomResourcesSize {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := rs.CustomResources[k]
		resourceExpression += fmt.Sprintf("-%s-%d", k, v)
	}

	keys = make([]string, 0, len(rs.CustomResourcesSpec))
	for k := range rs.CustomResourcesSpec {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := rs.CustomResourcesSpec[k]
		resourceExpression += fmt.Sprintf("-%s-%v", k, v)
	}
	if rs.InvokeLabel != "" {
		resourceExpression += fmt.Sprintf("-invoke-label-%s", rs.InvokeLabel)
	}
	resourceExpression += fmt.Sprintf("-ephemeral-storage-%v", rs.EphemeralStorage)
	return resourceExpression
}

// ResSpecKey is a representation of ResourceSpecification which can be used as key of map
type ResSpecKey struct {
	CPU                 int64
	Memory              int64
	EphemeralStorage    int
	CustomResources     string
	CustomResourcesSpec string
	InvokeLabel         string
}

// String returns ResSpecKey as string
func (rsk *ResSpecKey) String() string {
	return fmt.Sprintf("cpu-%d-mem-%d-storage-%d-cstRes-%s-cstResSpec-%s-invokeLabel-%s", rsk.CPU, rsk.Memory,
		rsk.EphemeralStorage, rsk.CustomResources, rsk.CustomResourcesSpec, rsk.InvokeLabel)
}

// ToResSpec convert ResSpecKey to ResourceSpecification
func (rsk *ResSpecKey) ToResSpec() *ResourceSpecification {
	cstRes := map[string]int64{}
	err := json.Unmarshal([]byte(rsk.CustomResources), &cstRes)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal to customResources error %s", err.Error())
	}
	cstResSpec := map[string]interface{}{}
	err = json.Unmarshal([]byte(rsk.CustomResourcesSpec), &cstResSpec)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal to customResourceSpec error %s", err.Error())
	}
	return &ResourceSpecification{
		CPU:                 rsk.CPU,
		Memory:              rsk.Memory,
		EphemeralStorage:    rsk.EphemeralStorage,
		CustomResources:     cstRes,
		CustomResourcesSpec: cstResSpec,
		InvokeLabel:         rsk.InvokeLabel,
	}
}
