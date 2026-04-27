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

// Package signalmanager -
package signalmanager

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/aliasroute"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// signalInstance -
type signalInstance struct {
	*types.Instance
	signalProcessors map[int]*signalProcessor
	Logger           api.FormatLogger
}

// PrepareSchedulerArg -
func PrepareSchedulerArg() ([]byte, error) {
	schedulerInfo := registry.GlobalRegistry.FaaSSchedulerRegistry.GetAllSchedulerInfo()
	schedulerData, err := json.Marshal(schedulerInfo)
	if err != nil {
		return nil, err
	}
	return schedulerData, nil
}

type getDataFunc func() ([]byte, error)

type signalProcessor struct {
	InstanceId string
	HasSignal  bool
	IsRunning  bool
	StopChan   chan struct{}
	SignalNo   int
	TenantId   string
	getDataFunc
	sync.RWMutex
	Logger   api.FormatLogger
	killFunc func(instanceID string, payload []byte) error
}

// manager -
type manager struct {
	instances map[string]*signalInstance
	lock      sync.RWMutex
	Logger    api.FormatLogger
	killFunc  func(instanceID string, signal int, payload []byte) error
}

var signalManager *manager
var once sync.Once

// GetSignalManager -
func GetSignalManager() *manager {
	once.Do(
		func() {
			signalManager = &manager{
				instances: make(map[string]*signalInstance),
				Logger:    log.GetLogger(),
			}
		})
	return signalManager
}

// SetKillFunc -
func (sm *manager) SetKillFunc(killFunc func(instanceID string, signal int, payload []byte) error) {
	sm.killFunc = killFunc
}

// SignalInstance -
func (sm *manager) SignalInstance(instance *types.Instance, signalNo int) {
	if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG {
		return
	}
	needProcessSignal := map[int]struct{}{
		constant.KillSignalAliasUpdate:         {},
		constant.KillSignalFaaSSchedulerUpdate: {},
	}

	if _, ok := needProcessSignal[signalNo]; !ok {
		sm.Logger.Warnf("no need process this signalNo: %d, instanceId: %sm", signalNo, instance.InstanceID)
		return
	}
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if sm.killFunc == nil {
		sm.Logger.Errorf("killFunc not set")
		return
	}
	sInstance, ok := sm.instances[instance.InstanceID]
	if !ok {
		tenantId := urnutils.GetTenantFromFuncKey(instance.FuncKey)
		if tenantId == "" {
			sm.Logger.Errorf("instance: tenantId parse failed, funcKey: %s, instanceId is %s",
				instance.FuncKey, instance.InstanceID)
			return
		}

		sInstance = &signalInstance{
			Instance: instance,
			Logger: log.GetLogger().With(zap.Any("funcKey", instance.FuncKey),
				zap.Any("instanceId", instance.InstanceID), zap.Any("tenantId", tenantId)),
			signalProcessors: make(map[int]*signalProcessor, 2),
		}
		sInstance.signalProcessors[constant.KillSignalAliasUpdate] = &signalProcessor{
			InstanceId: instance.InstanceID,
			StopChan:   make(chan struct{}),
			SignalNo:   constant.KillSignalAliasUpdate,
			TenantId:   tenantId,
			getDataFunc: func() ([]byte, error) {
				return aliasroute.MarshalTenantAliasList(tenantId)
			},
			killFunc: func(instanceID string, payload []byte) error {
				return sm.killFunc(instanceID, constant.KillSignalAliasUpdate, payload)
			},
			RWMutex: sync.RWMutex{},
			Logger: sInstance.Logger.With(zap.Any("signal", constant.KillSignalAliasUpdate),
				zap.Any("update alias", "")),
		}
		sInstance.signalProcessors[constant.KillSignalFaaSSchedulerUpdate] = &signalProcessor{
			InstanceId:  instance.InstanceID,
			StopChan:    make(chan struct{}),
			SignalNo:    constant.KillSignalFaaSSchedulerUpdate,
			getDataFunc: PrepareSchedulerArg,
			killFunc: func(instanceID string, payload []byte) error {
				return sm.killFunc(instanceID, constant.KillSignalFaaSSchedulerUpdate, payload)
			},
			RWMutex: sync.RWMutex{},
			Logger: sInstance.Logger.With(zap.Any("signal", constant.KillSignalFaaSSchedulerUpdate),
				zap.Any("update faasscheduler", "")),
		}
		sm.instances[instance.InstanceID] = sInstance
	}

	processor, ok := sInstance.signalProcessors[signalNo]
	if !ok {
		sInstance.Logger.Warnf("abnormal!, no signalNo: %d in processors", signalNo) // 通常不会走到这里
		return
	}
	processor.Lock()
	defer processor.Unlock()
	processor.HasSignal = true
	if !processor.IsRunning {
		processor.IsRunning = true
		go processor.signalInstance(uuid.New().String())
	}
}

// RemoveInstance -
func (sm *manager) RemoveInstance(instanceId string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sInstance, ok := sm.instances[instanceId]
	if !ok {
		return
	}
	for _, p := range sInstance.signalProcessors {
		utils.SafeCloseChannel(p.StopChan)
	}
	sm.Logger.Infof("remove instance: %s", instanceId)
	delete(sm.instances, instanceId)
}

func (si *signalProcessor) signalInstance(randomId string) {
	logger := si.Logger.With(zap.Any("uuid", randomId))
	isRetry := false
	retryInterval := 100 * time.Millisecond // 间隔时间初始值
	logger.Infof("begin signal instance")
	defer logger.Infof("signal instance over")
	for {
		si.Lock()
		if !si.HasSignal && !isRetry {
			si.IsRunning = false
			si.Unlock()
			return
		}
		si.HasSignal = false
		si.Unlock()

		select {
		case <-si.StopChan:
			si.Lock()
			si.IsRunning = false
			si.HasSignal = false
			si.Unlock()
			logger.Infof("instance removed, exit signalProcessor")
			return
		default:
			data, err := si.getDataFunc()
			if err != nil {
				logger.Errorf("get data for signal instance failed, err: %s", err.Error())
				isRetry = false
				continue
			}
			if err := si.killFunc(si.InstanceId, data); err != nil {
				logger.Errorf("failed to signal instance, error:%s", err.Error())
				// instance not found, the instance may have been killed
				if strings.Contains(err.Error(), "instance not found") {
					isRetry = false
					GetSignalManager().RemoveInstance(si.InstanceId)
					continue
				}
				time.Sleep(retryInterval)
				retryInterval *= 2                  // 翻倍
				if retryInterval >= 5*time.Minute { // 间隔时间最大值
					retryInterval = 5 * time.Minute // 间隔时间最大值
				}
				isRetry = true
			} else {
				retryInterval = 100 * time.Millisecond // 间隔时间初始值
				isRetry = false
			}
		}

	}
}
