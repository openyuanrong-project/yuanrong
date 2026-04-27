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

// Package microservicescheduler -
package microservicescheduler

import (
	"fmt"
	"math"
	"sync"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	randomThreadIDLength = 4
	// LeastConnections -
	LeastConnections = "LC"
	// RoundRobin -
	RoundRobin = "RR"
)

type instanceObserver struct {
	callback func(interface{})
}

type instanceElement struct {
	instance     *types.Instance
	requestCount int
}

// MicroServiceScheduler will schedule instance according to concurrency usage for reserved instance
type MicroServiceScheduler struct {
	instanceScaler scaler.InstanceScaler
	instanceQueue  []*instanceElement
	funcKeyWithRes string
	SchedulePolicy string
	curIndex       int
	sync.RWMutex
}

// NewMicroServiceScheduler creates MicroServiceScheduler
func NewMicroServiceScheduler(funcKeyWithRes string, schedulePolicy string) scheduler.InstanceScheduler {
	microServiceScheduler := &MicroServiceScheduler{
		funcKeyWithRes: funcKeyWithRes,
		instanceQueue:  make([]*instanceElement, 0, utils.DefaultSliceSize),
		SchedulePolicy: schedulePolicy,
		curIndex:       0,
	}
	return microServiceScheduler
}

// GetInstanceNumber gets instance number inside instance queue
func (ms *MicroServiceScheduler) GetInstanceNumber(onlySelf bool) int {
	ms.RLock()
	insNum := len(ms.instanceQueue)
	ms.RUnlock()
	return insNum
}

// CheckInstanceExist checks if instance exist
func (ms *MicroServiceScheduler) CheckInstanceExist(instance *types.Instance) bool {
	ms.RLock()
	exist := ms.getInstance(instance.InstanceID) != nil
	ms.RUnlock()
	return exist
}

// AcquireInstance acquires an instance chosen by instanceScheduler
func (ms *MicroServiceScheduler) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation,
	error) {
	ms.Lock()
	defer ms.Unlock()
	var (
		instance *types.Instance
		err      error
	)
	if len(ms.instanceQueue) == 0 {
		return nil, scheduler.ErrNoInsAvailable
	}
	switch ms.SchedulePolicy {
	case RoundRobin:
		instance, err = ms.selectInstanceWithRR()
	case LeastConnections:
		instance, err = ms.selectInstanceWithLC()
	default:
		instance, err = ms.selectInstanceWithRR()
	}
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, scheduler.ErrNoInsAvailable
	}
	return &types.InstanceAllocation{
		Instance:     instance,
		AllocationID: fmt.Sprintf("%s-%s", instance.InstanceID, utils.GenRandomString(randomThreadIDLength)),
	}, nil
}

// ReleaseInstance releases an instance to instanceScheduler
func (ms *MicroServiceScheduler) ReleaseInstance(thread *types.InstanceAllocation) error {
	ms.Lock()
	defer ms.Unlock()
	if ms.SchedulePolicy == LeastConnections {
		for _, element := range ms.instanceQueue {
			if element.instance.InstanceID == thread.Instance.InstanceID {
				if element.requestCount > 0 {
					element.requestCount--
				}
				break
			}
		}
	}
	return nil
}

// AddInstance adds an instance to instanceScheduler
func (ms *MicroServiceScheduler) AddInstance(instance *types.Instance) error {
	ms.Lock()
	defer ms.Unlock()
	if getInstance := ms.getInstance(instance.InstanceID); getInstance != nil {
		return scheduler.ErrInsAlreadyExist
	}
	switch instance.InstanceStatus.Code {
	case int32(constant.KernelInstanceStatusRunning):
		ms.addInstance(instance)
	default:
		log.GetLogger().Warnf("ignore unexpected instance %s with status code %d", instance.InstanceID,
			instance.InstanceStatus.Code)
		return scheduler.ErrInternal
	}
	return nil
}

// PopInstance pops an instance from instanceScheduler
func (ms *MicroServiceScheduler) PopInstance(force bool) *types.Instance {
	ms.Lock()
	defer ms.Unlock()
	var instance *types.Instance
	if instance = ms.popInstance(); instance == nil {
		return nil
	}
	return instance
}

// DelInstance deletes an instance from instanceScheduler
func (ms *MicroServiceScheduler) DelInstance(instance *types.Instance) error {
	ms.Lock()
	defer ms.Unlock()
	var (
		// concurrentNum may be updated and this instance could be an old one, so we get inQueIns from internal
		inQueIns *types.Instance
	)
	if inQueIns = ms.getInstance(instance.InstanceID); inQueIns == nil {
		return scheduler.ErrInsNotExist
	}
	ms.delInstance(instance.InstanceID)
	return nil
}

// ConnectWithInstanceScaler connects instanceScheduler with an instanceScaler, currently connects with replicaScaler
func (ms *MicroServiceScheduler) ConnectWithInstanceScaler(instanceScaler scaler.InstanceScaler) {
	return
}

// HandleFuncSpecUpdate handles funcSpec update comes from ETCD
func (ms *MicroServiceScheduler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
}

// HandleInstanceUpdate handles instance update comes from ETCD
func (ms *MicroServiceScheduler) HandleInstanceUpdate(instance *types.Instance) {
	logger := log.GetLogger().With(zap.Any("funcKey", ms.funcKeyWithRes), zap.Any("instance", instance.InstanceID),
		zap.Any("instanceStatus", instance.InstanceStatus.Code))
	ms.RLock()
	exist := ms.getInstance(instance.InstanceID) != nil
	ms.RUnlock()
	switch instance.InstanceStatus.Code {
	case int32(constant.KernelInstanceStatusRunning):
		if exist {
			logger.Warnf("no need to add instance repeatedly")
		} else {
			logger.Infof("add new instance from instance update")
			if err := ms.AddInstance(instance); err != nil {
				logger.Errorf("failed to add instance error %s", err.Error())
			}
		}
	case int32(constant.KernelInstanceStatusEvicting):
		if exist {
			err := ms.DelInstance(instance)
			if err != nil {
				logger.Errorf("failed to delete evicting instance %s", err.Error())
			}
		}
	default:
	}
}

// HandleCreateError handles create error
func (ms *MicroServiceScheduler) HandleCreateError(createErr error) {
}

// SignalAllInstances sends signal to all instances
func (ms *MicroServiceScheduler) SignalAllInstances(signalFunc scheduler.SignalInstanceFunc) {
}

// Destroy destroys instanceScheduler
func (ms *MicroServiceScheduler) Destroy() {
}

func (ms *MicroServiceScheduler) getInstance(targetID string) *types.Instance {
	for _, element := range ms.instanceQueue {
		if element.instance.InstanceID == targetID {
			return element.instance
		}
	}
	return nil
}

func (ms *MicroServiceScheduler) addInstance(instance *types.Instance) {
	ms.instanceQueue = append(ms.instanceQueue, &instanceElement{instance: instance})
}

func (ms *MicroServiceScheduler) delInstance(targetID string) {
	targetIndex := -1
	for index, element := range ms.instanceQueue {
		if element.instance.InstanceID == targetID {
			targetIndex = index
			break
		}
	}
	if targetIndex == -1 {
		return
	}
	ms.instanceQueue = append(ms.instanceQueue[0:targetIndex], ms.instanceQueue[targetIndex+1:]...)
}

func (ms *MicroServiceScheduler) popInstance() *types.Instance {
	queLen := len(ms.instanceQueue)
	if queLen == 0 {
		return nil
	}
	element := ms.instanceQueue[queLen-1]
	ms.instanceQueue = ms.instanceQueue[:queLen-1]
	return element.instance
}

func (ms *MicroServiceScheduler) selectInstanceWithLC() (*types.Instance, error) {
	var (
		chosenIns    *instanceElement
		chosenReqNum = math.MaxInt32
	)
	for _, element := range ms.instanceQueue {
		if element.requestCount < chosenReqNum {
			chosenReqNum = element.requestCount
			chosenIns = element
		}
	}
	if chosenIns != nil {
		chosenIns.requestCount++
		return chosenIns.instance, nil
	}
	return nil, scheduler.ErrNoInsAvailable
}

func (ms *MicroServiceScheduler) selectInstanceWithRR() (*types.Instance, error) {
	if ms.curIndex >= len(ms.instanceQueue) {
		ms.curIndex = ms.curIndex % len(ms.instanceQueue)
	}
	element := ms.instanceQueue[ms.curIndex]
	ms.curIndex++
	return element.instance, nil
}

// HandleFuncOwnerUpdate -
func (ms *MicroServiceScheduler) HandleFuncOwnerUpdate(isFuncOwner bool) {
}

// ReassignInstanceWhenGray -
func (ms *MicroServiceScheduler) ReassignInstanceWhenGray(ratio int) {
	return
}
