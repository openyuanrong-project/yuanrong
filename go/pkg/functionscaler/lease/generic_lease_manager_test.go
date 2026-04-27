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

// Package lease -
package lease

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestInstanceLeaseHolder(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	instance := &types.Instance{
		InstanceID:    "test-instance",
		ConcurrentNum: 100,
	}
	leaseHolder := newInstanceLeaseHolder(instance)
	assert.Equal(t, true, leaseHolder.enable)
	thread1 := &types.InstanceAllocation{
		Instance:     instance,
		AllocationID: fmt.Sprintf("%s-1", instance.InstanceID),
	}
	thread1Released := false
	err := leaseHolder.createLease(thread1, 500*time.Millisecond, func() {
		thread1Released = true
	})
	assert.Nil(t, err)
	time.Sleep(300 * time.Millisecond)
	err = leaseHolder.extendLease(thread1)
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, false, thread1Released)
	time.Sleep(600 * time.Millisecond)
	assert.Equal(t, true, thread1Released)
	err = leaseHolder.releaseLease(thread1)
	assert.Nil(t, err)
}

func TestHandleInstanceDelete(t *testing.T) {
	convey.Convey("HandleInstanceDelete", t, func() {
		manager := NewGenericLeaseManager("funcKey")
		instance := &types.Instance{InstanceID: "instanceID"}
		thread := &types.InstanceAllocation{Instance: instance}
		lease, err := manager.CreateInstanceLease(nil, 5*time.Second, func() {
		})
		convey.So(err, convey.ShouldNotBeNil)
		lease, err = manager.CreateInstanceLease(thread, 5*time.Second, func() {
		})
		convey.So(err, convey.ShouldBeNil)
		convey.So(lease, convey.ShouldNotBeNil)
		manager.HandleInstanceDelete(instance)
	})
}

func TestCleanAllLeases(t *testing.T) {
	convey.Convey("CleanAllLeases", t, func() {
		manager := NewGenericLeaseManager("funcKey")
		instance := &types.Instance{InstanceID: "instanceID"}
		thread := &types.InstanceAllocation{Instance: instance}
		lease, err := manager.CreateInstanceLease(thread, 5*time.Second, func() {
		})
		convey.So(err, convey.ShouldBeNil)
		convey.So(lease, convey.ShouldNotBeNil)
		manager.CleanAllLeases()
	})
}

func TestGenericInstanceLeaseManager(t *testing.T) {
	lm := &GenericInstanceLeaseManager{leaseHolders: make(map[string]*instanceLeaseHolder)}
	instance := &types.Instance{InstanceID: "instanceID"}
	insAlloc := &types.InstanceAllocation{Instance: instance}
	lm.leaseHolders["instance1"] = newInstanceLeaseHolder(instance)
	convey.Convey("Test GenericInstanceLeaseManager", t, func() {
		convey.Convey("test extendLease", func() {
			lm.CreateInstanceLease(insAlloc, 1*time.Second, func() {})
			err := lm.extendLease(insAlloc)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test releaseLease", func() {
			lm.CreateInstanceLease(insAlloc, 1*time.Second, func() {})
			err := lm.releaseLease(insAlloc)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGenericInstanceLease(t *testing.T) {
	manager := NewGenericLeaseManager("funcKey1")
	instance := &types.Instance{InstanceID: "instance1"}
	insAlloc := &types.InstanceAllocation{AllocationID: "allocation1", Instance: instance}
	lease, _ := manager.CreateInstanceLease(insAlloc, 1*time.Second, func() {})
	convey.Convey("Test GenericInstanceLease", t, func() {
		convey.Convey("test extendLease", func() {
			interval := lease.GetInterval()
			convey.So(interval, convey.ShouldEqual, 1*time.Second)
		})
		convey.Convey("test Extend", func() {
			err := lease.Extend()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test Release", func() {
			err := lease.Release()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
