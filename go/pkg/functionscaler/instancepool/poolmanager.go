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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/stateinstance"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	// SLA time should not be shorter than cold start time
	minSLATime = time.Duration(500) * time.Millisecond
)

type faasManagerInfo struct {
	funcKey    string
	instanceID string
}

// PoolManager manages instance pools of different functions
type PoolManager struct {
	faasManagerInfo faasManagerInfo
	instancePool    map[string]InstancePool
	// instanceRecord is used to find instancePool for a specific instance watched from etcd, because router etcd will
	instanceConfigRecord map[string]map[string]*instanceconfig.Configuration
	stateLeaseManager    map[string]*stateinstance.Leaser // key instanceID
	abnormalInstanceMap  *utils.TimeoutMap
	leaseInterval        time.Duration
	stopCh               <-chan struct{}
	sync.RWMutex
}

// GetLeaseInterval it's for state instance
func (pm *PoolManager) GetLeaseInterval() time.Duration {
	return pm.leaseInterval
}

func getScaleDownWindow() time.Duration {
	scaleUpWindow := time.Duration(config.GlobalConfig.AutoScaleConfig.SLAQuota) * time.Millisecond
	if scaleUpWindow < minSLATime {
		scaleUpWindow = minSLATime
	}
	scaleDownTime := time.Duration(config.GlobalConfig.AutoScaleConfig.ScaleDownTime) * time.Millisecond
	if scaleDownTime < scaleUpWindow {
		scaleDownTime = scaleUpWindow
	}
	return scaleDownTime
}

// NewPoolManager creates a PoolManager
func NewPoolManager(stopCh <-chan struct{}) *PoolManager {
	leaseInterval := time.Duration(config.GlobalConfig.LeaseSpan) * time.Millisecond
	if leaseInterval < types.MinLeaseInterval {
		leaseInterval = types.MinLeaseInterval
	}
	pm := &PoolManager{
		faasManagerInfo:      faasManagerInfo{},
		instancePool:         make(map[string]InstancePool, utils.DefaultMapSize),
		instanceConfigRecord: make(map[string]map[string]*instanceconfig.Configuration, utils.DefaultMapSize),
		stateLeaseManager:    make(map[string]*stateinstance.Leaser),
		abnormalInstanceMap:  utils.NewTimeoutMap(time.Minute),
		leaseInterval:        leaseInterval,
		stopCh:               stopCh,
	}
	return pm
}

// RecoverInstancePool recover instancePool data
// if fail to recover, see faaSScheduler as restarting
// precondition: make sure that FunctionRegistry.funcSpecs is up-to-date
func (pm *PoolManager) RecoverInstancePool() {
	s := state.GetState()
	log.GetLogger().Infof("ready to recovery instance pool")
	var wg sync.WaitGroup
	registry.GlobalRegistry.InstanceRegistry.WaitForETCDList()
	var deleteFunctions []*types.FunctionSpecification
	for funcKey, val := range s.InstancePool {
		var InstancePoolState *types.InstancePoolState
		commonUtils.DeepCopyObj(val, &InstancePoolState)
		funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
		// guarantee that the recovered data is the same as etcd data
		var deleteFuncFlag bool
		if funcSpec == nil {
			// it means function has been deleted while recovering
			// faasscheduler need to receive etcd delete event to delete function
			deleteFuncFlag = true
			funcSpec = &types.FunctionSpecification{FuncKey: funcKey}
			deleteFunctions = append(deleteFunctions, funcSpec)
		}
		pm.RLock()
		_, exist := pm.instancePool[funcSpec.FuncKey]
		pm.RUnlock()
		if !exist {
			if _, err := pm.processInstancePoolCreate(funcSpec); err != nil {
				continue
			}
		}
		log.GetLogger().Infof("now recover function :%s", funcSpec.FuncKey)
		wg.Add(1)
		go func() {
			log.GetLogger().Infof("recover func: %s, isStateful: %v, deleteFuncFlag: %v",
				funcSpec.FuncKey, funcSpec.FuncMetaData.IsStatefulFunction, deleteFuncFlag)
			pm.instancePool[funcSpec.FuncKey].RecoverInstance(funcSpec, InstancePoolState, deleteFuncFlag, &wg)
			pm.recoverStateLeaser(InstancePoolState.StateInstance, funcSpec)
		}()
	}
	wg.Wait()
	for _, function := range deleteFunctions {
		pm.HandleFunctionEvent(registry.SubEventTypeDelete, function)
	}
}

func (pm *PoolManager) recoverStateLeaser(stateInstance map[string]*types.Instance,
	funcSpec *types.FunctionSpecification,
) {
	for stateID, instance := range stateInstance {
		pool := pm.instancePool[funcSpec.FuncKey]
		if pool == nil {
			log.GetLogger().Errorf("func %s, stateid %s: pool is nil! can't create leaser manager!",
				funcSpec.FuncKey, stateID)
		} else {
			if instance.InstanceStatus.Code == int32(-1) {
				log.GetLogger().Warnf("instance of stateID %s is existed!", stateID)
				continue
			}
			log.GetLogger().Infof("func %s, stateId %s: create leaser manager for recovery, downwin %v",
				funcSpec.FuncKey, stateID, getScaleDownWindow())
			leaser := stateinstance.NewLeaser(funcSpec.InstanceMetaData.ConcurrentNum,
				pool.DeleteStateInstance, stateID, instance.InstanceID, getScaleDownWindow())
			pm.Lock()
			leaser.Recover()
			pm.stateLeaseManager[instance.InstanceID] = leaser
			pm.Unlock()
		}
	}
}

// GetAndDeleteState delete state and return whether the state exists
func (pm *PoolManager) GetAndDeleteState(stateID string, funcKey string, funcSpec *types.FunctionSpecification,
	logger api.FormatLogger,
) bool {
	pm.Lock()
	pool, exist := pm.instancePool[funcKey]
	if !exist {
		var err error
		if funcSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyOneshot {
			pool, err = NewOneShotInstancePool(funcSpec, pm.faasManagerInfo)
		} else {
			pool, err = NewGenericInstancePool(funcSpec, pm.faasManagerInfo)
		}
		if err != nil {
			pm.Unlock()
			return false
		}
		pm.instancePool[funcKey] = pool
	}
	pm.Unlock()

	logger.Infof("getAndDeleteStateInstance, stateKey %s", stateID)
	return pool.GetAndDeleteState(stateID)
}

// CreateInstance will create an instance of a specific function
func (pm *PoolManager) CreateInstance(insCrtReq *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
	pm.RLock()
	pool, exist := pm.instancePool[insCrtReq.FuncSpec.FuncKey]
	pm.RUnlock()
	if !exist {
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, "pool not exist")
	}
	return pool.CreateInstance(insCrtReq)
}

// DeleteInstance will delete an instance of a specific function
func (pm *PoolManager) DeleteInstance(instance *types.Instance) snerror.SNError {
	pm.RLock()
	pool, exist := pm.instancePool[instance.FuncKey]
	pm.RUnlock()
	if !exist {
		return snerror.New(statuscode.FuncMetaNotFoundErrCode, "pool not exist")
	}
	return pool.DeleteInstance(instance)
}

// AcquireInstanceThread will acquire a instance thread of a specific function
func (pm *PoolManager) AcquireInstanceThread(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation,
	snerror.SNError,
) {
	pm.RLock()
	pool, exist := pm.instancePool[insAcqReq.FuncSpec.FuncKey]
	pm.RUnlock()
	if !exist {
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, "pool not exist")
	}
	insAlloc, acquireErr := pool.AcquireInstance(insAcqReq)
	if acquireErr != nil {
		return nil, acquireErr
	}
	return insAlloc, nil
}

// ReleaseStateThread -
func (pm *PoolManager) ReleaseStateThread(thread *types.InstanceAllocation) error {
	leaser := pm.stateLeaseManager[thread.Instance.InstanceID]
	if leaser == nil {
		return errors.New("leaser is nil")
	}
	leaseID, err := getLeaseIdFromAllocationId(thread.AllocationID)
	if err != nil {
		return err
	}
	leaser.ReleaseLease(leaseID)
	return nil
}

// RetainStateThread -
func (pm *PoolManager) RetainStateThread(thread *types.InstanceAllocation) error {
	leaser := pm.stateLeaseManager[thread.Instance.InstanceID]
	if leaser == nil {
		return errors.New("leaser is nil")
	}
	leaseID, err := getLeaseIdFromAllocationId(thread.AllocationID)
	if err != nil {
		return err
	}
	return leaser.RetainLease(leaseID, pm.leaseInterval)
}

func getLeaseIdFromAllocationId(allocationId string) (int, error) {
	parts := strings.Split(allocationId, "-")
	if len(parts) < 1 { // %s-stateThread%d, eg:0600a7ba-cfc0-4a00-8000-0000000004a1-stateThread1
		return 0, fmt.Errorf("thread.AllocationID: %s invalid", allocationId)
	}
	numStr := strings.Replace(parts[len(parts)-1], "stateThread", "", 1)
	leaseID, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("thread.AllocationID: %s invalid, err %v", allocationId, err)
	}
	return leaseID, nil
}

// ReleaseInstanceThread will release a instance thread of a specific function
func (pm *PoolManager) ReleaseInstanceThread(insAlloc *types.InstanceAllocation) {
	instance := insAlloc.Instance
	pm.RLock()
	pool, exist := pm.instancePool[instance.FuncKey]
	pm.RUnlock()
	if !exist {
		log.GetLogger().Errorf("instance pool for function %s doesn't exist", instance.FuncKey)
		return
	}
	pool.ReleaseInstance(insAlloc)
}

func (pm *PoolManager) QuerySession(funcKey string, sessionID string) (string, error) {
	pm.RLock()
	pool, exist := pm.instancePool[funcKey]
	pm.RUnlock()
	if !exist {
		return "", fmt.Errorf("pool not exist for function %s", funcKey)
	}
	return pool.QuerySession(sessionID)
}

// ReleaseAbnormalInstance will release an abnormal instance of a specific function
func (pm *PoolManager) ReleaseAbnormalInstance(instance *types.Instance, logger api.FormatLogger) {
	pm.RLock()
	pool, exist := pm.instancePool[instance.FuncKey]
	pm.RUnlock()
	if !exist {
		logger.Warnf("instance pool for function %s doesn't exist", instance.FuncKey)
		return
	}
	if _, ok := pm.abnormalInstanceMap.Get(instance.InstanceID); ok {
		logger.Debugf("instance %s has stored into abnormal instance map, skip delete",
			instance.InstanceID)
		return
	}
	pm.abnormalInstanceMap.Set(instance.InstanceID, nil, time.Minute)
	pool.HandleInstanceEvent(registry.SubEventTypeRemove, instance)
}

// HandleFunctionEvent handles function event
func (pm *PoolManager) HandleFunctionEvent(eventType registry.EventType, funcSpec *types.FunctionSpecification) {
	log.GetLogger().Infof("handling function event type %s for function %s", eventType, funcSpec.FuncKey)
	if eventType == registry.SubEventTypeUpdate {
		pm.RLock()
		pool, poolExist := pm.instancePool[funcSpec.FuncKey]
		pm.RUnlock()
		var err error
		if !poolExist {
			if pool, err = pm.processInstancePoolCreate(funcSpec); err != nil {
				return
			}
		}
		if poolExist {
			pool.HandleFunctionEvent(eventType, funcSpec)
		}
		// 注意：这里的读锁需要把HandleInstanceConfigEvent包含在内
		pm.RLock()
		insConfigs, insConfigExist := pm.instanceConfigRecord[funcSpec.FuncKey]
		if insConfigExist {
			for _, insConfig := range insConfigs {
				pool.HandleInstanceConfigEvent(eventType, insConfig)
			}
		}
		pm.RUnlock()
	}
	if eventType == registry.SubEventTypeDelete {
		pm.RLock()
		pool, poolExist := pm.instancePool[funcSpec.FuncKey]
		pm.RUnlock()
		if poolExist {
			pm.processInstancePoolDelete(funcSpec)
			pool.HandleFunctionEvent(eventType, funcSpec)
		}
		metrics.ClearMetricsForFunction(funcSpec)
	}

	if eventType == registry.SubEventTypeSynced {
		log.GetLogger().Infof("send function synced event")
		registry.GlobalRegistry.FunctionRegistry.FinishEtcdList()
	}
}

func (pm *PoolManager) processInstancePoolCreate(funcSpec *types.FunctionSpecification) (InstancePool, error) {
	if funcSpec.FuncMetaData.VPCTriggerImage != "" &&
		config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeKernel {
		pm.RLock()
		faasMgrInfo := pm.faasManagerInfo
		pm.RUnlock()
		go handlePullTriggerCreate(faasMgrInfo, funcSpec)
	}
	var pool InstancePool
	var err error
	if funcSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyOneshot {
		pool, err = NewOneShotInstancePool(funcSpec, pm.faasManagerInfo)
	} else {
		pool, err = NewGenericInstancePool(funcSpec, pm.faasManagerInfo)
	}
	if err != nil {
		log.GetLogger().Errorf("failed to create instance pool of function %s error %s", funcSpec.FuncKey, err.Error())
		return nil, err
	}
	pm.Lock()
	pm.instancePool[funcSpec.FuncKey] = pool
	pm.Unlock()
	return pool, nil
}

func (pm *PoolManager) processInstancePoolDelete(funcSpec *types.FunctionSpecification) {
	if funcSpec.FuncMetaData.VPCTriggerImage != "" &&
		config.GlobalConfig.InstanceOperationBackend == constant.BackendTypeKernel {
		pm.RLock()
		faasMgrInfo := pm.faasManagerInfo
		pm.RUnlock()
		go handlePullTriggerDelete(faasMgrInfo, funcSpec)
	}
	pm.Lock()
	delete(pm.instancePool, funcSpec.FuncKey)
	delete(pm.instanceConfigRecord, funcSpec.FuncKey)
	pm.Unlock()
	state.Update(&types.InstancePoolStateInput{
		FuncKey: funcSpec.FuncKey,
	}, types.StateDelete)
}

// HandleInstanceEvent handles instance event
func (pm *PoolManager) HandleInstanceEvent(eventType registry.EventType, insSpec *commonTypes.InstanceSpecification) {
	if eventType == registry.SubEventTypeSynced {
		for _, pool := range pm.instancePool {
			pool.HandleInstanceEvent(eventType, nil)
		}
		return
	}
	items := strings.Split(insSpec.Function, utils.FuncKeyDelimiter)
	if len(items) != utils.ValidFuncKeyLen {
		return
	}
	if utils.IsFaaSManager(items[1]) {
		if eventType != registry.SubEventTypeDelete {
			pm.faasManagerInfo.funcKey = insSpec.Function
			pm.faasManagerInfo.instanceID = insSpec.InstanceID
			log.GetLogger().Infof("set faas manager info to %v", pm.faasManagerInfo)
			for _, pool := range pm.instancePool {
				pool.HandleFaaSManagerUpdate(pm.faasManagerInfo)
			}
		}
		return
	}
	logger := log.GetLogger().With(zap.Any("InstanceID", insSpec.InstanceID))
	logger.Infof("handling instance event, type %v, status %+v", eventType, insSpec.InstanceStatus)
	pm.Lock()
	pool, exist := pm.instancePool[insSpec.CreateOptions[types.FunctionKeyNote]]
	if !exist {
		logger.Warnf("instance pool for instance doesn't exist")
		if eventType != registry.SubEventTypeDelete {
			go DeleteUnexpectInstance(insSpec.ParentID, insSpec.InstanceID,
				insSpec.CreateOptions[types.FunctionKeyNote], logger)
		}
		pm.Unlock()
		return
	}
	pm.Unlock()
	instance := utils.BuildInstanceFromInsSpec(insSpec, pool.GetFuncSpec())
	switch eventType {
	case registry.SubEventTypeUpdate:
		metrics.EnsureLeaseRequestTotal(instance.MetricLabelValues)
		metrics.EnsureConcurrencyGaugeWithLabel(pool.GetFuncSpec().FuncKey,
			instance.ResKey.InvokeLabel, instance.MetricLabelValues)
	case registry.SubEventTypeDelete:
		metrics.ClearLeaseRequestTotal(instance.MetricLabelValues)
		metrics.ClearConcurrencyGaugeWithLabel(instance.MetricLabelValues)
	default:
		logger.Debugf("no need update prometheus metric")
	}
	pm.updateStateLeaseMgrForHandleInstanceEvent(eventType, instance)
	pool.HandleInstanceEvent(eventType, instance)
}

func (pm *PoolManager) updateStateLeaseMgrForHandleInstanceEvent(eventType registry.EventType,
	instance *types.Instance,
) {
	if eventType == registry.SubEventTypeDelete ||
		(eventType == registry.SubEventTypeUpdate && instance.InstanceStatus.Code == 6) { // 6 FATAL
		if stateLeaseManager, exist := pm.stateLeaseManager[instance.InstanceID]; exist {
			log.GetLogger().Infof("terminate state lease manager for instance %s, funckey %s",
				instance.InstanceID, instance.FuncKey)
			stateLeaseManager.Terminate()
			delete(pm.stateLeaseManager, instance.InstanceID)
		}
	}
}

// HandleSchedulerManaged current scheduler now is supposed to manage the scheduler's instances
func (pm *PoolManager) HandleSchedulerManaged(eventType registry.EventType,
	insSpec *commonTypes.InstanceSpecification,
) {
	log.GetLogger().Infof("handling scheduler managed event type %s, schedulerID:%s",
		eventType, insSpec.InstanceID)
	pm.Lock()
	for _, p := range pm.instancePool {
		p.handleManagedChange()
		p.HandleFaaSSchedulerEvent()
	}
	pm.Unlock()
}

// HandleRolloutRatioChange 监听灰度比例变化
func (pm *PoolManager) HandleRolloutRatioChange(ratio int) {
	log.GetLogger().Infof("handling scheduler ratio change %d", ratio)
	pm.Lock()
	defer pm.Unlock()
	for _, p := range pm.instancePool {
		p.handleRatioChange(ratio)
	}
}

// HandleInstanceConfigEvent handles instance config event
func (pm *PoolManager) HandleInstanceConfigEvent(eventType registry.EventType,
	insConfig *instanceconfig.Configuration,
) {
	logger := log.GetLogger().With(zap.Any("FuncKey", insConfig.FuncKey)).
		With(zap.Any("eventType", eventType)).
		With(zap.Any("InstanceLabel", insConfig.InstanceLabel))
	logger.Infof("handling instance config event")
	if eventType == registry.SubEventTypeSynced || eventType == registry.SubEventTypeCRSynced {
		for _, pool := range pm.instancePool {
			if eventType == registry.SubEventTypeSynced && !pool.GetFuncSpec().MetaFromCR ||
				eventType == registry.SubEventTypeCRSynced && pool.GetFuncSpec().MetaFromCR {
				pool.CleanOrphansInstanceQueue()
			}
		}
	}
	pm.RLock()
	pool, exist := pm.instancePool[insConfig.FuncKey]
	pm.RUnlock()
	if eventType == registry.SubEventTypeUpdate {
		pm.Lock()
		if _, ok := pm.instanceConfigRecord[insConfig.FuncKey]; !ok {
			pm.instanceConfigRecord[insConfig.FuncKey] = make(map[string]*instanceconfig.Configuration,
				constant.DefaultMapSize)
		}
		pm.instanceConfigRecord[insConfig.FuncKey][insConfig.InstanceLabel] = insConfig
		pm.Unlock()
		if !exist {
			logger.Warnf("instance pool for function doesn't exist")
			return
		}
		pool.HandleInstanceConfigEvent(eventType, insConfig)
	}
	if eventType == registry.SubEventTypeDelete {
		pm.Lock()
		if _, ok := pm.instanceConfigRecord[insConfig.FuncKey]; !ok {
			pm.Unlock()
			return
		}
		delete(pm.instanceConfigRecord[insConfig.FuncKey], insConfig.InstanceLabel)
		gi, ok := pm.instancePool[insConfig.FuncKey]
		pm.Unlock()
		if ok {
			gi.HandleInstanceConfigEvent(eventType, insConfig)
		}
	}
}

// HandleAliasEvent handles instance config event
func (pm *PoolManager) HandleAliasEvent(eventType registry.EventType, aliasUrn string) {
	logger := log.GetLogger().With(zap.Any("", "HandleAliasEvent")).With(zap.Any("urn", aliasUrn))
	logger.Infof("start")
	var wg sync.WaitGroup
	pm.RLock()
	for funcKey, pool := range pm.instancePool {
		tenantID := urnutils.GetTenantFromFuncKey(funcKey)
		logger.Infof("handle alias event for funcKey %s, poolName %s, tenantID %s", funcKey,
			pool.GetFuncSpec().FuncKey, tenantID)
		if urnutils.CheckAliasUrnTenant(tenantID, aliasUrn) {
			wg.Add(1)
			go func(p InstancePool) {
				defer wg.Done()
				p.HandleAliasEvent(eventType, aliasUrn)
			}(pool)
		}
	}
	pm.RUnlock()
	wg.Wait()
	logger.Infof("finish")
}

// ReportMetrics sends invoke metrics to instance pool of a specific function
func (pm *PoolManager) ReportMetrics(funcKey string, resKey resspeckey.ResSpecKey,
	insMetrics *types.InstanceThreadMetrics,
) {
	pm.RLock()
	pool, exist := pm.instancePool[funcKey]
	pm.RUnlock()
	if !exist {
		log.GetLogger().Errorf("failed to find pool for function %s", funcKey)
		return
	}
	pool.UpdateInvokeMetrics(resKey, insMetrics)
}

// CheckMinInsAndReport -
func (pm *PoolManager) CheckMinInsAndReport(stopCh <-chan struct{}) {
	go pm.checkMinInsRegularly(stopCh)
}

func getInstanceType(createOptions map[string]string) types.InstanceType {
	instanceType, ok := createOptions[types.InstanceTypeNote]
	if !ok {
		return types.InstanceTypeUnknown
	}
	return types.InstanceType(instanceType)
}
