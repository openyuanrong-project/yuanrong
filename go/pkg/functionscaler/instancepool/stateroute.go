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
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/stateinstance"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// StateInstance -
type StateInstance struct {
	StateID  string
	status   int // -1 init, 0 ok, 1 instance in sub health, 2 instance exit
	instance *types.Instance
}

const (
	// InstanceExit -
	InstanceExit = 2
	// InstanceAbnormal -
	InstanceAbnormal = 1
	// InstanceOk -
	InstanceOk = 0
)

// StateRoute -
type StateRoute struct {
	funcSpec           *types.FunctionSpecification
	stateRoute         map[string]*StateInstance
	stateLeaseManager  map[string]*stateinstance.Leaser
	stateConfig        commonTypes.StateConfig
	resSpec            *resspeckey.ResourceSpecification
	deleteInstanceFunc func(instance *types.Instance) error
	createInstanceFunc func(resSpec *resspeckey.ResourceSpecification, instanceType types.InstanceType,
		callerPodName string) (instance *types.Instance, err error)
	leaseInterval time.Duration
	logger        api.FormatLogger
	sync.RWMutex
	stateLocks sync.Map // key: stateID val: *sync.RWMutex
}

// Destroy -
func (sr *StateRoute) Destroy() {
	sr.Lock()
	defer sr.Unlock()
	instances := make([]string, 0)
	for _, v := range sr.stateRoute {

		if v.status != InstanceExit {
			instances = append(instances, fmt.Sprintf("%s:%d", v.instance.InstanceID, v.status))
			go func() {
				err := sr.deleteInstanceFunc(v.instance)
				if err != nil {
					sr.logger.Infof("delete instance:%s failed: %s", v.instance.InstanceID, err.Error())
				}
			}()
		}
	}
	sr.stateRoute = make(map[string]*StateInstance)
	sr.logger.Infof("destroy, instances: %s", strings.Join(instances, ","))
}

// HandleInstanceUpdate - handle etcd instance event
func (sr *StateRoute) HandleInstanceUpdate(instance *types.Instance) {
	if instance == nil || instance.InstanceType != types.InstanceTypeState {
		return
	}
	sr.Lock()

	// 通常认为实例已经添加进来了，仅考虑实例更新场景
	for stateID, stateInstance := range sr.stateRoute {
		if stateInstance.instance != nil && stateInstance.instance.InstanceID == instance.InstanceID {
			stateInstance.instance = instance
			oldStatus := stateInstance.status
			switch instance.InstanceStatus.Code {
			case int32(constant.KernelInstanceStatusSubHealth):
				stateInstance.status = InstanceAbnormal
			case int32(constant.KernelInstanceStatusRunning):
				stateInstance.status = InstanceOk
			default:
				stateInstance.status = InstanceAbnormal
			}
			sr.Unlock()
			sr.processStateRoute(stateInstance, stateID, stateUpdate)
			sr.logger.Infof("statusInstance:%s status %d->%d", instance.InstanceID, oldStatus, stateInstance.status)
			return
		}
	}
	sr.Unlock()
	sr.logger.Infof("not found stateInstance: %s, instanceStatus: %d", instance.InstanceID,
		instance.InstanceStatus.Code)
}

// GetAndDeleteState delete state and instance, return whether the state exists
func (sr *StateRoute) GetAndDeleteState(stateID string) bool {
	value, _ := sr.stateLocks.LoadOrStore(stateID, &sync.RWMutex{})
	stateLock, ok := value.(*sync.RWMutex)
	if !ok {
		sr.logger.Errorf("get statelock err, value is %v", value)
		return false
	}
	stateLock.Lock()
	stateInstance, exist := sr.stateRoute[stateID]
	sr.logger.Infof("stateRoute of %s exist is %v, stateInstance is %v", stateID, exist, stateInstance)
	if exist {
		sr.processStateRoute(stateInstance, stateID, stateDelete)
	}
	stateLock.Unlock()

	if stateInstance != nil && stateInstance.status != InstanceExit {
		err := sr.deleteInstanceFunc(stateInstance.instance)
		if err != nil {
			sr.logger.Errorf("failed to delete stateInstance %s error %s, instance status: %d",
				stateInstance.instance.InstanceID, err.Error(), stateInstance.status)
		}
	}
	return exist
}

// DeleteStateInstance called by ReleaseLease
func (sr *StateRoute) DeleteStateInstance(stateID string, instanceID string) {
	value, _ := sr.stateLocks.LoadOrStore(stateID, &sync.RWMutex{})
	stateLock, ok := value.(*sync.RWMutex)
	if !ok {
		sr.logger.Errorf("get statelock err, value is %v", value)
		return
	}
	stateLock.Lock()
	stateInstance, exist := sr.getStateInstance(stateID)
	if exist {
		if stateInstance.instance.InstanceID == instanceID {
			sr.processStateRoute(stateInstance, stateID, stateInstanceDelete)
		} else {
			sr.logger.Warnf("stateInstance is not matched %s:%s, stateID: %s",
				stateInstance.instance.InstanceID, instanceID, stateInstance.StateID)
		}

	}
	stateLock.Unlock()
	if stateInstance != nil {
		if stateInstance.instance.InstanceID != instanceID {
			sr.logger.Warnf("instanceID in lease manager and in state route are different: %s, %s, stateID: %s",
				instanceID, stateInstance.instance.InstanceID, stateID)
		} else {
			err := sr.deleteInstanceFunc(stateInstance.instance)
			if err != nil {
				sr.logger.Errorf("failed to delete stateInstance %s error %s", stateInstance.instance.InstanceID,
					err.Error())
			} else {
				sr.logger.Infof("DeleteStateInstance state %s, instance %s over", stateID, instanceID)
			}
		}
	}
}
func (sr *StateRoute) getStateInstance(stateID string) (*StateInstance, bool) {
	sr.RLock()
	defer sr.RUnlock()
	stateInstance, exist := sr.stateRoute[stateID]
	return stateInstance, exist
}

func (sr *StateRoute) setStateInstance(stateID string, stateInstance *StateInstance) {
	sr.Lock()
	defer sr.Unlock()
	sr.stateRoute[stateID] = stateInstance
}

func (sr *StateRoute) deleteStateInstance(stateID string) {
	sr.Lock()
	defer sr.Unlock()
	delete(sr.stateRoute, stateID)
}

func (sr *StateRoute) getStateIDByInstanceID(instanceID string) (string, bool) {
	sr.RLock()
	defer sr.RUnlock()
	for stateID, stateInstance := range sr.stateRoute {
		if stateInstance.instance != nil && stateInstance.instance.InstanceID == instanceID {
			return stateID, true
		}
	}
	return "", false
}

// DeleteStateInstanceByInstanceID  delete state route data by instance when instance is deleted
func (sr *StateRoute) DeleteStateInstanceByInstanceID(instanceID string) {
	stateID, exist := sr.getStateIDByInstanceID(instanceID)
	if !exist {
		sr.logger.Warnf("delete state route failed because instance %s is not in stateroute!", instanceID)
		return
	}
	stateInstance, exist := sr.getStateInstance(stateID)
	if !exist {
		sr.logger.Warnf("delete state route failed because stateinstance %s is not in stateroute!, stateID: %s",
			instanceID, stateID)
		return
	}
	sr.processStateRoute(stateInstance, stateID, stateInstanceDelete)
}

func (sr *StateRoute) processStateRoute(stateInstance *StateInstance, stateID string, opType string) {
	instancePoolStateInput := &types.InstancePoolStateInput{
		StateID: stateID,
	}
	if stateInstance.instance != nil {
		instancePoolStateInput.FuncKey = stateInstance.instance.FuncKey
		instancePoolStateInput.FuncSig = stateInstance.instance.FuncSig
		instancePoolStateInput.InstanceType = stateInstance.instance.InstanceType
		instancePoolStateInput.InstanceID = stateInstance.instance.InstanceID
		instancePoolStateInput.InstanceStatusCode = stateInstance.instance.InstanceStatus.Code

	}
	stateType := types.StateUpdate

	if opType == stateUpdate {
		sr.setStateInstance(stateID, stateInstance)
		stateType = types.StateUpdate
	} else if opType == stateDelete {
		sr.deleteStateInstance(stateID)
		sr.stateLocks.Delete(stateID)
		stateType = types.StateDelete
	} else if opType == stateInstanceDelete {
		if sr.stateConfig.LifeCycle == types.InstanceLifeCycleConsistentWithState {
			stateType = types.StateUpdate
			newInstance := *stateInstance
			newInstance.status = InstanceExit
			newInstance.instance = nil
			sr.setStateInstance(stateID, &newInstance)
			instancePoolStateInput.InstanceStatusCode = int32(constant.KernelInstanceStatusExited)
			stateType = types.StateUpdate
		} else {
			sr.deleteStateInstance(stateID)
			stateType = types.StateDelete
		}
	} else {
		sr.logger.Warnf("unknown stateType: %s", opType)
	}

	sr.logger.Infof("state update: %+v, type %v", instancePoolStateInput, stateType)
	state.Update(instancePoolStateInput, stateType)
}

func (sr *StateRoute) recover(instanceMap map[string]*types.Instance) {
	for stateID, instance := range instanceMap {
		statusCode := InstanceOk
		switch instance.InstanceStatus.Code {
		case int32(constant.KernelInstanceStatusExited):
			statusCode = InstanceExit
		case int32(constant.KernelInstanceStatusSubHealth):
			statusCode = InstanceAbnormal
		case int32(constant.KernelInstanceStatusRunning):
			statusCode = InstanceOk
		default:
			statusCode = InstanceAbnormal
		}
		sr.setStateInstance(stateID, &StateInstance{
			StateID:  stateID,
			status:   statusCode,
			instance: instance,
		})
		sr.logger.Infof("recover stateID:%s statusCode: %d, instanceID: %s, instanceStatusCode: %d",
			stateID, statusCode, instance.InstanceID, instance.InstanceStatus.Code)
	}
	sr.logger.Infof("recover stateRoute over")
}

func (sr *StateRoute) acquireStateInstanceThread(insThdApp *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, snerror.SNError) {
	logger := sr.logger.With(zap.Any("stateKey", insThdApp.StateID))
	value, _ := sr.stateLocks.LoadOrStore(insThdApp.StateID, &sync.RWMutex{})
	stateLock, ok := value.(*sync.RWMutex)
	if !ok {
		sr.logger.Errorf("get statelock err, value is %v", value)
		return nil, snerror.New(statuscode.FaaSSchedulerInternalErrCode, statuscode.FaaSSchedulerInternalErrMsg)
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	stateInstance, exist := sr.getStateInstance(insThdApp.StateID)

	if exist {
		if stateInstance != nil && stateInstance.status == InstanceOk {
			logger.Infof("state stateInstance existed in stateRoute, instanceID: %s", stateInstance.instance.InstanceID)
			lease, err := sr.generateStateLease(stateInstance)
			if err != nil {
				return nil, err
			}
			return &types.InstanceAllocation{
				Instance:     stateInstance.instance,
				AllocationID: fmt.Sprintf("%s-stateThread%d", stateInstance.instance.InstanceID, lease.ID),
			}, nil
		}
		if stateInstance != nil && stateInstance.status == InstanceAbnormal {
			logger.Infof("state stateInstance existed in stateRoute, but abnormal, instanceID: %s",
				stateInstance.instance.InstanceID)
			return nil, snerror.New(statuscode.InstanceStatusAbnormalCode, statuscode.InstanceStatusAbnormalMsg)
		}
		// The slave function stateInstance is destroyed, no longer repulsed, and the state should be deleted
		if sr.stateConfig.LifeCycle == types.InstanceLifeCycleConsistentWithState {
			logger.Infof("state stateInstance is destroyed and return 4028")
			return nil, snerror.New(statuscode.StateInstanceNotExistedErrCode, statuscode.StateInstanceNotExistedErrMsg)
		}
	}

	var resSpec *resspeckey.ResourceSpecification
	if utils.IsResSpecEmpty(insThdApp.ResSpec) {
		resSpec = sr.resSpec
	} else {
		resSpec = insThdApp.ResSpec
	}

	instance, err := sr.createInstanceFunc(resSpec, types.InstanceTypeState, insThdApp.CallerPodName)
	if err == nil {
		stateInstance = &StateInstance{
			StateID:  insThdApp.StateID,
			status:   InstanceOk,
			instance: instance,
		}
		logger.Infof("new statInstance, instanceId: %s", instance.InstanceID)
		sr.processStateRoute(stateInstance, insThdApp.StateID, stateUpdate)
		lease, err := sr.generateStateLease(stateInstance)
		if err != nil {
			return nil, err
		}
		return &types.InstanceAllocation{
			Instance:     instance,
			AllocationID: fmt.Sprintf("%s-stateThread%d", stateInstance.instance.InstanceID, lease.ID),
		}, nil
	}
	return nil, snerror.New(statuscode.NoInstanceAvailableErrCode, err.Error())
}

func (sr *StateRoute) generateStateLease(stateInstance *StateInstance) (*stateinstance.Lease, snerror.SNError) {
	sr.Lock()
	leaser := sr.stateLeaseManager[stateInstance.instance.InstanceID]
	if leaser == nil {
		leaser = stateinstance.NewLeaser(sr.funcSpec.InstanceMetaData.ConcurrentNum,
			sr.DeleteStateInstance, stateInstance.StateID, stateInstance.instance.InstanceID, getScaleDownWindow())
		sr.stateLeaseManager[stateInstance.instance.InstanceID] = leaser
	}
	sr.Unlock()
	lease, err := leaser.AcquireLease(sr.leaseInterval)
	if err != nil {
		log.GetLogger().Errorf("failed to generate state lease for instance %s error %s",
			stateInstance.instance.InstanceID, err.Error())
		if snErr, ok := err.(snerror.SNError); ok {
			return nil, snErr
		}
		return nil, snerror.New(statuscode.StatusInternalServerError, err.Error())
	}
	return lease, nil
}
