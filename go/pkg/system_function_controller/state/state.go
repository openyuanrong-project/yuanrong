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

// Package state -
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

// ControllerState add the status to be saved here.
type ControllerState struct {
	FaaSControllerConfig types.Config `json:"FaaSControllerConfig" valid:"optional"`
	// FaasInstance:  type(frontend/scheduler/manager) -- instanceID
	FaasInstance map[string]map[string]struct{} `json:"FaasInstance" valid:"optional"`
}

const (
	defaultHandlerQueueSize = 100
	faasInstanceTypeNum     = 3

	faasInstanceTagsLen       = 2
	faasInstanceTagsOptIndex  = 0
	faasInstanceTagsTypeIndex = 1
)

var (
	// controllerState -
	controllerState = &ControllerState{
		FaasInstance: make(map[string]map[string]struct{}, faasInstanceTypeNum),
	}
	controllerStateLock    sync.RWMutex
	controllerHandlerQueue *state.Queue
	stateKey               = ""
	allFaaSStateKey        = "/faas/state/recover"
)

func init() {
	stateKey = fmt.Sprintf("/faas/state/recover/faascontroller/%s", os.Getenv("INSTANCE_ID"))
}

// InitState -
func InitState(schedulerExclusivity []string) {
	if controllerHandlerQueue != nil {
		return
	}
	controllerState.FaasInstance[types.FaasFrontendInstanceState] = map[string]struct{}{}
	controllerState.FaasInstance[types.FaasSchedulerInstanceState] = map[string]struct{}{}
	for _, tenantID := range schedulerExclusivity {
		controllerState.FaasInstance[types.FaasSchedulerInstanceState+tenantID] = map[string]struct{}{}
	}
	controllerState.FaasInstance[types.FaasManagerInstanceState] = map[string]struct{}{}
	controllerHandlerQueue = state.NewStateQueue(defaultHandlerQueueSize)
	if controllerHandlerQueue == nil {
		return
	}
	go controllerHandlerQueue.Run(updateState)
	controllerStateLock.Lock()
	defer controllerStateLock.Unlock()
	stateBytes, err := controllerHandlerQueue.GetState(stateKey)
	if err != nil {
		log.GetLogger().Errorf("Failed to get state from etcd err: %v", err)
		return
	}
	err = json.Unmarshal(stateBytes, controllerState)
	if err != nil {
		log.GetLogger().Errorf("unmarshal controller state error %s", err.Error())
		return
	}
}

// GetState -
func GetState() ControllerState {
	controllerStateLock.RLock()
	defer controllerStateLock.RUnlock()
	return *controllerState
}

// SetState -
func SetState(byte []byte) error {
	return json.Unmarshal(byte, controllerState)
}

// GetStateByte is used to obtain the local state
func GetStateByte() ([]byte, error) {
	if controllerHandlerQueue == nil {
		return nil, fmt.Errorf("controllerHandlerQueue is not initialized")
	}
	controllerStateLock.RLock()
	defer controllerStateLock.RUnlock()
	stateBytes, err := controllerHandlerQueue.GetState(stateKey)
	if err != nil {
		return nil, err
	}
	log.GetLogger().Debugf("get state from etcd controllerState: %v", string(stateBytes))
	return stateBytes, nil
}

// DeleteStateByte -
func DeleteStateByte() error {
	if controllerHandlerQueue == nil {
		return fmt.Errorf("controllerHandlerQueue is not initialized")
	}
	controllerStateLock.Lock()
	defer controllerStateLock.Unlock()
	return controllerHandlerQueue.DeleteState(allFaaSStateKey)
}

func updateState(value interface{}, tags ...string) {
	if controllerHandlerQueue == nil {
		log.GetLogger().Errorf("controller state controllerHandlerQueue is nil")
		return
	}
	controllerStateLock.Lock()
	defer controllerStateLock.Unlock()
	switch v := value.(type) {
	case types.Config:
		controllerState.FaaSControllerConfig = v
		log.GetLogger().Infof("update controller state for controller config")
	case string:
		if len(tags) != faasInstanceTagsLen {
			log.GetLogger().Errorf("failed to operate the FaasInstance, tags length: %d", len(tags))
			return
		}
		if tags[faasInstanceTagsOptIndex] == types.StateUpdate {
			controllerState.FaasInstance[tags[faasInstanceTagsTypeIndex]][v] = struct{}{}
		} else if tags[faasInstanceTagsOptIndex] == types.StateDelete {
			delete(controllerState.FaasInstance[tags[faasInstanceTagsTypeIndex]], v)
		} else {
			log.GetLogger().Errorf("failed to operate the FaasInstance, opt is error %s", tags[0])
			return
		}
	default:
		log.GetLogger().Warnf("unknown data type for ControllerState")
		return
	}

	state, err := json.Marshal(controllerState)
	if err != nil {
		log.GetLogger().Errorf("get controller state error %s", err.Error())
		return
	}
	if err = controllerHandlerQueue.SaveState(state, stateKey); err != nil {
		log.GetLogger().Errorf("save controller state error %s", err.Error())
	}
	log.GetLogger().Infof("update controller state successfully")
}

// Update is used to write controller state to the cache queue
func Update(value interface{}, tags ...string) {
	if controllerHandlerQueue == nil {
		log.GetLogger().Errorf("controller state controllerHandlerQueue is nil")
		return
	}
	if err := controllerHandlerQueue.Push(value, tags...); err != nil {
		log.GetLogger().Errorf("failed to push state to state queue: %s", err.Error())
	}
}
