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

// Package requestqueue -
package requestqueue

import (
	"errors"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/queue"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
	"yuanrong.org/kernel/pkg/functionscaler/workermanager"
)

const (
	defaultTriggerChSize = 10000
	timeoutAccuracy      = time.Duration(500) * time.Millisecond
)

var (
	// DefaultRequestTimeout is defined here for the convenience of mocking
	DefaultRequestTimeout   = time.Duration(30) * time.Second
	errInsThdReqReachMaxNum = errors.New("instance thread request reach max number")
)

// ScheduleFunction -
type ScheduleFunction func(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, error)

// PendingInsAcqReq is the instance acquire request
type PendingInsAcqReq struct {
	ResultChan  chan *PendingInsAcqRsp
	CreatedTime time.Time
	InsAcqReq   *types.InstanceAcquireRequest
}

// PendingInsAcqRsp is the instance acquire response
type PendingInsAcqRsp struct {
	InsAlloc *types.InstanceAllocation
	Error    error
}

// InsAcqReqQueue is queue to store instance thread requests and also responsible for scheduling them
type InsAcqReqQueue struct {
	queue              *queue.FifoQueue
	maxQueueLen        int
	insNum             int
	funcKeyWithRes     string
	recoverableError   error
	unrecoverableError error
	requestTimeout     time.Duration
	triggerCh          chan struct{}
	stopCh             chan struct{}
	*sync.RWMutex

	schFuncMap sync.Map
}

// NewInsAcqReqQueue creates a InsAcqReqQueue
func NewInsAcqReqQueue(funcKeyWithRes string, requestTimeout time.Duration) *InsAcqReqQueue {
	if requestTimeout < DefaultRequestTimeout {
		requestTimeout = DefaultRequestTimeout
	}
	insThdReqQue := &InsAcqReqQueue{
		queue:          queue.NewFifoQueue(nil),
		maxQueueLen:    config.GlobalConfig.AutoScaleConfig.BurstScaleNum,
		funcKeyWithRes: funcKeyWithRes,
		requestTimeout: requestTimeout,
		triggerCh:      make(chan struct{}, defaultTriggerChSize),
		stopCh:         make(chan struct{}),
		RWMutex:        &sync.RWMutex{},
	}

	go insThdReqQue.TimeoutReqHandleLoop()
	return insThdReqQue
}

// HandleInsNumUpdate -
func (iq *InsAcqReqQueue) HandleInsNumUpdate(InsNumDiff int) {
	iq.Lock()
	iq.insNum += InsNumDiff
	iq.Unlock()
}

// Len returns length of queue
func (iq *InsAcqReqQueue) Len() int {
	iq.RLock()
	l := iq.queue.Len()
	iq.RUnlock()
	return l
}

// AddRequest adds request into queue
func (iq *InsAcqReqQueue) AddRequest(insThdReq *PendingInsAcqReq) error {
	needTriggerSchedule := false
	iq.Lock()
	defer iq.Unlock()
	if iq.unrecoverableError != nil {
		return iq.unrecoverableError
	}
	if iq.queue.Len() == 0 {
		needTriggerSchedule = true
	}
	if iq.queue.Len() >= iq.maxQueueLen {
		return errInsThdReqReachMaxNum
	}
	err := iq.queue.PushBack(insThdReq)
	if err != nil {
		return err
	}
	metrics.OnPendingRequestAdd(insThdReq.InsAcqReq)
	if needTriggerSchedule {
		iq.TriggerSchedule()
	}
	return nil
}

// RegisterSchFunc register schFunc for schedule instance
func (iq *InsAcqReqQueue) RegisterSchFunc(schFuncKey string, schFunc ScheduleFunction) {
	scheCh := make(chan struct{}, defaultTriggerChSize)
	iq.schFuncMap.Store(schFuncKey, scheCh)
	go func(schFunc ScheduleFunction) {
		for {
			select {
			case _, ok := <-scheCh:
				if !ok {
					return
				}
				iq.realScheduleRequest(schFunc)
			case <-iq.stopCh:
				return
			}
		}
	}(schFunc)
}

// ScheduleRequest schedules requests to instance threads
func (iq *InsAcqReqQueue) ScheduleRequest(schFuncKey string) {
	ch, loaded := iq.schFuncMap.Load(schFuncKey)
	if !loaded {
		log.GetLogger().Errorf("schFunc has not register, skip")
		return
	}
	scheCh, ok := ch.(chan struct{})
	if !ok {
		log.GetLogger().Errorf("schFunc type error, skip")
		return
	}
	if len(scheCh) == 0 {
		scheCh <- struct{}{}
	}
}

// realScheduleRequest schedules requests to instance threads
func (iq *InsAcqReqQueue) realScheduleRequest(schFunc ScheduleFunction) {
	for {
		iq.Lock()
		obj := iq.queue.PopFront()
		if obj == nil {
			iq.Unlock()
			return
		}
		pendingRequest, ok := obj.(*PendingInsAcqReq)
		if !ok {
			iq.Unlock()
			continue
		}
		iq.Unlock()
		insAlloc, err := schFunc(pendingRequest.InsAcqReq)
		if err != nil {
			iq.Lock()
			// PushBack 不会返回err 不需要处理
			_ = iq.queue.PushBack(pendingRequest)
			iq.Unlock()
			return
		}
		metrics.OnPendingRequestRelease(pendingRequest.InsAcqReq)
		pendingRequest.ResultChan <- &PendingInsAcqRsp{
			InsAlloc: insAlloc,
			Error:    nil,
		}
	}
}

// TriggerSchedule triggers request scheduling
func (iq *InsAcqReqQueue) TriggerSchedule() {
	select {
	case iq.triggerCh <- struct{}{}:
	default:
		log.GetLogger().Warnf("trigger channel is blocked in request queue")
	}
}

// TimeoutReqHandleLoop is a loop to continually handle timeout request in queue
func (iq *InsAcqReqQueue) TimeoutReqHandleLoop() {
	timeoutCh := make(<-chan time.Time, 1)
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		select {
		case <-iq.stopCh:
			log.GetLogger().Warnf("stop schedule instance thread for function %s now", iq.funcKeyWithRes)
			err := snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
			iq.Lock()
			iq.ClearReqQueueWithError(err)
			iq.Unlock()
			return
		case _, ok := <-iq.triggerCh:
			if !ok {
				log.GetLogger().Warnf("trigger channel is closed, stop schedule request now")
				return
			}
		case <-timeoutCh:
		}
		var nextReqTimeout time.Duration
		currentTime := time.Now()
		iq.Lock()
		for iq.queue.Len() != 0 {
			obj := iq.queue.Front()
			insThdReq, ok := obj.(*PendingInsAcqReq)
			if !ok {
				iq.queue.PopFront()
				continue
			}
			reqHasWaitTime := currentTime.Sub(insThdReq.CreatedTime)
			if (reqHasWaitTime + timeoutAccuracy).Milliseconds() >= iq.requestTimeout.Milliseconds() {
				err := scheduler.ErrInsReqTimeout
				if iq.recoverableError != nil {
					err = iq.recoverableError
				}
				metrics.OnPendingRequestRelease(insThdReq.InsAcqReq)
				insThdReq.ResultChan <- &PendingInsAcqRsp{
					InsAlloc: nil,
					Error:    err,
				}
				iq.queue.PopFront()
				continue
			}
			// request queue is a fifo queue based on enqueue time of request, if request at front can't be scheduled
			// or not yet timed out then there is no need to go further
			nextReqTimeout = iq.requestTimeout - reqHasWaitTime
			break
		}
		// if queue is not empty then we start a timer to handle request timeout
		if iq.queue.Len() != 0 {
			timer = time.NewTimer(nextReqTimeout)
			timeoutCh = timer.C
		}
		iq.Unlock()
	}
}

// ClearReqQueueWithError - without lock,should lock/unlock outside
func (iq *InsAcqReqQueue) ClearReqQueueWithError(err error) {
	for iq.queue.Len() != 0 {
		obj := iq.queue.Front()
		insThdReq, ok := obj.(*PendingInsAcqReq)
		if !ok {
			iq.queue.PopFront()
			continue
		}
		metrics.OnPendingRequestRelease(insThdReq.InsAcqReq)
		insThdReq.ResultChan <- &PendingInsAcqRsp{
			InsAlloc: nil,
			Error:    err,
		}
		iq.queue.PopFront()
	}
}

// HandleCreateError will return unrecoverable error immediately and save recoverable error for further return
func (iq *InsAcqReqQueue) HandleCreateError(createError error) {
	select {
	case <-iq.stopCh:
		log.GetLogger().Warnf("stop schedule instance thread for function %s now", iq.funcKeyWithRes)
		return
	default:
		if config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeFG && createError != nil {
			iq.handlerWorkMangerError(createError)
			return
		}
		if createError == nil {
			iq.Lock()
			iq.unrecoverableError = nil
			iq.recoverableError = nil
			iq.Unlock()
			return
		}
		log.GetLogger().Warnf("instance request queue for function %s handling create error %s", iq.funcKeyWithRes,
			createError.Error())
		if utils.IsUnrecoverableError(createError) {
			log.GetLogger().Warnf("set unrecoverable create error %s for function %s", createError.Error(),
				iq.funcKeyWithRes)
			iq.Lock()
			iq.unrecoverableError = createError
			if iq.insNum == 0 {
				iq.ClearReqQueueWithError(createError)
			}
			iq.Unlock()
			return
		}
		log.GetLogger().Warnf("set recoverable create error %s for function %s", createError.Error(),
			iq.funcKeyWithRes)
		iq.Lock()
		iq.recoverableError = createError
		iq.Unlock()
	}
}

func (iq *InsAcqReqQueue) handlerWorkMangerError(createError error) {
	iq.Lock()
	currentInstanceNum := iq.insNum
	iq.Unlock()
	if currentInstanceNum == 0 && !workermanager.NeedTryError(createError) {
		log.GetLogger().Errorf("worker manager return error %s for function %s, returning all request now",
			createError.Error(), iq.funcKeyWithRes)
		iq.Lock()
		iq.ClearReqQueueWithError(createError)
		iq.Unlock()
		return
	}
	log.GetLogger().Warnf("set delay return instance create error %v for function %s",
		createError, iq.funcKeyWithRes)
	iq.Lock()
	iq.recoverableError = createError
	iq.Unlock()
}

// UpdateRequestTimeout handles request timeout update
func (iq *InsAcqReqQueue) UpdateRequestTimeout(requestTimeout time.Duration) {
	iq.Lock()
	defer iq.Unlock()
	if requestTimeout == iq.requestTimeout {
		return
	}
	if requestTimeout < DefaultRequestTimeout {
		requestTimeout = DefaultRequestTimeout
	}
	iq.requestTimeout = requestTimeout
}

// Stop stops InsAcqReqQueue
func (iq *InsAcqReqQueue) Stop() {
	commonUtils.SafeCloseChannel(iq.stopCh)
}
