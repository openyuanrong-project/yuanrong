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
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/workermanager"
)

func createInstanceForFG(request createInstanceRequest) (*types.Instance, error) {
	createInstanceTraceID := uuid.New().String()
	logger := log.GetLogger().With(zap.Any("funcKey", request.funcSpec.FuncKey),
		zap.Any("traceID", createInstanceTraceID))
	createBeginTime := time.Now()
	ScaleOutParam := &workermanager.ScaleUpParam{
		TraceID:     createInstanceTraceID,
		FunctionKey: request.funcSpec.FuncKey,
		Timeout:     request.createTimeout,
		CPU:         int(request.resKey.CPU),
		Memory:      int(request.resKey.Memory),
	}
	wmInstance, err := workermanager.ScaleUpInstance(ScaleOutParam)
	if err != nil {
		createErr := err
		logger.Errorf("createErr is %v , instance %s, cost: %s", createErr, wmInstance, time.Since(createBeginTime))
		return nil, createErr
	}
	instance := buildInstance(wmInstance.InstanceID, request)
	instance.InstanceIP = wmInstance.IP
	instance.InstancePort = wmInstance.Port
	instance.NodeIP = wmInstance.OwnerIP
	instance.ResKey.CPU = wmInstance.Resource.Runtime.CPULimit
	instance.ResKey.Memory = wmInstance.Resource.Runtime.MemoryLimit
	instance.NodePort = constant.BusProxyHTTPPort
	logger.Infof("succeed to create instance %s, cost: %s, instance: %+v", wmInstance.InstanceID,
		time.Since(createBeginTime), instance)
	return instance, nil
}

func deleteInstanceForFG(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo,
	instance *types.Instance) error {
	log.GetLogger().Debugf("start to delete instance %s for function %s",
		instance.InstanceID, funcSpec.FuncKey)
	err := workermanager.ScaleDownInstance(instance.InstanceID, funcSpec.FuncKey, "")
	if err != nil {
		log.GetLogger().Errorf("failed to delete instance %s for function %s,err: %s ",
			instance.InstanceID, funcSpec.FuncKey, err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to delete instance %s for function %s",
		instance.InstanceID, funcSpec.FuncKey)
	return nil
}

func deleteInstanceByIDForFG(instanceID, funcKey string) error {
	log.GetLogger().Debugf("start to delete instance %s for function %s",
		instanceID, funcKey)
	err := workermanager.ScaleDownInstance(instanceID, funcKey, "")
	if err != nil {
		log.GetLogger().Errorf("failed to delete instance %s for function %s,err: %s ",
			instanceID, funcKey, err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to delete instance %s for function %s",
		instanceID, funcKey)
	return nil
}

func signalInstanceForFG(instance *types.Instance, signal int) {}
