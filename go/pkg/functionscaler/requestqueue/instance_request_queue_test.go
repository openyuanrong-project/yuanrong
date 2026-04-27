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
package requestqueue

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/queue"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/workermanager"
)

var testFuncSpec = &types.FunctionSpecification{
	FuncMetaSignature: "123",
	InstanceMetaData: commonTypes.InstanceMetaData{
		ConcurrentNum: 2,
	},
	FuncCtx: context.TODO(),
}

func TestMain(m *testing.M) {
	config.GlobalConfig = types.Configuration{
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      500,
			ScaleDownTime: 1000,
			BurstScaleNum: 1000,
		},
		LeaseSpan: 500,
	}
	m.Run()
}

func TestNewInsAcqReqQueue(t *testing.T) {
	q := NewInsAcqReqQueue("testFuncKey", 10*time.Second)
	assert.Equal(t, false, q == nil)
	q.Stop()
}

func TestHandleInsNumUpdate(t *testing.T) {
	q := NewInsAcqReqQueue("testFuncKey", 10*time.Second)
	q.HandleInsNumUpdate(1)
	assert.Equal(t, 1, q.insNum)
	q.Stop()
}

func TestLen(t *testing.T) {
	q := NewInsAcqReqQueue("testFuncKey", 10*time.Second)
	assert.Equal(t, 0, q.Len())
	q.Stop()
}

func TestScheduleRequest(t *testing.T) {
	q := NewInsAcqReqQueue("testFuncKey", 10*time.Second)
	ch := make(chan string, 1)
	q.RegisterSchFunc("scheFuncFail", func(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, error) {
		ch <- "scheFuncFail"
		return nil, errors.New("some error")
	})
	q.RegisterSchFunc("scheFuncSucc", func(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, error) {
		ch <- "scheFuncSucc"
		return &types.InstanceAllocation{}, nil
	})
	req := &PendingInsAcqReq{ResultChan: make(chan *PendingInsAcqRsp, 1)}
	q.AddRequest(req)
	q.ScheduleRequest("scheFuncFail")
	failOut := <-ch
	assert.Equal(t, "scheFuncFail", failOut)
	q.ScheduleRequest("scheFuncSucc")
	succOut := <-ch
	assert.Equal(t, "scheFuncSucc", succOut)
	q.Stop()
}

func TestReturnCreateError(t *testing.T) {
	stopChan := make(chan struct{})
	iq := &InsAcqReqQueue{
		stopCh:  stopChan,
		queue:   queue.NewFifoQueue(nil),
		RWMutex: &sync.RWMutex{},
	}
	createErr := snerror.New(statuscode.UserFuncEntryNotFoundErrCode, "entry not found")
	getRespCh := make(chan *PendingInsAcqRsp, 3)
	getResp := &PendingInsAcqReq{
		ResultChan: getRespCh,
	}
	iq.queue.PushBack(getResp)
	convey.Convey("get error success", t, func() {
		iq.HandleCreateError(createErr)
		var err error
		select {
		case itq := <-getRespCh:
			err = itq.Error
		default:
			err = nil
		}
		convey.So(err.Error(), convey.ShouldEqual, "entry not found")
	})
	close(stopChan)
	convey.Convey("chan stop", t, func() {
		iq.HandleCreateError(createErr)
		var err error
		select {
		case itq := <-getRespCh:
			err = itq.Error
		default:
			err = nil
		}
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestTimeoutReqHandleLoop(t *testing.T) {
	var insThdReq1RspTime time.Time
	var insThdReq2RspTime time.Time
	var insThdReq3RspTime time.Time
	var insThdReq4RspTime time.Time

	insThdReqQue := &InsAcqReqQueue{
		funcKeyWithRes: "funcKeyWithRes",
		queue:          queue.NewFifoQueue(nil),
		maxQueueLen:    1000,
		requestTimeout: time.Duration(1) * time.Second,
		triggerCh:      make(chan struct{}, defaultTriggerChSize),
		stopCh:         make(chan struct{}),
		RWMutex:        &sync.RWMutex{},
	}
	go insThdReqQue.TimeoutReqHandleLoop()
	time.Sleep(10 * time.Millisecond)
	insThdReq1 := &PendingInsAcqReq{
		CreatedTime: time.Now(),
		ResultChan:  make(chan *PendingInsAcqRsp, 1),
	}
	insThdReq2 := &PendingInsAcqReq{
		CreatedTime: time.Now(),
		ResultChan:  make(chan *PendingInsAcqRsp, 1),
	}

	insThdReqQue.AddRequest(insThdReq1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(insThdReq *PendingInsAcqReq) {
		defer wg.Done()
		insThdRsp := <-insThdReq.ResultChan
		insThdReq1RspTime = time.Now()
		assert.True(t, insThdReq1RspTime.Sub(insThdReq.CreatedTime) <= insThdReqQue.requestTimeout+50*time.Millisecond)
		assert.Equal(t, insThdRsp.Error.Error(), scheduler.ErrInsReqTimeout.Error())
	}(insThdReq1)
	insThdReqQue.AddRequest(insThdReq2)
	wg.Add(1)
	go func(insThdReq *PendingInsAcqReq) {
		defer wg.Done()
		insThdRsp := <-insThdReq.ResultChan
		insThdReq2RspTime = time.Now()
		assert.True(t, insThdReq2RspTime.Sub(insThdReq.CreatedTime) <= insThdReqQue.requestTimeout+50*time.Millisecond)
		assert.Equal(t, insThdRsp.Error.Error(), scheduler.ErrInsReqTimeout.Error())
	}(insThdReq2)

	time.Sleep(450 * time.Millisecond)
	insThdReq3 := &PendingInsAcqReq{
		CreatedTime: time.Now(),
		ResultChan:  make(chan *PendingInsAcqRsp, 1),
	}
	insThdReqQue.AddRequest(insThdReq3)
	wg.Add(1)
	go func(insThdReq *PendingInsAcqReq) {
		defer wg.Done()
		insThdRsp := <-insThdReq.ResultChan
		insThdReq3RspTime = time.Now()
		assert.True(t, insThdReq3RspTime.Sub(insThdReq.CreatedTime) <= insThdReqQue.requestTimeout+50*time.Millisecond)
		assert.Equal(t, insThdRsp.Error.Error(), scheduler.ErrInsReqTimeout.Error())
	}(insThdReq3)

	time.Sleep(100 * time.Millisecond)
	insThdReq4 := &PendingInsAcqReq{
		CreatedTime: time.Now(),
		ResultChan:  make(chan *PendingInsAcqRsp, 1),
	}
	insThdReqQue.AddRequest(insThdReq4)
	wg.Add(1)
	go func(insThdReq *PendingInsAcqReq) {
		defer wg.Done()
		insThdRsp := <-insThdReq.ResultChan
		insThdReq4RspTime = time.Now()
		assert.True(t, insThdReq3RspTime.Sub(insThdReq.CreatedTime) <= insThdReqQue.requestTimeout+50*time.Millisecond)
		assert.Equal(t, insThdRsp.Error.Error(), scheduler.ErrInsReqTimeout.Error())
	}(insThdReq4)
	wg.Wait()
	assert.True(t, insThdReq1RspTime.Sub(insThdReq2RspTime) <= 10*time.Millisecond)
	assert.True(t, insThdReq2RspTime.Sub(insThdReq3RspTime) <= 10*time.Millisecond)
	assert.True(t, insThdReq4RspTime.Sub(insThdReq3RspTime) >= 500*time.Millisecond)
}

func TestHandlerWorkMangerError(t *testing.T) {
	tests := []struct {
		name              string
		insNum            int
		createError       error
		needTryError      bool
		expectClearCalled bool
		expectRecoverable bool
	}{
		{
			name:              "Current instance number is 0, error not retryable",
			insNum:            0,
			createError:       errors.New("non-retryable error"),
			needTryError:      false,
			expectClearCalled: true,
			expectRecoverable: false,
		},
		{
			name:              "Current instance number is not 0, error is retryable",
			insNum:            1,
			createError:       errors.New("retryable error"),
			needTryError:      true,
			expectClearCalled: false,
			expectRecoverable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			iq := &InsAcqReqQueue{
				insNum:           tt.insNum,
				funcKeyWithRes:   "test-func-key",
				recoverableError: nil,
				RWMutex:          &sync.RWMutex{},
			}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			clearCalled := false

			patches.ApplyMethod(reflect.TypeOf(iq), "ClearReqQueueWithError",
				func(_ *InsAcqReqQueue, err error) {
					clearCalled = true
				})

			patches.ApplyFunc(workermanager.NeedTryError,
				func(err error) bool {
					return tt.needTryError
				})

			iq.handlerWorkMangerError(tt.createError)

			assert.Equal(t, tt.expectClearCalled, clearCalled, "ClearReqQueueWithError should be called accordingly")
			if tt.expectRecoverable {
				assert.Equal(t, tt.createError, iq.recoverableError, "recoverableError should be set accordingly")
			} else {
				assert.Nil(t, iq.recoverableError, "recoverableError should be nil")
			}
		})
	}
}

func TestUpdateRequestTimeout(t *testing.T) {
	q := NewInsAcqReqQueue("testFuncKey", 10*time.Second)
	q.UpdateRequestTimeout(20 * time.Second)
	assert.Equal(t, DefaultRequestTimeout, q.requestTimeout)
	q.Stop()
}
