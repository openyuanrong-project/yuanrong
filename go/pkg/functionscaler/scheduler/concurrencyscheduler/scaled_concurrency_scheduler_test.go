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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/queue"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

func TestInstanceQueueWithBuffer(t *testing.T) {
	insQue := &instanceQueueWithBuffer{
		queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(2)),
		buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
		idFunc: getInstanceID,
	}
	instance1 := &types.Instance{InstanceID: "instance1", ConcurrentNum: 2}
	insElem1 := &instanceElement{
		instance: instance1,
	}
	insElem1.initThreadMap()
	insQue.PushBack(insElem1)
	instance2 := &types.Instance{InstanceID: "instance2", ConcurrentNum: 2}
	insElem2 := &instanceElement{
		instance: instance2,
	}
	insElem2.initThreadMap()
	insQue.PushBack(insElem2)
	assert.Equal(t, 2, insQue.Len())
	insQue.DelByID(instance1.InstanceID)
	insList := make([]string, 0, 2)
	insQue.Range(func(obj interface{}) bool {
		insList = append(insList, obj.(*instanceElement).instance.InstanceID)
		return true
	})
	assert.Contains(t, insList, instance2.InstanceID)
	assert.NotContains(t, insList, instance1.InstanceID)
	insQue.SortedRange(func(obj interface{}) bool {
		insList = append(insList, obj.(*instanceElement).instance.InstanceID)
		return true
	})
	assert.Contains(t, insList, instance2.InstanceID)
	assert.NotContains(t, insList, instance1.InstanceID)
	insElem1 = &instanceElement{
		instance: instance1,
	}
	insElem1.initThreadMap()
	insQue.PushBack(insElem1)
	insQue.UpdateObjByID("instance1", &instanceElement{
		instance:  instance1,
		threadMap: make(map[string]struct{}, 0),
	})
	popIns1 := insQue.PopFront().(*instanceElement)
	assert.Equal(t, instance2.InstanceID, popIns1.instance.InstanceID)
	popIns2 := insQue.PopFront().(*instanceElement)
	assert.Equal(t, instance1.InstanceID, popIns2.instance.InstanceID)

	instance3 := &types.Instance{InstanceID: "instance3", ConcurrentNum: 2}
	err := insQue.UpdateObjByID("instance3", &instanceElement{
		instance:  instance3,
		threadMap: make(map[string]struct{}, 0),
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), queue.ErrObjectNotFound.Error())

	insQue.buffer = append(insQue.buffer, &instanceElement{
		instance: instance3,
	})
	err = insQue.UpdateObjByID("instance3", &instanceElement{
		instance:  instance3,
		threadMap: make(map[string]struct{}, 0),
	})
	assert.Nil(t, err)
}

func TestNewScaledConcurrencyScheduler(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 1000*time.Millisecond)
	scs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	assert.NotNil(t, scs)
}

func TestGetReqQueLen(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 1000*time.Millisecond)
	scs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue).(*ScaledConcurrencyScheduler)
	assert.Equal(t, 0, scs.GetReqQueLen())
}

func TestAcquireInstanceScaled(t *testing.T) {
	defer gomonkey.ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "IsFuncOwner", func(
		*selfregister.SchedulerProxy, string) bool {
		return true
	}).Reset()
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 1000*time.Millisecond)
	scs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	index := 0
	scs.ConnectWithInstanceScaler(&fakeInstanceScaler{
		scaleUpFunc: func() {
			index++
			fmt.Printf("fakeInstanceScaler add instance %d start\n", index)
			scs.AddInstance(&types.Instance{
				InstanceID:     fmt.Sprintf("instance%d", index),
				ConcurrentNum:  2,
				ResKey:         resspeckey.ResSpecKey{},
				InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
			})
			fmt.Printf("fakeInstanceScaler add instance %d finish\n", index)
		},
	})
	_, err := scs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	acqIns1, err := scs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := scs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	scs.HandleCreateError(nil)
	acqIns3, err := scs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns3.Instance.InstanceID)
	scs.ReleaseInstance(&types.InstanceAllocation{Instance: &types.Instance{InstanceID: "instance1"}})
	scs.ReleaseInstance(&types.InstanceAllocation{Instance: &types.Instance{InstanceID: "instance2"}})
	acqIns4, err := scs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns4.Instance.InstanceID)
	scs.PopInstance(true)
	scs.PopInstance(true)
	snErr := snerror.New(4001, "some error")
	scs.HandleCreateError(snErr)
	_, err = scs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, snErr, err)
}

func TestPopInstanceScaled(t *testing.T) {
	defer gomonkey.ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "IsFuncOwner", func(
		*selfregister.SchedulerProxy, string) bool {
		return true
	}).Reset()
	config.GlobalConfig.AutoScaleConfig.BurstScaleNum = 1000
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 50*time.Millisecond)
	scs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue).(*ScaledConcurrencyScheduler)
	scs.ConnectWithInstanceScaler(&fakeInstanceScaler{})
	scs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	scs.AcquireInstance(&types.InstanceAcquireRequest{})
	popIns1 := scs.PopInstance(false)
	assert.Nil(t, popIns1)
	scs.AcquireInstance(&types.InstanceAcquireRequest{})
	popIns2 := scs.PopInstance(true)
	assert.Equal(t, "instance1", popIns2.InstanceID)
}

func TestHandleFuncSpecUpdateScaled(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 50*time.Millisecond)
	rcs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	rcs.ConnectWithInstanceScaler(&fakeInstanceScaler{})
	rcs.HandleFuncSpecUpdate(&types.FunctionSpecification{
		InstanceMetaData: commontypes.InstanceMetaData{
			ConcurrentNum: 4,
		},
	})
}

func TestAddInstancePublishScaled(t *testing.T) {
	defer gomonkey.ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "IsFuncOwner", func(
		*selfregister.SchedulerProxy, string) bool {
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
	rcs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	rcs.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 2,
		ResKey:        resspeckey.ResSpecKey{},
		InstanceStatus: commontypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
		},
	})
	time.Sleep(time.Millisecond)
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

func TestDestroy(t *testing.T) {
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 50*time.Millisecond)
	rcs := NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	rcs.Destroy()
}

func Test_instanceQueueWithBuffer_DelByID(t *testing.T) {
	convey.Convey("test DelByID", t, func() {
		insQue := &instanceQueueWithBuffer{
			queue:  queue.NewPriorityQueue(getInstanceID, priorityFuncForScaledInstance(2)),
			buffer: make([]*instanceElement, 0, utils.DefaultSliceSize),
			idFunc: getInstanceID,
		}
		insQue.PushBack(&instanceElement{
			instance: &types.Instance{
				InstanceID: "2",
			},
			threadMap: nil,
		})
		convey.So(insQue.DelByID("1"), convey.ShouldNotBeNil)
		insQue.PushBack(&instanceElement{
			instance: &types.Instance{
				InstanceID: "1",
			},
			threadMap: nil,
		})
		convey.So(insQue.DelByID("1"), convey.ShouldBeNil)
	})

}
