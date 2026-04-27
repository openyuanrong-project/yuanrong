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

// Package instancequeue -
package instancequeue

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/concurrencyscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

type fakeInstanceScheduler struct {
	insQue     []*types.Instance
	index      int
	acquireErr error
	releaseErr error
}

func (f *fakeInstanceScheduler) ReassignInstanceWhenGray(ratio int) {
}

func (f *fakeInstanceScheduler) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, error) {
	if f.acquireErr != nil {
		return nil, f.acquireErr
	}
	instance := &types.Instance{
		InstanceID: fmt.Sprintf("instance%d", f.index),
	}
	thread := &types.InstanceAllocation{
		Instance:     instance,
		AllocationID: fmt.Sprintf("%s-thd1", instance.InstanceID),
	}
	f.index++
	return thread, nil
}

func (f *fakeInstanceScheduler) HandleFuncOwnerUpdate(isManaged bool) {
	// TODO implement me
}

func (f *fakeInstanceScheduler) GetInstanceNumber(onlySelf bool) int {
	return len(f.insQue)
}

func (f *fakeInstanceScheduler) ReleaseInstance(thread *types.InstanceAllocation) error {
	if f.releaseErr != nil {
		return f.releaseErr
	}
	return nil
}

func (f *fakeInstanceScheduler) AddInstance(instance *types.Instance) error {
	f.insQue = append(f.insQue, instance)
	return nil
}

func (f *fakeInstanceScheduler) PopInstance(force bool) *types.Instance {
	if len(f.insQue) == 0 {
		return nil
	}
	instance := f.insQue[len(f.insQue)-1]
	f.insQue = f.insQue[:len(f.insQue)-1]
	return instance
}

func (f *fakeInstanceScheduler) DelInstance(instance *types.Instance) error {
	index := -1
	for i, item := range f.insQue {
		if item.InstanceID == instance.InstanceID {
			index = i
			break
		}
	}
	if index == -1 {
		return scheduler.ErrInsNotExist
	}
	f.insQue = append(f.insQue[:index], f.insQue[index+1:]...)
	return nil
}

func (f *fakeInstanceScheduler) ConnectWithInstanceScaler(instanceScaler scaler.InstanceScaler) {
	return
}

func (f *fakeInstanceScheduler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
}

func (f *fakeInstanceScheduler) HandleInstanceUpdate(instance *types.Instance) {
	f.insQue = append(f.insQue, instance)
}

func (f *fakeInstanceScheduler) HandleCreateError(createErr error) {
}

func (f *fakeInstanceScheduler) SignalAllInstances(signalFunc scheduler.SignalInstanceFunc) {
	for _, item := range f.insQue {
		signalFunc(item)
	}
}

func (f *fakeInstanceScheduler) Destroy() {
}

type fakeInstanceScaler struct {
	enable bool
	insNum int
}

func (f *fakeInstanceScaler) SetFuncOwner(isManaged bool) {
}

func (f *fakeInstanceScaler) SetEnable(enable bool) {
	f.enable = enable
}

func (f *fakeInstanceScaler) TriggerScale() {
}

func (f *fakeInstanceScaler) CheckScaling() bool {
	return false
}

func (f *fakeInstanceScaler) UpdateCreateMetrics(coldStartTime time.Duration) {
}

func (f *fakeInstanceScaler) HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int) {
}

func (f *fakeInstanceScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
}

func (f *fakeInstanceScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
}

func (f *fakeInstanceScaler) HandleCreateError(createError error) {
}

func (f *fakeInstanceScaler) GetExpectInstanceNumber() int {
	return f.insNum
}

func (f *fakeInstanceScaler) Destroy() {
}

func TestNewScaledInstanceQueue(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	assert.NotNil(t, q)
}

func TestAcquireInstance(t *testing.T) {
	funcCtx, cancelFunc := context.WithCancel(context.TODO())
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx: funcCtx,
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := metrics.NewBucketMetricsCollector("testFunction", "500-500")
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	q.SetInstanceScheduler(&fakeInstanceScheduler{index: 1})
	insThd, err := q.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", insThd.Instance.InstanceID)
	someErr := snerror.New(1234, "some error")
	q.SetInstanceScheduler(&fakeInstanceScheduler{acquireErr: someErr})
	_, err = q.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, someErr, err)
	q.updating = true
	acqChan := make(chan error, 1)
	go func() {
		_, err = q.AcquireInstance(&types.InstanceAcquireRequest{})
		acqChan <- err
	}()
	time.Sleep(1 * time.Millisecond)
	select {
	case <-acqChan:
		t.Errorf("should not acquire instance thread")
	default:
	}
	q.Cond.Broadcast()
	time.Sleep(1 * time.Millisecond)
	select {
	case <-acqChan:
	default:
		t.Errorf("should acquire instance thread")
	}
	cancelFunc()
	_, err = q.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, statuscode.FuncMetaNotFoundErrCode, err.Code())
}

func TestReleaseInstance(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	q.SetInstanceScheduler(&fakeInstanceScheduler{})
	err := q.ReleaseInstance(&types.InstanceAllocation{AllocationID: "testThread"})
	assert.Nil(t, err)
	someErr := snerror.New(1234, "some error")
	q.SetInstanceScheduler(&fakeInstanceScheduler{releaseErr: someErr})
	err = q.ReleaseInstance(&types.InstanceAllocation{AllocationID: "testThread"})
	assert.Equal(t, someErr, err)
}

func TestHandleInstanceDelete(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
		},
	}
	q.SetInstanceScheduler(instanceScheduler)
	q.HandleInstanceDelete(&types.Instance{
		InstanceID: "instance1",
	})
	assert.Equal(t, 0, len(instanceScheduler.insQue))
}

func TestHandleFuncSpecUpdate(t *testing.T) {
	defer ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {}).Reset()
	defer ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}).Reset()
	deleteFunc := func(ins *types.Instance) error {
		return nil
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncKey:           "testFunction",
			FuncMetaSignature: "funcSig1",
			InstanceMetaData: commontypes.InstanceMetaData{
				ConcurrentNum: 1,
			},
		},
		ResKey: resspeckey.ResSpecKey{
			CPU:    500,
			Memory: 500,
		},
		DeleteInstanceFunc: deleteFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
		},
	}
	stop := make(chan struct{})
	registry.InitRegistry(stop)
	instanceScaler := &fakeInstanceScaler{}
	q.SetInstanceScheduler(instanceScheduler)
	q.SetInstanceScaler(instanceScaler)
	q.HandleFuncSpecUpdate(&types.FunctionSpecification{
		FuncMetaSignature: "funcSig2",
		ResourceMetaData: commontypes.ResourceMetaData{
			CPU:    300,
			Memory: 300,
		},
		InstanceMetaData: commontypes.InstanceMetaData{
			ConcurrentNum: 2,
		},
	})
	assert.Equal(t, "funcSig2", q.funcSig)
	assert.Equal(t, "testFunction-cpu-500-mem-500-storage-0-cstRes--cstResSpec--invokeLabel-", q.funcKeyWithRes)
	assert.Equal(t, 2, q.concurrentNum)
	assert.Equal(t, 0, len(instanceScheduler.insQue))
	assert.Equal(t, true, instanceScaler.enable)
	q.instanceType = types.InstanceTypeReserved
	q.HandleFuncSpecUpdate(&types.FunctionSpecification{
		FuncMetaSignature: "funcSig3",
		ResourceMetaData: commontypes.ResourceMetaData{
			CPU:    300,
			Memory: 300,
		},
		InstanceMetaData: commontypes.InstanceMetaData{
			ConcurrentNum: 4,
		},
	})
	assert.Equal(t, "funcSig3", q.funcSig)
	assert.Equal(t, 4, q.concurrentNum)
	assert.Equal(t, 0, len(instanceScheduler.insQue))
	assert.Equal(t, true, instanceScaler.enable)
	close(stop)
}

func TestHandleFaultyInstance(t *testing.T) {
	var delIns *types.Instance
	deleteFunc := func(ins *types.Instance) error {
		delIns = ins
		return nil
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType:       types.InstanceTypeScaled,
		FuncSpec:           &types.FunctionSpecification{},
		ResKey:             resspeckey.ResSpecKey{},
		DeleteInstanceFunc: deleteFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
		},
	}
	q.SetInstanceScheduler(instanceScheduler)
	q.HandleFaultyInstance(&types.Instance{
		InstanceID: "instance1",
	})
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, "instance1", delIns.InstanceID)

	q.HandleFaultyInstance(&types.Instance{InstanceID: "instance2"}) // 是为了构造在scheduler刚启动后，收到被状态为fatal的函数实例更新事件
	time.Sleep(1 * time.Millisecond)
	// 目的是为了测试，即使本地缓存没有，也应该要调用deleteFunc来删除etcd里残留的数据
	assert.Equal(t, "instance2", delIns.InstanceID)
}

func TestHandleAliasUpdate(t *testing.T) {
	var signalIns *types.Instance
	var signalNum int
	signalFunc := func(ins *types.Instance, sig int) {
		signalIns = ins
		signalNum = sig
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType:       types.InstanceTypeScaled,
		FuncSpec:           &types.FunctionSpecification{},
		ResKey:             resspeckey.ResSpecKey{},
		SignalInstanceFunc: signalFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
		},
	}
	q.SetInstanceScheduler(instanceScheduler)

	defer ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "CheckFuncOwner", func(sp *selfregister.SchedulerProxy, funcKey string) (string, bool) {
		return "", true
	}).Reset()

	q.HandleAliasUpdate()
	assert.Equal(t, "instance1", signalIns.InstanceID)
	assert.Equal(t, constant.KillSignalAliasUpdate, signalNum)
}

func TestGetInstanceNumber(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	q.SetInstanceScheduler(&fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
			{
				InstanceID: "instance2",
			},
			{
				InstanceID: "instance3",
			},
		},
	})
	assert.Equal(t, 3, q.GetInstanceNumber(true))
}

func TestRecoverInstance(t *testing.T) {
	var delIns *types.Instance
	deleteFunc := func(ins *types.Instance) error {
		delIns = ins
		return nil
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncMetaSignature: "funcSig1",
		},
		ResKey:             resspeckey.ResSpecKey{},
		DeleteInstanceFunc: deleteFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{}
	q.SetInstanceScheduler(instanceScheduler)
	instanceScaler := &fakeInstanceScaler{enable: false}
	q.SetInstanceScaler(instanceScaler)
	q.RecoverInstance(map[string]*types.Instance{
		"instance1": {
			InstanceID: "instance1",
			FuncSig:    "funcSig1",
			InstanceStatus: commontypes.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusSubHealth),
			},
		},
		"instance2": {
			InstanceID: "instance2",
			FuncSig:    "funcSig2",
			InstanceStatus: commontypes.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusRunning),
			},
		},
	})
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, "instance2", delIns.InstanceID)
	assert.Equal(t, false, instanceScaler.enable)
}

func TestDestroy(t *testing.T) {
	var delIns *types.Instance
	deleteFunc := func(ins *types.Instance) error {
		delIns = ins
		return nil
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncMetaSignature: "funcSig1",
		},
		ResKey:             resspeckey.ResSpecKey{},
		DeleteInstanceFunc: deleteFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{
			{
				InstanceID: "instance1",
			},
		},
	}
	instanceScaler := &fakeInstanceScaler{}
	q.SetInstanceScheduler(instanceScheduler)
	q.SetInstanceScaler(instanceScaler)
	q.Destroy()
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, "instance1", delIns.InstanceID)
}

func TestHandleInstanceSync(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncMetaSignature: "funcSig1",
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{
		insQue: []*types.Instance{},
	}
	instanceScaler := &fakeInstanceScaler{}
	q.SetInstanceScheduler(instanceScheduler)
	q.SetInstanceScaler(instanceScaler)
	cnt := 0
	assert.Equal(t, cnt, 0)
	gomonkey.ApplyFunc((*concurrencyscheduler.ScaledConcurrencyScheduler).HandleInstanceUpdate, func(_ *concurrencyscheduler.ScaledConcurrencyScheduler, instance *types.Instance) { cnt++ })
	q.HandleInstanceUpdate(&types.Instance{InstanceID: "aaa"})
	DisableCreateRetry()
	q.EnableInstanceScale()
	assert.Equal(t, retryDelayFactor, 2)
}

func TestStartScaleUpWorker(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncMetaSignature: "funcSig1",
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	q.SetInstanceScheduler(&fakeInstanceScheduler{})
	q.SetInstanceScaler(&fakeInstanceScaler{})
	q.Destroy()
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, true, q.insCreateQueue.stopped)
}

func TestScaleUpProcess(t *testing.T) {
	createFunc := func(string, string, types.InstanceType, resspeckey.ResSpecKey, []byte) (*types.Instance, error) {
		return &types.Instance{}, nil
	}
	var delIns *types.Instance
	deleteFunc := func(ins *types.Instance) error {
		delIns = ins
		return nil
	}
	retryDelayLimit = 1 * time.Second
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx:           context.TODO(),
			FuncMetaSignature: "funcSig1",
		},
		ResKey:             resspeckey.ResSpecKey{},
		CreateInstanceFunc: createFunc,
		DeleteInstanceFunc: deleteFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{}
	instanceScaler := &fakeInstanceScaler{}
	q.SetInstanceScheduler(instanceScheduler)
	q.SetInstanceScaler(instanceScaler)
	// no error
	callCount := 0
	callCountMutex := new(sync.Mutex)
	callback := func(i int) {
		callCountMutex.Lock()
		callCount++
		callCountMutex.Unlock()
	}
	q.ScaleUpHandler(3, callback)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, callCount)
	// instance nil & error not nil
	patchCreateFunc := ApplyFunc(createFunc, func(string, string, types.InstanceType, resspeckey.ResSpecKey, []byte) (*types.Instance, error) {
		return nil, snerror.New(4001, "user error")
	})
	q.ScaleUpHandler(1, callback)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 4, callCount)
	patchCreateFunc.Reset()
	// instance not nil & error not nil
	patchCreateFunc = ApplyFunc(createFunc, func(string, string, types.InstanceType, resspeckey.ResSpecKey, []byte) (*types.Instance, error) {
		return &types.Instance{InstanceID: "instance1"}, snerror.New(4001, "user error")
	})
	q.ScaleUpHandler(1, callback)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 5, callCount)
	assert.Equal(t, "instance1", delIns.InstanceID)
	patchCreateFunc.Reset()
}

func TestScaleDownProcess(t *testing.T) {
	var delIns *types.Instance
	deleteFunc := func(ins *types.Instance) error {
		delIns = ins
		return nil
	}
	createFunc := func(string, string, types.InstanceType, resspeckey.ResSpecKey, []byte) (
		*types.Instance, error) {
		return nil, nil
	}
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx:           context.TODO(),
			FuncMetaSignature: "funcSig1",
		},
		ResKey:             resspeckey.ResSpecKey{},
		DeleteInstanceFunc: deleteFunc,
		CreateInstanceFunc: createFunc,
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	instanceScheduler := &fakeInstanceScheduler{}
	instanceScaler := &fakeInstanceScaler{}
	q.SetInstanceScheduler(instanceScheduler)
	q.SetInstanceScaler(instanceScaler)
	callCount := 0
	callback := func(i int) {
		callCount++
	}
	q.ScaleDownHandler(1, callback)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 1, callCount)
	assert.Nil(t, delIns)
	instanceScheduler.insQue = append(instanceScheduler.insQue, &types.Instance{InstanceID: "instance1"})
	q.scaleDownInstance(callback)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, "instance1", delIns.InstanceID)
	assert.Equal(t, 0, instanceScheduler.GetInstanceNumber(true))
	delIns = nil
	q.insCreateQueue.push(&InstanceCreateRequest{callback: callback})
	q.ScaleDownHandler(1, callback)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 3, callCount)
	assert.Nil(t, delIns)
}

func TestHandleFuncOwnerChange(t *testing.T) {
	setFuncOwner := false
	patches := []*Patches{
		ApplyFunc((*selfregister.SchedulerProxy).CheckFuncOwner, func(_ *selfregister.SchedulerProxy, _ string) (string, bool) {
			return "", setFuncOwner
		}),
		ApplyFunc((*concurrencyscheduler.ScaledConcurrencyScheduler).HandleFuncOwnerUpdate, func(
			_ *concurrencyscheduler.ScaledConcurrencyScheduler, _ bool) {
			return
		}),
		ApplyFunc((*scaler.AutoScaler).SetEnable, func(_ *scaler.AutoScaler, _ bool) {
			return
		}),
	}
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx:           context.TODO(),
			FuncMetaSignature: "funcSig1",
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	q.instanceScheduler = &concurrencyscheduler.ScaledConcurrencyScheduler{}
	q.instanceScaler = &scaler.AutoScaler{}
	q.HandleFuncOwnerChange(func(sessionInfo *types.SessionInfo, instance *types.Instance) {})
	setFuncOwner = true
	q.HandleFuncOwnerChange(func(sessionInfo *types.SessionInfo, instance *types.Instance) {})
	assert.Equal(t, true, q.isFuncOwner)
}
func TestHandleRatioUpdate(t *testing.T) {
	patches := []*Patches{}
	expectRatio := 0
	patches = append(patches, ApplyFunc(
		(*concurrencyscheduler.ScaledConcurrencyScheduler).ReassignInstanceWhenGray,
		func(s *concurrencyscheduler.ScaledConcurrencyScheduler, ratio int) {
			expectRatio = ratio
		},
	))
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx:           context.TODO(),
			FuncMetaSignature: "funcSig1",
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 1000*time.Millisecond)
	q.instanceScheduler = concurrencyscheduler.NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue)
	q.instanceScaler = &scaler.AutoScaler{}

	q.HandleRatioUpdate(50)
	assert.Equal(t, 50, expectRatio)
}

func TestScaledInstanceQueue_HandleFaaSSchedulerUpdate(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx:           context.TODO(),
			FuncMetaSignature: "funcSig1",
		},
		ResKey: resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	q := NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	scheduler := &fakeInstanceScheduler{index: 1, insQue: make([]*types.Instance, 0)}
	scheduler.insQue = append(scheduler.insQue, &types.Instance{InstanceID: "instance1"})
	q.SetInstanceScheduler(scheduler)
	result := false
	defer gomonkey.ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "CheckFuncOwner", func(
		*selfregister.SchedulerProxy, string) (string, bool) {
		return "", result
	}).Reset()

	flag := false
	q.signalInstanceFunc = func(instance *types.Instance, i int) {
		flag = true
	}
	q.HandleFaaSSchedulerUpdate()
	assert.Equal(t, flag, true)

	result = true
	q.HandleFaaSSchedulerUpdate()
	assert.Equal(t, flag, true)
}
