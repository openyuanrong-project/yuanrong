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

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	ascendResourceD910B             = "huawei.com/ascend-1980"
	ascendResourceD910BInstanceType = "instanceType"
)

// ConvertToResSpecKey converts ResourceSpecification to ResSpecKey
func ConvertToResSpecKey(resSpec *ResourceSpecification) ResSpecKey {
	// for Go 1.7+ version, json.Marshal sorts the keys of map, same kv pairs will get same serialization result
	var (
		cstResExp     string
		cstResSpecExp string
	)
	if resSpec.CustomResources != nil && len(resSpec.CustomResources) != 0 {
		cstResBytes, err := json.Marshal(resSpec.CustomResources)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal customResources %#v error %s", resSpec.CustomResources, err.Error())
		}
		cstResExp = string(cstResBytes)
	}
	if len(cstResExp) != 0 && resSpec.CustomResourcesSpec != nil && len(resSpec.CustomResourcesSpec) != 0 {
		cstResSpecBytes, err := json.Marshal(resSpec.CustomResourcesSpec)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal customResourcesSpec %#v error %s", resSpec.CustomResourcesSpec,
				err.Error())
		}
		cstResSpecExp = string(cstResSpecBytes)
	}
	return ResSpecKey{
		CPU:                 resSpec.CPU,
		Memory:              resSpec.Memory,
		EphemeralStorage:    resSpec.EphemeralStorage,
		CustomResources:     cstResExp,
		CustomResourcesSpec: cstResSpecExp,
		InvokeLabel:         resSpec.InvokeLabel,
	}
}

// GetResKeyFromStr -
func GetResKeyFromStr(note string) (ResSpecKey, error) {
	resSpec := &ResourceSpecification{}
	err := json.Unmarshal([]byte(note), resSpec)
	if err != nil {
		return ResSpecKey{}, err
	}
	return ConvertToResSpecKey(resSpec), nil
}

// ConvertResourceMetaDataToResSpec will convert resource metadata
func ConvertResourceMetaDataToResSpec(resMeta types.ResourceMetaData) *ResourceSpecification {
	customResources := map[string]int64{}
	if resMeta.CustomResources != "" {
		if err := json.Unmarshal([]byte(resMeta.CustomResources), &customResources); err != nil {
			log.GetLogger().Warnf("failed to unmarshal custom resources %s, err: %s",
				resMeta.CustomResources, err.Error())
		}
	}
	customResourcesSpec := make(map[string]interface{})
	// npu tag may be unspecified and be updated to 376T, default value is needed to be set, otherwise reserved instance
	// will be recreated
	err := json.Unmarshal([]byte(resMeta.CustomResourcesSpec), &customResourcesSpec)
	if resMeta.CustomResourcesSpec != "" && err != nil {
		log.GetLogger().Warnf("failed to unmarshal custom resourcesSpec: %s,  err: %s",
			resMeta.CustomResourcesSpec, err.Error())
	}
	if _, ok := customResources[ascendResourceD910B]; ok {
		if _, ok := customResourcesSpec[ascendResourceD910BInstanceType]; !ok {
			customResourcesSpec[ascendResourceD910BInstanceType] = "376T"
		}
	}
	return &ResourceSpecification{
		CPU:                 resMeta.CPU,
		Memory:              resMeta.Memory,
		CustomResources:     customResources,
		CustomResourcesSpec: customResourcesSpec,
		EphemeralStorage:    resMeta.EphemeralStorage,
	}
}
