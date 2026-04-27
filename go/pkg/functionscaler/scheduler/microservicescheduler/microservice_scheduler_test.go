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

// Package leastconnectionscheduler -
package microservicescheduler

import (
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

func (f *fakeInstanceScaler) SetFuncOwner(isManaged bool) {

}

func TestNewMicroServiceScheduler(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin)
	assert.NotNil(t, ms)
}

func TestGetInstanceNumber(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin)
	ms.AddInstance(&types.Instance{
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		ResKey:         resspeckey.ResSpecKey{},
	})
	assert.Equal(t, 1, ms.GetInstanceNumber(true))
}

func TestAcquireInstanceWithRR(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin)
	_, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns2.Instance.InstanceID)
	acqIns3, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance3", acqIns3.Instance.InstanceID)
	acqIns4, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns4.Instance.InstanceID)
	acqIns5, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns5.Instance.InstanceID)
	acqIns6, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance3", acqIns6.Instance.InstanceID)
}

func TestAcquireInstanceWithLeastConnections(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", LeastConnections)
	_, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Equal(t, scheduler.ErrNoInsAvailable, err)
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance3",
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	acqIns1, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
	acqIns2, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns2.Instance.InstanceID)
	acqIns3, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance3", acqIns3.Instance.InstanceID)
	ms.ReleaseInstance(acqIns3)
	acqIns4, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance3", acqIns4.Instance.InstanceID)
	ms.ReleaseInstance(acqIns2)
	acqIns5, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance2", acqIns5.Instance.InstanceID)
	ms.ReleaseInstance(acqIns1)
	acqIns6, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns6.Instance.InstanceID)
}

func TestAddInstance(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin).(*MicroServiceScheduler)
	err := ms.AddInstance(&types.Instance{
		InstanceID:    "instance1",
		ConcurrentNum: 1,
		ResKey:        resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInternal, err)
	err = ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
	err = ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Equal(t, scheduler.ErrInsAlreadyExist, err)
	err = ms.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	assert.Nil(t, err)
}

func TestPopInstance(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin).(*MicroServiceScheduler)
	popIns1 := ms.PopInstance(false)
	assert.Nil(t, popIns1)
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	popIns3 := ms.PopInstance(false)
	assert.Equal(t, "instance2", popIns3.InstanceID)
	popIns4 := ms.PopInstance(false)
	assert.Equal(t, "instance1", popIns4.InstanceID)
}

func TestDelInstance(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin).(*MicroServiceScheduler)
	err := ms.DelInstance(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Equal(t, scheduler.ErrInsNotExist, err)
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance2",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})
	err = ms.DelInstance(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
	err = ms.DelInstance(&types.Instance{
		InstanceID: "instance2",
		ResKey:     resspeckey.ResSpecKey{},
	})
	assert.Nil(t, err)
}

func TestHandleInstanceUpdate(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin)
	fs := &fakeInstanceScaler{}
	ms.ConnectWithInstanceScaler(fs)
	ms.HandleInstanceUpdate(&types.Instance{
		InstanceID: "instance1",
		ResKey:     resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
		},
	})
	acqIns1, err := ms.AcquireInstance(&types.InstanceAcquireRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "instance1", acqIns1.Instance.InstanceID)
}

func TestCheckInstanceExist(t *testing.T) {
	ms := NewMicroServiceScheduler("testFunction", RoundRobin).(*MicroServiceScheduler)
	instance := &types.Instance{
		InstanceID: "instance1",
	}
	// 测试实例不存在的情况
	assert.False(t, ms.CheckInstanceExist(instance))
	ms.AddInstance(&types.Instance{
		InstanceID:     "instance1",
		ConcurrentNum:  1,
		ResKey:         resspeckey.ResSpecKey{},
		InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
	})

	// 测试实例存在的情况
	assert.True(t, ms.CheckInstanceExist(instance))
}
