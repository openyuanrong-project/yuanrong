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

// Package concurrencyscheduler -
package concurrencyscheduler

import (
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/queue"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	checkScalingInterval = 50 * time.Millisecond
)

// ReservedConcurrencyScheduler will schedule instance according to concurrency usage for reserved instance
type ReservedConcurrencyScheduler struct {
	basicConcurrencyScheduler
	instanceScaler      scaler.InstanceScaler
	checkScalingTimeout time.Duration
	createErr           error
}

// NewReservedConcurrencyScheduler creates ReservedConcurrencyScheduler
func NewReservedConcurrencyScheduler(funcSpec *types.FunctionSpecification, resKey resspeckey.ResSpecKey,
	requestTimeout time.Duration, insThdReqQueue *requestqueue.InsAcqReqQueue) scheduler.InstanceScheduler {
	instanceQueue := queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance)
	otherQueue := queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance)
	reservedConcurrencyScheduler := &ReservedConcurrencyScheduler{
		basicConcurrencyScheduler: newBasicConcurrencyScheduler(funcSpec, resKey, types.InstanceTypeReserved,
			instanceQueue, otherQueue),
		checkScalingTimeout: requestTimeout,
	}
	reservedConcurrencyScheduler.insAcqReqQueue = insThdReqQueue
	insThdReqQueue.RegisterSchFunc("reserveScheduleFunc", reservedConcurrencyScheduler.scheduleRequest)
	reservedConcurrencyScheduler.addObservers(scheduler.AvailInsThdTopic, func(data interface{}) {
		availInsThdDiff, ok := data.(int)
		// schedule request even if availInsThdDiff == 0, because some instance thread changes like session doesn't
		// affect availInsThdDiff
		if !ok || availInsThdDiff < 0 {
			return
		}
		insThdReqQueue.ScheduleRequest("reserveScheduleFunc")
	})
	log.GetLogger().Infof("succeed to create ReservedConcurrencyScheduler for function %s isFuncOwner %t",
		reservedConcurrencyScheduler.funcKeyWithRes, reservedConcurrencyScheduler.isFuncOwner)
	return reservedConcurrencyScheduler
}

// AcquireInstance acquires an instance chosen by instanceScheduler
func (rcs *ReservedConcurrencyScheduler) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, error) {
	var (
		insAlloc   *types.InstanceAllocation
		acquireErr error
		recordErr  error
	)
	insAlloc, acquireErr = rcs.basicConcurrencyScheduler.AcquireInstance(insAcqReq)
	if acquireErr != nil && acquireErr != scheduler.ErrNoInsAvailable {
		return nil, acquireErr
	}
	if acquireErr == scheduler.ErrNoInsAvailable && rcs.shouldTriggerColdStart() {
		// 这里如果是静态函数，则会触发到wisecloudscaler，触发一次nuwa cold start，如果不是静态函数，则会走到replicascaler，没有其他影响
		rcs.publishInsThdEvent(scheduler.TriggerScaleTopic, nil)
		if len(insAcqReq.DesignateInstanceID) != 0 {
			pendingRequest := &requestqueue.PendingInsAcqReq{
				CreatedTime: time.Now(),
				ResultChan:  make(chan *requestqueue.PendingInsAcqRsp, 1),
				InsAcqReq:   insAcqReq,
			}
			if err := rcs.insAcqReqQueue.AddRequest(pendingRequest); err != nil {
				return nil, err
			}
			insAcqRsp := <-pendingRequest.ResultChan
			return insAcqRsp.InsAlloc, insAcqRsp.Error
		}
		if config.GlobalConfig.DisableReplicaScaler {
			return nil, acquireErr
		}
		// 没租约可用且预留实例个数已经全部拉起来了，此时应该立即返回错误
		if rcs.GetInstanceNumber(true) >= rcs.instanceScaler.GetExpectInstanceNumber() {
			return nil, acquireErr
		}
		// block until instanceScaler finishes scaling reserved instance
		var createErr error
		ticker := time.NewTicker(checkScalingInterval)
		timer := time.NewTimer(rcs.checkScalingTimeout)
		defer ticker.Stop()
		defer timer.Stop()
	loop:
		for {
			select {
			case <-ticker.C:
				if insAlloc, acquireErr = rcs.basicConcurrencyScheduler.AcquireInstance(insAcqReq); insAlloc != nil {
					return insAlloc, nil
				}
				// 扩容出来的实例的租约瞬间获取完了，但仍有部分请求未获取到租约，则直接返回去扩弹性实例
				if rcs.GetInstanceNumber(true) == rcs.instanceScaler.GetExpectInstanceNumber() {
					return nil, scheduler.ErrNoInsAvailable
				}
				rcs.RLock()
				createErr = rcs.createErr
				rcs.RUnlock()
				if createErr != nil {
					recordErr = createErr
				}
				// 用户错误直接返回
				if createSnErr, ok := createErr.(snerror.SNError); ok {
					if snerror.IsUserError(createSnErr) {
						return nil, rcs.createErr
					}
				}
			case <-timer.C:
				break loop
			}
		}
		// 超时后若等待过程中出现过错误则返回此错误
		if recordErr != nil {
			return nil, recordErr
		}
		return nil, scheduler.ErrNoInsAvailable
	}
	return insAlloc, acquireErr
}

// PopInstance pops an instance, will block and wait for instance which is creating
func (rcs *ReservedConcurrencyScheduler) PopInstance(force bool) *types.Instance {
	wait := rcs.instanceScaler.CheckScaling()
	if force {
		wait = false
	}
	insElem := rcs.popInstanceElement(forward, nil, wait)
	if insElem == nil {
		return nil
	}
	return insElem.instance
}

// ConnectWithInstanceScaler connects instanceScheduler with an instanceScaler
func (rcs *ReservedConcurrencyScheduler) ConnectWithInstanceScaler(instanceScaler scaler.InstanceScaler) {
	rcs.addObservers(scheduler.TriggerScaleTopic, func(data interface{}) {
		instanceScaler.TriggerScale()
	})
	// check if instanceScaler is a concurrencyInstanceScaler type in future and return error if otherwise
	rcs.instanceScaler = instanceScaler
	rcs.addObservers(scheduler.InUseInsThdTopic, func(data interface{}) {
		inUsedInsThdDiff, ok := data.(int)
		if !ok {
			return
		}
		instanceScaler.HandleInsThdUpdate(inUsedInsThdDiff, 0)
	})
	rcs.addObservers(scheduler.TotalInsThdTopic, func(data interface{}) {
		totalInsThdDiff, ok := data.(int)
		if !ok {
			return
		}
		instanceScaler.HandleInsThdUpdate(0, totalInsThdDiff)
		rcs.insAcqReqQueue.HandleInsNumUpdate(totalInsThdDiff / rcs.concurrentNum)
	})
	return
}

// HandleFuncSpecUpdate handles funcSpec update comes from ETCD
func (rcs *ReservedConcurrencyScheduler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	rcs.Lock()
	rcs.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	rcs.basicConcurrencyScheduler.HandleFuncSpecUpdate(funcSpec)
	rcs.Unlock()
}

// HandleCreateError handles create error
func (rcs *ReservedConcurrencyScheduler) HandleCreateError(createErr error) {
	rcs.Lock()
	rcs.createErr = createErr
	rcs.Unlock()
}

// priorityFuncForReservedInstance is the priority function for reserved instance which will average requests among all
// instances since they will not be scaled down
func priorityFuncForReservedInstance(obj interface{}) (int, error) {
	insElem, ok := obj.(*instanceElement)
	if ok {
		weight := len(insElem.threadMap)
		if !insElem.isNewInstance {
			weight -= insElem.instance.ConcurrentNum
		}
		return weight, nil
	}
	return -1, scheduler.ErrTypeConvertFail
}
