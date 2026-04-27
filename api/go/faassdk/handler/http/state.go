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

// Package http -
package http

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/common/alarm"
	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/log"
)

const (
	defaultStateMapCapacity = 10
	defStateVal             = "1"
	cOK                     = 0 // same with datasystem K_OK
	cNotFound               = 3 // same with datasystem K_NOT_FOUND
)

const (
	// StateExistedErrCode -
	StateExistedErrCode = 4027
	// StateExistedErrMsg -
	StateExistedErrMsg = "state cannot be created repeatedly"
	// StateNotExistedErrCode -
	StateNotExistedErrCode = 4026
	// StateNotExistedErrMsg -
	StateNotExistedErrMsg = "state does not exist"
	// StateInstanceNotExistedErrCode -
	StateInstanceNotExistedErrCode = 4028
	// StateInstanceNotExistedErrMsg -
	StateInstanceNotExistedErrMsg = "state instance not existed"
	// DataSystemInternalErrCode -
	DataSystemInternalErrCode = 4030
	// DataSystemInternalErrMsg -
	DataSystemInternalErrMsg = "internal system error"
	// StateInstanceNoLease -
	StateInstanceNoLease = 4025
	// StateInstanceNoLeaseMsg -
	StateInstanceNoLeaseMsg = "maximum number of leases reached"
	// FaaSSchedulerInternalErrCode -
	FaaSSchedulerInternalErrCode = 4029
	// FaaSSchedulerInternalErrMsg -
	FaaSSchedulerInternalErrMsg = "internal system error"
	// InvalidState -
	InvalidState = 4040
	// InvalidStateErrMsg -
	InvalidStateErrMsg = "invalid state, expect not blank"
)

var (
	once     sync.Once
	instance *stateManager
)

type stateManager struct {
	stateMap map[string]*api.InstanceAllocation // key stateKey, value lease
	dsClinet api.KvClient
	muteMap  sync.RWMutex
}

// GetStateManager -
func GetStateManager(dsClient api.KvClient) *stateManager {
	once.Do(func() {
		instance = &stateManager{
			stateMap: make(map[string]*api.InstanceAllocation, defaultStateMapCapacity),
			dsClinet: dsClient,
			muteMap:  sync.RWMutex{},
		}
	})
	return instance
}

func (sm *stateManager) genStateKey(funcKey, stateID string) string {
	return fmt.Sprintf("/sn/state/function/%s/state/%s", funcKey, stateID)
}

func (sm *stateManager) newState(funcKey, stateID, traceID string) error {
	err := sm.getState(funcKey, stateID, traceID)
	if err == nil {
		return api.ErrorInfo{
			Code: StateExistedErrCode,
			Err:  fmt.Errorf(StateExistedErrMsg),
		}
	}

	stateKey := sm.genStateKey(funcKey, stateID)
	param := api.SetParam{
		WriteMode: 1, // WRITE_THROUGH_L2_CACHE
		TTLSecond: 0,
		Existence: 0, // should be 1 (NX) after datasystem issue is resolved
		CacheType: 0,
	}
	runtime.LockOSThread()
	sm.dsClinet.SetTraceID(traceID)
	ret := sm.dsClinet.KVSet(stateKey, []byte(defStateVal), param)
	runtime.UnlockOSThread()
	log.GetLogger().Infof("set state to datasystem %d, %v, traceid(%s)", ret.Code, ret.Err, traceID)
	switch ret.Code {
	case cOK:
		return nil
	case cNotFound:
		return api.ErrorInfo{
			Code: StateExistedErrCode,
			Err:  fmt.Errorf(StateExistedErrMsg),
		}
	default:
		alarmDetail := &alarm.Detail{
			SourceTag: os.Getenv(constants.PodNameEnvKey) + "|" + os.Getenv(constants.PodIPEnvKey) +
				"|" + os.Getenv(constants.ClusterName) + "|" + os.Getenv(constants.HostIPEnvKey),
			OpType: alarm.GenerateAlarmLog,
			Details: fmt.Sprintf("new State failed, datasystem error: %v, stateKey: %s, statefuncKey: %s",
				err, stateID, funcKey),
			StartTimestamp: int(time.Now().Unix()),
			EndTimestamp:   0,
		}
		alarmInfo := &alarm.LogAlarmInfo{
			AlarmID:    "NewStateFailedInDataSystem00001",
			AlarmName:  "NewStateFailedInDataSystem",
			AlarmLevel: alarm.Level2,
		}
		alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
		return api.ErrorInfo{
			Code: DataSystemInternalErrCode,
			Err:  fmt.Errorf(DataSystemInternalErrMsg),
		}
	}
}

func (sm *stateManager) getState(funcKey, stateID, traceID string) error {
	stateKey := sm.genStateKey(funcKey, stateID)
	runtime.LockOSThread()
	sm.dsClinet.SetTraceID(traceID)
	_, ret := sm.dsClinet.KVGet(stateKey)
	runtime.UnlockOSThread()
	log.GetLogger().Infof("get state from datasystem %d, %v, traceid(%s)", ret.Code, ret.Err, traceID)
	switch ret.Code {
	case cOK:
		return nil
	case cNotFound:
		return api.ErrorInfo{
			Code: StateNotExistedErrCode,
			Err:  fmt.Errorf(StateNotExistedErrMsg),
		}
	default:
		return api.ErrorInfo{
			Code: DataSystemInternalErrCode,
			Err:  fmt.Errorf(DataSystemInternalErrMsg),
		}
	}
}

func (sm *stateManager) delState(funcKey, stateID, traceID string) error {
	err := sm.getState(funcKey, stateID, traceID)
	if err != nil {
		return api.ErrorInfo{
			Code: StateNotExistedErrCode,
			Err:  fmt.Errorf(StateNotExistedErrMsg),
		}
	}
	stateKey := sm.genStateKey(funcKey, stateID)
	runtime.LockOSThread()
	sm.dsClinet.SetTraceID(traceID)
	ret := sm.dsClinet.KVDel(stateKey)
	runtime.UnlockOSThread()
	log.GetLogger().Infof("del state from datasystem %d, %v, traceid(%s)", ret.Code, ret.Err, traceID)
	switch ret.Code {
	case cOK:
		return nil
	default:
		alarmDetail := &alarm.Detail{
			SourceTag: os.Getenv(constants.PodNameEnvKey) + "|" + os.Getenv(constants.PodIPEnvKey) +
				"|" + os.Getenv(constants.ClusterName) + "|" + os.Getenv(constants.HostIPEnvKey),
			OpType: alarm.GenerateAlarmLog,
			Details: fmt.Sprintf("terminate State failed, datasystem error: %v, stateKey: %s, statefuncKey: %s",
				err, stateID, funcKey),
			StartTimestamp: int(time.Now().Unix()),
			EndTimestamp:   0,
		}
		alarmInfo := &alarm.LogAlarmInfo{
			AlarmID:    "TerminateStateFailedInDataSystem00001",
			AlarmName:  "TerminateStateFailedInDataSystem",
			AlarmLevel: alarm.Level2,
		}
		alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
		return api.ErrorInfo{
			Code: DataSystemInternalErrCode,
			Err:  fmt.Errorf(DataSystemInternalErrMsg),
		}
	}
}

func (sm *stateManager) getInstance(funcKey, stateID string) *api.InstanceAllocation {
	sm.muteMap.RLock()
	defer sm.muteMap.RUnlock()
	stateKey := sm.genStateKey(funcKey, stateID)
	return sm.stateMap[stateKey]
}

func (sm *stateManager) addInstance(funcKey, stateID string, lease *api.InstanceAllocation) {
	sm.muteMap.Lock()
	defer sm.muteMap.Unlock()
	stateKey := sm.genStateKey(funcKey, stateID)
	sm.stateMap[stateKey] = lease
}

func (sm *stateManager) delInstance(funcKey, stateID string) {
	sm.muteMap.Lock()
	defer sm.muteMap.Unlock()
	stateKey := sm.genStateKey(funcKey, stateID)
	delete(sm.stateMap, stateKey)
}
