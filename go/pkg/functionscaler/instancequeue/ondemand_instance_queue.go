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

// Package instancequeue -
package instancequeue

import (
	"context"
	"fmt"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// OnDemandInstanceQueue -
type OnDemandInstanceQueue struct {
	funcSpec           *types.FunctionSpecification
	instanceMap        map[string]*types.Instance
	insNameMap         map[string]*types.Instance
	instanceType       types.InstanceType
	resKey             resspeckey.ResSpecKey
	funcCtx            context.Context
	funcKey            string
	funcSig            string
	funcKeyWithRes     string
	createInstanceFunc CreateInstanceFunc
	deleteInstanceFunc DeleteInstanceFunc
	signalInstanceFunc SignalInstanceFunc
	sync.RWMutex
}

// NewOnDemandInstanceQueue -
func NewOnDemandInstanceQueue(config *InsQueConfig) *OnDemandInstanceQueue {
	return &OnDemandInstanceQueue{
		funcSpec:           config.FuncSpec,
		resKey:             config.ResKey,
		instanceMap:        make(map[string]*types.Instance, utils.DefaultMapSize),
		insNameMap:         make(map[string]*types.Instance, utils.DefaultMapSize),
		instanceType:       config.InstanceType,
		funcCtx:            config.FuncSpec.FuncCtx,
		funcKey:            config.FuncSpec.FuncKey,
		funcSig:            config.FuncSpec.FuncMetaSignature,
		funcKeyWithRes:     fmt.Sprintf("%s-%s", config.FuncSpec.FuncKey, config.ResKey.String()),
		createInstanceFunc: config.CreateInstanceFunc,
		deleteInstanceFunc: config.DeleteInstanceFunc,
		signalInstanceFunc: config.SignalInstanceFunc,
	}
}

// AcquireInstance will acquire an instance
func (oi *OnDemandInstanceQueue) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation,
	snerror.SNError) {
	select {
	case <-oi.funcCtx.Done():
		log.GetLogger().Errorf("function is deleted, can not acquire instance InsAlloc")
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
	}
	var instance *types.Instance
	instanceByName := oi.insNameMap[insAcqReq.InstanceName]
	instanceByID := oi.instanceMap[insAcqReq.DesignateInstanceID]
	if instanceByName != nil {
		instance = instanceByName
	} else if instanceByID != nil {
		instance = instanceByID
	} else {
		return nil, snerror.New(statuscode.InstanceNotFoundErrCode, statuscode.InstanceNotFoundErrMsg)
	}
	return &types.InstanceAllocation{
		AllocationID: instance.InstanceID,
		Instance:     instance,
	}, nil
}

// CreateInstance -
func (oi *OnDemandInstanceQueue) CreateInstance(insCrtReq *types.InstanceCreateRequest) (*types.Instance,
	snerror.SNError) {
	var (
		instance  *types.Instance
		createErr error
	)
	oi.RLock()
	functionSignature := oi.funcSig
	oi.RUnlock()
	instance, createErr = oi.createInstanceFunc(insCrtReq.TraceID, insCrtReq.InstanceName, oi.instanceType, oi.resKey,
		insCrtReq.CreateEvent)
	if createErr != nil {
		log.GetLogger().Errorf("failed to create instance for function %s error %s", oi.funcKeyWithRes,
			createErr.Error())
	}
	select {
	case _, ok := <-oi.funcCtx.Done():
		if !ok {
			log.GetLogger().Warnf("function %s is deleted, killing instance now", oi.funcKey)
			createErr = ErrFunctionDeleted
		}
	default:
		// in case of function signature change during instance creating
		oi.RLock()
		checkFunctionSignature := oi.funcSig
		oi.RUnlock()
		if functionSignature != checkFunctionSignature {
			log.GetLogger().Errorf("function signature changes while creating instance for function %s, "+
				"killing instance now", oi.funcKeyWithRes)
			createErr = ErrFuncSigMismatch
		}
	}
	if createErr != nil {
		if instance != nil {
			log.GetLogger().Warnf("killing failed created instance %s for function %s", instance.InstanceID,
				oi.funcKeyWithRes)
			go oi.DeleteInstance(instance)
		}
		return nil, buildSnError(createErr)
	}
	oi.Lock()
	oi.instanceMap[instance.InstanceID] = instance
	if len(insCrtReq.InstanceName) != 0 {
		oi.insNameMap[insCrtReq.InstanceName] = instance
	}
	oi.Unlock()
	return instance, nil
}

// DeleteInstance -
func (oi *OnDemandInstanceQueue) DeleteInstance(instance *types.Instance) snerror.SNError {
	oi.Lock()
	delete(oi.instanceMap, instance.InstanceID)
	if len(instance.InstanceName) != 0 {
		delete(oi.insNameMap, instance.InstanceName)
	}
	oi.Unlock()
	return buildSnError(oi.deleteInstanceFunc(instance))
}

// HandleInstanceUpdate -
func (oi *OnDemandInstanceQueue) HandleInstanceUpdate(instance *types.Instance) {
	log.GetLogger().Infof("handling instance update of function %s instanceID %s instanceName %s", oi.funcKeyWithRes,
		instance.InstanceID, oi.funcKeyWithRes)
}

// HandleInstanceDelete -
func (oi *OnDemandInstanceQueue) HandleInstanceDelete(instance *types.Instance) {
	log.GetLogger().Infof("handling instance delete of function %s instanceID %s instanceName %s", oi.funcKeyWithRes,
		instance.InstanceID, oi.funcKeyWithRes)
	oi.Lock()
	delete(oi.instanceMap, instance.InstanceID)
	delete(oi.insNameMap, instance.InstanceName)
	oi.Unlock()
}

// HandleFuncSpecUpdate -
func (oi *OnDemandInstanceQueue) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	log.GetLogger().Infof("handling funcSpec update of function %s", oi.funcKeyWithRes)
	if oi.funcSig != funcSpec.FuncMetaSignature {
		log.GetLogger().Warnf("function %s signature changes from %s to %s", oi.funcKeyWithRes, oi.funcSig,
			funcSpec.FuncMetaSignature)
		deleteList := make([]*types.Instance, len(oi.insNameMap))
		oi.Lock()
		for _, instance := range oi.instanceMap {
			deleteList = append(deleteList, instance)
		}
		oi.instanceMap = make(map[string]*types.Instance, utils.DefaultMapSize)
		oi.insNameMap = make(map[string]*types.Instance, utils.DefaultMapSize)
		oi.Unlock()
		for _, instance := range deleteList {
			go oi.deleteInstanceFunc(instance)
		}
	}
}

// Destroy -
func (oi *OnDemandInstanceQueue) Destroy() {
}

// GetInstanceNumber -
func (oi *OnDemandInstanceQueue) GetInstanceNumber(onlySelf bool) int {
	oi.RLock()
	insNum := len(oi.insNameMap)
	oi.RUnlock()
	return insNum
}
