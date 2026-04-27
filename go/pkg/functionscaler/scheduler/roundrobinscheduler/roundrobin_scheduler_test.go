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

// Package roundrobinscheduler -
package roundrobinscheduler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

type fakeInstanceScaler struct {
	timer          *time.Timer
	expectInsNum   int
	inUseInsThdNum int
	totalInsThdNum int
	scaling        bool
	createErr      error
	scaleUpFunc    func()
}

func (f *fakeInstanceScaler) SetFuncOwner(isManaged bool) {
}

func (f *fakeInstanceScaler) SetEnable(enable bool) {
}

func (f *fakeInstanceScaler) TriggerScale() {
	go func() {
		time.Sleep(10 * time.Millisecond)
		f.scaleUpFunc()
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
	f.inUseInsThdNum += inUseInsThdDiff
	f.totalInsThdNum += totalInsThdDiff
}

func (f *fakeInstanceScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
}

func (f *fakeInstanceScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
}

func (f *fakeInstanceScaler) HandleCreateError(createError error) {
	f.createErr = createError
}

func (f *fakeInstanceScaler) GetExpectInstanceNumber() int {
	return f.expectInsNum
}

func (f *fakeInstanceScaler) Destroy() {
}

func TestNewRoundRobinScheduler(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	assert.NotNil(t, rs)
}

func TestGetInstanceNumber(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	rs.AddInstance(&types.Instance{
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
		ResKey:         resspeckey.ResSpecKey{},
	})
	assert.Equal(t, 1, rs.GetInstanceNumber(true))
}

func TestAcquireInstance(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	fs := &fakeInstanceScaler{}
	rs.ConnectWithInstanceScaler(fs)
	_, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	_, err = rs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	fs.expectInsNum = 3
	acqIns1, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns2.Instance.InstanceID)
	acqIns3, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance3", acqIns3.Instance.InstanceID)
	acqIns4, err := rs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance1"})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns4.Instance.InstanceID)
	m := make(map[string]int, 0)
	for i := 0; i < 10000; i++ {
		acqIns, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
		assert.Nil(t, err)
		_, ok := m[acqIns.AllocationID]
		if ok {
			assert.Errorf(t, fmt.Errorf("task repeat, AllocationID: %s", acqIns.AllocationID), "")
			return
		}
		m[acqIns.AllocationID] = 1
	}
}

func TestAddInstance(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond).(*RoundRobinScheduler)
	checkTotalInsThd := 0
	rs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	err := rs.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 1,
		ResKey:        resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInternal, err)
	err = rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, checkTotalInsThd)
	err = rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Equal(t, scheduler.ErrInsAlreadyExist, err)
	assert.Equal(t, 1, checkTotalInsThd)
	err = rs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, checkTotalInsThd)
	err = rs.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, checkTotalInsThd)
	assert.Equal(t, 1, len(rs.subHealthInstance))
	assert.Equal(t, 2, len(rs.instanceQueue))
}

func TestPopInstance(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond).(*RoundRobinScheduler)
	checkTotalInsThd := 0
	rs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	popIns1 := rs.PopInstance(false)
	assert.Nil(t, popIns1)
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	popIns2 := rs.PopInstance(false)
	assert.Equal(t, "instance3", popIns2.InstanceID)
	assert.Equal(t, 2, checkTotalInsThd)
	popIns3 := rs.PopInstance(false)
	assert.Equal(t, "instance2", popIns3.InstanceID)
	assert.Equal(t, 1, checkTotalInsThd)
	popIns4 := rs.PopInstance(false)
	assert.Equal(t, "instance1", popIns4.InstanceID)
	assert.Equal(t, 0, checkTotalInsThd)
}

func TestDelInstance(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond).(*RoundRobinScheduler)
	checkTotalInsThd := 0
	rs.addObservers(scheduler.TotalInsThdTopic, func(obj interface{}) {
		delta := obj.(int)
		checkTotalInsThd += delta
	})
	err := rs.DelInstance(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusSubHealth)},
	})
	err = rs.DelInstance(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, checkTotalInsThd)
	err = rs.DelInstance(&types.Instance{
		InstanceID: "instance2",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, checkTotalInsThd)
}

func TestHandleInstanceUpdate(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	fs := &fakeInstanceScaler{}
	rs.ConnectWithInstanceScaler(fs)
	rs.HandleInstanceUpdate(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
		},
	})
	rs.HandleInstanceUpdate(&types.Instance{
		InstanceID: "instance2",
		ResKey:     resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusSubHealth),
		},
	})
	fs.expectInsNum = 2
	acqIns1, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns2.Instance.InstanceID)
	_, err = rs.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "instance2"})
	assert.Equal(t, scheduler.ErrInsSubHealthy, err)
	rs.HandleInstanceUpdate(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusSubHealth),
		},
	})
	rs.HandleInstanceUpdate(&types.Instance{
		InstanceID: "instance2",
		ResKey:     resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
		},
	})
	acqIns3, err := rs.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns3.Instance.InstanceID)
}

func TestConnectWithInstanceScaler(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	instanceScaler := &fakeInstanceScaler{}
	rs.ConnectWithInstanceScaler(instanceScaler)
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  2,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Equal(t, 2, instanceScaler.totalInsThdNum)
	someErr := errors.New("some error")
	rs.HandleCreateError(someErr)
	assert.Equal(t, someErr, instanceScaler.createErr)
}

func TestSignalAllInstances(t *testing.T) {
	rs := NewRoundRobinScheduler("testFunction", true, 10*time.Millisecond)
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	rs.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	insIDList := make([]string, 0, 2)
	rs.SignalAllInstances(func(instance *types.Instance) {
		insIDList = append(insIDList, instance.InstanceID)
	})
	assert.Contains(t, insIDList, "instance1")
	assert.Contains(t, insIDList, "instance2")
}

func TestRoundRobinScheduler_CheckInstanceExist(t *testing.T) {
	tests := []struct {
		name          string
		healthyQueue  []*types.Instance
		subHealthMap  map[string]*types.Instance
		checkInstance *types.Instance
		expectExist   bool
	}{
		{
			name: "instance_exists_in_healthy_queue",
			healthyQueue: []*types.Instance{
				{InstanceID: "instance-1"},
				{InstanceID: "instance-2"},
			},
			subHealthMap:  make(map[string]*types.Instance),
			checkInstance: &types.Instance{InstanceID: "instance-1"},
			expectExist:   true,
		},
		{
			name: "instance_exists_in_subhealth_map",
			healthyQueue: []*types.Instance{
				{InstanceID: "instance-2"},
			},
			subHealthMap: map[string]*types.Instance{
				"instance-1": {InstanceID: "instance-1"},
			},
			checkInstance: &types.Instance{InstanceID: "instance-1"},
			expectExist:   true,
		},
		{
			name: "instance_exists_in_both",
			healthyQueue: []*types.Instance{
				{InstanceID: "instance-1"},
			},
			subHealthMap: map[string]*types.Instance{
				"instance-1": {InstanceID: "instance-1"},
			},
			checkInstance: &types.Instance{InstanceID: "instance-1"},
			expectExist:   true,
		},
		{
			name: "instance_not_exist",
			healthyQueue: []*types.Instance{
				{InstanceID: "instance-2"},
			},
			subHealthMap: map[string]*types.Instance{
				"instance-3": {InstanceID: "instance-3"},
			},
			checkInstance: &types.Instance{InstanceID: "instance-1"},
			expectExist:   false,
		},
		{
			name:          "empty_scheduler",
			healthyQueue:  []*types.Instance{},
			subHealthMap:  make(map[string]*types.Instance),
			checkInstance: &types.Instance{InstanceID: "instance-1"},
			expectExist:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RoundRobinScheduler{
				instanceQueue:     tt.healthyQueue,
				subHealthInstance: tt.subHealthMap,
			}

			exist := rs.CheckInstanceExist(tt.checkInstance)

			assert.Equal(t, tt.expectExist, exist,
				"CheckInstanceExist() result mismatch for case: %s", tt.name)
		})
	}
}
