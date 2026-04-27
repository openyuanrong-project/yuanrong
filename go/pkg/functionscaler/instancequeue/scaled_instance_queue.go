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
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/concurrencyscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/microservicescheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/roundrobinscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

var (
	initialErrorDelay           = 1 * time.Second
	retryDelayLimit             = 16 * time.Second
	retryDelayFactor            = 2
	dataSystemFeatureUsedStream = "stream"
)

// ScaledInstanceQueue stores instances and handles scaling automatically
type ScaledInstanceQueue struct {
	funcSpec           *types.FunctionSpecification
	metricsCollector   metrics.Collector
	instanceScheduler  scheduler.InstanceScheduler
	instanceScaler     scaler.InstanceScaler
	insCreateQueue     *InstanceCreateQueue
	instanceType       types.InstanceType
	resKey             resspeckey.ResSpecKey
	funcCtx            context.Context
	funcKey            string
	funcSig            string
	funcKeyWithRes     string
	concurrentNum      int
	updating           bool
	isFuncOwner        bool
	stopCh             chan struct{}
	createInstanceFunc CreateInstanceFunc
	deleteInstanceFunc DeleteInstanceFunc
	signalInstanceFunc SignalInstanceFunc
	*sync.Cond
}

// NewScaledInstanceQueue -
func NewScaledInstanceQueue(config *InsQueConfig, metricsCollector metrics.Collector) *ScaledInstanceQueue {
	funcSpec := config.FuncSpec
	instanceQueue := &ScaledInstanceQueue{
		funcSpec:           funcSpec,
		resKey:             config.ResKey,
		insCreateQueue:     NewInstanceCreateQueue(),
		metricsCollector:   metricsCollector,
		instanceType:       config.InstanceType,
		funcCtx:            funcSpec.FuncCtx,
		funcKey:            funcSpec.FuncKey,
		funcSig:            funcSpec.FuncMetaSignature,
		funcKeyWithRes:     fmt.Sprintf("%s-%s", funcSpec.FuncKey, config.ResKey.String()),
		concurrentNum:      utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum),
		stopCh:             make(chan struct{}),
		createInstanceFunc: config.CreateInstanceFunc,
		deleteInstanceFunc: config.DeleteInstanceFunc,
		signalInstanceFunc: config.SignalInstanceFunc,
		Cond:               sync.NewCond(&sync.Mutex{}),
		isFuncOwner:        selfregister.GlobalSchedulerProxy.IsFuncOwner(funcSpec.FuncKey),
	}
	go instanceQueue.startScaleUpWorker()
	return instanceQueue
}

// DisableCreateRetry -
func DisableCreateRetry() {
	initialErrorDelay = 1 * time.Second
	retryDelayLimit = 1 * time.Second
	retryDelayFactor = 2
}

// SetInstanceScheduler sets instanceScheduler
func (si *ScaledInstanceQueue) SetInstanceScheduler(instanceScheduler scheduler.InstanceScheduler) {
	si.instanceScheduler = instanceScheduler
}

// GetInstanceScheduler return queue's isntance scheduler
func (si *ScaledInstanceQueue) GetInstanceScheduler() scheduler.InstanceScheduler {
	if si.instanceScheduler != nil {
		return si.instanceScheduler
	}
	return nil
}

// GetSchedulerPolicy get current scheduelr policy
func (si *ScaledInstanceQueue) GetSchedulerPolicy() string {
	if _, ok := si.instanceScheduler.(*microservicescheduler.MicroServiceScheduler); ok {
		return types.InstanceSchedulePolicyMicroService
	}
	if _, ok := si.instanceScheduler.(*roundrobinscheduler.RoundRobinScheduler); ok {
		return types.InstanceSchedulePolicyRoundRobin
	}
	return types.InstanceSchedulePolicyConcurrency
}

// ReconnectWithScaler change scheduelr
func (si *ScaledInstanceQueue) ReconnectWithScaler() {
	si.instanceScheduler.ConnectWithInstanceScaler(si.instanceScaler)
}

// SetInstanceScaler sets instanceScaler
func (si *ScaledInstanceQueue) SetInstanceScaler(instanceScaler scaler.InstanceScaler) {
	si.instanceScaler = instanceScaler
}

// AcquireInstance will acquire one instance
func (si *ScaledInstanceQueue) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, snerror.SNError) {
	select {
	case <-si.funcCtx.Done():
		log.GetLogger().Errorf("function is deleted, can not acquire instance InsAlloc")
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
		si.metricsCollector.UpdateInvokeRequests(1)
		si.Cond.L.Lock()
		if si.updating {
			si.Wait()
		}
		si.Cond.L.Unlock()
		insAlloc, err := si.instanceScheduler.AcquireInstance(insAcqReq)
		return insAlloc, buildSnError(err)
	}
}

// ReleaseInstance will release an instance (this is not a necessary method, need to be removed in future)
func (si *ScaledInstanceQueue) ReleaseInstance(thread *types.InstanceAllocation) snerror.SNError {
	return buildSnError(si.instanceScheduler.ReleaseInstance(thread))
}

// HandleInstanceUpdate handles instance update comes from ETCD
func (si *ScaledInstanceQueue) HandleInstanceUpdate(instance *types.Instance) {
	logger := log.GetLogger().With(zap.Any("instanceID", instance.InstanceID), zap.Any("funcKey", si.funcKeyWithRes))
	logger.Infof("handling instance: %s update, pod id: %s, deployment name: %s, status code is %d, error code is %d",
		instance.InstanceID, instance.PodID, instance.PodDeploymentName, instance.InstanceStatus.Code,
		instance.InstanceStatus.ErrorCode)
	si.instanceScheduler.HandleInstanceUpdate(instance)
	wiseScale, ok := si.instanceScaler.(*scaler.WiseCloudScaler)
	if ok && constant.InstanceStatus(instance.InstanceStatus.Code) == constant.KernelInstanceStatusSubHealth &&
		(instance.InstanceStatus.ErrorCode == constant.KernelDataSystemUnavailable ||
			instance.InstanceStatus.ErrorCode == constant.KernelNPUFAULTErrCode) {
		go wiseScale.DelNuwaPod(instance)
	}
}

// HandleInstanceDelete handles instance delete comes from ETCD
func (si *ScaledInstanceQueue) HandleInstanceDelete(instance *types.Instance) {
	logger := log.GetLogger().With(zap.Any("instanceID", instance.InstanceID), zap.Any("funcKey", si.funcKeyWithRes))
	logger.Infof("handling instance delete")
	if err := si.instanceScheduler.DelInstance(instance); err != nil {
		logger.Errorf("failed to delete Instance error %s", err.Error())
	}
}

// HandleFuncSpecUpdate handles instance metadata update
// 1. if concurrentNum changes, clean all instances since we can't modify runtime concurrency here
// 2. if resourceMetadata changes, only reserved instance will handle
func (si *ScaledInstanceQueue) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	logger := log.GetLogger().With(zap.Any("funcKeyWithRes", si.funcKeyWithRes))
	logger.Infof("handling funcSpec update")
	needUpdate := false
	si.Cond.L.Lock()
	if si.funcSig != funcSpec.FuncMetaSignature {
		log.GetLogger().Warnf("signature changes from %s to %s", si.funcSig, funcSpec.FuncMetaSignature)
		si.funcSig = funcSpec.FuncMetaSignature
		// only reserved instance needs to handle resource update
		if si.instanceType == types.InstanceTypeReserved ||
			funcSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyPredict {
			si.resKey = resspeckey.ConvertToResSpecKey(resspeckey.ConvertResourceMetaDataToResSpec(
				funcSpec.ResourceMetaData))
		}
		// reset create error when funcSpec is updated since user code may be modified and unrecoverable error may be
		// resolved
		si.instanceScheduler.HandleCreateError(nil)
		si.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
		needUpdate = true
		si.updating = true
	}
	si.Cond.L.Unlock()
	if needUpdate {
		if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
			logger.Infof("cleaning old instances")
			si.cleanInstances()
		}
		logger.Infof("updating instanceScheduler and instanceScaler")
		si.instanceScheduler.HandleFuncSpecUpdate(funcSpec)
		si.instanceScaler.HandleFuncSpecUpdate(funcSpec)
		// if unrecoverable createError happens, instanceScaler may stop, try again since function is updated
		si.instanceScaler.SetEnable(si.isFuncOwner)
		si.Cond.L.Lock()
		si.updating = false
		si.Cond.L.Unlock()
		si.Cond.Broadcast()
	}
}

// HandleInstanceSync recover session record in scheduler
func (si *ScaledInstanceQueue) HandleInstanceSync(fn utils.RecoverSessionCallback) {
	if sc, ok := si.instanceScheduler.(*concurrencyscheduler.ScaledConcurrencyScheduler); ok {
		sc.RecoverSessionRecordFromDataSystem(fn)
	}
	if sc, ok := si.instanceScheduler.(*concurrencyscheduler.ReservedConcurrencyScheduler); ok {
		sc.RecoverSessionRecordFromDataSystem(fn)
	}
}

// HandleInsConfigUpdate updates instance configuration
func (si *ScaledInstanceQueue) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
	si.instanceScaler.HandleInsConfigUpdate(insConfig)
}

// EnableInstanceScale enable scaler scale replica
func (si *ScaledInstanceQueue) EnableInstanceScale() {
	si.L.Lock()
	isFuncOwner := si.isFuncOwner
	si.L.Unlock()
	si.instanceScaler.SetEnable(isFuncOwner)
}

// HandleFaultyInstance will remove a faulty instance
func (si *ScaledInstanceQueue) HandleFaultyInstance(instance *types.Instance) {
	log.GetLogger().Warnf("handling faulty instance %s for function %s", instance.InstanceID, si.funcKeyWithRes)
	if err := si.instanceScheduler.DelInstance(instance); err != nil {
		log.GetLogger().Warnf("delete Instance from instance queue failed, err: %v", err)
	}
	si.deleteInstance(instance)
}

// HandleAliasUpdate -
func (si *ScaledInstanceQueue) HandleAliasUpdate() {
	if !selfregister.GlobalSchedulerProxy.IsFuncOwner(si.funcKey) {
		log.GetLogger().Infof("no %s funcOwner, skip update alias to instance", si.funcKey)
		return
	}
	log.GetLogger().Infof("%s funcOwner, begin update alias to instance", si.funcKey)
	si.instanceScheduler.SignalAllInstances(func(instance *types.Instance) {
		si.signalInstanceFunc(instance, constant.KillSignalAliasUpdate)
	})
}

// HandleFaaSSchedulerUpdate -
func (si *ScaledInstanceQueue) HandleFaaSSchedulerUpdate() {
	// 去除对函数owner的判断，防止这样的场景，两个scheduler a和b，b由于接收事件延后，先推送了旧的事件给实例，但是在新的事件到来时，改变了scheduler对这个函数的owner，导致不推新的事件，导致脏数据残留
	log.GetLogger().Infof("begin update scheduler to %s's instance", si.funcKey)
	si.instanceScheduler.SignalAllInstances(func(instance *types.Instance) {
		si.signalInstanceFunc(instance, constant.KillSignalFaaSSchedulerUpdate)
	})
}

// GetInstanceNumber will get current instance number
func (si *ScaledInstanceQueue) GetInstanceNumber(onlySelf bool) int {
	return si.instanceScheduler.GetInstanceNumber(onlySelf)
}

// RecoverInstance recover instances from scheduler state
func (si *ScaledInstanceQueue) RecoverInstance(instanceMap map[string]*types.Instance) {
	si.instanceScaler.SetEnable(false)
	for _, instanceState := range instanceMap {
		si.instanceScheduler.HandleInstanceUpdate(instanceState)
		if si.funcSig != instanceState.FuncSig {
			si.HandleFaultyInstance(instanceState)
		}
	}
}

// Destroy will destroy instance queue and its components
func (si *ScaledInstanceQueue) Destroy() {
	log.GetLogger().Infof("destroy instance queue type %s for function %s", si.instanceType, si.funcKeyWithRes)
	commonUtils.SafeCloseChannel(si.stopCh)
	si.insCreateQueue.destroy()
	si.instanceScheduler.Destroy()
	si.instanceScaler.Destroy()
	si.cleanInstances()
	log.GetLogger().Debugf("destroy instance queue type %s for function %s completed", si.instanceType,
		si.funcKeyWithRes)
}

func (si *ScaledInstanceQueue) startScaleUpWorker() {
	for {
		select {
		case _, ok := <-si.stopCh:
			if !ok {
				log.GetLogger().Warnf("stop scale up worker for function %s", si.funcKeyWithRes)
				return
			}
		default:
		}
		// getForCreate blocks until a createReq is pushed in or insCreateQueue is destroyed
		if createReq := si.insCreateQueue.getForCreate(); createReq != nil {
			go si.scaleUpInstanceWithRetry(createReq.callback)
		}
	}
}

// ScaleUpHandler handles instance scale up planed by instanceScaler
func (si *ScaledInstanceQueue) ScaleUpHandler(insNum int, callback scaler.ScaleUpCallback) {
	for i := 0; i < insNum; i++ {
		si.insCreateQueue.push(&InstanceCreateRequest{callback: callback})
	}
	if insNum > 0 {
		log.GetLogger().Debugf("succeed to submit %d instance to scale up for function %s", insNum, si.funcKeyWithRes)
	}
}

func (si *ScaledInstanceQueue) scaleUpInstanceWithRetry(callback scaler.ScaleUpCallback) {
	var (
		instance  *types.Instance
		createErr error
	)
	var retryDelay time.Duration
	for retryDelay = initialErrorDelay; retryDelay <= retryDelayLimit; retryDelay *= time.Duration(retryDelayFactor) {
		select {
		case _, ok := <-si.funcCtx.Done():
			if !ok {
				log.GetLogger().Warnf("function %s is deleted, stop scale up instance now", si.funcKey)
				return
			}
		default:
		}
		instance, createErr = si.createInstance()
		if createErr == nil && instance != nil {
			break
		}
		if si.instanceType == types.InstanceTypeReserved &&
			config.GlobalConfig.ScaleRetryConfig.ReservedInstanceAlwaysRetry {
			continue
		}
		if utils.IsUnrecoverableError(createErr) {
			break
		}
		log.GetLogger().Warnf("failed to create type %s instance for function %s createErr %v should be retried "+
			"after %.0fs", si.instanceType, si.funcKey, createErr, retryDelay.Seconds())
		time.Sleep(retryDelay)
	}
	// offset pendingInsThdNum pre-increased in instanceScaler before add instance into instanceScheduler
	callback(1)
	if createErr != nil || instance == nil {
		// retry may finish after function is updated, handle create error again to trigger scale again
		si.instanceScaler.HandleCreateError(createErr)
		si.instanceScheduler.HandleCreateError(createErr)
		log.GetLogger().Errorf("failed to create instance after retry %s s for function %s create error %v",
			retryDelay, si.funcKeyWithRes, createErr)
		return
	}
	if si.signalInstanceFunc != nil {
		si.signalInstanceFunc(instance, constant.KillSignalAliasUpdate)
		si.signalInstanceFunc(instance, constant.KillSignalFaaSSchedulerUpdate)
	}
	// instance event may come before or after scale up process, so ErrInsAlreadyExist is ok
	if err := si.instanceScheduler.AddInstance(instance); err != nil && err != scheduler.ErrInsAlreadyExist {
		log.GetLogger().Errorf("failed to add instance to instanceScheduler for function %s error %s",
			si.funcKeyWithRes, err.Error())
	}
}

func (si *ScaledInstanceQueue) createInstance() (instance *types.Instance, createErr error) {
	defer func() {
		// nil should also be handled by instanceScaler and instanceScheduler
		si.instanceScaler.HandleCreateError(createErr)
		si.instanceScheduler.HandleCreateError(createErr)
	}()
	si.Cond.L.Lock()
	functionSignature := si.funcSig
	si.Cond.L.Unlock()
	startTime := time.Now()
	instance, createErr = si.createInstanceFunc("", "", si.instanceType, si.resKey, nil)
	if createErr != nil {
		log.GetLogger().Errorf("failed to create instance for function %s, error: %s", si.funcKeyWithRes,
			createErr.Error())
	} else {
		si.instanceScaler.UpdateCreateMetrics(time.Now().Sub(startTime))
	}
	select {
	case _, ok := <-si.funcCtx.Done():
		if !ok {
			log.GetLogger().Warnf("function %s is deleted, killing instance now", si.funcKey)
			createErr = ErrFunctionDeleted
		}
	default:
		// in case of function signature change during instance creating
		si.Cond.L.Lock()
		checkFunctionSignature := si.funcSig
		si.Cond.L.Unlock()
		if functionSignature != checkFunctionSignature {
			log.GetLogger().Errorf("function signature changes while creating instance for function %s, "+
				"killing instance now", si.funcKeyWithRes)
			createErr = ErrFuncSigMismatch
		}
	}
	if createErr != nil && instance != nil {
		log.GetLogger().Warnf("killing failed created instance %s for function %s", instance.InstanceID,
			si.funcKeyWithRes)
		go si.deleteInstance(instance)
	}
	return instance, createErr
}

// ScaleDownHandler handles instance scale down planed by instanceScaler, consider to handle delete retry in future
func (si *ScaledInstanceQueue) ScaleDownHandler(insNum int, callback scaler.ScaleDownCallback) {
	for i := 0; i < insNum; i++ {
		go func() {
			si.scaleDownInstance(callback)
		}()
	}
}

// consider to add retry with backoff process in future
func (si *ScaledInstanceQueue) scaleDownInstance(callback scaler.ScaleDownCallback) {
	select {
	case _, ok := <-si.funcCtx.Done():
		if !ok {
			log.GetLogger().Warnf("function %s is deleted, stop scale down instance now", si.funcKey)
			return
		}
	default:
	}
	// offset pendingInsThdNum pre-decreased in instanceScaler after pop instance from instanceScheduler
	defer callback(1)
	if createReq := si.insCreateQueue.getForCancel(); createReq != nil {
		return
	}
	if si.funcSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyPredict {
		if instance := si.instanceScheduler.PopInstance(true); instance != nil {
			si.deleteInstance(instance)
		}
	} else {
		if instance := si.instanceScheduler.PopInstance(false); instance != nil {
			si.deleteInstance(instance)
		}
	}
}

func (si *ScaledInstanceQueue) deleteInstance(instance *types.Instance) {
	log.GetLogger().Infof("deleting instance %s for function %s", instance.InstanceID, si.funcKeyWithRes)
	if scaler, isWiseCloudScaler := si.instanceScaler.(*scaler.WiseCloudScaler); isWiseCloudScaler {
		log.GetLogger().Warnf("deleting nuwa instance %s for function %s", instance.InstanceID,
			si.funcKeyWithRes)
		go scaler.DelNuwaPod(instance)
		return
	}
	if err := si.deleteInstanceFunc(instance); err != nil {
		log.GetLogger().Errorf("failed to delete instance %s function %s error %s", instance.InstanceID,
			si.funcKeyWithRes, err.Error())
	}
}

func (si *ScaledInstanceQueue) cleanInstances() {
	log.GetLogger().Infof("cleaning all instances for function %s", si.funcKeyWithRes)
	var instance *types.Instance
	for {
		if instance = si.instanceScheduler.PopInstance(true); instance != nil {
			go si.deleteInstance(instance)
		} else {
			break
		}
	}
}

// HandleFuncOwnerChange -
func (si *ScaledInstanceQueue) HandleFuncOwnerChange(fn utils.RecoverSessionCallback) {
	isFuncOwner := selfregister.GlobalSchedulerProxy.IsFuncOwner(si.funcKey)
	si.Cond.L.Lock()
	if si.isFuncOwner == isFuncOwner {
		si.Cond.L.Unlock()
		log.GetLogger().Infof("function owner change of %s doesn't affect this instance queue", si.funcKey)
		return
	}
	log.GetLogger().Infof("function owner of %s changes from %t to %t", si.funcKey, si.isFuncOwner, isFuncOwner)
	si.isFuncOwner = isFuncOwner
	si.Cond.L.Unlock()
	si.instanceScaler.SetEnable(false)
	// reassign in instanceScheduler first then reset instanceScaler
	si.instanceScheduler.HandleFuncOwnerUpdate(isFuncOwner)
	si.instanceScaler.SetEnable(isFuncOwner)
	// after owner change, need recover session info from datasystem
	if isFuncOwner {
		si.HandleInstanceSync(fn)
	}
}

// HandleRatioUpdate -
func (si *ScaledInstanceQueue) HandleRatioUpdate(ratio int) {
	si.Cond.L.Lock()
	isFuncOwner := si.isFuncOwner
	si.Cond.L.Unlock()
	if isFuncOwner {
		si.instanceScheduler.ReassignInstanceWhenGray(ratio)
	}
}
