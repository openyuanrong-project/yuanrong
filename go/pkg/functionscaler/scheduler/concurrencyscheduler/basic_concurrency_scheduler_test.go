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
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/datasystemclient"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/queue"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/lease"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

type fakeInstanceScaler struct {
	timer           *time.Timer
	scaling         bool
	scaleUpFunc     func()
	targetRsvInsNum int
}

func (f *fakeInstanceScaler) SetFuncOwner(isManaged bool) {
}

func (f *fakeInstanceScaler) SetEnable(enable bool) {
}

func (f *fakeInstanceScaler) TriggerScale() {
	go func() {
		time.Sleep(10 * time.Millisecond)
		if f.scaleUpFunc != nil {
			f.scaleUpFunc()
		}
	}()
}

func (f *fakeInstanceScaler) CheckScaling() bool {
	if f.timer == nil {
		return false
	}
	select {
	case <-f.timer.C:
		f.scaling = false
		return false
	default:
		return f.scaling
	}
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
	return f.targetRsvInsNum
}

func (f *fakeInstanceScaler) Destroy() {
}

func TestMain(m *testing.M) {
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {}),
		gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetCAEMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}),
		gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	config.GlobalConfig = types.Configuration{}
	config.GlobalConfig.AutoScaleConfig = types.AutoScaleConfig{
		SLAQuota:      1000,
		ScaleDownTime: 1000,
		BurstScaleNum: 100000,
	}
	config.GlobalConfig.LeaseSpan = 500
	config.GlobalConfig.EnableSessionRecover = true
	registry.InitRegistry(make(chan struct{}))
	m.Run()
}

func TestNewBasicConcurrencyScheduler(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, "", nil, nil)
	assert.NotNil(t, bcs)
}

func TestGetInstanceNumber(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	err := bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
	getNum := bcs.GetInstanceNumber(true)
	assert.Equal(t, 0, getNum)
	bcs.isFuncOwner = true
	err = bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Nil(t, err)
	getNum = bcs.GetInstanceNumber(true)
	assert.Equal(t, 1, getNum)
}

func TestAcquireInstanceBasic(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	assert.Nil(t, acqIns1)
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	assert.Equal(t, 1, checkInUseInsThd)
	assert.Equal(t, 1, checkAvailInsThd)
	acqIns3, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns3.Instance.InstanceID)
	assert.Equal(t, 2, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	acqIns4, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	assert.Nil(t, acqIns4)
	defer gomonkey.ApplyFunc((*lease.GenericInstanceLeaseManager).CreateInstanceLease,
		func(_ *lease.GenericInstanceLeaseManager,
			insAlloc *types.InstanceAllocation, interval time.Duration, callback func()) (types.InstanceLease, error) {
			return nil, errors.New("some error")
		}).Reset()
	bcs.ReleaseInstance(acqIns3)
	_, err = bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.NotNil(t, err)
}

func TestAcquireInstanceOtherQueue(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = false
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	assert.Nil(t, acqIns1)
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	assert.Equal(t, 0, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	acqIns3, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns3.Instance.InstanceID)
	assert.Equal(t, 0, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	acqIns4, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	assert.Nil(t, acqIns4)
}

func TestAcquireInstanceWithSession(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	//bcs.isFuncOwner = true
	bcs.HandleFuncOwnerUpdate(true)
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: 2,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns1.Instance.InstanceID)
	assert.Equal(t, 2, checkInUseInsThd)
	assert.Equal(t, 6, checkAvailInsThd)
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: 2,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns2.Instance.InstanceID)
	assert.Equal(t, 2, checkInUseInsThd)
	assert.Equal(t, 6, checkAvailInsThd)
	acqIns3, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: 2,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns3.Instance.InstanceID)
	assert.Equal(t, 3, checkInUseInsThd)
	assert.Equal(t, 5, checkAvailInsThd)
}

func TestReleaseInstance(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	err := bcs.ReleaseInstance(&types.InstanceAllocation{
		Instance: &types.Instance{
			InstanceID: "instance3",
		},
	})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	acqIns1, _ := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	err = bcs.ReleaseInstance(acqIns1)
	assert.Nil(t, err)
	assert.Equal(t, 0, checkInUseInsThd)
	assert.Equal(t, 2, checkAvailInsThd)
	err = bcs.ReleaseInstance(&types.InstanceAllocation{
		Instance: &types.Instance{
			InstanceID: "instance2",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, checkAvailInsThd)
}

func TestReleaseInstanceWithSession(t *testing.T) {
	mockTimer := time.NewTimer(100 * time.Millisecond)
	defer gomonkey.ApplyFunc(time.NewTimer, func(d time.Duration) *time.Timer {
		mockTimer.Reset(100 * time.Millisecond)
		return mockTimer
	}).Reset()
	defer gomonkey.ApplyFunc((*lease.GenericInstanceLeaseManager).CreateInstanceLease,
		func(_ *lease.GenericInstanceLeaseManager,
			insAlloc *types.InstanceAllocation, interval time.Duration, callback func()) (types.InstanceLease, error) {
			return nil, nil
		}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			SessionTTL:  1,
			Concurrency: 2,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	err = bcs.ReleaseInstance(acqIns1)
	assert.Nil(t, err)
	assert.Equal(t, 2, checkInUseInsThd)
	assert.Equal(t, 2, checkAvailInsThd)
	time.Sleep(50 * time.Millisecond)
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			SessionTTL:  1,
			Concurrency: 2,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	time.Sleep(50 * time.Millisecond)
	err = bcs.ReleaseInstance(acqIns2)
	assert.Nil(t, err)
	assert.Equal(t, 2, checkInUseInsThd)
	assert.Equal(t, 2, checkAvailInsThd)
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, 0, checkInUseInsThd)
	assert.Equal(t, 4, checkAvailInsThd)
}

func TestAgentSessionOverAcquireShouldNotReportLeaseMetrics(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	acquireCount := 0
	releaseCount := 0
	patches.ApplyFunc(metrics.OnAcquireLease, func(_ *types.InstanceAllocation) {
		acquireCount++
	})
	patches.ApplyFunc(metrics.OnReleaseLease, func(_ *types.InstanceAllocation) {
		releaseCount++
	})

	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey: "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 1,
		},
		ExtendedMetaData: commonTypes.ExtendedMetaData{
			EnableAgentSession: true,
		},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))

	insElem := &instanceElement{
		instance: &types.Instance{
			InstanceID:     "instance1",
			ConcurrentNum:  1,
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		},
		threadMap: map[string]struct{}{},
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	record := &sessionRecord{
		ctx:           ctx,
		sessionID:     "session1",
		availThdMap:   map[string]struct{}{},
		allocThdMap:   map[string]struct{}{"thread-1": {}},
		overAcqThdMap: make(map[string]struct{}),
		insElem:       insElem,
		cancelFunc:    cancelFunc,
	}
	bcs.sessionManager.addSession(record.sessionID, record)

	insAlloc, err := bcs.createOverAcqThread(record, &types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID: "session1",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, acquireCount)
	assert.Len(t, record.overAcqThdMap, 1)

	err = bcs.releaseInstanceThreadWithSession(bcs.selfInstanceQueue, insElem, insAlloc)
	assert.Nil(t, err)
	assert.Equal(t, 0, releaseCount)
	assert.Len(t, record.overAcqThdMap, 0)
}

func TestAddInstance(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkAvailInsThd := 0
	checkTotalInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	err := bcs.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 2,
		ResKey:        resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInternal, err)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkTotalInsThd)
	err = bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	err = bcs.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 2,
		ResKey:        resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInsAlreadyExist, err)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	err = bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	err = bcs.AddInstance(&types.Instance{
		InstanceID:    "instance2",
		ConcurrentNum: 2,
		ResKey:        resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInsAlreadyExist, err)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)

	// evicting实例能添加进去，但是指标不上报
	err = bcs.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
}

func TestDelInstance(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkAvailInsThd := 0
	checkInUsedInsThd := 0
	checkTotalInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUsedInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance_evicting",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	err := bcs.DelInstance(&types.Instance{
		InstanceID: "instance3",
	})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	err = bcs.DelInstance(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	err = bcs.DelInstance(&types.Instance{
		InstanceID: "instance2",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 0, checkTotalInsThd)

	// evicting实例能正常删除，并且不影响指标
	err = bcs.DelInstance(&types.Instance{
		InstanceID: "instance_evicting",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 0, checkTotalInsThd)

}

func TestPopInstanceElement(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkAvailInsThd := 0
	checkInUsedInsThd := 0
	checkTotalInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUsedInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	popIns1 := bcs.popInstanceElement(forward, nil, false)
	assert.Nil(t, popIns1)
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance_evicting",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	popIns2 := bcs.popInstanceElement(forward, nil, false)
	assert.Equal(t, "instance2", popIns2.instance.InstanceID)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	popIns3 := bcs.popInstanceElement(forward, func(element *instanceElement) bool { return false }, false)
	assert.Nil(t, popIns3)
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	popIns4 := bcs.popInstanceElement(forward, func(element *instanceElement) bool { return true }, false)
	assert.Equal(t, "instance1", popIns4.instance.InstanceID)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 0, checkTotalInsThd)

	// evicting实例仅供绑定会话的申请租约请求使用，不干涉扩缩容逻辑，因此无法pop该实例
	popIns5 := bcs.popInstanceElement(forward, func(element *instanceElement) bool { return true }, false)
	assert.Nil(t, popIns5)
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkInUsedInsThd)
	assert.Equal(t, 0, checkTotalInsThd)
}

func TestSignalAllInstances(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance_evicting",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	insIDList := make([]string, 0, 3)
	bcs.SignalAllInstances(func(instance *types.Instance) {
		insIDList = append(insIDList, instance.InstanceID)
	})
	assert.Contains(t, insIDList, "instance1")
	assert.Contains(t, insIDList, "instance2")
	// evicting实例可能还需要给会话请求使用，因此仍然需要被signal
	assert.Contains(t, insIDList, "instance_evicting")
}

func TestHandleInstanceUpdate(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkAvailInsThd := 0
	checkTotalInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	_, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.Equal(t, scheduler.ErrInsSubHealthy, err)
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	selfregister.GlobalSchedulerProxy.Add(&commonTypes.InstanceInfo{InstanceName: "scheduler1"}, "")
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:        "instance3",
		ConcurrentNum:     2,
		ResKey:            resspeckey.ResSpecKey{},
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		CreateSchedulerID: "scheduler1",
		Permanent:         true,
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:        "instance3",
		ConcurrentNum:     2,
		ResKey:            resspeckey.ResSpecKey{},
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
		CreateSchedulerID: "scheduler1",
		Permanent:         true,
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:        "instance3",
		ConcurrentNum:     2,
		ResKey:            resspeckey.ResSpecKey{},
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		CreateSchedulerID: "scheduler1",
		Permanent:         true,
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
}

func TestHandleInstanceUpdate_withEvictingInstance(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.isFuncOwner = true
	checkAvailInsThd := 0
	checkTotalInsThd := 0
	checkInUseInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 4, checkTotalInsThd)
	assert.Equal(t, 0, checkInUseInsThd)

	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	assert.Equal(t, 0, checkInUseInsThd)

	_, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.NotNil(t, err)

	assert.Equal(t, 2, checkAvailInsThd)
	assert.Equal(t, 2, checkTotalInsThd)
	assert.Equal(t, 0, checkInUseInsThd)

	bcs.HandleInstanceUpdate(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	})
	assert.Equal(t, 0, checkAvailInsThd)
	assert.Equal(t, 0, checkTotalInsThd)
	assert.Equal(t, 0, checkInUseInsThd)
}

func Test_basicConcurrencyScheduler_ReassignInstance(t *testing.T) {
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	checkAvailInsThd := 0
	checkTotalInsThd := 0
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	instance1 := &types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	instance2 := &types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	}
	instance_evicting := &types.Instance{
		InstanceID:     "instance_evicting",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
	}
	convey.Convey("test HandleFuncOwnerUpdate", t, func() {
		convey.Convey("become owner", func() {
			checkAvailInsThd = 0
			checkTotalInsThd = 0
			bcs.isFuncOwner = false
			bcs.AddInstance(instance1)
			bcs.AddInstance(instance2)
			bcs.AddInstance(instance_evicting)
			defer bcs.DelInstance(instance_evicting)
			bcs.HandleFuncOwnerUpdate(true)
			assert.Equal(t, 2, checkAvailInsThd)
			assert.Equal(t, 4, checkTotalInsThd)
			assert.True(t, bcs.selfInstanceQueue.GetByID(instance_evicting.InstanceID) != nil)
			assert.True(t, bcs.otherInstanceQueue.GetByID(instance_evicting.InstanceID) != nil)
			bcs.DelInstance(instance1)
			bcs.DelInstance(instance2)
		})
		convey.Convey("resign owner", func() {
			checkAvailInsThd = 0
			checkTotalInsThd = 0
			bcs.isFuncOwner = true
			bcs.AddInstance(instance1)
			bcs.AddInstance(instance2)
			bcs.AddInstance(instance_evicting)
			defer bcs.DelInstance(instance_evicting)
			bcs.HandleFuncOwnerUpdate(false)
			assert.Equal(t, 0, checkAvailInsThd)
			assert.Equal(t, 0, checkTotalInsThd)
			assert.True(t, bcs.selfInstanceQueue.GetByID(instance_evicting.InstanceID) != nil)
			assert.True(t, bcs.otherInstanceQueue.GetByID(instance_evicting.InstanceID) != nil)
			bcs.DelInstance(instance1)
			bcs.DelInstance(instance2)
		})
		convey.Convey("no change", func() {
			checkAvailInsThd = 0
			checkTotalInsThd = 0
			bcs.isFuncOwner = true
			bcs.AddInstance(instance1)
			bcs.AddInstance(instance2)
			bcs.AddInstance(instance_evicting)
			defer bcs.DelInstance(instance_evicting)
			bcs.HandleFuncOwnerUpdate(true)
			assert.Equal(t, 2, checkAvailInsThd)
			assert.Equal(t, 4, checkTotalInsThd)
			assert.True(t, bcs.selfInstanceQueue.GetByID(instance_evicting.InstanceID) != nil)
			bcs.DelInstance(instance1)
			bcs.DelInstance(instance2)
		})
	})
}

func Test_basicConcurrencyScheduler_scheduleRequest(t *testing.T) {
	convey.Convey("test scheduleRequest", t, func() {
		convey.Convey("baseline", func() {
			bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
				FuncKey:          "testFunction",
				InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
			}, resspeckey.ResSpecKey{}, "",
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
			bcs.isFuncOwner = false
			p := gomonkey.ApplyFunc((*basicConcurrencyScheduler).acquireInstanceInternal,
				func(_ *basicConcurrencyScheduler,
					queue queue.Queue, request *types.InstanceAcquireRequest) (*types.InstanceAllocation, error) {
					return &types.InstanceAllocation{
						Instance: &types.Instance{
							InstanceType: "bbb",
							InstanceID:   "ccc",
						},
						AllocationID: "aaa",
					}, nil
				})
			defer p.Reset()
			insAlloc, err := bcs.scheduleRequest(&types.InstanceAcquireRequest{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc.Instance.InstanceID, convey.ShouldEqual, "ccc")
		})
		convey.Convey("acquire failed", func() {
			bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
				FuncKey:          "testFunction",
				InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
			}, resspeckey.ResSpecKey{}, "",
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
			bcs.isFuncOwner = false
			p := gomonkey.ApplyFunc((*basicConcurrencyScheduler).acquireInstanceInternal,
				func(_ *basicConcurrencyScheduler,
					queue queue.Queue, request *types.InstanceAcquireRequest) (*types.InstanceAllocation, error) {
					return nil, fmt.Errorf("error")
				})
			defer p.Reset()
			insAlloc, err := bcs.scheduleRequest(&types.InstanceAcquireRequest{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(insAlloc, convey.ShouldBeNil)
		})
		convey.Convey("session bind", func() {
			bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
				FuncKey:          "testFunction",
				InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 1},
			}, resspeckey.ResSpecKey{}, "",
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
				queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
			bcs.isFuncOwner = true
			bcs.AddInstance(&types.Instance{
				InstanceID:    "instance1",
				ConcurrentNum: 1,
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusRunning),
				},
			})
			insAcqReq := &types.InstanceAcquireRequest{
				InstanceSession: commonTypes.InstanceSessionConfig{
					SessionID:   "123",
					SessionTTL:  10,
					Concurrency: 1,
				},
			}
			insAlloc1, err := bcs.AcquireInstance(insAcqReq)
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc1.Instance.InstanceID, convey.ShouldEqual, "instance1")
			_, err = bcs.scheduleRequest(insAcqReq)
			convey.So(err, convey.ShouldNotBeNil)
			err = insAlloc1.Lease.Release()
			convey.So(err, convey.ShouldBeNil)
			insAlloc2, err := bcs.scheduleRequest(insAcqReq)
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc2.Instance.InstanceID, convey.ShouldEqual, "instance1")
		})
	})
}

// 测试初始10个instance，分配、更新ratio后分配，两个scheduler self和other队列对应相等
func TestReassignInstancesGray(t *testing.T) {
	mainScheduler := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	mainScheduler.isFuncOwner = true

	grayScheduler := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	grayScheduler.isFuncOwner = true

	for i := 1; i <= 10; i++ {
		instance := &types.Instance{
			InstanceID:     fmt.Sprintf("instance%d", i),
			ConcurrentNum:  i,
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		}
		assert.NoError(t, mainScheduler.AddInstance(instance))
		assert.NoError(t, grayScheduler.AddInstance(instance))
	}

	config.GlobalConfig.EnableRollout = true
	selfregister.IsRollingOut = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
		selfregister.IsRollingOut = false
	}()

	selfregister.IsRolloutObject = false
	mainScheduler.ReassignInstanceWhenGray(50)
	assert.Equal(t, 5, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 5, mainScheduler.otherInstanceQueue.Len())

	selfregister.IsRolloutObject = true
	grayScheduler.ReassignInstanceWhenGray(50)
	assert.Equal(t, mainScheduler.selfInstanceQueue.Len(), grayScheduler.otherInstanceQueue.Len())
	mainScheduler.selfInstanceQueue.Range(func(obj interface{}) bool {
		insElem, _ := obj.(*instanceElement)
		insElemIn2 := grayScheduler.otherInstanceQueue.GetByID(insElem.instance.InstanceID)
		assert.NotNil(t, insElemIn2)
		return true
	})

	selfregister.IsRolloutObject = false
	mainScheduler.ReassignInstanceWhenGray(70)
	assert.Equal(t, 3, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 7, mainScheduler.otherInstanceQueue.Len())

	selfregister.IsRolloutObject = true
	grayScheduler.ReassignInstanceWhenGray(70)
	assert.Equal(t, 7, grayScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 3, grayScheduler.otherInstanceQueue.Len())
}

// 10个节点10%灰度 9个节点到10个节点删除、增加触发重分配
func TestReassignInstancesGrayWhenAddOrDelReassign(t *testing.T) {
	mainScheduler := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	mainScheduler.isFuncOwner = true

	grayScheduler := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	grayScheduler.isFuncOwner = true
	// 当前(9,0)
	for i := 1; i <= 9; i++ {
		instance := &types.Instance{
			InstanceID:     fmt.Sprintf("instance%d", i),
			ConcurrentNum:  2,
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		}
		assert.NoError(t, mainScheduler.AddInstance(instance))
		assert.NoError(t, grayScheduler.AddInstance(instance))
	}
	config.GlobalConfig.EnableRollout = true
	selfregister.IsRollingOut = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
		selfregister.IsRollingOut = false
	}()
	// main重分配
	selfregister.IsRolloutObject = false
	mainScheduler.ReassignInstanceWhenGray(10)

	assert.Equal(t, 9, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 0, mainScheduler.otherInstanceQueue.Len())
	// gray重分配
	selfregister.IsRolloutObject = true
	grayScheduler.ReassignInstanceWhenGray(10)

	// 2个sc各再加入一个 判断先加入了self (10,0) 应该自动变成(9,1)
	instance := &types.Instance{
		InstanceID:     fmt.Sprintf("instance%d", 10),
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}

	instance9 := &types.Instance{
		InstanceID:     fmt.Sprintf("instance%d", 9),
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}

	// main加入
	selfregister.IsRolloutObject = false
	assert.NoError(t, mainScheduler.AddInstance(instance))
	// gray加入
	selfregister.IsRolloutObject = true
	assert.NoError(t, grayScheduler.AddInstance(instance))

	assert.Equal(t, mainScheduler.selfInstanceQueue.Len(), grayScheduler.otherInstanceQueue.Len())
	assert.Equal(t, 9, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 1, mainScheduler.otherInstanceQueue.Len())
	mainScheduler.selfInstanceQueue.Range(func(obj interface{}) bool {
		insElem, _ := obj.(*instanceElement)
		insElemIn2 := grayScheduler.otherInstanceQueue.GetByID(insElem.instance.InstanceID)
		assert.NotNil(t, insElemIn2)
		return true
	})
	//确定9 hash最大被分到了other
	assert.Equal(t, "instance9", mainScheduler.otherInstanceQueue.Front().(*instanceElement).instance.InstanceID)

	// 假设删除9
	selfregister.IsRolloutObject = false
	assert.NoError(t, mainScheduler.DelInstance(&types.Instance{
		InstanceID: "instance9",
	}))
	selfregister.IsRolloutObject = true
	assert.NoError(t, grayScheduler.DelInstance(&types.Instance{
		InstanceID: "instance9",
	}))
	// 这里应该不会触发reassign
	assert.Equal(t, 9, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 0, mainScheduler.otherInstanceQueue.Len())

	// 但是如果加入9，然后删除1，会触发reassign
	// 以update的方式加入
	// main重新加入9 - 不会触发reassign
	selfregister.IsRolloutObject = false
	mainScheduler.HandleInstanceUpdate(instance9)

	// gray重新加入9
	selfregister.IsRolloutObject = true
	grayScheduler.HandleInstanceUpdate(instance9)

	// 验证9还是被分配到other
	assert.Equal(t, "instance9", mainScheduler.otherInstanceQueue.Front().(*instanceElement).instance.InstanceID)

	// 删除1 -应该触发重分配(8,1)->(9,0)
	selfregister.IsRolloutObject = false
	assert.NoError(t, mainScheduler.DelInstance(&types.Instance{
		InstanceID: "instance1",
	}))
	selfregister.IsRolloutObject = true
	assert.NoError(t, grayScheduler.DelInstance(&types.Instance{
		InstanceID: "instance1",
	}))
	assert.Equal(t, 9, mainScheduler.selfInstanceQueue.Len())
	assert.Equal(t, 0, mainScheduler.otherInstanceQueue.Len())
}

// 测试空指针防御
func TestReassignInstancesGrayBothQueuesInitiallyEmpty(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	selfregister.IsRollingOut = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
		selfregister.IsRollingOut = false
	}()
	scheduler1 := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	scheduler1.isFuncOwner = true
	// 空状态下reassign无空指针
	scheduler1.ReassignInstanceWhenGray(50)
	assert.Equal(t, 0, scheduler1.selfInstanceQueue.Len())
	assert.Equal(t, 0, scheduler1.otherInstanceQueue.Len())
	// 测试空状态下增加删除无空指针
	instance := &types.Instance{
		InstanceID:     fmt.Sprintf("instance%d", 10),
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	// 当前是旧sc
	selfregister.IsRolloutObject = false
	scheduler1.HandleInstanceUpdate(instance)
	assert.Equal(t, 1, scheduler1.selfInstanceQueue.Len())
	assert.NoError(t, scheduler1.DelInstance(&types.Instance{
		InstanceID: "instance10",
	}))

	// 测试边界
	scheduler1.HandleInstanceUpdate(instance)
	scheduler1.ReassignInstanceWhenGray(0)
	assert.Equal(t, 1, scheduler1.selfInstanceQueue.Len())

	// 全部灰度后
	scheduler1.ReassignInstanceWhenGray(100)
	// 加入other，已经有了报错
	assert.Error(t, scheduler1.AddInstance(instance))
	assert.Equal(t, 1, scheduler1.otherInstanceQueue.Len())

	// 当前是新sc
	selfregister.IsRolloutObject = true
	scheduler1.HandleInstanceUpdate(instance)
	scheduler1.ReassignInstanceWhenGray(0)
	assert.Error(t, scheduler1.AddInstance(instance))
	assert.Equal(t, 1, scheduler1.otherInstanceQueue.Len())

	// 全部灰度后
	scheduler1.ReassignInstanceWhenGray(100)
	// 加入self，已经有了报错
	assert.Error(t, scheduler1.AddInstance(instance))
	assert.Equal(t, 1, scheduler1.selfInstanceQueue.Len())
}

func TestReassignInstancesGrayWhenFixedOtherInstance(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	selfregister.IsRollingOut = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
		selfregister.IsRollingOut = false
	}()

	scheduler1 := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	instance := &types.Instance{
		InstanceID:        "instance1",
		ConcurrentNum:     2,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	selfregister.IsRolloutObject = false
	assert.NoError(t, scheduler1.AddInstance(instance))
	assert.Equal(t, 0, scheduler1.selfInstanceQueue.Len())
	assert.Equal(t, 1, scheduler1.otherInstanceQueue.Len())

	selfregister.IsRolloutObject = false
	scheduler1.ReassignInstanceWhenGray(50)
	assert.Equal(t, 0, scheduler1.selfInstanceQueue.Len())
	assert.Equal(t, 1, scheduler1.otherInstanceQueue.Len())
}

func TestInstanceQueueWithSubHealthAndEvictingRecord(t *testing.T) {
	convey.Convey("Test instanceQueueWithSubHealthAndEvictingRecord", t, func() {
		// 创建 mock queue 和 mock instanceElement
		newInstanceQueue := func() *instanceQueueWithSubHealthAndEvictingRecord {
			mockQueue := queue.NewFifoQueue(getInstanceID)
			mockSubHealthRecord := make(map[string]*instanceElement)
			mockEvictingRecord := make(map[string]*instanceElement)
			return &instanceQueueWithSubHealthAndEvictingRecord{
				instanceQueue:   mockQueue,
				subHealthRecord: mockSubHealthRecord,
				evictingRecord:  mockEvictingRecord,
			}
		}
		iq := newInstanceQueue()

		insID := "test-instance-id"
		insElem := &instanceElement{
			instance: &types.Instance{
				InstanceID: insID,
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusRunning),
				},
			},
		}

		convey.Convey("PushBack, instance already exists", func() {
			iq = newInstanceQueue()
			iq.subHealthRecord[insID] = insElem
			err := iq.PushBack(insElem)
			convey.So(err, convey.ShouldEqual, scheduler.ErrInsAlreadyExist)
		})

		convey.Convey("PopSubHealth, subHealthRecord is empty", func() {
			iq = newInstanceQueue()
			result := iq.PopSubHealth()
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("GetByID should return instance from subHealthRecord", func() {
			iq = newInstanceQueue()
			iq.subHealthRecord[insID] = insElem
			result := iq.GetByID(insID)
			convey.So(result, convey.ShouldEqual, insElem)
		})

		convey.Convey("DelByID", func() {
			iq = newInstanceQueue()
			iq.PushBack(insElem)

			err := iq.DelByID(insID)
			assert.NoError(t, err)
			assert.False(t, iq.subHealthRecord[insID] != nil)
			assert.False(t, iq.evictingRecord[insID] != nil)
			assert.Nil(t, iq.instanceQueue.GetByID(insID))
		})

		convey.Convey("complex", func() {
			iq = newInstanceQueue()
			err := iq.PushBack(insElem)
			convey.So(err, convey.ShouldBeNil)
			convey.So(iq.instanceQueue.GetByID(insElem.instance.InstanceID) == nil, convey.ShouldBeFalse)
			convey.So(len(iq.subHealthRecord), convey.ShouldEqual, 0)
			convey.So(len(iq.evictingRecord), convey.ShouldEqual, 0)
			convey.So(iq.Len(), convey.ShouldEqual, 1)

			insElem.instance.InstanceStatus.Code = int32(constant.KernelInstanceStatusSubHealth)
			err = iq.UpdateObjByID(insID, insElem)
			convey.So(err, convey.ShouldBeNil)
			convey.So(iq.instanceQueue.GetByID(insElem.instance.InstanceID) == nil, convey.ShouldBeTrue)
			_, ok := iq.subHealthRecord[insID]
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(iq.evictingRecord), convey.ShouldEqual, 0)
			convey.So(iq.Len(), convey.ShouldEqual, 1)

			insElem.instance.InstanceStatus.Code = int32(constant.KernelInstanceStatusEvicting)
			err = iq.UpdateObjByID(insID, insElem)
			convey.So(err, convey.ShouldBeNil)
			convey.So(iq.instanceQueue.GetByID(insElem.instance.InstanceID) == nil, convey.ShouldBeTrue)
			_, ok = iq.evictingRecord[insID]
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(iq.subHealthRecord), convey.ShouldEqual, 0)
			convey.So(iq.Len(), convey.ShouldEqual, 0)
		})
	})
}

// newTestQueue 是一个辅助函数，用于创建一个带有预填充数据的测试实例
func newTestQueue(main, sub, evicting map[string]*instanceElement) *instanceQueueWithSubHealthAndEvictingRecord {
	// 注意：这里假设 instanceQueueWithSubHealthAndEvictingRecord 有一个可以接收 mockInstances 的构造函数
	// 或者它的成员变量可以直接赋值。如果您的实现不同，需要调整这部分。
	// 为了演示，我们假设可以这样初始化：
	instanceQueue := queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance)
	for _, ins := range main {
		instanceQueue.PushBack(ins)
	}
	q := &instanceQueueWithSubHealthAndEvictingRecord{
		// 假设 instanceQueue 是一个实现了 Range 和 SortedRange 的类型
		// 这里我们用一个 mockInstances 来模拟它
		instanceQueue:   instanceQueue,
		subHealthRecord: sub,
		evictingRecord:  evicting,
	}
	return q
}

func getMockInstanceMap(instanceNames []string, statusCode int32) map[string]*instanceElement {
	instanceMap := make(map[string]*instanceElement)
	for _, name := range instanceNames {
		ins := &instanceElement{
			instance: &types.Instance{
				InstanceStatus: commonTypes.InstanceStatus{
					Code: statusCode,
				},
				ResKey:     resspeckey.ResSpecKey{},
				InstanceID: name,
			},
			threadMap: make(map[string]struct{}),
		}
		instanceMap[name] = ins
	}
	return instanceMap
}

// --- Test Cases for Range ---
func TestInstanceQueue_Range(t *testing.T) {
	convey.Convey("TestInstanceQueue_Range", t, func() {
		testInstanceQueueWithSubHealthAndEvictingRecord := newTestQueue(
			getMockInstanceMap([]string{"a", "b", "c", "d", "e"}, int32(constant.KernelInstanceStatusRunning)),
			getMockInstanceMap([]string{"d", "e", "f"}, int32(constant.KernelInstanceStatusSubHealth)),
			getMockInstanceMap([]string{"e", "f", "g"}, int32(constant.KernelInstanceStatusEvicting)))

		var testInstance *instanceElement

		vistedStr := "d visited"
		testInstanceId := "d"
		testFunc := func(obj interface{}) bool {
			ins, ok := obj.(*instanceElement)
			if !ok {
				return true
			}
			ins.instance.FuncKey = vistedStr
			if ins.instance.InstanceID == testInstanceId {
				testInstance = ins
				return false
			}
			return true
		}
		testInstanceQueueWithSubHealthAndEvictingRecord.Range(testFunc)
		convey.So(testInstance, convey.ShouldNotBeNil)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		for _, ins := range testInstanceQueueWithSubHealthAndEvictingRecord.subHealthRecord {
			convey.So(ins.instance.FuncKey, convey.ShouldNotEqual, vistedStr)
		}
		for _, ins := range testInstanceQueueWithSubHealthAndEvictingRecord.evictingRecord {
			convey.So(ins.instance.FuncKey, convey.ShouldNotEqual, vistedStr)
		}

		vistedStr = "d sortrange visited"
		testInstanceQueueWithSubHealthAndEvictingRecord.SortedRange(testFunc)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		for _, ins := range testInstanceQueueWithSubHealthAndEvictingRecord.subHealthRecord {
			convey.So(ins.instance.FuncKey, convey.ShouldNotEqual, vistedStr)
		}
		for _, ins := range testInstanceQueueWithSubHealthAndEvictingRecord.evictingRecord {
			convey.So(ins.instance.FuncKey, convey.ShouldNotEqual, vistedStr)
		}

		testInstanceId = "f"
		testInstanceQueueWithSubHealthAndEvictingRecord.Range(testFunc)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		convey.So(testInstance.instance.InstanceStatus.Code, convey.ShouldEqual, int32(constant.KernelInstanceStatusSubHealth))
		testInstance = nil
		testInstanceQueueWithSubHealthAndEvictingRecord.SortedRange(testFunc)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		convey.So(testInstance.instance.InstanceStatus.Code, convey.ShouldEqual, int32(constant.KernelInstanceStatusSubHealth))

		testInstanceId = "g"
		testInstanceQueueWithSubHealthAndEvictingRecord.Range(testFunc)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		convey.So(testInstance.instance.InstanceStatus.Code, convey.ShouldEqual, int32(constant.KernelInstanceStatusEvicting))
		testInstance = nil
		testInstanceQueueWithSubHealthAndEvictingRecord.SortedRange(testFunc)
		convey.So(testInstance.instance.InstanceID, convey.ShouldEqual, testInstanceId)
		convey.So(testInstance.instance.InstanceStatus.Code, convey.ShouldEqual, int32(constant.KernelInstanceStatusEvicting))
	})
}

func TestRecoverSessionRecordFromDataSystem(t *testing.T) {
	config.GlobalConfig.EnableSessionRecover = true
	defer gomonkey.ApplyFunc(datasystemclient.KVGetWithRetry, func(key string, option *datasystemclient.Option, traceID string) ([]byte, error) {
		return []byte("{\"38b03220-7f67-473e-8000-000000000030\":[{\"sessionID\":\"bbbbb\",\"sessionTTL\":3600,\"concurrency\":3},{\"sessionID\":\"ccccc\",\"sessionTTL\":3600,\"concurrency\":3}]}"), nil
	}).Reset()

	config.GlobalConfig.DataSystemConfig = types.DataSystemConfig{
		CurrentCluster:  "",
		UploadWriteMode: 0,
		UploadTTLSec:    10,
	}
	sc := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	sc.HandleFuncOwnerUpdate(true)
	instance := &types.Instance{
		InstanceID:        "38b03220-7f67-473e-8000-000000000030",
		ConcurrentNum:     6,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	assert.NoError(t, sc.AddInstance(instance))
	i := 0
	defer gomonkey.ApplyFunc(datasystemclient.KVPutWithRetry, func(key string, value []byte, option *datasystemclient.Option, traceID string) error {
		i++
		return fmt.Errorf("put failed")
	}).Reset()
	defer gomonkey.ApplyFunc(datasystemclient.KVDelWithRetry, func(key string, option *datasystemclient.Option, traceID string) error {
		i++
		return fmt.Errorf("del failed")
	}).Reset()
	sc.RecoverSessionRecordFromDataSystem(func(sessInfo *types.SessionInfo, instance *types.Instance) {})
	assert.Equal(t, 2, len(sc.sessionManager.sessionMap))
	sc.sessionManager.delSession("bbbbb")
	sc.sessionManager.delSession("ccccc")
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 4, i)
}

func TestRecoverSessionRecordFromDataSystemInGray(t *testing.T) {
	config.GlobalConfig.EnableSessionRecover = true
	defer gomonkey.ApplyFunc(datasystemclient.KVGetWithRetry, func(key string, option *datasystemclient.Option, traceID string) ([]byte, error) {
		return []byte("{\"38b03220-7f67-473e-8000-000000000030\":[{\"SchedulerID\":\"1.1.1.1\",\"sessionID\":\"bbbbb\",\"sessionTTL\":3600,\"concurrency\":3},{\"SchedulerID\":\"2.2.2.2\",\"sessionID\":\"ccccc\",\"sessionTTL\":3600,\"concurrency\":3}]}"), nil
	}).Reset()
	delCnt := 0
	putCnt := 0
	defer gomonkey.ApplyFunc(datasystemclient.KVPutWithRetry, func(key string, value []byte, option *datasystemclient.Option, traceID string) error {
		putCnt++
		return nil
	}).Reset()
	defer gomonkey.ApplyFunc(datasystemclient.KVDelWithRetry, func(key string, option *datasystemclient.Option, traceID string) error {
		delCnt++
		return nil
	}).Reset()
	selfregister.IsRollingOut = true
	defer func() {
		selfregister.IsRollingOut = false
	}()
	config.GlobalConfig.DataSystemConfig = types.DataSystemConfig{
		CurrentCluster:  "",
		UploadWriteMode: 0,
		UploadTTLSec:    10,
	}
	os.Setenv("POD_IP", "1.1.1.1")
	defer os.Unsetenv("POD_IP")
	sc := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	sc.HandleFuncOwnerUpdate(true)
	instance := &types.Instance{
		InstanceID:        "38b03220-7f67-473e-8000-000000000030",
		ConcurrentNum:     6,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	assert.NoError(t, sc.AddInstance(instance))
	sc.RecoverSessionRecordFromDataSystem(func(sessInfo *types.SessionInfo, instance *types.Instance) {})
	assert.Equal(t, 1, len(sc.sessionManager.sessionMap))
	_, exist := sc.sessionManager.getSession("bbbbb")
	assert.Equal(t, exist, true)
	_, exist = sc.sessionManager.getSession("ccccc")
	assert.Equal(t, exist, false)
	insId := sc.sessionManager.queryInsBySessionFromDS("ccccc")
	assert.Equal(t, insId, "38b03220-7f67-473e-8000-000000000030")
	insId = sc.sessionManager.queryInsBySessionFromDS("aaaaa")
	assert.Equal(t, insId, "")
	sc.sessionManager.triggerDeleteSessionRecord()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, delCnt, 0)
	assert.Equal(t, putCnt, 2)
}

func TestRecoverSessionRecordSwitch(t *testing.T) {
	config.GlobalConfig.EnableSessionRecover = false
	sc := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	sc.HandleFuncOwnerUpdate(true)
	instance := &types.Instance{
		InstanceID:        "38b03220-7f67-473e-8000-000000000030",
		ConcurrentNum:     6,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	assert.NoError(t, sc.AddInstance(instance))
	i := 0
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		i++
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		i++
		return
	}).Reset()
	sc.sessionManager.triggerSaveSessionRecord()
	sc.sessionManager.triggerSaveSessionRecord()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, i)
}

func TestSaveSessionRecord(t *testing.T) {
	config.GlobalConfig.EnableSessionRecover = true
	sc := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 2},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	sc.HandleFuncOwnerUpdate(true)
	instance := &types.Instance{
		InstanceID:        "38b03220-7f67-473e-8000-000000000030",
		ConcurrentNum:     6,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	assert.NoError(t, sc.AddInstance(instance))

	i := 0
	defer gomonkey.ApplyFunc(datasystemclient.KVPutWithRetry, func(key string, value []byte, option *datasystemclient.Option, traceID string) error {
		i++
		assert.Equal(t, "sessioncache--6f273ad8e3999bbc", key)
		return nil
	}).Reset()
	sc.sessionManager.addSession("sessionid", &sessionRecord{
		ttl:         0,
		concurrency: 0,
		sessionID:   "sessionid",
		insElem: &instanceElement{
			instance: instance,
		},
	})
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, i)
}

func TestAcquireInstanceWithSessionFullConcurrency(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.HandleFuncOwnerUpdate(true)
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: -1,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns.Instance.InstanceID)
	assert.Equal(t, 4, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	record, exist := bcs.sessionManager.getSession("session1")
	assert.True(t, exist)
	assert.Equal(t, 4, record.concurrency)
}

func TestAcquireInstanceWithSessionFullConcurrencyInsufficient(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.HandleFuncOwnerUpdate(true)
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: -1,
		},
	})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	assert.Nil(t, acqIns2)
}

func TestAcquireReleaseSessionFullConcurrency(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.HandleFuncOwnerUpdate(true)
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: -1,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	record, exist := bcs.sessionManager.getSession("session1")
	assert.True(t, exist)
	assert.Equal(t, 3, len(record.availThdMap))
	acqIns2, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			Concurrency: -1,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	assert.Equal(t, 2, len(record.availThdMap))
	err = bcs.ReleaseInstance(acqIns1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(record.availThdMap))
	err = bcs.ReleaseInstance(acqIns2)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(record.availThdMap))
}

func TestAcquireInstanceWithSessionFullConcurrencyMonopolyChoosesFullyIdleInstance(t *testing.T) {
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.HandleFuncOwnerUpdate(true)
	assert.NoError(t, bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}))
	assert.NoError(t, bcs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}))
	// Occupy one thread on instance2 so it is no longer fully idle.
	normalAlloc, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.NoError(t, err)
	assert.Equal(t, "instance2", normalAlloc.Instance.InstanceID)
	fullAlloc, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session-full",
			Concurrency: -1,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "instance1", fullAlloc.Instance.InstanceID)
	record, exist := bcs.sessionManager.getSession("session-full")
	assert.True(t, exist)
	assert.Equal(t, 4, record.concurrency)
	assert.Equal(t, 4, len(record.allocThdMap))
}

func TestSessionFullConcurrencyTTLExpire(t *testing.T) {
	mockTimer := time.NewTimer(100 * time.Millisecond)
	defer gomonkey.ApplyFunc(time.NewTimer, func(d time.Duration) *time.Timer {
		mockTimer.Reset(100 * time.Millisecond)
		return mockTimer
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).saveSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).deleteSessionRecordToDataSystem, func(_ *sessionManager) {
		return
	}).Reset()
	defer gomonkey.ApplyFunc((*sessionManager).loadSessionFromDataSystem, func(_ *sessionManager) map[string][]commonTypes.InstanceSessionConfig {
		return map[string][]commonTypes.InstanceSessionConfig{}
	}).Reset()
	bcs := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	bcs.HandleFuncOwnerUpdate(true)
	checkInUseInsThd := 0
	checkAvailInsThd := 0
	bcs.addObservers(scheduler.InUseInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkInUseInsThd += delta
	})
	bcs.addObservers(scheduler.AvailInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkAvailInsThd += delta
	})
	bcs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  4,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns, err := bcs.AcquireInstance(&types.InstanceAcquireRequest{
		InstanceSession: commonTypes.InstanceSessionConfig{
			SessionID:   "session1",
			SessionTTL:  1,
			Concurrency: -1,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns.Instance.InstanceID)
	assert.Equal(t, 4, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	err = bcs.ReleaseInstance(acqIns)
	assert.Nil(t, err)
	assert.Equal(t, 4, checkInUseInsThd)
	assert.Equal(t, 0, checkAvailInsThd)
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, 0, checkInUseInsThd)
	assert.Equal(t, 4, checkAvailInsThd)
	_, exist := bcs.sessionManager.getSession("session1")
	assert.False(t, exist)
}

func TestRecoverSessionRecordFullConcurrency(t *testing.T) {
	config.GlobalConfig.EnableSessionRecover = true
	defer gomonkey.ApplyFunc(datasystemclient.KVGetWithRetry, func(key string, option *datasystemclient.Option, traceID string) ([]byte, error) {
		return []byte("{\"instance1\":[{\"sessionID\":\"session1\",\"sessionTTL\":3600,\"concurrency\":4}]}"), nil
	}).Reset()
	config.GlobalConfig.DataSystemConfig = types.DataSystemConfig{
		CurrentCluster:  "",
		UploadWriteMode: 0,
		UploadTTLSec:    10,
	}
	sc := newBasicConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 4},
	}, resspeckey.ResSpecKey{}, "",
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance),
		queue.NewPriorityQueue(getInstanceID, priorityFuncForReservedInstance))
	sc.HandleFuncOwnerUpdate(true)
	instance := &types.Instance{
		InstanceID:        "instance1",
		ConcurrentNum:     4,
		ResKey:            resspeckey.ResSpecKey{},
		Permanent:         true,
		CreateSchedulerID: "abc",
		InstanceStatus:    commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	}
	assert.NoError(t, sc.AddInstance(instance))
	defer gomonkey.ApplyFunc(datasystemclient.KVPutWithRetry, func(key string, value []byte, option *datasystemclient.Option, traceID string) error {
		return nil
	}).Reset()
	defer gomonkey.ApplyFunc(datasystemclient.KVDelWithRetry, func(key string, option *datasystemclient.Option, traceID string) error {
		return nil
	}).Reset()
	sc.RecoverSessionRecordFromDataSystem(func(sessInfo *types.SessionInfo, instance *types.Instance) {})
	assert.Equal(t, 1, len(sc.sessionManager.sessionMap))
	record, exist := sc.sessionManager.getSession("session1")
	assert.True(t, exist)
	assert.Equal(t, 4, record.concurrency)
	assert.Equal(t, 4, len(record.allocThdMap))
	sc.sessionManager.delSession("session1")
}
