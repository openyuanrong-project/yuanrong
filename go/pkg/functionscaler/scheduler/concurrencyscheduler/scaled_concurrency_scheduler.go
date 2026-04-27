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
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

type instanceQueueWithBuffer struct {
	queue  *queue.PriorityQueue
	buffer []*instanceElement
	idFunc func(interface{}) string
}

// Front -
func (p *instanceQueueWithBuffer) Front() interface{} {
	if p.queue.Len() != 0 {
		return p.queue.Front()
	}
	if len(p.buffer) == 0 {
		return nil
	}
	return p.buffer[0]
}

// Back -
func (p *instanceQueueWithBuffer) Back() interface{} {
	if p.queue.Len() != 0 {
		return p.queue.Back()
	}
	length := len(p.buffer)
	if length == 0 {
		return nil
	}
	item := p.buffer[length-1]
	return item
}

// PopFront -
func (p *instanceQueueWithBuffer) PopFront() interface{} {
	if p.queue.Len() != 0 {
		return p.queue.PopFront()
	}
	if len(p.buffer) == 0 {
		return nil
	}
	item := p.buffer[0]
	p.buffer = p.buffer[1:]
	return item
}

// PopBack -
func (p *instanceQueueWithBuffer) PopBack() interface{} {
	if p.queue.Len() != 0 {
		return p.queue.PopBack()
	}
	length := len(p.buffer)
	if length == 0 {
		return nil
	}
	item := p.buffer[length-1]
	p.buffer = p.buffer[:length-1]
	return item
}

// PushBack - 在调用pushBack之前，需确保该obj不在instanceQueueWithBuffer中
func (p *instanceQueueWithBuffer) PushBack(obj interface{}) error {
	insElem, ok := obj.(*instanceElement)
	if !ok {
		return scheduler.ErrTypeConvertFail
	}
	if len(insElem.threadMap) == 0 { // threadMap == 0 实例需要放入buffer中
		p.buffer = append(p.buffer, insElem)
		return nil
	}
	return p.queue.PushBack(obj)
}

// GetByID -
func (p *instanceQueueWithBuffer) GetByID(objID string) interface{} {
	get := p.queue.GetByID(objID)
	if get != nil {
		return get
	}
	for _, item := range p.buffer {
		if p.idFunc(item) == objID {
			return item
		}
	}
	return nil
}

// DelByID -
func (p *instanceQueueWithBuffer) DelByID(objID string) error {
	err := p.queue.DelByID(objID)
	if err == nil {
		return nil
	}
	if err != queue.ErrObjectNotFound {
		return err
	}
	index := -1
	for i, item := range p.buffer {
		if p.idFunc(item) == objID {
			index = i
			break
		}
	}
	if index == -1 {
		return queue.ErrObjectNotFound
	}
	p.buffer = append(p.buffer[0:index], p.buffer[index+1:]...)
	return nil
}

// Range iterates item in queue and process item with given function
func (p *instanceQueueWithBuffer) Range(f func(obj interface{}) bool) bool {
	if !p.queue.Range(f) {
		return false
	}
	for _, item := range p.buffer {
		if !f(item) {
			return false
		}
	}
	return true
}

// SortedRange iterates item in queue and process item with given function in order
func (p *instanceQueueWithBuffer) SortedRange(f func(obj interface{}) bool) bool {
	if !p.queue.SortedRange(f) {
		return false
	}
	for _, item := range p.buffer {
		if !f(item) {
			return false
		}
	}
	return true
}

// UpdateObjByID has some cases which should be aware:
//  1. for instance who will have no available thread after updating, we should move it from queue to buffer
//  2. for instance had no available thread before and will have available thread after updating, we should mov it from
//     buffer to queue
func (p *instanceQueueWithBuffer) UpdateObjByID(objID string, obj interface{}) error {
	insElem, ok := obj.(*instanceElement)
	if !ok {
		return scheduler.ErrTypeConvertFail
	}
	updateErr := p.queue.UpdateObjByID(objID, obj)
	// transfer instance from queue to buffer if there is no available thread in it, transfer instance from buffer to
	// queue if there is available thread in it
	if updateErr == nil && len(insElem.threadMap) == 0 {
		err := p.queue.DelByID(insElem.instance.InstanceID)
		if err != nil {
			log.GetLogger().Errorf("failed to remove instance %s from scaled instance queue error %s",
				insElem.instance.InstanceID, err.Error())
			return err
		}
		p.buffer = append(p.buffer, insElem)
		return nil
	} else if updateErr == queue.ErrObjectNotFound {
		index := -1
		for i, item := range p.buffer {
			if p.idFunc(item) == objID {
				index = i
				p.buffer[index] = insElem
			}
		}
		if index == -1 {
			return queue.ErrObjectNotFound
		}
		if len(insElem.threadMap) != 0 {
			p.buffer = append(p.buffer[0:index], p.buffer[index+1:]...)
			err := p.queue.PushBack(insElem)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return updateErr
}

// Len -
func (p *instanceQueueWithBuffer) Len() int {
	return p.queue.Len() + len(p.buffer)
}

// ScaledConcurrencyScheduler will schedule instance according to concurrency usage for scaled instance
type ScaledConcurrencyScheduler struct {
	basicConcurrencyScheduler
}

// NewScaledConcurrencyScheduler creates ScaledConcurrencyScheduler
func NewScaledConcurrencyScheduler(funcSpec *types.FunctionSpecification, resKey resspeckey.ResSpecKey,
	insThdReqQueue *requestqueue.InsAcqReqQueue) scheduler.InstanceScheduler {
	concurrentNum := utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	instanceQueue := &instanceQueueWithBuffer{
		queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(concurrentNum)),
		buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
		idFunc: getInstanceID,
	}
	otherQueue := &instanceQueueWithBuffer{
		queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(concurrentNum)),
		buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
		idFunc: getInstanceID,
	}
	scaledConcurrencyScheduler := &ScaledConcurrencyScheduler{
		basicConcurrencyScheduler: newBasicConcurrencyScheduler(funcSpec, resKey, types.InstanceTypeScaled,
			instanceQueue, otherQueue),
	}
	scaledConcurrencyScheduler.insAcqReqQueue = insThdReqQueue
	insThdReqQueue.RegisterSchFunc("scaledScheduleFunc", scaledConcurrencyScheduler.scheduleRequest)
	scaledConcurrencyScheduler.addObservers(scheduler.AvailInsThdTopic, func(data interface{}) {
		availInsThdDiff, ok := data.(int)
		// schedule request even if availInsThdDiff == 0, because some instance thread changes like session doesn't
		// affect availInsThdDiff
		if !ok || availInsThdDiff < 0 {
			return
		}
		insThdReqQueue.ScheduleRequest("scaledScheduleFunc")
	})
	scaledConcurrencyScheduler.addObservers(scheduler.CreateErrorTopic, func(data interface{}) {
		if data == nil {
			insThdReqQueue.HandleCreateError(nil)
			return
		}
		createError, ok := data.(error)
		if !ok {
			return
		}
		insThdReqQueue.HandleCreateError(createError)
	})
	log.GetLogger().Infof("succeed to create ScaledConcurrencyScheduler for function %s isFuncOwner %t",
		scaledConcurrencyScheduler.funcKeyWithRes, scaledConcurrencyScheduler.isFuncOwner)
	return scaledConcurrencyScheduler
}

// GetReqQueLen will get instance request queue length
func (scs *ScaledConcurrencyScheduler) GetReqQueLen() int {
	return scs.insAcqReqQueue.Len()
}

// AcquireInstance acquires an instance chosen by instanceScheduler
func (scs *ScaledConcurrencyScheduler) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, error) {
	insAlloc, acquireErr := scs.basicConcurrencyScheduler.AcquireInstance(insAcqReq)
	if acquireErr != nil && acquireErr != scheduler.ErrNoInsAvailable {
		return nil, acquireErr
	}
	if acquireErr == scheduler.ErrNoInsAvailable && !insAcqReq.SkipWaitPending && scs.shouldTriggerColdStart() {
		// if this scheduler is not the funcOwner, only scale in the case of traffic limit of the original funcOwner
		if !scs.IsFuncOwner() && !insAcqReq.TrafficLimited {
			return nil, acquireErr
		}
		pendingRequest := &requestqueue.PendingInsAcqReq{
			CreatedTime: time.Now(),
			ResultChan:  make(chan *requestqueue.PendingInsAcqRsp, 1),
			InsAcqReq:   insAcqReq,
		}
		if err := scs.insAcqReqQueue.AddRequest(pendingRequest); err != nil {
			return nil, err
		}
		scs.publishInsThdEvent(scheduler.TriggerScaleTopic, nil)
		insAcqRsp := <-pendingRequest.ResultChan
		return insAcqRsp.InsAlloc, insAcqRsp.Error
	}
	return insAlloc, acquireErr
}

// PopInstance pops an instance, set force to false may return nil if there is no instance with full concurrency
// available
func (scs *ScaledConcurrencyScheduler) PopInstance(force bool) *types.Instance {
	var insElem *instanceElement
	if force {
		insElem = scs.popInstanceElement(backward, nil, false)
	} else {
		insElem = scs.popInstanceElement(backward, shouldPopScaledInstance, false)
	}

	if insElem == nil {
		return nil
	}
	return insElem.instance
}

// ConnectWithInstanceScaler connects instanceScheduler with an instanceScaler
func (scs *ScaledConcurrencyScheduler) ConnectWithInstanceScaler(instanceScaler scaler.InstanceScaler) {
	// check if instanceScaler is a concurrencyInstanceScaler type in future and return error if otherwise
	scs.addObservers(scheduler.TriggerScaleTopic, func(data interface{}) {
		instanceScaler.TriggerScale()
	})
	scs.addObservers(scheduler.InUseInsThdTopic, func(data interface{}) {
		inUsedInsThdDiff, ok := data.(int)
		if !ok {
			return
		}
		instanceScaler.HandleInsThdUpdate(inUsedInsThdDiff, 0)
	})
	scs.addObservers(scheduler.TotalInsThdTopic, func(data interface{}) {
		totalInsThdDiff, ok := data.(int)
		if !ok {
			return
		}
		instanceScaler.HandleInsThdUpdate(0, totalInsThdDiff)
		scs.insAcqReqQueue.HandleInsNumUpdate(totalInsThdDiff / scs.concurrentNum)
	})
	return
}

// HandleFuncSpecUpdate handles funcSpec update comes from ETCD
func (scs *ScaledConcurrencyScheduler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	scs.Lock()
	concurrentNum := utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	if scs.concurrentNum != concurrentNum {
		scs.concurrentNum = concurrentNum
		scs.selfInstanceQueue = &instanceQueueWithObserver{
			instanceQueueWithSubHealthAndEvictingRecord: instanceQueueWithSubHealthAndEvictingRecord{
				instanceQueue: &instanceQueueWithBuffer{
					queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(scs.concurrentNum)),
					buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
					idFunc: getInstanceID,
				},
				subHealthRecord: make(map[string]*instanceElement, utils.DefaultMapSize),
				evictingRecord:  make(map[string]*instanceElement, utils.DefaultMapSize),
			},
			insAvailThdCount:  make(map[string]int, utils.DefaultMapSize),
			pubAvailTopicFunc: func(data int) { scs.publishInsThdEvent(scheduler.AvailInsThdTopic, data) },
			pubInUseTopicFunc: func(data int) { scs.publishInsThdEvent(scheduler.InUseInsThdTopic, data) },
			pubTotalTopicFunc: func(data int) { scs.publishInsThdEvent(scheduler.TotalInsThdTopic, data) },
		}
		scs.otherInstanceQueue = &instanceQueueWithSubHealthAndEvictingRecord{
			instanceQueue: &instanceQueueWithBuffer{
				queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(scs.concurrentNum)),
				buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
				idFunc: getInstanceID,
			},
			subHealthRecord: make(map[string]*instanceElement, utils.DefaultMapSize),
			evictingRecord:  make(map[string]*instanceElement, utils.DefaultMapSize),
		}
	}
	scs.Unlock()
}

// HandleCreateError handles create error
func (scs *ScaledConcurrencyScheduler) HandleCreateError(createErr error) {
	scs.publishInsThdEvent(scheduler.CreateErrorTopic, createErr)
}

// Destroy destroys instanceScheduler
func (scs *ScaledConcurrencyScheduler) Destroy() {
	scs.basicConcurrencyScheduler.Destroy()
}

// priorityFuncForScaledInstance will create priority function which put instance with less concurrentNum in front. for
// scaled instance, aggregate requests on busy instances will benefit the scaled down process which improves resource
// utilization
func priorityFuncForScaledInstance(concurrency int) func(obj interface{}) (int, error) {
	return func(obj interface{}) (int, error) {
		insElem, ok := obj.(*instanceElement)
		if ok {
			weight := concurrency - len(insElem.threadMap)
			if !insElem.isNewInstance {
				weight -= concurrency
			}
			return weight, nil
		}
		return -1, scheduler.ErrTypeConvertFail
	}
}

func shouldPopScaledInstance(insElem *instanceElement) bool {
	return len(insElem.threadMap) == insElem.instance.ConcurrentNum
}
