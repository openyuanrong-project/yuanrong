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
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
)

// ManagerState add the status to be saved here.
type ManagerState struct {
	PatKeyList       map[string]map[string]struct{} `json:"PatKeyList" valid:"optional"`
	RemoteClientList map[string]struct{}            `json:"RemoteClientList" valid:"optional"`
}

const (
	defaultHandlerQueueSize = 2000
	stateKey                = "/faas/state/recover/faasmanager"
)

var (
	// managerState -
	managerState = &ManagerState{
		PatKeyList:       make(map[string]map[string]struct{}),
		RemoteClientList: map[string]struct{}{},
	}
	// ManagerStateLock -
	ManagerStateLock    sync.RWMutex
	managerHandlerQueue *state.Queue
)

// InitState -
func InitState() {
	if managerHandlerQueue != nil {
		return
	}
	managerHandlerQueue = state.NewStateQueue(defaultHandlerQueueSize)
	if managerHandlerQueue == nil {
		return
	}
	go managerHandlerQueue.Run(updateState)
}

// SetState -
func SetState(byte []byte) error {
	return json.Unmarshal(byte, managerState)
}

// GetState -
func GetState() *ManagerState {
	ManagerStateLock.RLock()
	defer ManagerStateLock.RUnlock()
	return managerState
}

// GetStateByte is used to obtain the local state
func GetStateByte() ([]byte, error) {
	if managerHandlerQueue == nil {
		return nil, fmt.Errorf("managerHandlerQueue is not initialized")
	}
	ManagerStateLock.RLock()
	defer ManagerStateLock.RUnlock()
	stateBytes, err := managerHandlerQueue.GetState(stateKey)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(stateBytes, managerState); err != nil {
		log.GetLogger().Errorf("update managerState error :%s", err.Error())
	}
	log.GetLogger().Debugf("get state from etcd managerState: %v", string(stateBytes))
	return stateBytes, nil
}

func updateState(value interface{}, tags ...string) {
	if managerHandlerQueue == nil {
		log.GetLogger().Errorf("manager state managerHandlerQueue is nil")
		return
	}
	ManagerStateLock.Lock()
	defer ManagerStateLock.Unlock()
	switch v := value.(type) {
	case map[string]map[string]struct{}:
		managerState.PatKeyList = v
		log.GetLogger().Infof("update manager state for PatKeyList")
	case map[string]struct{}:
		managerState.RemoteClientList = v
		log.GetLogger().Infof("update manager state for RemoteClientList")
	default:
		log.GetLogger().Warnf("unknown data type for ManagerState")
		return
	}

	state, err := json.Marshal(managerState)
	if err != nil {
		log.GetLogger().Errorf("get manager state error %s", err.Error())
		return
	}
	if err = managerHandlerQueue.SaveState(state, stateKey); err != nil {
		log.GetLogger().Errorf("save manager state error: %s", err.Error())
		return
	}
	log.GetLogger().Infof("update manager state: %v", string(state))
}

// Update is used to write manger state to the cache queue
func Update(value interface{}, tags ...string) {
	if managerHandlerQueue == nil {
		log.GetLogger().Errorf("manager state managerHandlerQueue is nil")
		return
	}
	if err := managerHandlerQueue.Push(value, tags...); err != nil {
		log.GetLogger().Errorf("failed to push state to state queue: %s", err.Error())
	}
}
