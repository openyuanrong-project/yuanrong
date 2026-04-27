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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/lease"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewReservedConcurrencyScheduler(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("", 100*time.Millisecond)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, 0, InsThdReqQueue)
	assert.NotNil(t, rcs)
}

func TestAcquireInstanceReservedNew(t *testing.T) {
	config.GlobalConfig.LeaseSpan = 5000
	defer func() {
		config.GlobalConfig.LeaseSpan = 0
	}()
	defer gomonkey.ApplyGlobalVar(&requestqueue.DefaultRequestTimeout, 100*time.Millisecond).Reset()
	defer gomonkey.ApplyFunc((*selfregister.SchedulerProxy).IsFuncOwner, func(_ *selfregister.SchedulerProxy,
		funcKey string) bool {
		return true
	}).Reset()
	defer gomonkey.ApplyMethod(reflect.TypeOf(&fakeInstanceScaler{}), "GetExpectInstanceNumber",
		func(f *fakeInstanceScaler) int {
			return 1
		}).Reset()
	defer gomonkey.ApplyFunc((*lease.GenericInstanceLeaseManager).CreateInstanceLease,
		func(_ *lease.GenericInstanceLeaseManager,
			insAlloc *types.InstanceAllocation, interval time.Duration, callback func()) (types.InstanceLease, error) {
			return nil, nil
		}).Reset()
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("", 100*time.Millisecond)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, 50*time.Millisecond, InsThdReqQueue)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{
		scaling: true,
		timer:   time.NewTimer(100 * time.Millisecond),
	})

	rcs.HandleCreateError(errors.New("some error"))
	_, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, "some error", err.Error())
	rcs.HandleCreateError(snerror.New(4011, "user error"))
	_, err = rcs.AcquireInstance(&types.InstanceAcquireRequest{})

	rcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rcs.HandleCreateError(nil)
	insAlloc1, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Equal(t, nil, err)
	assert.Equal(t, "instance1", insAlloc1.Instance.InstanceID)
	_, err = rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Equal(t, false, err == nil)
}

func TestAcquireInstanceReserved(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("", 10)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, 50*time.Millisecond, InsThdReqQueue)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{
		scaling: true,
		timer:   time.NewTimer(100 * time.Millisecond),
	})
	_, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{
		scaling: true,
		timer:   time.NewTimer(10 * time.Millisecond),
	})
	_, err = rcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	_, err = rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	rcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	rcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns2, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns2.Instance.InstanceID)
	acqIns3, err := rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns3.Instance.InstanceID)
	_, err = rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance3"})
	assert.Equal(t, "instance does not exist in queue", err.Error())

	rc := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, 150*time.Millisecond, InsThdReqQueue)
	rc.ConnectWithInstanceScaler(&fakeInstanceScaler{
		scaling:         true,
		timer:           time.NewTimer(100 * time.Millisecond),
		targetRsvInsNum: 2,
	})
	rc.HandleCreateError(errors.New("resource not enough"))
	_, err = rc.AcquireInstance(&types.InstanceAcquireRequest{})
	go func() {
		time.Sleep(80 * time.Millisecond)
		rc.HandleCreateError(nil)
	}()
	assert.Equal(t, "resource not enough", err.Error())
}

func TestPopInstanceReserved(t *testing.T) {
	defer gomonkey.ApplyFunc((*selfregister.SchedulerProxy).IsFuncOwner, func(_ *selfregister.SchedulerProxy,
		funcKey string) bool {
		return true
	}).Reset()
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("", 10)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, 50*time.Millisecond, InsThdReqQueue)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{})
	popIns1 := rcs.PopInstance(false)
	assert.Nil(t, popIns1)
	rcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	popIns2 := rcs.PopInstance(false)
	assert.Equal(t, "instance2", popIns2.InstanceID)
}

func TestHandleFuncSpecUpdateReserved(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("", 10)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, 50*time.Millisecond, InsThdReqQueue)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{})
	rcs.HandleFuncSpecUpdate(&types.FunctionSpecification{
		InstanceMetaData: commontypes.InstanceMetaData{
			ConcurrentNum: 4,
		},
	})
}

func TestAddInstancePublishReserved(t *testing.T) {
	defer gomonkey.ApplyFunc((*selfregister.SchedulerProxy).IsFuncOwner, func(_ *selfregister.SchedulerProxy,
		funcKey string) bool {
		return true
	}).Reset()
	config.GlobalConfig.AutoScaleConfig.BurstScaleNum = 1000
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 50*time.Millisecond)
	insThdReq1 := &requestqueue.PendingInsAcqReq{
		ResultChan: make(chan *requestqueue.PendingInsAcqRsp, 1),
		InsAcqReq:  &types.InstanceAcquireRequest{},
	}
	insThdReq2 := &requestqueue.PendingInsAcqReq{
		ResultChan: make(chan *requestqueue.PendingInsAcqRsp, 1),
		InsAcqReq:  &types.InstanceAcquireRequest{},
	}
	InsThdReqQueue.AddRequest(insThdReq1)
	InsThdReqQueue.AddRequest(insThdReq2)
	rcs := NewReservedConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, 50*time.Millisecond, InsThdReqQueue)
	rcs.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 2,
		ResKey:        resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
		},
	})
	time.Sleep(10 * time.Millisecond)
	select {
	case insThd := <-insThdReq1.ResultChan:
		assert.Equal(t, "instance1", insThd.InsAlloc.Instance.InstanceID)
	default:
		t.Errorf("should get instance from result channel")
	}
	select {
	case insThd := <-insThdReq2.ResultChan:
		assert.Equal(t, "instance1", insThd.InsAlloc.Instance.InstanceID)
	default:
		t.Errorf("should get instance from result channel")
	}
}
