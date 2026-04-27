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
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	funcInfoMinLen = 3
	// InstanceScalePolicyStaticFunction is the schedule policy for static function
	InstanceScalePolicyStaticFunction = "staticFunction"
)

// GetFuncMetaSignature will calculate function signature based on essentials
func GetFuncMetaSignature(metaInfo *types.FunctionMetaInfo, filterFlag bool) string {
	// static function set revisionID as signature
	if metaInfo.InstanceMetaData.ScalePolicy == InstanceScalePolicyStaticFunction {
		return metaInfo.FuncMetaData.RevisionID
	}
	metaInfoCopy := &types.FunctionMetaInfo{}
	if err := DeepCopyObj(metaInfo, metaInfoCopy); err != nil {
		return "invalid function meta info"
	}
	if filterFlag {
		metaInfoFieldFilter(metaInfoCopy)
	}
	metaInfoCopy.FuncMetaData.FuncID = ""
	metaInfoCopy.FuncMetaData.Type = ""
	metaInfoCopy.FuncMetaData.EnableCloudDebug = ""
	metaInfoCopy.FuncMetaData.Dependencies = ""
	metaInfoCopy.FuncMetaData.CodeSize = 0
	metaInfoCopy.FuncMetaData.CodeSha512 = ""
	metaInfoCopy.FuncMetaData.FunctionType = ""
	metaInfoCopy.FuncMetaData.Tags = nil
	metaInfoCopy.FuncMetaData.FunctionDescription = ""
	metaInfoCopy.FuncMetaData.FunctionUpdateTime = ""
	metaInfoCopy.InstanceMetaData.ScalePolicy = ""
	metaInfoCopy.InstanceMetaData.MaxInstance = 0
	metaInfoCopy.InstanceMetaData.MinInstance = 0
	metaInfoCopy.ExtendedMetaData.DynamicConfig.UpdateTime = ""
	metaInfoCopy.ExtendedMetaData.DynamicConfig.ConfigContent = []types.KV{}
	metaInfoCopy.ExtendedMetaData.StrategyConfig = types.StrategyConfig{}
	metaInfoCopy.ExtendedMetaData.ExtendConfig = ""
	metaInfoCopy.ExtendedMetaData.EnterpriseProjectID = ""
	metaInfoCopy.ExtendedMetaData.AsyncConfigLoaded = false
	metaInfoCopy.ExtendedMetaData.NetworkController = types.NetworkController{}
	metaInfoCopy.ResourceMetaData.CustomResourcesSpec =
		getCustomResourceSpec(metaInfo.ResourceMetaData.CustomResources, metaInfo.ResourceMetaData.CustomResourcesSpec)
	data, err := json.Marshal(metaInfoCopy)
	if err != nil {
		return "invalid function meta info"
	}
	return FnvHash(string(data))
}
func getCustomResourceSpec(customResources string, customResourceSpec string) string {
	// customResources为空，customResourceSpec必然为空
	if customResources == "" {
		return ""
	}
	customResourcesJSON := make(map[string]int64)
	customResourcesSpecJSON := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(customResources), &customResourcesJSON)

	err2 := json.Unmarshal([]byte(customResourceSpec), &customResourcesSpecJSON)
	if err1 != nil || (err2 != nil && customResourceSpec != "") {
		return ""
	}
	for k := range customResourcesJSON {
		if k == "huawei.com/ascend-1980" {
			_, ok := customResourcesSpecJSON["instanceType"]
			if !ok {
				customResourcesSpecJSON["instanceType"] = "376T"
			}
			break
		}
	}
	v, err3 := json.Marshal(customResourcesSpecJSON)
	if err3 != nil {
		return ""
	}
	return string(v)
}

func metaInfoFieldFilter(metaInfoCopy *types.FunctionMetaInfo) {
	metaInfoCopy.FuncMetaData.Service = ""
	metaInfoCopy.S3MetaData = types.S3MetaData{}

	metaInfoCopy.EnvMetaData = types.EnvMetaData{}

	metaInfoCopy.ResourceMetaData.EnableDynamicMemory = false
	metaInfoCopy.ResourceMetaData.EnableTmpExpansion = false
	metaInfoCopy.ResourceMetaData.GpuMemory = 0
	metaInfoCopy.ResourceMetaData.EphemeralStorage = 0

	metaInfoCopy.ExtendedMetaData.ImageName = ""
	if metaInfoCopy.ExtendedMetaData.VpcConfig != nil {
		metaInfoCopy.ExtendedMetaData.VpcConfig.Xrole = ""
	}
	metaInfoCopy.ExtendedMetaData.UserAgency = types.UserAgency{}
}

// FnvHash a hash function
func FnvHash(s string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return ""
	}

	// for 2 <= base <= 36. The result uses the lower-case letters 'a' to 'z'
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

// DeepCopyObj deal with src and dst
func DeepCopyObj(src interface{}, dst interface{}) error {
	if dst == nil {
		return fmt.Errorf("dst cannot be nil")
	}
	if src == nil {
		return fmt.Errorf("src cannot be nil")
	}

	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("unable to marshal src: %s", err)
	}

	err = json.Unmarshal(bytes, dst)
	if err != nil {
		return fmt.Errorf("unable to unmarshal into dst: %s", err)
	}
	return nil
}

// SetFuncMetaDynamicConfEnable will calculate DynamicConfig and set DynamicConfig.Enabled
func SetFuncMetaDynamicConfEnable(metaInfo *types.FunctionMetaInfo) {
	// The DynamicConfig.Enabled will use for calculate function signature.
	// When DynamicConfig.Enabled changes, the instance will be restarted.
	// If function version is not latest,DynamicConfig.Enabled will never change
	if len(metaInfo.ExtendedMetaData.DynamicConfig.UpdateTime) == 0 {
		metaInfo.ExtendedMetaData.DynamicConfig.Enabled = false
		return
	}
	//
	if metaInfo.FuncMetaData.Version == constant.DefaultURNVersion &&
		len(metaInfo.ExtendedMetaData.DynamicConfig.ConfigContent) == 0 {
		metaInfo.ExtendedMetaData.DynamicConfig.Enabled = false
		return
	}
	metaInfo.ExtendedMetaData.DynamicConfig.Enabled = true
}

// ParseFuncKey parse funcKey with format "tenantID/funcName/funcVersion" or "tenantID/funcName/funcVersion/CPU-memory"
func ParseFuncKey(funcKey string) (string, string, string) {
	funcInfo := strings.Split(funcKey, "/")
	if len(funcInfo) < funcInfoMinLen {
		return "", "", ""
	}
	return funcInfo[0], funcInfo[1], funcInfo[2]
}

// GetAPIType -
func GetAPIType(BusinessType string) api.ApiType {
	if BusinessType == constant.BusinessTypeServe {
		return api.ServeApi
	}
	return api.FaaSApi
}
