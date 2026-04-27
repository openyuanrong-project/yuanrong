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

// Package instancepool -
package instancepool

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

type OneShotInstanceLease struct {
	insAlloc *types.InstanceAllocation
	interval time.Duration
	callback func(insAlloc *types.InstanceAllocation)
}

// Extend will extend lease
func (ol *OneShotInstanceLease) Extend() error {
	return nil
}

// Release will release lease
func (ol *OneShotInstanceLease) Release() error {
	ol.callback(ol.insAlloc)
	return nil
}

// GetInterval will return interval of lease
func (ol *OneShotInstanceLease) GetInterval() time.Duration {
	return ol.interval
}

// OneShotInstancePool implements one-shot instance pool strategy:
// - Creates a new instance for each acquire request
// - Automatically deletes instance after use
// - Does not maintain instance pool or queue
// - Suitable for strict isolation requirements
type OneShotInstancePool struct {
	funcSpec          *types.FunctionSpecification
	defaultResSpec    *resspeckey.ResourceSpecification
	defaultResKey     resspeckey.ResSpecKey
	faasManagerInfo   faasManagerInfo
	createTimeout     time.Duration
	activeInstances   map[string]*oneShotInstanceRecord
	currentPoolLabel  string
	defaultPoolLabel  string
	functionSignature string
	sync.RWMutex
}

// oneShotInstanceRecord tracks a one-shot instance
type oneShotInstanceRecord struct {
	instance     *types.Instance
	allocationID string
	createTime   time.Time
	cleanupTimer *time.Timer
}

// NewOneShotInstancePool creates a new OneShotInstancePool
func NewOneShotInstancePool(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo) (InstancePool, error) {
	log.GetLogger().Infof("create one-shot instance pool for function %s", funcSpec.FuncKey)

	defaultResSpec := resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData)
	defaultResKey := resspeckey.ConvertToResSpecKey(defaultResSpec)

	pool := &OneShotInstancePool{
		funcSpec:          funcSpec,
		defaultResSpec:    defaultResSpec,
		defaultResKey:     defaultResKey,
		faasManagerInfo:   faasManagerInfo,
		createTimeout:     utils.GetCreateTimeout(funcSpec),
		activeInstances:   make(map[string]*oneShotInstanceRecord),
		currentPoolLabel:  funcSpec.InstanceMetaData.PoolLabel,
		defaultPoolLabel:  funcSpec.InstanceMetaData.PoolLabel,
		functionSignature: funcSpec.FuncMetaSignature,
	}

	return pool, nil
}

// CreateInstance creates a new one-shot instance (not typically used directly)
func (op *OneShotInstancePool) CreateInstance(insCrtReq *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
	log.GetLogger().Infof("creating one-shot instance for function %s with instanceName %s",
		op.funcSpec.FuncKey, insCrtReq.InstanceName)

	select {
	case <-op.funcSpec.FuncCtx.Done():
		log.GetLogger().Errorf("function %s is deleted, cannot create instance", op.funcSpec.FuncKey)
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
	}

	resSpec := op.defaultResSpec
	if !utils.IsResSpecEmpty(insCrtReq.ResSpec) {
		resSpec = insCrtReq.ResSpec
	}
	resKey := resspeckey.ConvertToResSpecKey(resSpec)

	instance, err := op.createInstanceInternal(insCrtReq.InstanceName, resKey, insCrtReq.CreateEvent)
	if err != nil {
		return nil, snerror.New(statuscode.StatusInternalServerError, err.Error())
	}

	return instance, nil
}

// DeleteInstance deletes a one-shot instance
func (op *OneShotInstancePool) DeleteInstance(instance *types.Instance) snerror.SNError {
	log.GetLogger().Infof("deleting one-shot instance %s for function %s",
		instance.InstanceID, op.funcSpec.FuncKey)

	select {
	case <-op.funcSpec.FuncCtx.Done():
		log.GetLogger().Errorf("function %s is deleted, cannot delete instance", op.funcSpec.FuncKey)
		return snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
	}

	if err := op.deleteInstanceInternal(instance); err != nil {
		return snerror.New(statuscode.StatusInternalServerError, err.Error())
	}

	return nil
}

// AcquireInstance creates a new instance for each request
func (op *OneShotInstancePool) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError) {
	logger := log.GetLogger().With(zap.Any("traceId", insAcqReq.TraceID), zap.Any("funcKey", op.funcSpec.FuncKey))
	logger.Infof("acquiring one-shot instance")

	select {
	case <-op.funcSpec.FuncCtx.Done():
		logger.Errorf("function is deleted, cannot acquire one-shot instance")
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
	}

	// Determine resource specification
	resSpec := op.defaultResSpec
	if !utils.IsResSpecEmpty(insAcqReq.ResSpec) {
		resSpec = insAcqReq.ResSpec
	}
	resKey := resspeckey.ConvertToResSpecKey(resSpec)

	// Create new instance
	instance, err := op.createInstanceInternal("", resKey, []byte{})
	if err != nil {
		logger.Errorf("failed to create one-shot instance: %v", err)
		return nil, snerror.New(statuscode.StatusInternalServerError, err.Error())
	}

	// Track active instance with auto-cleanup
	op.Lock()
	record := &oneShotInstanceRecord{
		instance:     instance,
		allocationID: instance.InstanceID,
		createTime:   time.Now(),
	}

	// Schedule automatic cleanup
	timeout := time.Duration(op.funcSpec.FuncMetaData.Timeout) * time.Second
	gracePeriod := 30 * time.Second
	cleanupTime := timeout + gracePeriod

	record.cleanupTimer = time.AfterFunc(cleanupTime, func() {
		op.autoCleanupInstance(instance.InstanceID, instance)
	})

	op.activeInstances[instance.InstanceID] = record
	op.Unlock()
	insAlloc := &types.InstanceAllocation{
		AllocationID: instance.InstanceID,
		Instance:     instance,
	}

	newLease := &OneShotInstanceLease{
		insAlloc: insAlloc,
		interval: cleanupTime,
		callback: op.ReleaseInstance,
	}
	insAlloc.Lease = newLease

	logger.Infof("successfully created one-shot instance %s for function %s",
		instance.InstanceID, op.funcSpec.FuncKey)
	return insAlloc, nil
}

// ReleaseInstance deletes the instance after use
func (op *OneShotInstancePool) ReleaseInstance(insAlloc *types.InstanceAllocation) {
	if insAlloc == nil || insAlloc.Instance == nil {
		log.GetLogger().Warnf("invalid instance allocation for function %s", op.funcSpec.FuncKey)
		return
	}

	logger := log.GetLogger().With(zap.Any("funcKey", op.funcSpec.FuncKey),
		zap.Any("instanceID", insAlloc.Instance.InstanceID),
		zap.Any("allocationID", insAlloc.AllocationID))

	logger.Infof("releasing one-shot instance")

	// Remove from active tracking and cancel cleanup timer
	op.Lock()
	record, exists := op.activeInstances[insAlloc.AllocationID]
	if exists {
		delete(op.activeInstances, insAlloc.AllocationID)
		if record.cleanupTimer != nil {
			record.cleanupTimer.Stop()
		}
	}
	op.Unlock()

	if !exists {
		logger.Warnf("allocation not found in active instances")
	}

	// Delete instance asynchronously
	go func() {
		if err := op.deleteInstanceInternal(insAlloc.Instance); err != nil {
			logger.Errorf("failed to delete one-shot instance: %v", err)
		} else {
			logger.Infof("successfully deleted one-shot instance")
		}
	}()
}

// autoCleanupInstance automatically cleans up instances that exceeded timeout
func (op *OneShotInstancePool) autoCleanupInstance(allocationID string, instance *types.Instance) {
	logger := log.GetLogger().With(zap.Any("funcKey", op.funcSpec.FuncKey),
		zap.Any("instanceID", instance.InstanceID),
		zap.Any("allocationID", allocationID))

	op.Lock()
	record, exists := op.activeInstances[allocationID]
	if exists {
		delete(op.activeInstances, allocationID)
	}
	op.Unlock()

	if exists {
		logger.Warnf("one-shot instance exceeded timeout, auto-cleaning up")
		if err := op.deleteInstanceInternal(record.instance); err != nil {
			logger.Errorf("failed to auto-cleanup instance: %v", err)
		}
	}
}

// createInstanceInternal handles actual instance creation
func (op *OneShotInstancePool) createInstanceInternal(instanceName string, resKey resspeckey.ResSpecKey,
	createEvent []byte,
) (*types.Instance, error) {
	op.RLock()
	createRequest := createInstanceRequest{
		funcSpec:        op.funcSpec,
		poolLabel:       op.currentPoolLabel,
		createTimeout:   op.createTimeout,
		faasManagerInfo: op.faasManagerInfo,
		resKey:          resKey,
		instanceName:    instanceName,
		createEvent:     createEvent,
		instanceType:    types.InstanceTypeScaled, // Use scaled type for one-shot instances
	}
	op.RUnlock()

	return CreateInstance(createRequest)
}

// deleteInstanceInternal handles actual instance deletion
func (op *OneShotInstancePool) deleteInstanceInternal(instance *types.Instance) error {
	op.RLock()
	currentFaasManagerInfo := op.faasManagerInfo
	op.RUnlock()

	return DeleteInstance(op.funcSpec, currentFaasManagerInfo, instance)
}

// HandleFunctionEvent handles function events
func (op *OneShotInstancePool) HandleFunctionEvent(eventType registry.EventType, funcSpec *types.FunctionSpecification) {
	logger := log.GetLogger().With(zap.Any("funcKey", op.funcSpec.FuncKey), zap.Any("eventType", eventType))
	logger.Infof("handling function event")

	switch eventType {
	case registry.SubEventTypeUpdate:
		op.handleFunctionUpdate(funcSpec)
	case registry.SubEventTypeDelete:
		op.handleFunctionDelete()
	}
}

// handleFunctionUpdate handles function update event
func (op *OneShotInstancePool) handleFunctionUpdate(funcSpec *types.FunctionSpecification) {
	op.Lock()
	oldSig := op.functionSignature
	op.funcSpec = funcSpec
	op.defaultResSpec = resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData)
	op.defaultResKey = resspeckey.ConvertToResSpecKey(op.defaultResSpec)
	op.createTimeout = utils.GetCreateTimeout(funcSpec)
	op.defaultPoolLabel = funcSpec.InstanceMetaData.PoolLabel
	op.currentPoolLabel = funcSpec.InstanceMetaData.PoolLabel
	op.functionSignature = funcSpec.FuncMetaSignature
	op.Unlock()

	if oldSig != funcSpec.FuncMetaSignature {
		log.GetLogger().Warnf("function %s signature changed from %s to %s, cleaning up all one-shot instances",
			op.funcSpec.FuncKey, oldSig, funcSpec.FuncMetaSignature)
		op.cleanupAllInstances()
	}
}

// handleFunctionDelete handles function delete event
func (op *OneShotInstancePool) handleFunctionDelete() {
	log.GetLogger().Infof("function %s deleted, cleaning up all one-shot instances", op.funcSpec.FuncKey)
	op.cleanupAllInstances()
}

// cleanupAllInstances cleans up all active instances
func (op *OneShotInstancePool) cleanupAllInstances() {
	op.Lock()
	instancesToDelete := make([]*oneShotInstanceRecord, 0, len(op.activeInstances))
	for _, record := range op.activeInstances {
		if record.cleanupTimer != nil {
			record.cleanupTimer.Stop()
		}
		instancesToDelete = append(instancesToDelete, record)
	}
	op.activeInstances = make(map[string]*oneShotInstanceRecord)
	op.Unlock()

	// Delete all instances
	for _, record := range instancesToDelete {
		go op.deleteInstanceInternal(record.instance)
	}
}

// HandleAliasEvent handles alias event (no-op for one-shot pool)
func (op *OneShotInstancePool) HandleAliasEvent(eventType registry.EventType, aliasUrn string) {
	log.GetLogger().Infof("handling alias event for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// HandleFaaSSchedulerEvent handles FaaS scheduler event (no-op for one-shot pool)
func (op *OneShotInstancePool) HandleFaaSSchedulerEvent() {
	log.GetLogger().Infof("handling FaaS scheduler event for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// HandleInstanceEvent handles instance event
func (op *OneShotInstancePool) HandleInstanceEvent(eventType registry.EventType, instance *types.Instance) {
	logger := log.GetLogger().With(zap.Any("funcKey", op.funcSpec.FuncKey),
		zap.Any("instanceID", instance.InstanceID),
		zap.Any("eventType", eventType))

	logger.Infof("handling instance event")

	switch eventType {
	case registry.SubEventTypeDelete:
		// Remove from tracking if present
		op.Lock()
		for allocID, record := range op.activeInstances {
			if record.instance.InstanceID == instance.InstanceID {
				delete(op.activeInstances, allocID)
				if record.cleanupTimer != nil {
					record.cleanupTimer.Stop()
				}
				logger.Infof("removed one-shot instance from active tracking")
				break
			}
		}
		op.Unlock()
	case registry.SubEventTypeUpdate:
		logger.Infof("instance update event (no action needed for one-shot)")
	}
}

// HandleInstanceConfigEvent handles instance config event (no-op for one-shot pool)
func (op *OneShotInstancePool) HandleInstanceConfigEvent(eventType registry.EventType, insConfig *instanceconfig.Configuration) {
	log.GetLogger().Infof("handling instance config event for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// UpdateInvokeMetrics updates invoke metrics (no-op for one-shot pool)
func (op *OneShotInstancePool) UpdateInvokeMetrics(resKey resspeckey.ResSpecKey, insMetrics *types.InstanceThreadMetrics) {
	// One-shot instances don't need metrics collection
}

// HandleFaaSManagerUpdate handles FaaS manager update
func (op *OneShotInstancePool) HandleFaaSManagerUpdate(faasManagerInfo faasManagerInfo) {
	op.Lock()
	op.faasManagerInfo = faasManagerInfo
	op.Unlock()
}

// GetFuncSpec returns the function specification
func (op *OneShotInstancePool) GetFuncSpec() *types.FunctionSpecification {
	op.RLock()
	funcSpec := op.funcSpec
	op.RUnlock()
	return funcSpec
}

// RecoverInstance recovers instance pool (no-op for one-shot pool)
func (op *OneShotInstancePool) RecoverInstance(funcSpec *types.FunctionSpecification,
	instancePoolState *types.InstancePoolState, deleteFunc bool, wg *sync.WaitGroup,
) {
	defer wg.Done()
	log.GetLogger().Infof("recover instance for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// GetAndDeleteState gets and deletes state (not applicable for one-shot pool)
func (op *OneShotInstancePool) GetAndDeleteState(stateID string) bool {
	return false
}

// DeleteStateInstance deletes state instance (not applicable for one-shot pool)
func (op *OneShotInstancePool) DeleteStateInstance(stateID string, instanceID string) {
	// One-shot instances don't support state
}

// QuerySession resolves session ID to instance ID. One-shot pool does not record sessions
// (unlike GenericInstancePool.sessionRecordMap), so lookups always miss.
func (op *OneShotInstancePool) QuerySession(sessionID string) (string, error) {
	op.RLock()
	defer op.RUnlock()
	return "", fmt.Errorf("session %s not found", sessionID)
}

// handleManagedChange handles managed change (no-op for one-shot pool)
func (op *OneShotInstancePool) handleManagedChange() {
	log.GetLogger().Infof("handling managed change for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// handleRatioChange handles ratio change (no-op for one-shot pool)
func (op *OneShotInstancePool) handleRatioChange(ratio int) {
	log.GetLogger().Infof("handling ratio change for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}

// CleanOrphansInstanceQueue cleans orphan instance queue (no-op for one-shot pool)
func (op *OneShotInstancePool) CleanOrphansInstanceQueue() {
	log.GetLogger().Infof("cleaning orphan instance queue for one-shot pool %s (no-op)", op.funcSpec.FuncKey)
}
