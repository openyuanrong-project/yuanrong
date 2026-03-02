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

// Package registry -
package registry

import (
	"encoding/json"
	"fmt"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	functionEtcdPrefix                 = "/sn/function"
	instanceEtcdPrefix                 = "/sn/instance"
	userAgencyEtcdPrefix               = "/sn/agency/"
	instancesInfoEtcdPrefix            = "/instances"
	keySeparator                       = "/"
	validEtcdKeyLenForCluster          = 6
	validEtcdKeyLenForFrontend         = 7
	validEtcdKeyLenForAlias            = 10
	validEtcdKeyLenForFunction         = 11
	validEtcdKeyLenForInstance         = 14
	validEtcdKeyLenForAgency           = 11
	validEtcdKeyLenForInsConfig        = 12
	validEtcdKeyLenForInsWithLabelConf = 14
	validEtcdKeyLenForQuota1           = 7
	validEtcdKeyLenForQuota2           = 8
	quotaKeyIndex                      = 2
	instanceMetadataKeyIndex1          = 6
	instanceMetadataKeyIndex2          = 7
	clusterFrontendClusterIndex        = 4
	clusterFrontendIPIndex             = 5
	tenantValueIndex                   = 6
	funcNameValueIndex                 = 8
	versionValueIndex                  = 10
	instanceIDValueIndex               = 13
	instanceKeyIndex                   = 2
	aliasKeyIndex                      = 2
	tenantKeyIndex                     = 5
	functionKeyIndex                   = 7
	executorKeyIndex                   = 8
	agencyKeyIndex1                    = 2
	agencyKeyIndex2                    = 9
	agencyBusinessKeyIndex             = 3
	agencyTenantKeyIndex               = 5
	agencyDomainKeyIndex               = 7
	insInfoKeyIndex                    = 1
	insInfoClusterKeyIndex             = 4
	insInfoClusterValueIndex           = 5
	insInfoTenantKeyIndex              = 6
	insInfoFunctionKeyIndex            = 8
	insInfoTenantValueIndex            = 7
	insInfoFuncNameValueIndex          = 9
	insInfoVersionValueIndex           = 11
	insInfoLabelKeyIndex               = 12
	insInfoLabelValueIndex             = 13
	functionClusterKeyIdx              = 5
)

const (
	// SubEventTypeUpdate is update type of subscribe event
	SubEventTypeUpdate EventType = "update"
	// SubEventTypeDelete is delete type of subscribe event
	SubEventTypeDelete EventType = "delete"
	// SubEventTypeAdd is add type of subscribe event
	SubEventTypeAdd EventType = "add"
	// SubEventTypeSynced is synced type of subscribe event
	SubEventTypeSynced EventType = "synced"
	// SubEventTypeCRSynced -
	SubEventTypeCRSynced EventType = "crSynced" // todo 待后续修正
	// SubEventTypeRemove is remove type of instance event
	SubEventTypeRemove      EventType = "remove"
	defaultEphemeralStorage           = 512
)

var (
	// GlobalRegistry is the global registry
	GlobalRegistry *Registry
)

// EventType defines registry event type
type EventType string

// SubEvent contains event published to subscribers
type SubEvent struct {
	EventType
	EventMsg interface{}
}

// Registry watches etcd and builds registry cache based on etcd watch
type Registry struct {
	FaaSSchedulerRegistry     *FaasSchedulerRegistry
	FunctionRegistry          *FunctionRegistry
	AgentRegistry             *AgentRegistry
	InstanceRegistry          *InstanceRegistry
	FaaSManagerRegistry       *FaaSManagerRegistry
	InstanceConfigRegistry    *InstanceConfigRegistry
	AliasRegistry             *AliasRegistry
	FunctionAvailableRegistry *FunctionAvailableRegistry
	FaaSFrontendRegistry      *FaaSFrontendRegistry
	TenantQuotaRegistry       *TenantQuotaRegistry
	RolloutRegistry           *RolloutRegistry
	stopCh                    <-chan struct{}
}

// InitRegistry will initialize registry
func InitRegistry(stopCh <-chan struct{}) error {
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:          NewFunctionRegistry(stopCh),
		AgentRegistry:             NewAgentRegistry(stopCh),
		InstanceRegistry:          NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
		InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
		AliasRegistry:             NewAliasRegistry(stopCh),
		FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
		FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
		TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
		RolloutRegistry:           NewRolloutRegistry(stopCh),
	}
	GlobalRegistry.FaaSSchedulerRegistry.initWatcher(etcd3.GetRouterEtcdClient())
	GlobalRegistry.RolloutRegistry.initWatcher(etcd3.GetMetaEtcdClient())
	return nil
}

// ProcessETCDList before watch etcd event, list etcd kv first
func ProcessETCDList() {
	routerEtcdClient := etcd3.GetRouterEtcdClient()
	metaEtcdClient := etcd3.GetMetaEtcdClient()
	// Serial Execution
	if GlobalRegistry != nil {
		GlobalRegistry.FunctionRegistry.initWatcher(metaEtcdClient)
		GlobalRegistry.AliasRegistry.initWatcher(metaEtcdClient)
		if GlobalRegistry.AgentRegistry != nil {
			GlobalRegistry.AgentRegistry.RunWatcher()
		}
		GlobalRegistry.InstanceRegistry.initWatcher(routerEtcdClient)
		GlobalRegistry.FaaSManagerRegistry.initWatcher(routerEtcdClient)
		GlobalRegistry.InstanceConfigRegistry.initWatcher(metaEtcdClient)
		GlobalRegistry.FunctionAvailableRegistry.initWatcher(metaEtcdClient)
		GlobalRegistry.FaaSFrontendRegistry.initWatcher(metaEtcdClient)
		GlobalRegistry.TenantQuotaRegistry.initWatcher(metaEtcdClient)
	}
}

// StartRegistry will start registry
func StartRegistry() {
	if GlobalRegistry == nil {
		log.GetLogger().Errorf("faaSScheduler registry is nil")
		return
	}
	if GlobalRegistry.FaaSSchedulerRegistry != nil {
		GlobalRegistry.FaaSSchedulerRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("faaSScheduler registry is nil")
	}
	if GlobalRegistry.FunctionRegistry != nil {
		GlobalRegistry.FunctionRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("function registry is nil")
	}
	if GlobalRegistry.InstanceRegistry != nil {
		GlobalRegistry.InstanceRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("instance registry is nil")
	}
	if GlobalRegistry.FaaSManagerRegistry != nil {
		GlobalRegistry.FaaSManagerRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("faas manager registry is nil")
	}
	if GlobalRegistry.InstanceConfigRegistry != nil {
		GlobalRegistry.InstanceConfigRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("instances info registry is nil")
	}
	if GlobalRegistry.AliasRegistry != nil {
		GlobalRegistry.AliasRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("instances info registry is nil")
	}
	if GlobalRegistry.FunctionAvailableRegistry != nil {
		GlobalRegistry.FunctionAvailableRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("function available clusters registry is nil")
	}
	if GlobalRegistry.FaaSFrontendRegistry != nil {
		GlobalRegistry.FaaSFrontendRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("frontend instance registry is nil")
	}
	if GlobalRegistry.TenantQuotaRegistry != nil {
		GlobalRegistry.TenantQuotaRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("tenant instance registry is nil")
	}
	if GlobalRegistry.RolloutRegistry != nil {
		GlobalRegistry.RolloutRegistry.RunWatcher()
	} else {
		log.GetLogger().Errorf("rollout registry is nil")
	}
	commonUtils.ClearStringMemory(config.GlobalConfig.MetaETCDConfig.Password)
	commonUtils.ClearStringMemory(config.GlobalConfig.RouterETCDConfig.Password)
}

// GetFuncSpec will get function specification
func (r *Registry) GetFuncSpec(funcKey string) *types.FunctionSpecification {
	if r.AgentRegistry != nil {
		funcSpec := r.AgentRegistry.getFuncSpec(funcKey)
		if funcSpec != nil {
			return funcSpec
		}
	}
	return r.FunctionRegistry.getFuncSpec(funcKey)
}

// GetInstance will get instance
func (r *Registry) GetInstance(instanceID string) *types.Instance {
	return r.InstanceRegistry.GetInstance(instanceID)
}

// FetchSilentFuncSpec will get silent function specification
func (r *Registry) FetchSilentFuncSpec(funcKey string) *types.FunctionSpecification {
	return r.FunctionRegistry.fetchSilentFuncSpec(funcKey)
}

// SubscribeFuncSpec will add subscriber for function registry
func (r *Registry) SubscribeFuncSpec(subChan chan SubEvent) {
	r.FunctionRegistry.addSubscriberChan(subChan)
}

// SubscribeAgentRunInfo -
func (r *Registry) SubscribeAgentRunInfo(subChan chan SubEvent) {
	r.AgentRegistry.addSubscriberChan(subChan)
}

// SubscribeInsSpec will add subscriber for instance registry
func (r *Registry) SubscribeInsSpec(subChan chan SubEvent) {
	r.InstanceRegistry.addSubscriberChan(subChan)
	r.FaaSManagerRegistry.addSubscriberChan(subChan)
}

// SubscribeInsConfig will add subscriber for instanceConfig registry
func (r *Registry) SubscribeInsConfig(subChan chan SubEvent) {
	r.InstanceConfigRegistry.addSubscriberChan(subChan)
}

// SubscribeAliasSpec will add subscriber for instanceConfig registry
func (r *Registry) SubscribeAliasSpec(subChan chan SubEvent) {
	r.AliasRegistry.addSubscriberChan(subChan)
}

// SubscribeSchedulerProxy will add subscriber for scheduler registry
func (r *Registry) SubscribeSchedulerProxy(subChan chan SubEvent) {
	r.FaaSSchedulerRegistry.addSubscriberChan(subChan)
}

// SubscribeRolloutConfig will add subscriber for rollout config
func (r *Registry) SubscribeRolloutConfig(subChan chan SubEvent) {
	r.RolloutRegistry.addSubscriberChan(subChan)
}

// GetFuncKeyFromEtcdKey will get funcKey from etcd key
func GetFuncKeyFromEtcdKey(etcdKey string) string {
	items := strings.Split(etcdKey, keySeparator)
	if len(items) != validEtcdKeyLenForFunction {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", items[tenantValueIndex], items[funcNameValueIndex], items[versionValueIndex])
}

// GetFuncMetaInfoFromEtcdValue will get FunctionMetaInfo from etcd value
func GetFuncMetaInfoFromEtcdValue(etcdValue []byte) *commonTypes.FunctionMetaInfo {
	funcMetaInfo := &commonTypes.FunctionMetaInfo{}
	err := json.Unmarshal(etcdValue, funcMetaInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal etcd value to function meta info %s", err.Error())
		return nil
	}
	// set default value
	if config.GlobalConfig.Scenario == types.ScenarioFunctionGraph && strings.Contains(funcMetaInfo.FuncMetaData.Runtime,
		types.CustomContainerRuntimeType) && funcMetaInfo.ResourceMetaData.EphemeralStorage == 0 {
		funcMetaInfo.ResourceMetaData.EphemeralStorage = int(config.GlobalConfig.EphemeralStorage)
		if npu, _ := utils.GetNpuTypeAndInstanceTypeFromStr(funcMetaInfo.ResourceMetaData.CustomResources,
			funcMetaInfo.ResourceMetaData.CustomResourcesSpec); npu != "" &&
			config.GlobalConfig.NpuEphemeralStorage != 0 {
			funcMetaInfo.ResourceMetaData.EphemeralStorage = int(config.GlobalConfig.NpuEphemeralStorage)
		}
		if funcMetaInfo.ResourceMetaData.EphemeralStorage == 0 {
			funcMetaInfo.ResourceMetaData.EphemeralStorage = defaultEphemeralStorage
		}
	}
	// currently for instances of algorithm function in CaaS which utilizes NPU need to be scheduled with round-robin
	// policy
	if funcMetaInfo.FuncMetaData.BusinessType == constant.BusinessTypeCAE {
		funcMetaInfo.InstanceMetaData.SchedulePolicy = types.InstanceSchedulePolicyMicroService
	}
	return funcMetaInfo
}
