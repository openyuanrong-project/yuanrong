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

package instancepool

import (
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/signalmanager"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// CreateInstance -
func CreateInstance(request createInstanceRequest) (*types.Instance, error) {
	if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG {
		return createInstanceForFG(request)
	}
	return createInstanceForKernel(request)
}

// DeleteInstance -
func DeleteInstance(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo,
	instance *types.Instance) error {
	if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG {
		return deleteInstanceForFG(funcSpec, faasManagerInfo, instance)
	}
	return deleteInstanceForKernel(funcSpec, faasManagerInfo, instance)
}

// DeleteInstanceByID -
func DeleteInstanceByID(instanceID, funcKey string) error {
	if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG {
		return deleteInstanceByIDForFG(instanceID, funcKey)
	}
	return deleteInstanceByIDForKernel(instanceID, funcKey)
}

// DeleteUnexpectInstance -
func DeleteUnexpectInstance(parentID, instanceID, funcKey string, logger api.FormatLogger) {
	if parentID != selfregister.SelfInstanceID &&
		selfregister.GlobalSchedulerProxy.Contains(parentID) ||
		parentID == constant.WorkerManagerApplier ||
		strings.HasPrefix(parentID, constant.FunctionTaskApplier) ||
		parentID == constant.ASBResApplier ||
		parentID == constant.StaticInstanceApplier {
		return
	}
	logger.Warnf("instance is belong to this scheduler, but not found function meta, start to delete instance.")
	err := DeleteInstanceByID(instanceID, funcKey)
	if err != nil {
		logger.Errorf("failed to delete instance, err: %v", err)
	}
}

// SignalInstance -
func SignalInstance(instance *types.Instance, signal int) {
	if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG {
		signalInstanceForFG(instance, signal)
		return
	}
	signalmanager.GetSignalManager().SignalInstance(instance, signal)
}

func buildInstance(instanceID string, request createInstanceRequest) *types.Instance {
	return &types.Instance{
		InstanceType: request.instanceType,
		ResKey:       request.resKey,
		InstanceID:   instanceID,
		InstanceName: request.instanceName,
		ParentID:     selfregister.SelfInstanceID,
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
			Msg:  "",
		},
		FuncKey:       request.funcSpec.FuncKey,
		FuncSig:       request.funcSpec.FuncMetaSignature,
		ConcurrentNum: request.funcSpec.InstanceMetaData.ConcurrentNum,
	}
}
