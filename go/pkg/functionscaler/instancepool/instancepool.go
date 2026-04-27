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
	"context"
	"fmt"
	"reflect"
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
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/instancequeue"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/signalmanager"
	"yuanrong.org/kernel/pkg/functionscaler/stateinstance"
	"yuanrong.org/kernel/pkg/functionscaler/tenantquota"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	defaultSessionReaperInterval = 60 * time.Second
	faasManagerRequestTimeout    = 30 * time.Second
	vpcPullTriggerRequestTimeout = 30 * time.Second
	maxInstanceLimit             = 1000
	killSignalVal                = 1
	vpcOpCreatePATService        = "CreatePATService"
	vpcOpDeleteInstanceID        = "DeleteInstanceID"
	vpcOpCreatePullTrigger       = "CreatePullTrigger"
	vpcOpDeletePullTrigger       = "DeletePullTrigger"
	patProberInterval            = 5
	patProberTimeout             = 5
	patProberFailureThreshold    = 3
	resourcesCPU                 = "CPU"
	resourcesMemory              = "Memory"
	defaultEphemeralStorage      = 512
	resourcesEphemeralStorage    = "ephemeral-storage"
	podLabelInstanceType         = "instanceType"
	podLabelFuncName             = "funcName"
	podLabelIsPoolPod            = "isPoolPod"
	podLabelServiceID            = "serviceID"
	podLabelStandard             = "standard"
	podLabelTenantID             = "tenantID"
	podLabelVersion              = "version"
	podLabelSecurityGroup        = "securityGroup"
	executorFormat               = "default/0-system-faasExecutor%s/$latest"
	serveExecutor                = "default/0-system-serveExecutor/$latest"
	resSpecLen                   = 4
	faasExecutorStsCertPath      = "/opt/certs/%s/%s/%s.ini"
	defaultDelegateDirectoryInfo = "/tmp"
	base                         = 10
	defaultDirectoryLimit        = 512
	annotationFullFuncName       = "funcName"
)

const (
	invokeTypeEnvName   = "INVOKE_TYPE"
	invokeTypeEnvValue  = "faas"
	cceAscendAnnotation = "scheduling.cce.io/gpu-topology-placement"
)

const (
	biEvnTenantID        = "x-system-tenantId"
	biEvnFunctionName    = "x-system-functionName"
	biEvnFunctionVersion = "x-system-functionVersion"
	biEvnRegion          = "x-system-region"
	biEvnClusterID       = "x-system-clusterID"
	biEvnNodeIP          = "x-system-NODE_IP"
	biEvnPodName         = "x-system-podName"
	podIPEnv             = "POD_IP"
	podNameEnvNew        = "POD_NAME"
	hostIPEnv            = "HOST_IP"
	podNameEnv           = "PodName"
	podIDEnv             = "POD_ID"
)

const (
	stateInstanceDelete = "instanceDelete"
	stateDelete         = "delete"
	stateUpdate         = "update"
)

type createOption struct {
	traceID       string
	callerPodName string
}

// CreateParams is used to send config to runtime during instance initialization
type CreateParams struct {
	InstanceLabel     string `json:"instanceLabel,omitempty"`
	EventCreateParams `json:",inline"`
	HTTPCreateParams  `json:",inline"`
}

// EventCreateParams is used to send config to runtime during instance initialization
type EventCreateParams struct {
	UserInitEntry  string `json:"userInitEntry,omitempty"`
	UserCallEntry  string `json:"userCallEntry,omitempty"`
	UserStateEntry string `json:"userStateEntry,omitempty"`
}

// HTTPCreateParams is used to send config of http function during instance initialization
type HTTPCreateParams struct {
	Port      int    `json:"port,omitempty"`
	InitRoute string `json:"initRoute,omitempty"`
	CallRoute string `json:"callRoute,omitempty"`
}

type patSvcCreateResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type createInstanceRequest struct {
	createEvent     []byte
	traceID         string
	instanceName    string
	callerPodName   string
	poolLabel       string
	poolID          string
	funcSpec        *types.FunctionSpecification
	nuwaRuntimeInfo *wisecloudTypes.NuwaRuntimeInfo
	instanceType    types.InstanceType
	resKey          resspeckey.ResSpecKey
	instanceBuilder types.InstanceBuilder
	faasManagerInfo faasManagerInfo
	createTimeout   time.Duration
}

type sessionRecord struct {
	instance   *types.Instance
	sessionCtx context.Context
}

// InstancePool defines operations of instance pool
type InstancePool interface {
	CreateInstance(insCrtReq *types.InstanceCreateRequest) (*types.Instance, snerror.SNError)
	DeleteInstance(instance *types.Instance) snerror.SNError
	AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError)
	ReleaseInstance(instance *types.InstanceAllocation)
	HandleFunctionEvent(eventType registry.EventType, funcSpec *types.FunctionSpecification)
	HandleAliasEvent(eventType registry.EventType, aliasUrn string)
	HandleFaaSSchedulerEvent()
	HandleInstanceEvent(eventType registry.EventType, instance *types.Instance)
	HandleInstanceConfigEvent(eventType registry.EventType, insConfig *instanceconfig.Configuration)
	UpdateInvokeMetrics(resKey resspeckey.ResSpecKey, insMetrics *types.InstanceThreadMetrics)
	HandleFaaSManagerUpdate(faasManagerInfo faasManagerInfo)
	GetFuncSpec() *types.FunctionSpecification
	RecoverInstance(*types.FunctionSpecification, *types.InstancePoolState, bool, *sync.WaitGroup)
	GetAndDeleteState(stateID string) bool
	DeleteStateInstance(stateID string, instaceID string)
	handleManagedChange()
	handleRatioChange(ratio int)
	CleanOrphansInstanceQueue()
	QuerySession(sessionID string) (string, error)
}

// GenericInstancePool is a generic instance pool to manage instances of a specific function
type GenericInstancePool struct {
	FuncSpec              *types.FunctionSpecification
	defaultResSpec        *resspeckey.ResourceSpecification
	insConfig             map[resspeckey.ResSpecKey]*instanceconfig.Configuration
	metricsCollector      map[resspeckey.ResSpecKey]metrics.Collector
	insAcqReqQueue        map[resspeckey.ResSpecKey]*requestqueue.InsAcqReqQueue
	onDemandInstanceQueue map[resspeckey.ResSpecKey]*instancequeue.OnDemandInstanceQueue
	reservedInstanceQueue map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue
	scaledInstanceQueue   map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue
	sessionRecordMap      map[string]sessionRecord
	instanceSessionMap    map[string]map[string]struct{}
	stateInstanceID       sync.Map
	defaultResKey         resspeckey.ResSpecKey
	stateRoute            StateRoute
	faasManagerInfo       faasManagerInfo
	createTimeout         time.Duration
	sessionReaperInterval time.Duration
	stopCh                chan struct{}
	waitInsConfigChan     chan struct{}
	functionSignature     string
	currentPoolLabel      string
	defaultPoolLabel      string
	minScaleUpdatedTime   time.Time
	pendingInstanceNum    map[string]int
	minScaleAlarmSign     map[string]bool

	synced bool
	sync.RWMutex
	closeChanOnce sync.Once
}

// GetAndDeleteState delete state and instance, return whether the state exists
func (gi *GenericInstancePool) GetAndDeleteState(stateID string) bool {
	return gi.stateRoute.GetAndDeleteState(stateID)
}

// DeleteStateInstance called by ReleaseLease
func (gi *GenericInstancePool) DeleteStateInstance(stateID string, instanceID string) {
	gi.stateRoute.DeleteStateInstance(stateID, instanceID)
}

func (gi *GenericInstancePool) recoverStateRouteMap(stateInstanceMap map[string]*types.Instance) {
	gi.stateRoute.recover(stateInstanceMap)
}

// NewGenericInstancePool creates a GenericInstancePool
func NewGenericInstancePool(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo) (InstancePool,
	error) {
	log.GetLogger().Infof("create instance pool for function %s", funcSpec.FuncKey)
	defaultResSpec := resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData)
	defaultResKey := resspeckey.ConvertToResSpecKey(defaultResSpec)
	pool := &GenericInstancePool{
		FuncSpec:              funcSpec,
		defaultResSpec:        defaultResSpec,
		defaultPoolLabel:      funcSpec.InstanceMetaData.PoolLabel,
		currentPoolLabel:      funcSpec.InstanceMetaData.PoolLabel,
		insConfig:             make(map[resspeckey.ResSpecKey]*instanceconfig.Configuration, utils.DefaultMapSize),
		metricsCollector:      make(map[resspeckey.ResSpecKey]metrics.Collector, utils.DefaultMapSize),
		insAcqReqQueue:        make(map[resspeckey.ResSpecKey]*requestqueue.InsAcqReqQueue, utils.DefaultMapSize),
		onDemandInstanceQueue: make(map[resspeckey.ResSpecKey]*instancequeue.OnDemandInstanceQueue, utils.DefaultMapSize),
		reservedInstanceQueue: make(map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue, utils.DefaultMapSize),
		scaledInstanceQueue:   make(map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue, utils.DefaultMapSize),
		sessionRecordMap:      make(map[string]sessionRecord, utils.DefaultMapSize),
		instanceSessionMap:    make(map[string]map[string]struct{}, utils.DefaultMapSize),
		defaultResKey:         defaultResKey,
		createTimeout:         utils.GetCreateTimeout(funcSpec),
		sessionReaperInterval: defaultSessionReaperInterval,
		faasManagerInfo:       faasManagerInfo,
		functionSignature:     funcSpec.FuncMetaSignature,
		minScaleUpdatedTime:   time.Now(),
		waitInsConfigChan:     make(chan struct{}),
		pendingInstanceNum:    make(map[string]int),
		minScaleAlarmSign:     make(map[string]bool),
	}
	pool.stateRoute = StateRoute{
		funcSpec:           funcSpec,
		stateRoute:         make(map[string]*StateInstance, utils.DefaultMapSize),
		stateLeaseManager:  make(map[string]*stateinstance.Leaser, utils.DefaultMapSize),
		stateConfig:        funcSpec.FuncMetaData.StateConfig,
		resSpec:            defaultResSpec,
		deleteInstanceFunc: pool.deleteInstance,
		createInstanceFunc: pool.createInstanceAndAddCallerPodName,
		leaseInterval:      time.Duration(config.GlobalConfig.AutoScaleConfig.ScaleDownTime) * time.Millisecond,
		RWMutex:            sync.RWMutex{},
		logger:             log.GetLogger().With(zap.Any("funcKey", funcSpec.FuncKey)),
	}
	if reservedInstanceQueue, err := pool.createInstanceQueue(types.InstanceTypeReserved, defaultResKey); err == nil {
		pool.reservedInstanceQueue[defaultResKey] = reservedInstanceQueue.(*instancequeue.ScaledInstanceQueue)
	} else {
		return nil, err
	}
	if scaledInstanceQueue, err := pool.createInstanceQueue(types.InstanceTypeScaled, defaultResKey); err == nil {
		pool.scaledInstanceQueue[defaultResKey] = scaledInstanceQueue.(*instancequeue.ScaledInstanceQueue)
	} else {
		return nil, err
	}
	if onDemandInstanceQueue, err := pool.createInstanceQueue(types.InstanceTypeOnDemand, defaultResKey); err == nil {
		pool.onDemandInstanceQueue[defaultResKey] = onDemandInstanceQueue.(*instancequeue.OnDemandInstanceQueue)
	} else {
		return nil, err
	}
	go pool.instanceSessionReaper()
	return pool, nil
}

func (gi *GenericInstancePool) createInstanceQueue(instanceType types.InstanceType, resKey resspeckey.ResSpecKey) (
	instancequeue.InstanceQueue, snerror.SNError) {
	metricsCollector, exist := gi.metricsCollector[resKey]
	if !exist {
		metricsCollector = metrics.NewBucketMetricsCollector(gi.FuncSpec.FuncKey, resKey.String())
		gi.metricsCollector[resKey] = metricsCollector
	}
	insAcqReqQueue, exist := gi.insAcqReqQueue[resKey]
	if !exist {
		funcKeyWithRes := utils.GenFuncKeyWithRes(gi.FuncSpec.FuncKey, resKey.String())
		requestTimeout := utils.GetRequestTimeout(gi.FuncSpec)
		insAcqReqQueue = requestqueue.NewInsAcqReqQueue(funcKeyWithRes, requestTimeout)
		gi.insAcqReqQueue[resKey] = insAcqReqQueue
	}
	insQueConfig := &instancequeue.InsQueConfig{
		FuncSpec:           gi.FuncSpec,
		InsThdReqQueue:     insAcqReqQueue,
		InstanceType:       instanceType,
		ResKey:             resKey,
		MetricsCollector:   metricsCollector,
		CreateInstanceFunc: gi.createInstance,
		DeleteInstanceFunc: gi.deleteInstance,
		SignalInstanceFunc: gi.handleSignal,
	}
	instanceQueue, err := instancequeue.BuildInstanceQueue(insQueConfig, insAcqReqQueue, metricsCollector)
	if err != nil {
		log.GetLogger().Errorf("failed to create %s instance queue for function %s of resource %+v error %s",
			instanceType, gi.FuncSpec.FuncKey, resKey, err.Error())
		return nil, snerror.New(statuscode.StatusInternalServerError, err.Error())
	}
	log.GetLogger().Infof("create %s instance queue for function %s of resource %+v", instanceType,
		gi.FuncSpec.FuncKey, resKey)
	return instanceQueue, nil
}

// RecoverInstance recover instance pool, reserved pool and scaled pool
func (gi *GenericInstancePool) RecoverInstance(funcSpec *types.FunctionSpecification,
	instancePoolState *types.InstancePoolState, deleteFunc bool, wg *sync.WaitGroup) {
	defer wg.Done()
	// existInstanceIdMap got from etcd
	instanceIDMapsFromEtcd := registry.GlobalRegistry.InstanceRegistry.GetFunctionInstanceIDMap()
	StateInstanceMap := instancePoolState.StateInstance
	if len(StateInstanceMap) != 0 {
		filterStateInstanceMap(StateInstanceMap, instanceIDMapsFromEtcd, gi.FuncSpec.FuncKey)
		gi.recoverStateRouteMap(StateInstanceMap)
	}
}

// filterInstanceIDMap delete keys in filterMap but not in existsMap
func (gi *GenericInstancePool) filterInstanceIDMap(instanceMapFromState map[string]*types.Instance,
	instanceMapFromEtcd map[string]map[string]*commonTypes.InstanceSpecification, instanceType string) {
	if instanceMapFromEtcd == nil {
		for id := range instanceMapFromState {
			delete(instanceMapFromState, id)
		}
		return
	}
	// delete extra instance in state
	for id := range instanceMapFromState {
		var etcdIDMap map[string]*commonTypes.InstanceSpecification
		var ok bool
		if etcdIDMap, ok = instanceMapFromEtcd[gi.FuncSpec.FuncKey]; !ok {
			delete(instanceMapFromState, id)
			continue
		}
		if _, exist := etcdIDMap[id]; !exist {
			delete(instanceMapFromState, id)
		} else {
			oldStatus := instanceMapFromState[id].InstanceStatus
			instanceMapFromState[id].InstanceStatus = etcdIDMap[id].InstanceStatus
			log.GetLogger().Infof("instanceFromState instanceStatus update: oldStatus: %v -> newStatus: %v, "+
				"instanceId: %v", oldStatus, etcdIDMap[id].InstanceStatus, id)
		}
	}
	if etcdIDMap, ok := instanceMapFromEtcd[gi.FuncSpec.FuncKey]; ok {
		// kill new instance
		for id, insSpec := range etcdIDMap {
			if _, exist := instanceMapFromState[id]; !exist &&
				insSpec.CreateOptions[types.FunctionKeyNote] == gi.FuncSpec.FuncKey &&
				insSpec.CreateOptions[types.InstanceTypeNote] == instanceType {
				if err := gi.deleteInstance(&types.Instance{InstanceID: id}); err != nil {
					continue
				}
			}
		}
	}
}

// filterStateInstanceMap delete keys in filterMap but not in existsMap
func filterStateInstanceMap(filterMap map[string]*types.Instance,
	existsMap map[string]map[string]*commonTypes.InstanceSpecification, funcKey string) {
	if etcdIDMap, ok := existsMap[funcKey]; ok {
		for _, instance := range filterMap {
			if existInstance, exist := etcdIDMap[instance.InstanceID]; !exist {
				instance.InstanceStatus.Code = int32(constant.KernelInstanceStatusExited)
			} else {
				instance.InstanceStatus = existInstance.InstanceStatus
			}
		}
	}
}

// CreateInstance will create an instance
func (gi *GenericInstancePool) CreateInstance(insCrtReq *types.InstanceCreateRequest) (*types.Instance,
	snerror.SNError) {
	log.GetLogger().Infof("start to create instance for function %s with instanceName %s traceID %s",
		gi.FuncSpec.FuncKey, insCrtReq.InstanceName, insCrtReq.TraceID)
	select {
	case <-gi.FuncSpec.FuncCtx.Done():
		log.GetLogger().Errorf("function %s is deleted, can not create instance", gi.FuncSpec.FuncKey)
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
		var (
			resSpec *resspeckey.ResourceSpecification
			err     snerror.SNError
		)
		if utils.IsResSpecEmpty(insCrtReq.ResSpec) {
			resSpec = gi.defaultResSpec
		} else {
			resSpec = insCrtReq.ResSpec
		}
		resKey := resspeckey.ConvertToResSpecKey(resSpec)
		onDemandInstanceQueue, err := gi.acquireOnDemandInstanceQueue(resKey)
		if err != nil {
			log.GetLogger().Errorf("failed to acquire on-demand instance queue of function %s error %s",
				gi.FuncSpec.FuncKey, err.Error())
			return nil, err
		}
		return onDemandInstanceQueue.CreateInstance(insCrtReq)
	}
}

// DeleteInstance will delete an instance
func (gi *GenericInstancePool) DeleteInstance(instance *types.Instance) snerror.SNError {
	log.GetLogger().Infof("start to delete instance for function %s with instanceName %s", gi.FuncSpec.FuncKey,
		instance.InstanceName)
	select {
	case <-gi.FuncSpec.FuncCtx.Done():
		log.GetLogger().Errorf("function %s is deleted, can not delete instance", gi.FuncSpec.FuncKey)
		return snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
		gi.RLock()
		onDemandInstanceQueue, exist := gi.onDemandInstanceQueue[instance.ResKey]
		if !exist {
			gi.RUnlock()
			return snerror.New(statuscode.StatusInternalServerError, "on-demand instance queue not exist")
		}
		gi.RUnlock()
		return onDemandInstanceQueue.DeleteInstance(instance)
	}
}

func (gi *GenericInstancePool) acquireStateInstanceThread(insAcqReq *types.InstanceAcquireRequest) (
	*types.InstanceAllocation, snerror.SNError) {
	/* stateID +  satateful func -> ok
	nil stateID +  satateful func -> err
	stateID +  non-satateful func -> err
	nil stateID +  non-satateful func -> ok
	*/
	if !gi.FuncSpec.FuncMetaData.IsStatefulFunction || len(insAcqReq.StateID) == 0 {
		return nil, snerror.New(statuscode.StateMismatch, statuscode.StateMismatchErrMsg)
	}
	return gi.stateRoute.acquireStateInstanceThread(insAcqReq)
}

// AcquireInstance will acquire an instance
func (gi *GenericInstancePool) AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation,
	snerror.SNError) {
	logger := log.GetLogger().With(zap.Any("traceId", insAcqReq.TraceID), zap.Any("funcKey", gi.FuncSpec.FuncKey),
		zap.Any("designatedInstance", insAcqReq.DesignateInstanceID))
	logger.Debugf("acquire instance")
	select {
	case <-gi.FuncSpec.FuncCtx.Done():
		logger.Errorf("function is deleted, can not acquire instance")
		return nil, snerror.New(statuscode.FuncMetaNotFoundErrCode, statuscode.FuncMetaNotFoundErrMsg)
	default:
		if insAcqReq.StateID != "" {
			return gi.acquireStateInstanceThread(insAcqReq)
		}
		if utils.IsResSpecEmpty(insAcqReq.ResSpec) {
			return nil, snerror.New(statuscode.InternalErrorCode, statuscode.InternalErrorMessage)
		}
		var (
			insAlloc *types.InstanceAllocation
			err      snerror.SNError
		)
		defer func() {
			if insAlloc != nil && len(insAlloc.SessionInfo.SessionID) != 0 {
				gi.recordInstanceSession(insAlloc)
			}
		}()
		if len(insAcqReq.InstanceSession.SessionID) != 0 {
			gi.processInstanceSession(insAcqReq)
		}
		gi.currentPoolLabel = insAcqReq.PoolLabel
		resKey := resspeckey.ConvertToResSpecKey(insAcqReq.ResSpec)
		// 当label有值但是没有对于label的实例配置时，直接返回报错
		gi.RLock()
		if insAcqReq.ResSpec.InvokeLabel != DefaultInstanceLabel && gi.insConfig[resKey] == nil {
			gi.RUnlock()
			return nil, snerror.New(statuscode.InstanceLabelNotFoundErrCode, statuscode.InstanceLabelNotFoundErrMsg)
		}
		gi.RUnlock()
		logger = logger.With(zap.Any("resource", resKey))
		if len(insAcqReq.InstanceName) != 0 {
			onDemandInstanceQueue, err := gi.acquireOnDemandInstanceQueue(resKey)
			if err != nil {
				logger.Errorf("failed to acquire on-demand instance queue of function, error %s", err.Error())
				return nil, err
			}
			insAlloc, err = onDemandInstanceQueue.AcquireInstance(insAcqReq)
			return insAlloc, err
		}
		if !insAcqReq.TrafficLimited {
			reservedInstanceQueue, err := gi.acquireReservedInstanceQueue(resKey)
			if err != nil {
				logger.Errorf("failed to acquire reserved instance queue of function, error %s", err.Error())
				return nil, err
			}
			insAlloc, err = reservedInstanceQueue.AcquireInstance(insAcqReq)
			if insAlloc != nil {
				logger.Infof("acquired reserved instance thread %s of function", insAlloc.AllocationID)
				return insAlloc, nil
			}
			if err.Code() != statuscode.NoInstanceAvailableErrCode && err.Code() != statuscode.InstanceNotFoundErrCode {
				logger.Errorf("failed to acquire reserved instance of function, error %s", err.Error())
				return nil, err
			}
		}
		insAlloc, err = gi.acquireInstanceFromScaleQueueWithBackup(resKey, insAcqReq, logger)
		return insAlloc, err
	}
}

func (gi *GenericInstancePool) acquireInstanceFromScaleQueueWithBackup(resKey resspeckey.ResSpecKey,
	insAcqReq *types.InstanceAcquireRequest, logger api.FormatLogger) (*types.InstanceAllocation, snerror.SNError) {
	var backupQueue []instancequeue.InstanceQueue
	gi.RLock()
	for key, queue := range gi.reservedInstanceQueue {
		if resKey != key && resKey.InvokeLabel == key.InvokeLabel && queue.GetInstanceNumber(false) > 0 {
			backupQueue = append(backupQueue, queue)
		}
	}
	for key, queue := range gi.scaledInstanceQueue {
		if resKey != key && resKey.InvokeLabel == key.InvokeLabel && queue.GetInstanceNumber(false) > 0 {
			backupQueue = append(backupQueue, queue)
		}
	}
	gi.RUnlock()
	if len(backupQueue) > 0 {
		logger.Infof("has backup queue, will skip cold start")
		insAcqReq.SkipWaitPending = true
	}
	scaledInstanceQueue, err := gi.acquireScaleInstanceQueue(resKey)
	if err != nil {
		logger.Errorf("failed to acquire scaled instance queue of function, error %s", err.Error())
		return nil, err
	}
	insAlloc, err := scaledInstanceQueue.AcquireInstance(insAcqReq)
	if insAlloc == nil {
		for _, queue := range backupQueue {
			insAlloc, err = queue.AcquireInstance(insAcqReq)
			if insAlloc != nil {
				logger.Infof("acquired backup instance thread %s of function", insAlloc.AllocationID)
				return insAlloc, nil
			}
		}
		logger.Errorf("failed to acquire scaled instance of function, error %s", err.Error())
		return nil, err
	}
	logger.Infof("acquired scaled instance insAlloc %s of function", insAlloc.AllocationID)
	return insAlloc, nil
}

// ReleaseInstance will release an instance
func (gi *GenericInstancePool) ReleaseInstance(insAlloc *types.InstanceAllocation) {
	instance := insAlloc.Instance
	var err snerror.SNError
	switch instance.InstanceType {
	case types.InstanceTypeReserved:
		if gi.reservedInstanceQueue != nil && gi.reservedInstanceQueue[instance.ResKey] != nil {
			err = gi.reservedInstanceQueue[instance.ResKey].ReleaseInstance(insAlloc)
		}
	case types.InstanceTypeScaled:
		gi.RLock()
		scaledInstanceQueue, exist := gi.scaledInstanceQueue[instance.ResKey]
		gi.RUnlock()
		if !exist {
			err = snerror.New(statuscode.StatusInternalServerError,
				"instance queue with this resource doesn't exist")
			break
		}
		err = scaledInstanceQueue.ReleaseInstance(insAlloc)
	default:
		log.GetLogger().Errorf("unsupported instance type")
	}
	if err != nil && err.Code() == statuscode.InstanceNotFoundErrCode {
		gi.cleanInstanceSession(instance.InstanceID)
	}
	if err != nil {
		log.GetLogger().Errorf("failed to release instance insAlloc %s error %s", insAlloc.AllocationID, err.Error())
		return
	}
	log.GetLogger().Infof("released instance insAlloc %s of function %s resource %+v", insAlloc.AllocationID,
		gi.FuncSpec.FuncKey, instance.ResKey)
}

// HandleFaaSManagerUpdate handles faas manager update
func (gi *GenericInstancePool) HandleFaaSManagerUpdate(faasManagerInfo faasManagerInfo) {
	gi.Lock()
	oldInstanceID := gi.faasManagerInfo.instanceID
	gi.faasManagerInfo = faasManagerInfo
	var funcKey string
	if gi.FuncSpec != nil {
		funcKey = gi.FuncSpec.FuncKey
	}
	log.GetLogger().Infof("succeed to update faas-manager info for func %s, oldInstanceID: %s, "+
		"newInstanceID: %s", funcKey, oldInstanceID, faasManagerInfo.instanceID)
	gi.Unlock()
}

// HandleFunctionEvent handles function event
func (gi *GenericInstancePool) HandleFunctionEvent(eventType registry.EventType,
	funcSpec *types.FunctionSpecification) {
	log.GetLogger().Infof("handling event type %s for function %s", eventType, gi.FuncSpec.FuncKey)
	if eventType == registry.SubEventTypeUpdate {
		gi.HandleFunctionUpdateEvent(funcSpec)
	}
	if eventType == registry.SubEventTypeDelete {
		gi.HandleFunctionDeleteEvent()
	}
}

// HandleFunctionDeleteEvent -
func (gi *GenericInstancePool) HandleFunctionDeleteEvent() {
	for _, mc := range gi.metricsCollector {
		mc.Stop()
	}
	gi.RLock()
	if gi.reservedInstanceQueue != nil {
		for _, queue := range gi.reservedInstanceQueue {
			queue.Destroy()
		}
	}
	for _, queue := range gi.scaledInstanceQueue {
		queue.Destroy()
	}
	for _, insThdReqQueue := range gi.insAcqReqQueue {
		insThdReqQueue.Stop()
	}
	gi.RUnlock()
	gi.stateRoute.Destroy() // 租约怎么办
}

// HandleFunctionUpdateEvent -
func (gi *GenericInstancePool) HandleFunctionUpdateEvent(funcSpec *types.FunctionSpecification) {
	gi.Lock()
	// watch of instance need this funcSpec to set currentNum in instance struct, so this funcSpec needs to be
	// updated
	preResSpec := gi.defaultResSpec
	gi.FuncSpec = funcSpec
	gi.defaultResSpec = resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData)
	gi.defaultResKey = resspeckey.ConvertToResSpecKey(gi.defaultResSpec)
	if !reflect.DeepEqual(preResSpec, gi.defaultResSpec) && config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		reservedInstanceQueueMap := make(map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue, utils.DefaultMapSize)
		for resKey, configuration := range gi.insConfig {
			resSpec := gi.defaultResSpec.DeepCopy()
			resSpec.InvokeLabel = resKey.InvokeLabel
			newResKey := resspeckey.ConvertToResSpecKey(resSpec)
			reservedInstanceQueue, err := gi.createInstanceQueue(types.InstanceTypeReserved, newResKey)
			if err != nil {
				log.GetLogger().Errorf("failed to create reserved instance queue for function %s during function "+
					"update error %s", gi.FuncSpec.FuncKey, err.Error())
			} else {
				if queue, exist := gi.reservedInstanceQueue[resKey]; exist {
					log.GetLogger().Debugf("reservedInstanceQueue for function %s destroy start",
						gi.FuncSpec.FuncKey)
					queue.Destroy()
				}
				reservedInstanceQueueMap[newResKey] = reservedInstanceQueue.(*instancequeue.ScaledInstanceQueue)
				reservedInstanceQueueMap[newResKey].HandleInsConfigUpdate(configuration)
				reservedInstanceQueueMap[newResKey].EnableInstanceScale()
			}
		}
		gi.reservedInstanceQueue = reservedInstanceQueueMap
		scaledInstanceQueue, err := gi.createInstanceQueue(types.InstanceTypeScaled, gi.defaultResKey)
		if err != nil {
			log.GetLogger().Errorf("failed to create scaled instance queue for function %s during function "+
				"update error %s", gi.FuncSpec.FuncKey, err.Error())
		} else {
			gi.scaledInstanceQueue[gi.defaultResKey] = scaledInstanceQueue.(*instancequeue.ScaledInstanceQueue)
		}
	}
	gi.createTimeout = utils.GetCreateTimeout(funcSpec)
	gi.defaultPoolLabel = funcSpec.InstanceMetaData.PoolLabel
	gi.currentPoolLabel = funcSpec.InstanceMetaData.PoolLabel
	gi.minScaleUpdatedTime = time.Now()
	gi.Unlock()
	gi.RLock()
	if gi.reservedInstanceQueue != nil {
		for resKey, queue := range gi.reservedInstanceQueue {
			if queue.GetSchedulerPolicy() != getFuncSpecSchedulePolicy(gi.FuncSpec.InstanceMetaData.SchedulePolicy) {
				err := gi.resetInstanceScheduler(queue, resKey)
				if err != nil {
					log.GetLogger().Errorf("%s failed to reset instance scheduler, from %s to %s, err: %s",
						queue.GetSchedulerPolicy(), gi.FuncSpec.FuncKey, gi.FuncSpec.InstanceMetaData.SchedulePolicy,
						err.Error())
				}
			}
			queue.HandleFuncSpecUpdate(funcSpec)
		}
	}
	for resKey, queue := range gi.scaledInstanceQueue {
		if queue.GetSchedulerPolicy() != getFuncSpecSchedulePolicy(gi.FuncSpec.InstanceMetaData.SchedulePolicy) {
			err := gi.resetInstanceScheduler(queue, resKey)
			if err != nil {
				log.GetLogger().Errorf("%s failed to reset instance scheduler, from %s to %s, err: %s",
					queue.GetSchedulerPolicy(), gi.FuncSpec.FuncKey, gi.FuncSpec.InstanceMetaData.SchedulePolicy,
					err.Error())
			}
		}
		queue.HandleFuncSpecUpdate(funcSpec)
	}
	for _, insThdReqQueue := range gi.insAcqReqQueue {
		insThdReqQueue.UpdateRequestTimeout(utils.GetRequestTimeout(funcSpec))
	}
	gi.RUnlock()
	// todo 之后再考虑状态函数 函数元信息变更事件, 考虑两种情况，有状态无状态之间切换、有状态函数其他原信息变更
}

func getFuncSpecSchedulePolicy(schedulePolicy string) string {
	if schedulePolicy == "" {
		return types.InstanceSchedulePolicyConcurrency
	}
	return schedulePolicy
}

func (gi *GenericInstancePool) resetInstanceScheduler(instanceQueue *instancequeue.ScaledInstanceQueue,
	resKey resspeckey.ResSpecKey) error {
	log.GetLogger().Debugf("%s reset instance scheduler, from %s to %s", gi.FuncSpec.FuncKey,
		instanceQueue.GetSchedulerPolicy(), gi.FuncSpec.InstanceMetaData.SchedulePolicy)
	insAcqReqQueue, exist := gi.insAcqReqQueue[resKey]
	if !exist {
		funcKeyWithRes := utils.GenFuncKeyWithRes(gi.FuncSpec.FuncKey, resKey.String())
		requestTimeout := utils.GetRequestTimeout(gi.FuncSpec)
		insAcqReqQueue = requestqueue.NewInsAcqReqQueue(funcKeyWithRes, requestTimeout)
		gi.insAcqReqQueue[resKey] = insAcqReqQueue
	}
	instanceQueue.L.Lock()
	defer instanceQueue.L.Unlock()
	var currentInstance []*types.Instance
	oldInstanceScheduler := instanceQueue.GetInstanceScheduler()
	for {
		ins := oldInstanceScheduler.PopInstance(true)
		if ins == nil {
			break
		}
		currentInstance = append(currentInstance, ins)
	}
	oldInstanceScheduler.Destroy()
	err := instancequeue.AssembleScheduler(gi.FuncSpec.InstanceMetaData.SchedulePolicy, instanceQueue, insAcqReqQueue)
	if err != nil {
		return err
	}
	instanceQueue.ReconnectWithScaler()
	for _, instance := range currentInstance {
		instanceQueue.HandleInstanceUpdate(instance)
	}
	return nil
}

// HandleAliasEvent handles alias event
func (gi *GenericInstancePool) HandleAliasEvent(eventType registry.EventType, aliasUrn string) {
	log.GetLogger().Infof("pool %s handling event type %s for alias,urn:%s", gi.FuncSpec.FuncKey, eventType, aliasUrn)
	if eventType == registry.SubEventTypeUpdate || eventType == registry.SubEventTypeDelete {
		gi.updateInstanceAliasData()
	}
}

func (gi *GenericInstancePool) updateInstanceAliasData() {
	if gi.reservedInstanceQueue != nil {
		for _, queue := range gi.reservedInstanceQueue {
			queue.HandleAliasUpdate()
		}
	}
	for _, instanceQueue := range gi.scaledInstanceQueue {
		instanceQueue.HandleAliasUpdate()
	}
}

// HandleFaaSSchedulerEvent -
func (gi *GenericInstancePool) HandleFaaSSchedulerEvent() {
	if gi.reservedInstanceQueue != nil {
		for _, queue := range gi.reservedInstanceQueue {
			queue.HandleFaaSSchedulerUpdate()
		}
	}
	for _, instanceQueue := range gi.scaledInstanceQueue {
		instanceQueue.HandleFaaSSchedulerUpdate()
	}
}

// HandleInstanceEvent handles instance event
func (gi *GenericInstancePool) HandleInstanceEvent(eventType registry.EventType, instance *types.Instance) {
	if eventType == registry.SubEventTypeSynced {
		gi.handleInstanceSynced()
		return
	}
	instance.FuncKey = gi.FuncSpec.FuncKey
	logger := log.GetLogger().With(zap.Any("", "HandleInstanceEvent")).
		With(zap.Any("FuncKey", gi.FuncSpec.FuncKey)).
		With(zap.Any("ResKey", instance.ResKey)).
		With(zap.Any("InstanceID", instance.InstanceID)).With(zap.Any("eventType", eventType))
	// Reserved instance must have insConfig.
	insConfResKey := gi.defaultResKey
	insConfResKey.InvokeLabel = instance.ResKey.InvokeLabel
	gi.RLock()
	if _, ok := gi.insConfig[insConfResKey]; gi.synced && instance.InstanceType == types.InstanceTypeReserved && !ok {
		gi.RUnlock()
		go DeleteUnexpectInstance(instance.ParentID, instance.InstanceID, instance.FuncKey, logger)
		return
	}
	gi.RUnlock()
	logger.Infof("handling instance event")
	switch eventType {
	case registry.SubEventTypeUpdate:
		gi.handleInstanceUpdate(instance, logger)
	case registry.SubEventTypeDelete:
		gi.handleInstanceDelete(instance, logger)
	case registry.SubEventTypeRemove:
		gi.removeInstance(instance, logger)
	default:
		logger.Warnf("unsupported instance event type: %s", eventType)
	}
}

var defaultInstanceNeedProcessInstanceCodeMap = map[int32]struct{}{
	int32(constant.KernelInstanceStatusRunning):   {},
	int32(constant.KernelInstanceStatusSubHealth): {},
}

var reservedAndScaledInstanceNeedProcessInstanceCodeMap = map[int32]struct{}{
	int32(constant.KernelInstanceStatusRunning):   {},
	int32(constant.KernelInstanceStatusSubHealth): {},
	int32(constant.KernelInstanceStatusEvicting):  {}, // 为了支持绑定会话的实例，在优雅退出时，依旧能支持会话请求,
}

func getNeedProcessInstanceCodeMap(instanceType types.InstanceType) map[int32]struct{} {
	switch instanceType {
	case types.InstanceTypeReserved, types.InstanceTypeScaled:
		return reservedAndScaledInstanceNeedProcessInstanceCodeMap
	default:
		return defaultInstanceNeedProcessInstanceCodeMap
	}
}

var defaultFaultyInstanceStatusMap = map[int32]struct{}{
	int32(constant.KernelInstanceStatusFatal):          {},
	int32(constant.KernelInstanceStatusScheduleFailed): {},
	int32(constant.KernelInstanceStatusEvicted):        {},
	int32(constant.KernelInstanceStatusEvicting):       {},
	int32(constant.KernelInstanceStatusExiting):        {},
}

var reservedAndScaledFaultyInstanceStatusMap = map[int32]struct{}{
	int32(constant.KernelInstanceStatusFatal):          {},
	int32(constant.KernelInstanceStatusScheduleFailed): {},
	int32(constant.KernelInstanceStatusEvicted):        {},
	int32(constant.KernelInstanceStatusExiting):        {},
}

func getFaultyInstanceStatusMap(instanceType types.InstanceType) map[int32]struct{} {
	switch instanceType {
	case types.InstanceTypeReserved, types.InstanceTypeScaled:
		return reservedAndScaledFaultyInstanceStatusMap
	default:
		return defaultFaultyInstanceStatusMap
	}
}

func (gi *GenericInstancePool) handleInstanceUpdate(instance *types.Instance, logger api.FormatLogger) {
	instanceStatusCodeMap := getNeedProcessInstanceCodeMap(instance.InstanceType)
	if _, ok := instanceStatusCodeMap[instance.InstanceStatus.Code]; ok {
		if config.GlobalConfig.Scenario != types.ScenarioWiseCloud &&
			instance.FuncSig != gi.FuncSpec.FuncMetaSignature {
			logger.Errorf("handle event failed, function signature changes, killing instance now")
			gi.removeInstance(instance, logger)
			return
		}
		switch instance.InstanceType {
		case types.InstanceTypeOnDemand:
			onDemandInstanceQueue, err := gi.acquireOnDemandInstanceQueue(instance.ResKey)
			if err != nil {
				logger.Errorf("failed to acquire on-demand instance queue error %s", err.Error())
				break
			}
			onDemandInstanceQueue.HandleInstanceUpdate(instance)
		case types.InstanceTypeReserved:
			reservedInstanceQueue, err := gi.acquireReservedInstanceQueue(instance.ResKey)
			if err != nil {
				logger.Errorf("failed to acquire reserved instance queue error %s", err.Error())
				break
			}
			reservedInstanceQueue.HandleInstanceUpdate(instance)
		case types.InstanceTypeScaled:
			scaledInstanceQueue, err := gi.acquireScaleInstanceQueue(instance.ResKey)
			if err != nil {
				logger.Errorf("failed to acquire scaled instance queue error %s", err.Error())
				break
			}
			scaledInstanceQueue.HandleInstanceUpdate(instance)
		case types.InstanceTypeState:
			gi.stateRoute.HandleInstanceUpdate(instance)
		default:
			logger.Warnf("instance type %s update not implemented", instance.InstanceType)
		}
		gi.judgeExceedInstance(instance.ResKey, logger)
	}

	faultyInstanceStatusMap := getFaultyInstanceStatusMap(instance.InstanceType)
	if _, ok := faultyInstanceStatusMap[instance.InstanceStatus.Code]; ok {
		logger.Warnf("instance status is updated to %d, remove this instance now", instance.InstanceStatus.Code)
		gi.removeInstance(instance, logger)
	}
}

func (gi *GenericInstancePool) handleInstanceDelete(instance *types.Instance, logger api.FormatLogger) {
	signalmanager.GetSignalManager().RemoveInstance(instance.InstanceID)
	gi.cleanInstanceSession(instance.InstanceID)
	switch instance.InstanceType {
	case types.InstanceTypeOnDemand:
		gi.RLock()
		onDemandInstanceQueue, exist := gi.onDemandInstanceQueue[instance.ResKey]
		gi.RUnlock()
		if exist {
			onDemandInstanceQueue.HandleInstanceDelete(instance)
		}
	case types.InstanceTypeReserved:
		gi.RLock()
		reservedInstanceQueue, exist := gi.reservedInstanceQueue[instance.ResKey]
		gi.RUnlock()
		if exist {
			reservedInstanceQueue.HandleInstanceDelete(instance)
		}
	case types.InstanceTypeScaled:
		gi.RLock()
		scaledInstanceQueue, exist := gi.scaledInstanceQueue[instance.ResKey]
		gi.RUnlock()
		if exist {
			scaledInstanceQueue.HandleInstanceDelete(instance)
		}
	default:
		logger.Warnf("instance type %s update not implemented", instance.InstanceType)
	}
}

func (gi *GenericInstancePool) acquireOnDemandInstanceQueue(resKey resspeckey.ResSpecKey) (
	*instancequeue.OnDemandInstanceQueue, snerror.SNError) {
	var err snerror.SNError
	gi.RLock()
	instanceQueue, exist := gi.onDemandInstanceQueue[resKey]
	gi.RUnlock()
	if exist {
		return instanceQueue, err
	}
	gi.Lock()
	// 需要二次判断，防止重复创建
	instanceQueue, exist = gi.onDemandInstanceQueue[resKey]
	if !exist {
		log.GetLogger().Debugf("createInstanceQueue type OnDemand for function %s destroy start",
			gi.FuncSpec.FuncKey)
		var queue instancequeue.InstanceQueue
		queue, err = gi.createInstanceQueue(types.InstanceTypeOnDemand, resKey)
		if err == nil {
			instanceQueue = queue.(*instancequeue.OnDemandInstanceQueue)
			gi.onDemandInstanceQueue[resKey] = instanceQueue
		}
	}
	gi.Unlock()
	return instanceQueue, err
}

func (gi *GenericInstancePool) acquireReservedInstanceQueue(resKey resspeckey.ResSpecKey) (
	*instancequeue.ScaledInstanceQueue, snerror.SNError) {
	var snErr snerror.SNError
	gi.RLock()
	instanceQueue, exist := gi.reservedInstanceQueue[resKey]
	gi.RUnlock()
	if exist {
		return instanceQueue, snErr
	}
	// 需要createInstanceQueue时，从读锁升级为写锁
	gi.Lock()
	// 需要二次判断，防止重复创建
	instanceQueue, exist = gi.reservedInstanceQueue[resKey]
	if !exist {
		log.GetLogger().Debugf("createInstanceQueue type reserved for function %s destroy start",
			gi.FuncSpec.FuncKey)
		insQ, err := gi.createInstanceQueue(types.InstanceTypeReserved, resKey)
		if err == nil {
			gi.reservedInstanceQueue[resKey] = insQ.(*instancequeue.ScaledInstanceQueue)
			instanceQueue = insQ.(*instancequeue.ScaledInstanceQueue)
			if gi.insConfig[resKey] != nil {
				instanceQueue.HandleInsConfigUpdate(gi.insConfig[resKey])
			}
		}
		snErr = err
	}
	gi.Unlock()
	return instanceQueue, snErr
}

func (gi *GenericInstancePool) acquireScaleInstanceQueue(resKey resspeckey.ResSpecKey) (
	*instancequeue.ScaledInstanceQueue, snerror.SNError) {
	var snErr snerror.SNError
	gi.RLock()
	instanceQueue, exist := gi.scaledInstanceQueue[resKey]
	gi.RUnlock()
	if exist {
		return instanceQueue, snErr
	}
	// 需要createInstanceQueue时，从读锁升级为写锁
	gi.Lock()
	// 需要二次判断，防止重复创建
	instanceQueue, exist = gi.scaledInstanceQueue[resKey]
	if !exist {
		log.GetLogger().Debugf("createInstanceQueue type scaled for function %s destroy start",
			gi.FuncSpec.FuncKey)
		insQ, err := gi.createInstanceQueue(types.InstanceTypeScaled, resKey)
		if err == nil {
			gi.scaledInstanceQueue[resKey] = insQ.(*instancequeue.ScaledInstanceQueue)
			instanceQueue = insQ.(*instancequeue.ScaledInstanceQueue)
			if gi.insConfig[resKey] != nil {
				instanceQueue.HandleInsConfigUpdate(gi.insConfig[resKey])
			}
		}
		snErr = err
	}
	gi.Unlock()
	return instanceQueue, snErr
}

func (gi *GenericInstancePool) removeInstance(instance *types.Instance, logger api.FormatLogger) {
	logger.Infof("start to removed instance")
	signalmanager.GetSignalManager().RemoveInstance(instance.InstanceID)
	switch instance.InstanceType {
	case types.InstanceTypeReserved:
		gi.RLock()
		reservedInstanceQueue, exist := gi.reservedInstanceQueue[instance.ResKey]
		gi.RUnlock()
		// label没有了也应该要清理instance
		if !exist {
			logger.Errorf("reserved queue of function %s resource %+v doesn't exist", gi.FuncSpec.FuncKey,
				instance.ResKey)
			go DeleteUnexpectInstance(instance.ParentID, instance.InstanceID, instance.FuncKey, logger)
			break
		}
		reservedInstanceQueue.HandleFaultyInstance(instance)
	case types.InstanceTypeScaled:
		gi.RLock()
		scaledInstanceQueue, exist := gi.scaledInstanceQueue[instance.ResKey]
		gi.RUnlock()
		// label没有了也应该要清理instance
		if !exist {
			logger.Errorf("scaled queue of function %s resource %+v doesn't exist", gi.FuncSpec.FuncKey,
				instance.ResKey)
			go DeleteUnexpectInstance(instance.ParentID, instance.InstanceID, instance.FuncKey, logger)
			break
		}
		scaledInstanceQueue.HandleFaultyInstance(instance)
		// todo 以后这里要考虑删除对应的租约
		gi.stateRoute.DeleteStateInstanceByInstanceID(instance.InstanceID)
	default:
		logger.Errorf("unsupported instance type")
	}
	logger.Infof("succeed to remove instance")
}

func (gi *GenericInstancePool) handleInstanceSynced() {
	gi.Lock()
	wg := sync.WaitGroup{}
	for _, queue := range gi.reservedInstanceQueue {
		wg.Add(1)
		go func(queue *instancequeue.ScaledInstanceQueue) {
			queue.HandleInstanceSync(gi.recoverFuncCall)
			wg.Done()
		}(queue)
	}
	for _, queue := range gi.scaledInstanceQueue {
		wg.Add(1)
		go func(queue *instancequeue.ScaledInstanceQueue) {
			queue.HandleInstanceSync(gi.recoverFuncCall)
			wg.Done()
		}(queue)
	}
	gi.Unlock()
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	select {
	case <-c:
		log.GetLogger().Infof("handle all instance synced success")
	case <-time.After(time.Minute):
		log.GetLogger().Warnf("handle all instance synced timeout")
	}
}

// HandleInstanceConfigEvent updates instance configuration
func (gi *GenericInstancePool) HandleInstanceConfigEvent(eventType registry.EventType,
	insConfig *instanceconfig.Configuration) {
	logger := log.GetLogger().With(zap.Any("", "HandleInstanceConfigEvent")).
		With(zap.Any("FuncKey", gi.FuncSpec.FuncKey)).
		With(zap.Any("Label", insConfig.InstanceLabel))
	logger.Infof("handle start")
	// currently insConfig isn't stored with resSpec, resKey of all insConfig will be set with default resource
	resKey := resspeckey.ConvertToResSpecKey(gi.defaultResSpec)
	resKey.InvokeLabel = insConfig.InstanceLabel
	switch eventType {
	case registry.SubEventTypeUpdate:
		gi.Lock()
		if _, ok := gi.insConfig[resKey]; !ok ||
			gi.insConfig[resKey].InstanceMetaData.MinInstance != insConfig.InstanceMetaData.MinInstance {
			gi.minScaleUpdatedTime = time.Now()
		}
		gi.insConfig[resKey] = generateInstanceConfig(insConfig)
		gi.Unlock()
		// there is a checkpoint of insConfig in create function for no label instance, insConfig of labeled instances
		// will be checked in the process of label
		if len(insConfig.InstanceLabel) == 0 {
			gi.closeChanOnce.Do(func() { close(gi.waitInsConfigChan) })
		}
		gi.handleInstanceConfigUpdate(insConfig, resKey, logger)
		gi.judgeExceedInstance(resKey, logger)
	case registry.SubEventTypeDelete:
		gi.Lock()
		delete(gi.insConfig, resKey)
		gi.Unlock()
		gi.handleInstanceConfigDelete(insConfig, resKey, logger)
		metrics.ClearMetricsForFunctionInsConfig(gi.FuncSpec, resKey.InvokeLabel)
	default:
		logger.Warnf("unsupported instance config event type: %s", eventType)
	}
}

func (gi *GenericInstancePool) handleInstanceConfigUpdate(insConfig *instanceconfig.Configuration,
	resKey resspeckey.ResSpecKey, logger api.FormatLogger) {
	reservedInstanceQueue, err := gi.acquireReservedInstanceQueue(resKey)
	if err == nil {
		reservedInstanceQueue.HandleInsConfigUpdate(insConfig)
		reservedInstanceQueue.EnableInstanceScale()
	} else {
		logger.Errorf("acquire reserved instance queue failed, err %s", err.Error())
	}
	gi.RLock()
	for scaleResKey, queue := range gi.scaledInstanceQueue {
		if scaleResKey.InvokeLabel != resKey.InvokeLabel {
			continue
		}
		queue.HandleInsConfigUpdate(insConfig)
		queue.EnableInstanceScale()
	}
	gi.RUnlock()
}

func (gi *GenericInstancePool) handleInstanceConfigDelete(insConfig *instanceconfig.Configuration,
	resKey resspeckey.ResSpecKey, logger api.FormatLogger) {
	logger.Debugf("handleInstanceConfigDelete")
	insConfig.InstanceMetaData.MinInstance = 0
	insConfig.InstanceMetaData.MaxInstance = 0
	gi.RLock()
	reservedInstanceQueue, exist := gi.reservedInstanceQueue[resKey]
	gi.RUnlock()
	if exist {
		// labeled instance queue need to be destroyed, instance queue with no label updates with min/max=0/0
		if resKey.InvokeLabel != DefaultInstanceLabel {
			gi.Lock()
			delete(gi.reservedInstanceQueue, resKey)
			gi.Unlock()
			reservedInstanceQueue.Destroy()
		} else {
			reservedInstanceQueue.HandleInsConfigUpdate(insConfig)
			reservedInstanceQueue.EnableInstanceScale()
		}
	}
	// labeled instance queue need to be destroyed,  instance queue with no label takes no action, minInstance won't
	// influence scaled instance, maxInstance will be handled in createInstanceFunc
	for res, queue := range gi.scaledInstanceQueue {
		if res.InvokeLabel != DefaultInstanceLabel && res.InvokeLabel == resKey.InvokeLabel {
			gi.Lock()
			delete(gi.scaledInstanceQueue, resKey)
			gi.Unlock()
			queue.Destroy()
		}
	}
}

// CleanOrphansInstanceQueue destroy instance queue without instance config
func (gi *GenericInstancePool) CleanOrphansInstanceQueue() {
	gi.Lock()
	defer gi.Unlock()
	for resKey, queue := range gi.reservedInstanceQueue {
		if _, ok := gi.insConfig[resKey]; !ok && resKey.InvokeLabel != DefaultInstanceLabel {
			delete(gi.reservedInstanceQueue, resKey)
			queue.Destroy()
		}
	}
	for key, queue := range gi.scaledInstanceQueue {
		insConfResKey := gi.defaultResKey
		insConfResKey.InvokeLabel = key.InvokeLabel
		if _, ok := gi.insConfig[insConfResKey]; !ok {
			delete(gi.scaledInstanceQueue, key)
			queue.Destroy()
		}
	}
	for key, queue := range gi.onDemandInstanceQueue {
		insConfResKey := gi.defaultResKey
		insConfResKey.InvokeLabel = key.InvokeLabel
		if _, ok := gi.insConfig[insConfResKey]; !ok {
			delete(gi.onDemandInstanceQueue, key)
			queue.Destroy()
		}
	}
	gi.synced = true
}

func (gi *GenericInstancePool) QuerySession(sessionID string) (string, error) {
	gi.RLock()
	defer gi.RUnlock()

	record, exist := gi.sessionRecordMap[sessionID]
	if exist && record.instance != nil {
		return record.instance.InstanceID, nil
	}

	return "", fmt.Errorf("session %s not found", sessionID)
}

func generateInstanceConfig(insConf *instanceconfig.Configuration) *instanceconfig.Configuration {
	if insConf.InstanceMetaData.MinInstance < 0 {
		insConf.InstanceMetaData.MinInstance = 0
	}
	if insConf.InstanceMetaData.MaxInstance < 0 {
		insConf.InstanceMetaData.MaxInstance = maxInstanceLimit
	}
	return insConf
}

// UpdateInvokeMetrics sends invoke metrics of instance thread to autoScaler
func (gi *GenericInstancePool) UpdateInvokeMetrics(resKey resspeckey.ResSpecKey,
	InsThdMetrics *types.InstanceThreadMetrics) {
	gi.RLock()
	metricsCollector, exist := gi.metricsCollector[resKey]
	gi.RUnlock()
	if !exist {
		log.GetLogger().Errorf("update invoke metrics failed for function %s, resource %s doesn't exist",
			gi.FuncSpec.FuncKey, resKey)
		return
	}
	metricsCollector.UpdateInvokeMetrics(InsThdMetrics)
}

// GetFuncSpec will return the funcSpec of this pool
func (gi *GenericInstancePool) GetFuncSpec() *types.FunctionSpecification {
	gi.RLock()
	funcSpec := gi.FuncSpec
	gi.RUnlock()
	return funcSpec
}

func (gi *GenericInstancePool) getCurrentInstanceNum(resKey resspeckey.ResSpecKey) (int, int, snerror.SNError) {
	// insConfig is stored with default resource, labeled instances have their individual limit, all types of instances
	// are subject to the limit with default resource and no label
	if _, exist := gi.insConfig[resKey]; !exist {
		log.GetLogger().Warnf("insConfig of function %s for resource %+v doesn't exist", gi.FuncSpec.FuncKey, resKey)
		return 0, 0, snerror.New(statuscode.FunctionIsDisabled, fmt.Sprintf("function is disabled"))
	}
	var scaledNum, reservedNum int
	var scaledNumGlobal, reservedNumGlobal, pendingNumGlobal int

	for res, queue := range gi.reservedInstanceQueue {
		reservedNumGlobal += queue.GetInstanceNumber(true)
		// reserved instance with label can't be set with dynamic resource
		if res == resKey {
			reservedNum += queue.GetInstanceNumber(true)
		}
	}
	for res, queue := range gi.scaledInstanceQueue {
		scaledNumGlobal += queue.GetInstanceNumber(true)
		// scaled instance with label can be set with dynamic resource
		if res.InvokeLabel == resKey.InvokeLabel {
			scaledNum += queue.GetInstanceNumber(true)
		}
	}
	for _, v := range gi.pendingInstanceNum {
		pendingNumGlobal += v
	}
	sumForCurrentLabel := scaledNum + reservedNum + gi.pendingInstanceNum[resKey.InvokeLabel]
	sumForGlobal := scaledNumGlobal + reservedNumGlobal + pendingNumGlobal
	return sumForCurrentLabel, sumForGlobal, nil
}

// 判断实例是否可以继续扩容，需要判断同一个label的实例数之和是否超出label级的
// 最大实例限制，再判断同一个函数版本的实例数之和是否超出函数级的最大实例限制
func (gi *GenericInstancePool) checkScaleLimit(resKey resspeckey.ResSpecKey) snerror.SNError {
	insConfResKey := gi.defaultResKey
	insConfResKey.InvokeLabel = resKey.InvokeLabel
	sumForCurrentLabel, sumForGlobal, err := gi.getCurrentInstanceNum(insConfResKey)
	if err != nil {
		return err
	}
	limitForLabel := int(gi.insConfig[insConfResKey].InstanceMetaData.MaxInstance)
	limitForGlobal := int(gi.insConfig[gi.defaultResKey].InstanceMetaData.MaxInstance)

	reachLimitForLabel := sumForCurrentLabel >= limitForLabel
	reachLimitForGlobal := sumForGlobal >= limitForGlobal
	if reachLimitForLabel {
		log.GetLogger().Errorf("function %s reaches scale limit %d of label %s, pending num %d, current num %d, "+
			"stop creating", gi.FuncSpec.FuncKey, limitForLabel, resKey.InvokeLabel,
			gi.pendingInstanceNum[resKey.InvokeLabel], sumForCurrentLabel)
		return snerror.New(statuscode.ReachMaxInstancesCode, fmt.Sprintf("%s %d", statuscode.ReachMaxInstancesErrMsg,
			limitForLabel))
	}
	if reachLimitForGlobal {
		log.GetLogger().Errorf("function %s reaches general scale limit %d, current num %d, stop creating",
			gi.FuncSpec.FuncKey, sumForGlobal, limitForGlobal)
		return snerror.New(statuscode.ReachMaxInstancesCode, fmt.Sprintf("%s %d", statuscode.ReachMaxInstancesErrMsg,
			limitForGlobal))
	}
	return nil
}

func (gi *GenericInstancePool) checkTenantLimit(instanceType types.InstanceType) (bool, bool) {
	if !config.GlobalConfig.TenantInsNumLimitEnable {
		return false, false
	}
	tenantID := urnutils.GetTenantFromFuncKey(gi.FuncSpec.FuncKey)
	return tenantquota.IncreaseTenantInstanceNum(tenantID, instanceType == types.InstanceTypeReserved)
}

func (gi *GenericInstancePool) createInstanceAndAddCallerPodName(resSpec *resspeckey.ResourceSpecification,
	instanceType types.InstanceType, callerPodName string) (*types.Instance, error) {
	return gi.createInstanceFunc("", instanceType, gi.defaultResKey, nil, createOption{callerPodName: callerPodName})
}

func (gi *GenericInstancePool) createInstance(traceID string, insName string, instanceType types.InstanceType,
	resKey resspeckey.ResSpecKey, createEvent []byte) (*types.Instance, error) {
	return gi.createInstanceFunc(insName, instanceType, resKey, createEvent, createOption{traceID: traceID})
}

func (gi *GenericInstancePool) createInstanceFunc(insName string, instanceType types.InstanceType,
	resKey resspeckey.ResSpecKey, createEvent []byte, createOption createOption) (instance *types.Instance,
	createErr error) {
	logger := log.GetLogger().With(zap.Any("funcKey", gi.FuncSpec.FuncKey))
	// createInstance use insConfig, so gi.insConfig must exist
	<-gi.waitInsConfigChan
	gi.Lock()
	if config.GlobalConfig.InstanceOperationBackend != constant.BackendTypeFG {
		reachMaxOnDemandInsNum, reachMaxReversedInsNum := gi.checkTenantLimit(instanceType)
		logger.Debugf("checkTenantLimit completed, reachMaxOnDemandInsNum is %v, reachMaxReversedInsNum is %v",
			reachMaxOnDemandInsNum, reachMaxReversedInsNum)
		maxOnDemandInsNum, maxReversedInsNum := tenantquota.GetTenantCache().GetTenantQuotaNum(
			urnutils.GetTenantFromFuncKey(gi.FuncSpec.FuncKey))
		if reachMaxOnDemandInsNum {
			gi.Unlock()
			return nil, snerror.New(statuscode.ReachMaxOnDemandInstancesPerTenant, fmt.Sprintf("%s, limit %d",
				statuscode.ReachMaxInstancesPerTenantErrMsg, maxOnDemandInsNum))
		}
		if reachMaxReversedInsNum {
			logger.Warnf("reach max reversed instance num per tenant: %s, limit %d", urnutils.Anonymize(
				urnutils.GetTenantFromFuncKey(gi.FuncSpec.FuncKey)), maxReversedInsNum)
		}
		if err := gi.checkScaleLimit(resKey); err != nil {
			gi.Unlock()
			return nil, err
		}
	}
	gi.pendingInstanceNum[resKey.InvokeLabel]++
	gi.Unlock()
	defer func() {
		gi.Lock()
		if _, ok := gi.pendingInstanceNum[resKey.InvokeLabel]; ok {
			gi.pendingInstanceNum[resKey.InvokeLabel]--
			if gi.pendingInstanceNum[resKey.InvokeLabel] == 0 {
				delete(gi.pendingInstanceNum, resKey.InvokeLabel)
			}
		}
		gi.Unlock()
	}()

	gi.RLock()

	createRequest := createInstanceRequest{
		traceID:         createOption.traceID,
		funcSpec:        gi.FuncSpec,
		poolLabel:       gi.currentPoolLabel,
		createTimeout:   gi.createTimeout,
		faasManagerInfo: gi.faasManagerInfo,
		resKey:          resKey,
		instanceName:    insName,
		callerPodName:   createOption.callerPodName,
		createEvent:     createEvent,
		instanceType:    instanceType,
	}
	insConfResKey := gi.defaultResKey
	insConfResKey.InvokeLabel = resKey.InvokeLabel
	if insConf, exist := gi.insConfig[insConfResKey]; exist {
		createRequest.nuwaRuntimeInfo = &insConf.NuwaRuntimeInfo
	}
	gi.RUnlock()
	return CreateInstance(createRequest)
}

func (gi *GenericInstancePool) deleteInstance(instance *types.Instance) error {
	gi.RLock()
	currentFaasManagerInfo := gi.faasManagerInfo
	gi.RUnlock()
	return DeleteInstance(gi.FuncSpec, currentFaasManagerInfo, instance)
}

func (gi *GenericInstancePool) handleSignal(instance *types.Instance, signal int) {
	SignalInstance(instance, signal)
}

func (gi *GenericInstancePool) handleManagedChange() {
	gi.Lock()
	log.GetLogger().Debugf("HandleFuncOwnerChange for function %s start",
		gi.FuncSpec.FuncKey)
	for _, q := range gi.scaledInstanceQueue {
		q.HandleFuncOwnerChange(gi.recoverFuncCall)
	}
	for k, q := range gi.reservedInstanceQueue {
		q.HandleFuncOwnerChange(gi.recoverFuncCall)
		if _, ok := gi.insConfig[k]; ok {
			q.HandleInsConfigUpdate(gi.insConfig[k])
		}
	}
	gi.Unlock()
}

func (gi *GenericInstancePool) handleRatioChange(ratio int) {
	gi.Lock()
	for _, q := range gi.scaledInstanceQueue {
		q.HandleRatioUpdate(ratio)
	}
	for _, q := range gi.reservedInstanceQueue {
		q.HandleRatioUpdate(ratio)
	}
	gi.Unlock()
}

func (gi *GenericInstancePool) recordInstanceSession(insAlloc *types.InstanceAllocation) {
	gi.Lock()
	sessions, exist := gi.instanceSessionMap[insAlloc.Instance.InstanceID]
	if !exist {
		sessions = make(map[string]struct{}, utils.DefaultMapSize)
		gi.instanceSessionMap[insAlloc.Instance.InstanceID] = sessions
	}
	sessions[insAlloc.SessionInfo.SessionID] = struct{}{}
	gi.sessionRecordMap[insAlloc.SessionInfo.SessionID] = sessionRecord{
		instance:   insAlloc.Instance,
		sessionCtx: insAlloc.SessionInfo.SessionCtx,
	}
	gi.Unlock()
}

func (gi *GenericInstancePool) cleanInstanceSession(instanceID string) {
	gi.RLock()
	_, exist := gi.instanceSessionMap[instanceID]
	if !exist {
		gi.RUnlock()
		return
	}
	gi.RUnlock()
	gi.Lock()
	sessions, exist := gi.instanceSessionMap[instanceID]
	if exist {
		delete(gi.instanceSessionMap, instanceID)
		for sessionID, _ := range sessions {
			delete(gi.sessionRecordMap, sessionID)
		}
	}
	gi.Unlock()
}

func (gi *GenericInstancePool) processInstanceSession(insAcqReq *types.InstanceAcquireRequest) {
	gi.RLock()
	record, exist := gi.sessionRecordMap[insAcqReq.InstanceSession.SessionID]
	gi.RUnlock()
	if exist {
		select {
		case <-record.sessionCtx.Done():
			gi.Lock()
			delete(gi.sessionRecordMap, insAcqReq.InstanceSession.SessionID)
			sessions, exist := gi.instanceSessionMap[record.instance.InstanceID]
			if exist {
				delete(sessions, insAcqReq.InstanceSession.SessionID)
			}
			gi.Unlock()
		default:
			insAcqReq.DesignateInstanceID = record.instance.InstanceID
		}
	}
}

func (gi *GenericInstancePool) instanceSessionReaper() {
	// for LLT convenience
	if gi.FuncSpec == nil || gi.FuncSpec.FuncCtx == nil {
		return
	}
	ticker := time.NewTicker(gi.sessionReaperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case _, ok := <-gi.FuncSpec.FuncCtx.Done():
			if !ok {
				log.GetLogger().Warnf("instance pool stopped, stop instance session reaper now")
				return
			}
		}
		gi.Lock()
		for sessionID, record := range gi.sessionRecordMap {
			select {
			case <-record.sessionCtx.Done():
				delete(gi.sessionRecordMap, sessionID)
				sessions, exist := gi.instanceSessionMap[record.instance.InstanceID]
				if exist {
					delete(sessions, sessionID)
				}
			default:
			}
		}
		gi.Unlock()
	}
}

func (gi *GenericInstancePool) judgeExceedInstance(resKey resspeckey.ResSpecKey, logger api.FormatLogger) {
	if config.GlobalConfig.Scenario == types.ScenarioWiseCloud {
		return
	}
	insConfResKey := gi.defaultResKey
	insConfResKey.InvokeLabel = resKey.InvokeLabel
	gi.RLock()
	defer gi.RUnlock()
	_, ok1 := gi.insConfig[insConfResKey]
	_, ok2 := gi.insConfig[gi.defaultResKey]
	if !ok1 || !ok2 {
		logger.Errorf("instance config not exist, skip")
		return
	}
	sumForCurrentLabel, sumForGlobal, err := gi.getCurrentInstanceNum(insConfResKey)
	if err != nil {
		return
	}
	limitForLabel := int(gi.insConfig[insConfResKey].InstanceMetaData.MaxInstance)
	limitForGlobal := int(gi.insConfig[gi.defaultResKey].InstanceMetaData.MaxInstance)
	exceedLimitForLabel := sumForCurrentLabel > limitForLabel
	exceedLimitForGlobal := sumForGlobal > limitForGlobal
	if !exceedLimitForGlobal && !exceedLimitForLabel {
		return
	}
	instanceQueue, exist := gi.scaledInstanceQueue[resKey]
	if !exist {
		logger.Warnf("scaled instance queue not exist, delete default scale queue")
		return
	}
	scaleDiff := utils.IntMax(sumForCurrentLabel-limitForLabel, sumForGlobal-limitForGlobal)
	logger.Infof("start scale down exceed instance, %d", scaleDiff)
	instanceQueue.ScaleDownHandler(scaleDiff, func(i int) {
		logger.Infof("scale down exceed instance %d succeed", i)
	})
}

func (gi *GenericInstancePool) recoverFuncCall(sessInfo *types.SessionInfo, instance *types.Instance) {
	sessions, exist := gi.instanceSessionMap[instance.InstanceID]
	if !exist {
		sessions = make(map[string]struct{}, utils.DefaultMapSize)
		gi.instanceSessionMap[instance.InstanceID] = sessions
	}
	sessions[sessInfo.SessionID] = struct{}{}
	gi.sessionRecordMap[sessInfo.SessionID] = sessionRecord{
		instance:   instance,
		sessionCtx: sessInfo.SessionCtx,
	}
}
