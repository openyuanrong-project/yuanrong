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
	"strings"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

const (
	// ResourcesCPU -
	ResourcesCPU = "CPU"
	// ResourcesMemory -
	ResourcesMemory = "Memory"
)

// ExtractInfoFromEtcdKey extracts infomation from etcdKey
func ExtractInfoFromEtcdKey(etcdKey string, index int) string {
	items := strings.Split(etcdKey, constant.KeySeparator)
	if len(items) != constant.ValidEtcdKeyLenForInstance {
		return ""
	}
	if index > len(items)-1 {
		log.GetLogger().Errorf("index is out of range")
		return ""
	}
	return items[index]
}

// ExtractInstanceIDFromEtcdKey will extract instanceID from etcdKey
func ExtractInstanceIDFromEtcdKey(etcdKey string) string {
	items := strings.Split(etcdKey, constant.KeySeparator)
	if len(items) != constant.ValidEtcdKeyLenForInstance {
		return ""
	}
	return items[len(items)-1]
}

// GetInstanceSpecFromEtcdValue will get schedulerSpec from etcd value
func GetInstanceSpecFromEtcdValue(etcdValue []byte) *types.InstanceSpecification {
	specMeta := &types.InstanceSpecificationMeta{}
	err := json.Unmarshal(etcdValue, specMeta)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal etcd value to function specification %s", err.Error())
		return nil
	}
	spec := &types.InstanceSpecification{InstanceSpecificationMeta: *specMeta}
	return spec
}

// GetMinSleepTime -
func GetMinSleepTime(defaultSleepTime, maxSleepTime time.Duration) time.Duration {
	if defaultSleepTime.Seconds() <= maxSleepTime.Seconds() {
		return defaultSleepTime
	}
	return maxSleepTime
}

// CheckNeedRecover -
func CheckNeedRecover(newInstanceMeta types.InstanceSpecificationMeta) bool {
	// instance status from running to fatal
	return newInstanceMeta.InstanceStatus.Code == int(constant.KernelInstanceStatusFatal) ||
		newInstanceMeta.InstanceStatus.Code == int(constant.KernelInstanceStatusEvicting) ||
		newInstanceMeta.InstanceStatus.Code == int(constant.KernelInstanceStatusScheduleFailed) ||
		newInstanceMeta.InstanceStatus.Code == int(constant.KernelInstanceStatusEvicted)
}

// GetSchedulerConfigSignature -
func GetSchedulerConfigSignature(schedulerCfg *types.SchedulerConfig) string {
	var cfg types.SchedulerConfig
	err := utils.DeepCopyObj(schedulerCfg, &cfg)
	if err != nil {
		log.GetLogger().Warnf("Failed to copy scheduler config: %v", err)
		return ""
	}
	cfg.SchedulerNum = 0
	cfgStr, err := json.Marshal(cfg)
	if err != nil {
		log.GetLogger().Errorf("Failed to marshal scheduler config: %v", err)
		return ""
	}
	h := utils.FnvHash(string(cfgStr))
	return h
}

// GetFrontendConfigSignature -
func GetFrontendConfigSignature(frontendCfg *types.FrontendConfig) string {
	var cfg types.FrontendConfig
	err := utils.DeepCopyObj(frontendCfg, &cfg)
	if err != nil {
		log.GetLogger().Warnf("Failed to copy frontend config: %v", err)
		return ""
	}
	cfg.InstanceNum = 0
	cfgStr, err := json.Marshal(cfg)
	if err != nil {
		log.GetLogger().Errorf("Failed to marshal frontend config: %v", err)
		return ""
	}
	return utils.FnvHash(string(cfgStr))
}

// GetManagerConfigSignature -
func GetManagerConfigSignature(managerCfg *types.ManagerConfig) string {
	var cfg types.ManagerConfig
	err := utils.DeepCopyObj(managerCfg, &cfg)
	if err != nil {
		log.GetLogger().Warnf("Failed to copy manager config: %v", err)
		return ""
	}
	cfg.ManagerInstanceNum = 0
	cfgStr, err := json.Marshal(cfg)
	if err != nil {
		log.GetLogger().Errorf("Failed to marshal manager config: %v", err)
		return ""
	}
	return utils.FnvHash(string(cfgStr))
}
