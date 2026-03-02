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
	"errors"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/timewheel"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	defaultTimeWheelPace  = 5 * time.Millisecond
	defaultTimeWheelSlots = 100
)

var (
	errInstanceLeaseHolderNotEnable = errors.New("instance lease holder not enable")
	// ErrInstanceNotFound instance not found err
	ErrInstanceNotFound = errors.New("instance doesn't exist")
)

type instanceLeaseHolder struct {
	instance    *types.Instance
	timeWheel   timewheel.TimeWheel
	intervalMap map[string]time.Duration
	callbackMap map[string]func()
	enable      bool
	sync.RWMutex
}

func newInstanceLeaseHolder(instance *types.Instance) *instanceLeaseHolder {
	holder := &instanceLeaseHolder{
		instance:    instance,
		timeWheel:   timewheel.NewSimpleTimeWheel(defaultTimeWheelPace, defaultTimeWheelSlots),
		intervalMap: make(map[string]time.Duration, instance.ConcurrentNum),
		callbackMap: make(map[string]func(), instance.ConcurrentNum),
		enable:      true,
	}
	go holder.pollLease()
	return holder
}

func (il *instanceLeaseHolder) stop() {
	il.Lock()
	enable := il.enable
	il.enable = false
	il.Unlock()
	// make sure just stop once
	if enable {
		il.timeWheel.Stop()
		log.GetLogger().Warnf("stopping timeWheel for instance %s function %s",
			il.instance.InstanceID, il.instance.FuncKey)
	}
}

func (il *instanceLeaseHolder) pollLease() {
	for {
		il.RLock()
		// if instanceLeaseHolder stopped, break loop
		if !il.enable {
			log.GetLogger().Warnf("stopping leases for instance %s function %s",
				il.instance.InstanceID, il.instance.FuncKey)
			il.RUnlock()
			return
		}
		il.RUnlock()
		readyList, err := il.timeWheel.Wait()
		if err != nil {
			log.GetLogger().Warnf("time wheel wait, %s", err.Error())
		}
		if len(readyList) == 0 {
			continue
		}
		go func() {
			for _, allocationID := range readyList {
				log.GetLogger().Warnf("lease %s expires, now release", allocationID)
				il.Lock()
				if err := il.timeWheel.DelTask(allocationID); err != nil {
					log.GetLogger().Errorf("failed to delete task %s in time wheel", allocationID)
				}
				callback, exist := il.callbackMap[allocationID]
				delete(il.intervalMap, allocationID)
				delete(il.callbackMap, allocationID)
				il.Unlock()
				if exist {
					callback()
				}
			}
		}()
	}
}

func (il *instanceLeaseHolder) createLease(insAlloc *types.InstanceAllocation, interval time.Duration,
	callback func()) error {
	il.Lock()
	defer il.Unlock()
	if !il.enable {
		return errInstanceLeaseHolderNotEnable
	}
	err := il.timeWheel.AddTask(insAlloc.AllocationID, interval, 1)
	if err != nil {
		return err
	}
	il.intervalMap[insAlloc.AllocationID] = interval
	il.callbackMap[insAlloc.AllocationID] = callback
	return nil
}

func (il *instanceLeaseHolder) extendLease(insAlloc *types.InstanceAllocation) error {
	il.Lock()
	defer il.Unlock()
	if !il.enable {
		return errInstanceLeaseHolderNotEnable
	}
	interval, exist := il.intervalMap[insAlloc.AllocationID]
	if !exist {
		return errors.New("lease doesn't exist or released")
	}
	err := il.timeWheel.UpdateTask(insAlloc.AllocationID, interval, 1)
	if err != nil {
		return err
	}
	return nil
}

func (il *instanceLeaseHolder) releaseLease(insAlloc *types.InstanceAllocation) error {
	il.Lock()
	if !il.enable {
		il.Unlock()
		return errInstanceLeaseHolderNotEnable
	}
	callback, exist := il.callbackMap[insAlloc.AllocationID]
	delete(il.intervalMap, insAlloc.AllocationID)
	delete(il.callbackMap, insAlloc.AllocationID)
	err := il.timeWheel.DelTask(insAlloc.AllocationID)
	il.Unlock()
	if err != nil {
		return err
	}
	if exist {
		callback()
	}
	return nil
}

// GenericInstanceLeaseManager manages insAlloc leases of instances of a specific function
type GenericInstanceLeaseManager struct {
	leaseHolders map[string]*instanceLeaseHolder
	funcKey      string
	sync.RWMutex
}

// NewGenericLeaseManager creates a GenericInstanceLeaseManager
func NewGenericLeaseManager(funcKey string) InstanceLeaseManager {
	return &GenericInstanceLeaseManager{
		leaseHolders: make(map[string]*instanceLeaseHolder, utils.DefaultMapSize),
		funcKey:      funcKey,
	}
}

// CreateInstanceLease creates a lease for an instance insAlloc
func (gm *GenericInstanceLeaseManager) CreateInstanceLease(insAlloc *types.InstanceAllocation, interval time.Duration,
	callback func()) (types.InstanceLease, error) {
	if insAlloc == nil || insAlloc.Instance == nil {
		log.GetLogger().Errorf("invalid instance insAlloc")
		return nil, errors.New("invalid instance insAlloc")
	}
	instance := insAlloc.Instance
	gm.Lock()
	leaseHolder, exist := gm.leaseHolders[instance.InstanceID]
	if !exist {
		leaseHolder = newInstanceLeaseHolder(insAlloc.Instance)
		gm.leaseHolders[instance.InstanceID] = leaseHolder
	}
	gm.Unlock()
	if err := leaseHolder.createLease(insAlloc, interval, callback); err != nil {
		return nil, err
	}
	return &GenericInstanceLease{
		insAlloc: insAlloc,
		manager:  gm,
		interval: interval,
	}, nil
}

// HandleInstanceDelete handles instance delete
func (gm *GenericInstanceLeaseManager) HandleInstanceDelete(instance *types.Instance) {
	gm.Lock()
	leaseHolder, exist := gm.leaseHolders[instance.InstanceID]
	delete(gm.leaseHolders, instance.InstanceID)
	gm.Unlock()
	if !exist {
		return
	}
	leaseHolder.stop()
}

// CleanAllLeases cleans all leases
func (gm *GenericInstanceLeaseManager) CleanAllLeases() {
	gm.Lock()
	for instanceID, leaseHolder := range gm.leaseHolders {
		leaseHolder.stop()
		log.GetLogger().Infof("leases for instance %s function %s stopped", instanceID, gm.funcKey)
	}
	gm.leaseHolders = map[string]*instanceLeaseHolder{}
	gm.Unlock()
}

func (gm *GenericInstanceLeaseManager) extendLease(insAlloc *types.InstanceAllocation) error {
	gm.Lock()
	leaseHolder, exist := gm.leaseHolders[insAlloc.Instance.InstanceID]
	gm.Unlock()
	if !exist {
		return ErrInstanceNotFound
	}
	return leaseHolder.extendLease(insAlloc)
}

func (gm *GenericInstanceLeaseManager) releaseLease(insAlloc *types.InstanceAllocation) error {
	gm.Lock()
	leaseHolder, exist := gm.leaseHolders[insAlloc.Instance.InstanceID]
	gm.Unlock()
	if !exist {
		return ErrInstanceNotFound
	}
	return leaseHolder.releaseLease(insAlloc)
}

// GenericInstanceLease provides lease operations of a specified instance allocation
type GenericInstanceLease struct {
	insAlloc *types.InstanceAllocation
	manager  *GenericInstanceLeaseManager
	interval time.Duration
}

// Extend will extend lease
func (gl *GenericInstanceLease) Extend() error {
	if gl.manager == nil {
		return errors.New("lease manager doesn't exist")
	}
	return gl.manager.extendLease(gl.insAlloc)
}

// Release will release lease
func (gl *GenericInstanceLease) Release() error {
	if gl.manager == nil {
		return errors.New("lease manager doesn't exist")
	}
	return gl.manager.releaseLease(gl.insAlloc)
}

// GetInterval will return interval of lease
func (gl *GenericInstanceLease) GetInterval() time.Duration {
	return gl.interval
}
