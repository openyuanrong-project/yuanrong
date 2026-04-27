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
	"strconv"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	workersEtcdPrefix         = "/sn/workers"
	validEtcdKeyLenForWorkers = 13
	versionKeyIndex           = 9
	instanceIDFGValueIndex    = 12
	workersKeyIndex           = 2
)

// watcherFGFilter will filter version FG instance event from etcd event
func (ir *InstanceRegistry) watcherFGFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForWorkers {
		return true
	}
	if items[workersKeyIndex] != "workers" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" || items[versionKeyIndex] != "version" ||
		!strings.HasPrefix(items[executorKeyIndex], "0@") {
		return true
	}
	return false
}

// watcherFGHandler will handle FG instance event from etcd
func (ir *InstanceRegistry) watcherFGHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling instance event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("instance registry ready to receive etcd kv")
		ir.fgListDoneCh <- struct{}{}
		return
	}
	instanceID := GetInstanceIDFromFGEtcdKey(event.Key)
	if len(instanceID) == 0 {
		log.GetLogger().Warnf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	ir.handleEtcdEvent(event, instanceID, GetInsSpecFromEtcdKVForFG)
}

// GetInsSpecFromEtcdKVForFG gets InstanceSpecification from etcd key and value of instance
func GetInsSpecFromEtcdKVForFG(etcdKey string, etcdValue []byte) *commonTypes.InstanceSpecification {
	insSpecFG := &commonTypes.InstanceSpecificationFG{}
	insSpec := &commonTypes.InstanceSpecification{}
	// extract fields from etcd key
	keyFeilds := strings.Split(etcdKey, keySeparator)
	if len(keyFeilds) != validEtcdKeyLenForWorkers {
		log.GetLogger().Errorf("the etcdKey length doesn't match FG vesrsion!")
		return nil
	}
	tenantID := keyFeilds[6]
	functionName := keyFeilds[8]
	version := keyFeilds[10]
	insSpec.Function = tenantID + "/" + functionName + "/" + version
	if tenantID == "" || functionName == "" {
		log.GetLogger().Errorf("failed to get tenatID or functionName!,etcdKey %s", etcdKey)
		return nil
	}
	insSpec.Function = tenantID + "/" + functionName + "/" + version
	// extract fields from etcd value
	err := json.Unmarshal(etcdValue, insSpecFG)
	if err != nil {
		log.GetLogger().Errorf("funcKey %s,failed to unmarshal etcd value to instance specification %s",
			insSpec.Function, err.Error())
		return nil
	}
	if !buildCreateOptions(insSpecFG, insSpec) {
		log.GetLogger().Warnf("funcKey %s,applier %s is invalid,ignore this instance",
			insSpec.Function, insSpecFG.Applier)
		return nil
	}
	insSpec.StartTime = strconv.Itoa(insSpecFG.CreationTime)
	insSpec.RuntimeAddress = insSpecFG.NodeIP + ":" + insSpecFG.NodePort
	insSpec.FunctionProxyID = insSpecFG.InstanceIP + ":" + insSpecFG.InstancePort
	insSpec.ParentID = insSpecFG.Applier
	insSpec.InstanceStatus = commonTypes.InstanceStatus{
		Code: int32(constant.KernelInstanceStatusRunning),
		Msg:  "",
	}
	buildResource(insSpec, insSpecFG)
	return insSpec
}

func buildResource(insSpec *commonTypes.InstanceSpecification, insSpecFG *commonTypes.InstanceSpecificationFG) {
	var ephemeralStorage int
	funcSpec := GlobalRegistry.GetFuncSpec(insSpec.Function)
	if funcSpec != nil && funcSpec.ResourceMetaData.EphemeralStorage != 0 {
		ephemeralStorage = funcSpec.ResourceMetaData.EphemeralStorage
	}
	insSpec.Resources = commonTypes.Resources{
		Resources: map[string]commonTypes.Resource{
			constant.ResourceCPUName: {
				Name: constant.ResourceCPUName,
				Scalar: commonTypes.ValueScalar{
					Value: float64(insSpecFG.Resource.Runtime.CPULimit),
				},
			},
			constant.ResourceMemoryName: {
				Name: constant.ResourceMemoryName,
				Scalar: commonTypes.ValueScalar{
					Value: float64(insSpecFG.Resource.Runtime.MemoryLimit),
				},
			},
			constant.ResourceEphemeralStorage: {
				Name: constant.ResourceEphemeralStorage,
				Scalar: commonTypes.ValueScalar{
					Value: float64(ephemeralStorage),
				},
			},
		}}
}

func buildCreateOptions(insSpecFG *commonTypes.InstanceSpecificationFG,
	insSpec *commonTypes.InstanceSpecification) bool {
	if strings.HasPrefix(insSpecFG.Applier, constant.FaasSchedulerApplier) {
		insSpec.CreateOptions = map[string]string{
			types.InstanceTypeNote: string(types.InstanceTypeScaled),
			types.FunctionKeyNote:  insSpec.Function,
			types.SchedulerIDNote:  insSpecFG.Applier,
		}
		return true
	}
	if insSpecFG.Applier == constant.WorkerManagerApplier ||
		insSpecFG.BusinessType == constant.BusinessTypeCAE {
		insSpec.CreateOptions = map[string]string{
			types.InstanceTypeNote: string(types.InstanceTypeReserved),
			types.FunctionKeyNote:  insSpec.Function,
			types.SchedulerIDNote:  insSpecFG.Applier,
		}
		return true
	}
	return false
}

// GetInstanceIDFromFGEtcdKey gets instance id from etcd key of instance
func GetInstanceIDFromFGEtcdKey(etcdKey string) string {
	items := strings.Split(etcdKey, keySeparator)
	if len(items) != validEtcdKeyLenForWorkers {
		return ""
	}
	instanceID := items[instanceIDFGValueIndex]
	return instanceID
}
