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

// Package state is used to save and restore the scheduler state.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// SchedulerState add the status to be saved here.
type SchedulerState struct {
	InstancePool map[string]*types.InstancePoolState `json:"InstancePool" valid:"optional"` // funcKey - instance
}

// SchedulerEtcdRevision -
type SchedulerEtcdRevision struct {
	// it means etcd key /sn/instance/ revision, all operation on the key will update LastInstanceRevision
	LastInstanceRevision int32 `json:"lastInstanceRevision"`
}

const defaultHandlerQueueSize = 2000

var (
	schedulerStateLock sync.RWMutex
	// schedulerState -
	schedulerState = &SchedulerState{
		InstancePool: make(map[string]*types.InstancePoolState),
	}
	schedulerEtcdRev      *SchedulerEtcdRevision
	schedulerHandlerQueue *state.Queue
	stateKey              = ""
	stateRevKey           = ""
)

// RecoverConfig recover config
func RecoverConfig() error {
	log.GetLogger().Infof("GlobalConfig recovered")
	return nil
}

func init() {
	schedulerInstanceIDSelf := os.Getenv("INSTANCE_ID")
	stateKey = "/faas/state/recover/faasscheduler/" + schedulerInstanceIDSelf
	stateRevKey = "/faas/state/revision/faasscheduler/" + schedulerInstanceIDSelf
}

// InitState -
func InitState() {
	if config.GlobalConfig.StateDisable {
		log.GetLogger().Warnf("state is disable, skip init state")
		return
	}
	for k, v := range schedulerState.InstancePool {
		log.GetLogger().Infof("recover state: %s: %v", k, *v)
	}

	if schedulerHandlerQueue != nil {
		return
	}
	schedulerHandlerQueue = state.NewStateQueue(defaultHandlerQueueSize)
	if schedulerHandlerQueue == nil {
		return
	}
	go schedulerHandlerQueue.Run(updateState)
}

// SetState -
func SetState(byte []byte) error {
	return json.Unmarshal(byte, schedulerState)
}

// RecoverStateRev -
func RecoverStateRev() {
	if schedulerEtcdRev == nil && schedulerHandlerQueue != nil {
		stateRevBytes, err := schedulerHandlerQueue.GetState(stateRevKey)
		if err != nil {
			log.GetLogger().Warnf("failed to get stateRev from etcd, err:%s", err.Error())
			return
		}
		if err := json.Unmarshal(stateRevBytes, &schedulerEtcdRev); err != nil {
			log.GetLogger().Warnf("failed to unmarshal stateRev, err:%s", err.Error())
			return
		}
	}
}

// GetState -
func GetState() *SchedulerState {
	schedulerStateLock.RLock()
	defer schedulerStateLock.RUnlock()
	return schedulerState
}

// GetStateRev -
func GetStateRev() *SchedulerEtcdRevision {
	schedulerStateLock.RLock()
	defer schedulerStateLock.RUnlock()
	return schedulerEtcdRev
}

// GetStateByte is used to obtain the local state
func GetStateByte() ([]byte, error) {
	if schedulerHandlerQueue == nil {
		return nil, fmt.Errorf("schedulerHandlerQueue is not initialized")
	}
	schedulerStateLock.RLock()
	defer schedulerStateLock.RUnlock()
	stateBytes, err := schedulerHandlerQueue.GetState(stateKey)
	if err != nil {
		return nil, err
	}
	log.GetLogger().Debugf("get state from etcd schedulerState: %v", string(stateBytes))
	return stateBytes, nil
}

func updateState(value interface{}, tags ...string) {
	if schedulerHandlerQueue == nil {
		log.GetLogger().Errorf("scheduler state schedulerHandlerQueue is nil")
		return
	}
	var (
		stateBytes []byte
		updateKey  string
		err        error
	)
	schedulerStateLock.Lock()
	defer schedulerStateLock.Unlock()
	switch v := value.(type) {
	case *types.InstancePoolStateInput:
		// tags[0] as opt
		if len(tags) <= 0 {
			log.GetLogger().Errorf("failed to operate the instancePool, tags is empty")
			return
		}
		switch tags[0] {
		case types.StateUpdate:
			log.GetLogger().Infof("update scheduler state for instance queue")
			updateInstancePool(v)
		case types.StateDelete:
			log.GetLogger().Info("delete scheduler state for instance queue")
			deleteInstancePool(v)
		default:
			log.GetLogger().Errorf("failed to operate the instancePool, opt is error %s", tags[0])
			return
		}
		if stateBytes, err = json.Marshal(schedulerState); err != nil {
			log.GetLogger().Errorf("get scheduler state error %s", err.Error())
			return
		}
		updateKey = stateKey
	case int32:
		if schedulerEtcdRev == nil {
			schedulerEtcdRev = &SchedulerEtcdRevision{}
		}
		schedulerEtcdRev.LastInstanceRevision = v
		if stateBytes, err = json.Marshal(schedulerEtcdRev); err != nil {
			log.GetLogger().Errorf("get scheduler state error %s", err.Error())
			return
		}
		updateKey = stateRevKey
	default:
		log.GetLogger().Warnf("unknown data type for scheduler state")
		return
	}
	if len(stateBytes) <= 0 || updateKey == "" {
		return
	}
	if err = schedulerHandlerQueue.SaveState(stateBytes, updateKey); err != nil {
		log.GetLogger().Errorf("save scheduler state error: %s", err.Error())
		return
	}
	log.GetLogger().Infof("update scheduler state successfully")
}

// Update is used to write scheduler state to the cache queue
func Update(value interface{}, tags ...string) {
	if schedulerHandlerQueue == nil {
		return
	}
	if err := schedulerHandlerQueue.Push(value, tags...); err != nil {
		log.GetLogger().Errorf("failed to push state to state queue: %s", err.Error())
	}
}

// updateInstancePool adds an element to the queue specified by the tag
func updateInstancePool(value *types.InstancePoolStateInput) {
	if value.InstanceType == types.InstanceTypeState {
		if schedulerState.InstancePool[value.FuncKey].StateInstance[value.StateID] != nil {
			log.GetLogger().Warnf("state, func %s state %s already has instance %s",
				value.FuncKey, value.StateID,
				schedulerState.InstancePool[value.FuncKey].StateInstance[value.StateID].InstanceID)
		}
		schedulerState.InstancePool[value.FuncKey].StateInstance[value.StateID] = &types.Instance{
			InstanceStatus: commonTypes.InstanceStatus{Code: value.InstanceStatusCode},
			InstanceType:   types.InstanceTypeState,
			InstanceID:     value.InstanceID,
			FuncKey:        value.FuncKey,
			FuncSig:        value.FuncSig,
		}
		return
	}
	log.GetLogger().Errorf("failed to update the instancePool")
}

// deleteInstancePool deletes an element from the queue specified by the tag
func deleteInstancePool(value *types.InstancePoolStateInput) {
	if value.StateID != "" && value.InstanceType == types.InstanceTypeState {
		delete(schedulerState.InstancePool[value.FuncKey].StateInstance, value.StateID)
		log.GetLogger().Infof("state del state %s", value.StateID)
		return
	}
	delete(schedulerState.InstancePool, value.FuncKey)
}
