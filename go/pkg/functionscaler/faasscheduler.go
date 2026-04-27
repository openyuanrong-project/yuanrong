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

// Package functionscaler -
package functionscaler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/healthlog"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/common/faas_common/trafficlimit"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/crprocessor"
	"yuanrong.org/kernel/pkg/functionscaler/instancepool"
	"yuanrong.org/kernel/pkg/functionscaler/lease"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/rollout"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	defaultChanSize        = 1000
	minArgsNum             = 1
	validArgsNum           = 2
	libruntimeValidArgsNum = 4
	validInsOpLen          = 2
	waitForETCDList        = 10 * time.Millisecond
	frontendNodePort       = "31222"
	logFileName            = "faas-scheduler"
	stateFuncKeyLen        = 2
)

var (
	// insOpSeparator stands for separator of instance operation
	insOpSeparator = "#"
	// insOpCreate stands for instance create operation
	insOpCreate InstanceOperation = "create"
	// insOpDelete stands for instance delete operation
	insOpDelete InstanceOperation = "delete"
	// insOpAcquire stands for instance acquire operation
	insOpAcquire InstanceOperation = "acquire"
	// insOpRetain stands for instance retain operation
	insOpRetain InstanceOperation = "retain"
	// insOpBatchRetain stands for instance batch retain operation
	insOpBatchRetain InstanceOperation = "batchRetain"
	// insOpRelease stands for instance release operation
	insOpRelease InstanceOperation = "release"
	// insOpRelease stands for instance release operation
	insOpRollout InstanceOperation = "rollout"
	// insOpQuerySession stands for query session operation
	insOpQuerySession InstanceOperation = "querySession"
	// insOpUnknown stands for unknown instance operation
	insOpUnknown InstanceOperation = "unknown"
	// stateSplitStr -
	stateSplitStr = ";"

	// InstanceRequirementPoolLabel - key of poolLabel
	instanceRequirementPoolLabel = "poolLabel"
)

// InstanceOperation defines instance operations
type InstanceOperation string

// StateOperation defines state instance operations
type StateOperation string

// FaaSScheduler manages instances for faas functions
type FaaSScheduler struct {
	PoolManager     *instancepool.PoolManager
	funcSpecCh      chan registry.SubEvent
	insSpecCh       chan registry.SubEvent
	insConfigCh     chan registry.SubEvent
	aliasSpecCh     chan registry.SubEvent
	schedulerCh     chan registry.SubEvent
	rolloutConfigCh chan registry.SubEvent

	leaseInterval time.Duration

	allocRecord sync.Map
	sync.RWMutex
}

var globalFaaSScheduler *FaaSScheduler

// NewFaaSScheduler will create a FaaSScheduler
func NewFaaSScheduler(stopCh <-chan struct{}) *FaaSScheduler {
	leaseInterval := time.Duration(config.GlobalConfig.LeaseSpan) * time.Millisecond
	if leaseInterval < types.MinLeaseInterval {
		leaseInterval = types.MinLeaseInterval
	}
	go func() {
		if config.GlobalConfig.PprofAddr == "" {
			return
		}
		err := http.ListenAndServe(config.GlobalConfig.PprofAddr, nil)
		if err != nil {
			return
		}
	}()
	faasScheduler := &FaaSScheduler{
		PoolManager:     instancepool.NewPoolManager(stopCh),
		funcSpecCh:      make(chan registry.SubEvent, defaultChanSize),
		insSpecCh:       make(chan registry.SubEvent, defaultChanSize),
		insConfigCh:     make(chan registry.SubEvent, defaultChanSize),
		aliasSpecCh:     make(chan registry.SubEvent, defaultChanSize),
		schedulerCh:     make(chan registry.SubEvent, defaultChanSize),
		rolloutConfigCh: make(chan registry.SubEvent, defaultChanSize),
		leaseInterval:   leaseInterval,
	}
	setupAgentCRsManager(stopCh, faasScheduler)
	registry.GlobalRegistry.SubscribeFuncSpec(faasScheduler.funcSpecCh)
	registry.GlobalRegistry.SubscribeInsSpec(faasScheduler.insSpecCh)
	registry.GlobalRegistry.SubscribeInsConfig(faasScheduler.insConfigCh)
	registry.GlobalRegistry.SubscribeAliasSpec(faasScheduler.aliasSpecCh)
	registry.GlobalRegistry.SubscribeSchedulerProxy(faasScheduler.schedulerCh)
	registry.GlobalRegistry.SubscribeRolloutConfig(faasScheduler.rolloutConfigCh)
	go faasScheduler.processFunctionSubscription()
	go faasScheduler.processInstanceSubscription()
	go faasScheduler.processInstanceConfigSubscription()
	go faasScheduler.processAliasSpecSubscription()
	go faasScheduler.processSchedulerProxySubscription()
	go faasScheduler.processRolloutConfigSubscription()
	go healthlog.PrintHealthLog(stopCh, printInputLog, logFileName)
	if config.GlobalConfig.AlarmConfig.EnableAlarm {
		faasScheduler.PoolManager.CheckMinInsAndReport(stopCh)
	}
	if selfregister.IsRolloutObject {
		go faasScheduler.syncAllocRecordDuringRollout()
	}
	go metrics.InitServerMetric(stopCh)

	return faasScheduler
}

func setupAgentCRsManager(stopCh <-chan struct{}, faasScheduler *FaaSScheduler) {
	if os.Getenv(constant.EnableAgentCRDRegistry) != "" {
		agentRunResourceManager := crprocessor.NewAgentCRsManager(stopCh)
		registry.GlobalRegistry.SubscribeInsSpec(agentRunResourceManager.AgentCRInsCh)
		registry.GlobalRegistry.SubscribeSchedulerProxy(agentRunResourceManager.FaaSSchedulerProxyCh)
		registry.GlobalRegistry.SubscribeAgentRunInfo(agentRunResourceManager.AgentCRCh)
		agentRunResourceManager.AddFunctionSubscriberChan(faasScheduler.funcSpecCh)
		agentRunResourceManager.AddInstanceConfigSubscriberChan(faasScheduler.insConfigCh)
		agentRunResourceManager.StartLoop()
	}
}

// InitGlobalScheduler -
func InitGlobalScheduler(stopCh <-chan struct{}) {
	globalFaaSScheduler = NewFaaSScheduler(stopCh)
}

// GetGlobalScheduler -
func GetGlobalScheduler() *FaaSScheduler {
	return globalFaaSScheduler
}

// Recover before recover faaSScheduler, must wait StartList complete
func (fs *FaaSScheduler) Recover() {
	// wait for StartList completion
	for len(fs.funcSpecCh) != 0 {
		time.Sleep(waitForETCDList)
	}
	time.Sleep(waitForETCDList)
	fs.PoolManager.RecoverInstancePool()
}

func (fs *FaaSScheduler) processFunctionSubscription() {
	for {
		select {
		case event, ok := <-fs.funcSpecCh:
			if !ok {
				log.GetLogger().Warnf("function channel is closed")
				return
			}
			funcSpec, ok := event.EventMsg.(*types.FunctionSpecification)
			if !ok {
				log.GetLogger().Warnf("event message doesn't contain function specification")
				continue
			}
			fs.PoolManager.HandleFunctionEvent(event.EventType, funcSpec)
		}
	}
}

func (fs *FaaSScheduler) processInstanceSubscription() {
	for {
		select {
		case event, ok := <-fs.insSpecCh:
			if !ok {
				log.GetLogger().Warnf("instance channel is closed")
				return
			}
			insSpec, ok := event.EventMsg.(*commonTypes.InstanceSpecification)
			if !ok {
				log.GetLogger().Warnf("event message doesn't contain instance specification")
				continue
			}
			fs.PoolManager.HandleInstanceEvent(event.EventType, insSpec)
		}
	}
}

func (fs *FaaSScheduler) processInstanceConfigSubscription() {
	for {
		select {
		case event, ok := <-fs.insConfigCh:
			if !ok {
				log.GetLogger().Warnf("instances info channel is closed")
				return
			}
			insConfig, ok := event.EventMsg.(*instanceconfig.Configuration)
			if !ok {
				log.GetLogger().Warnf("event message doesn't contain instance specification")
				continue
			}
			fs.PoolManager.HandleInstanceConfigEvent(event.EventType, insConfig)
		}
	}
}

func (fs *FaaSScheduler) processAliasSpecSubscription() {
	for {
		select {
		case event, ok := <-fs.aliasSpecCh:
			if !ok {
				log.GetLogger().Warnf("instances info channel is closed")
				return
			}
			aliasUrn, ok := event.EventMsg.(string)
			if !ok {
				log.GetLogger().Warnf("event message doesn't contain instance specification")
				continue
			}
			fs.PoolManager.HandleAliasEvent(event.EventType, aliasUrn)
		}
	}
}

func (fs *FaaSScheduler) processSchedulerProxySubscription() {
	for {
		select {
		case event, ok := <-fs.schedulerCh:
			if !ok {
				log.GetLogger().Warnf("scheduler proxy channel is closed")
				return
			}
			if instanceSpec, assertOK := event.EventMsg.(*commonTypes.InstanceSpecification); assertOK {
				fs.PoolManager.HandleSchedulerManaged(event.EventType, instanceSpec)
			} else {
				log.GetLogger().Warnf("event message doesn't contain scheduler info")
				continue
			}
		}
	}
}

func (fs *FaaSScheduler) processRolloutConfigSubscription() {
	for {
		select {
		case event, ok := <-fs.rolloutConfigCh:
			if !ok {
				log.GetLogger().Warnf("scheduler proxy channel is closed")
				return
			}
			if ratio, ok := event.EventMsg.(int); ok {
				fs.PoolManager.HandleRolloutRatioChange(ratio)
			} else {
				log.GetLogger().Warnf("event message doesn't contain ratio info")
				continue
			}
		}
	}
}

// ProcessInstanceRequestLibruntime will handle acquire, release and retain of instance based on multi libruntime
func (fs *FaaSScheduler) ProcessInstanceRequestLibruntime(args []api.Arg, traceID string) ([]byte, error) {
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	insOp, targetName, extraData, eventData := parseInstanceOperation(args, traceID)
	startTime := time.Now()

	defer func() {
		logger.Infof("process of instance operation %s target %s cost %dms", insOp, targetName,
			time.Now().Sub(startTime).Milliseconds())
	}()
	result, err, shouldReply := fs.HandleRequestForward(insOp, args, traceID)
	if shouldReply {
		return result, err
	}
	var response interface{}
	switch insOp {
	case insOpCreate:
		response = fs.handleInstanceCreate(targetName, extraData, eventData, traceID)
	case insOpDelete:
		response = fs.handleInstanceDelete(targetName, extraData, traceID)
	case insOpAcquire:
		response = fs.handleInstanceAcquire(targetName, extraData, traceID)
	case insOpRelease:
		response = fs.handleInstanceRelease(targetName, extraData, traceID)
	case insOpRetain:
		response = fs.handleInstanceRetain(targetName, extraData, traceID)
	case insOpBatchRetain:
		response = fs.handleInstanceBatchRetain(targetName, extraData, traceID)
	case insOpRollout:
		response = fs.handleRollout(targetName, traceID)
	case insOpQuerySession:
		response = fs.handleQuerySession(targetName, extraData, traceID)
	default:
		logger.Warnf("unknown instance operation %s", insOp)
		response = generateInstanceResponse(nil, snerror.New(constant.UnsupportedOperationErrorCode,
			constant.UnsupportedOperationErrorMessage), startTime)
	}
	respData, err := json.Marshal(response)
	if err != nil {
		logger.Errorf("failed to marshal response of instance operation %s error %s", insOp, err.Error())
		return nil, err
	}
	return respData, nil
}

// HandleRequestForward return forward result and  shouldReply flag
func (fs *FaaSScheduler) HandleRequestForward(insOp InstanceOperation, args []api.Arg, traceID string) ([]byte, error,
	bool,
) {
	if !rollout.GetGlobalRolloutHandler().IsGaryUpdating {
		return []byte{}, nil, false
	}
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	switch insOp {
	case insOpCreate, insOpAcquire:
		if rollout.GetGlobalRolloutHandler().ShouldForwardRequest() {
			logger.Infof("gray updating forward %s request to instance %s", string(insOp),
				rollout.GetGlobalRolloutHandler().ForwardInstance)
			result, err := rollout.InvokeByInstanceId(args, rollout.GetGlobalRolloutHandler().ForwardInstance,
				traceID)
			if err != nil {
				// 调用另一个scheduler失败需要兜底
				return result, err, false
			}
			response := &commonTypes.InstanceResponse{}
			err = json.Unmarshal(result, response)
			if err != nil {
				return []byte{}, err, false
			}
			if response.ErrorCode == statuscode.NoInstanceAvailableErrCode ||
				response.ErrorCode == statuscode.InsThdReqTimeoutCode {
				logger.Infof("gray updating get no instance available error %s from instance %s",
					response.ErrorCode, rollout.GetGlobalRolloutHandler().ForwardInstance)
				return []byte{}, nil, false
			}
			return result, err, true
		}
	case insOpRelease, insOpRetain, insOpBatchRetain, insOpDelete:
		logger.Infof("gray updating forward %s request to instance %s", string(insOp),
			rollout.GetGlobalRolloutHandler().ForwardInstance)
		_, _ = rollout.InvokeByInstanceId(args, rollout.GetGlobalRolloutHandler().ForwardInstance, traceID)
		return []byte{}, nil, false
	default:
		logger.Warnf("unknown instance operation %s", insOp)
	}
	return []byte{}, nil, false
}

func (fs *FaaSScheduler) handleInstanceCreate(funcKey string, extraData, eventData []byte,
	traceID string,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	logger := log.GetLogger().With(zap.Any("traceID", traceID), zap.Any("funcKey", funcKey))
	funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
	if funcSpec == nil {
		logger.Errorf("failed to create instance, function %s doesn't exist", funcKey)
		return generateInstanceResponse(nil, snerror.New(statuscode.FuncMetaNotFoundErrCode,
			statuscode.FuncMetaNotFoundErrMsg), startTime)
	}
	dataInfo, err := parseExtraData(extraData)
	if err != nil {
		logger.Errorf("failed to parse extraData error :%v", err)
		return generateInstanceResponse(nil, err, startTime)
	}
	resSpec, err := getResourceSpecification(dataInfo.resourceData, dataInfo.invokeLabel, funcSpec)
	if err != nil {
		logger.Errorf("failed get resSpec error %v", err)
		return generateInstanceResponse(nil, err, startTime)
	}
	instance, err := fs.PoolManager.CreateInstance(&types.InstanceCreateRequest{
		TraceID:      traceID,
		FuncSpec:     funcSpec,
		ResSpec:      resSpec,
		InstanceName: dataInfo.designateInstanceName,
		CreateEvent:  eventData,
	})
	if err != nil {
		logger.Errorf("failed to create instance for function %s, error %s", funcSpec.FuncKey, err.Error())
		return generateInstanceResponse(nil, err, startTime)
	}
	return generateInstanceResponse(&types.InstanceAllocation{Instance: instance}, nil, startTime)
}

func (fs *FaaSScheduler) handleInstanceDelete(instanceID string, extraData []byte,
	traceID string,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	instance := registry.GlobalRegistry.GetInstance(instanceID)
	if instance == nil {
		return generateInstanceResponse(nil, snerror.New(statuscode.InstanceNotFoundErrCode,
			statuscode.InstanceNotFoundErrMsg), startTime)
	}
	logger := log.GetLogger().With(zap.Any("traceID", traceID), zap.Any("funcKey", instance.FuncKey))
	err := fs.PoolManager.DeleteInstance(instance)
	if err != nil {
		logger.Errorf("failed to delete instance for function %s, error %s", instance.FuncKey, err.Error())
		return generateInstanceResponse(nil, err, startTime)
	}
	return generateInstanceResponse(&types.InstanceAllocation{Instance: instance}, nil, startTime)
}

func (fs *FaaSScheduler) handleInstanceAcquire(targetName string, extraData []byte,
	traceID string,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	funcKey, stateID := parseStateOperation(targetName)
	logger := log.GetLogger().With(zap.Any("traceID", traceID), zap.Any("funcKey", funcKey),
		zap.Any("stateID", stateID))
	ownerSchedulerInstanceId, ok := selfregister.GlobalSchedulerProxy.CheckFuncOwner(funcKey)
	if !ok {
		logger.Errorf("non-owner faasscheduler, return owner faasscheduelr: %s", ownerSchedulerInstanceId)
		return generateInstanceResponse(nil, snerror.New(statuscode.AcquireNonOwnerSchedulerErrorCode,
			ownerSchedulerInstanceId), startTime)
	}
	funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
	if funcSpec == nil {
		logger.Errorf("failed to get instance, function %s doesn't exist", funcKey)
		return generateInstanceResponse(nil, snerror.New(statuscode.FuncMetaNotFoundErrCode,
			statuscode.FuncMetaNotFoundErrMsg), startTime)
	}

	needForward, endpoint, forwardErr := judgeForwardToOtherCluster(funcSpec.FuncMetaData.FunctionURN, logger)
	if forwardErr != nil {
		return generateInstanceResponse(nil, forwardErr, startTime)
	}
	if needForward {
		logger.Infof("request should forward to %s for %s", endpoint, funcSpec.FuncMetaData.FunctionURN)
		return generateInstanceResponse(nil, snerror.New(constant.AcquireLeaseVPCConflictErrorCode, endpoint),
			startTime)
	}

	if !trafficlimit.FuncTrafficLimit(funcKey) {
		logger.Warnf("handle instance acquire limited for function: %s", funcKey)
		return generateInstanceResponse(nil, snerror.New(constant.AcquireLeaseTrafficLimitErrorCode,
			constant.AcquireLeaseTrafficLimitErrorMessage), startTime)
	}
	var insAlloc *types.InstanceAllocation
	dataInfo, err := parseExtraData(extraData)
	if err != nil {
		logger.Errorf("failed to parse extraData error :%v", err)
		return generateInstanceResponse(nil, err, startTime)
	}
	resSpec, err := getResourceSpecification(dataInfo.resourceData, dataInfo.invokeLabel, funcSpec)
	if err != nil {
		logger.Errorf("failed get resSpec error %v", err)
		return generateInstanceResponse(nil, err, startTime)
	}
	logger.Infof("handling instance acquire for resSpec %v instanceID %s instanceSession %v traceID %s", resSpec,
		dataInfo.designateInstanceID, dataInfo.instanceSession, traceID)
	poolLabel := getPoolLabel(dataInfo.poolLabel, funcSpec.InstanceMetaData.PoolLabel)
	insAlloc, err = fs.PoolManager.AcquireInstanceThread(&types.InstanceAcquireRequest{
		FuncSpec:            funcSpec, // etcd
		ResSpec:             resSpec,  // args
		TraceID:             traceID,
		StateID:             stateID,
		PoolLabel:           poolLabel,
		InstanceName:        dataInfo.designateInstanceName,
		DesignateInstanceID: dataInfo.designateInstanceID,
		CallerPodName:       dataInfo.callerPodName,
		TrafficLimited:      dataInfo.trafficLimited,
		InstanceSession:     dataInfo.instanceSession,
	})
	if err != nil {
		logger.Errorf("failed to acquire instance of function %s traceID %s error %s", funcSpec.FuncKey, traceID,
			err.Error())
		return generateInstanceResponse(nil, err, startTime)
	}
	if insAlloc.Lease != nil {
		fs.allocRecord.Store(insAlloc.AllocationID, insAlloc)
	}
	logger.Infof("succeed to acquire instance %s of function %s traceID %s", insAlloc.AllocationID, funcSpec.FuncKey,
		traceID)
	return generateInstanceResponse(insAlloc, nil, startTime)
}

func (fs *FaaSScheduler) handleQuerySession(targetName string, extraData []byte,
	traceID string) *commonTypes.InstanceResponse {
	startTime := time.Now()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))

	dataInfo, err := parseExtraData(extraData)
	if err != nil {
		logger.Errorf("failed to parse extraData error :%v", err)
		return generateInstanceResponse(nil, err, startTime)
	}

	if len(dataInfo.instanceSession.SessionID) == 0 {
		logger.Errorf("sessionID is empty in query request")
		return generateInstanceResponse(nil, snerror.New(statuscode.InstanceSessionInvalidErrCode,
			"sessionID is empty"), startTime)
	}

	funcKey := targetName
	funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
	if funcSpec == nil {
		logger.Errorf("failed to get instance, function %s doesn't exist", funcKey)
		return generateInstanceResponse(nil, snerror.New(statuscode.FuncMetaNotFoundErrCode,
			statuscode.FuncMetaNotFoundErrMsg), startTime)
	}

	if !funcSpec.ExtendedMetaData.EnableAgentSession {
		logger.Errorf("AI Agent session is not enabled for function %s", funcKey)
		return generateInstanceResponse(nil, snerror.New(statuscode.AgentSessionNotEnabledErrCode,
			"AI Agent session not enabled"), startTime)
	}

	instanceID, queryErr := fs.PoolManager.QuerySession(funcKey, dataInfo.instanceSession.SessionID)
	if queryErr != nil {
		logger.Errorf("failed to query session %s for function %s: %v",
			dataInfo.instanceSession.SessionID, funcKey, queryErr)
		return generateInstanceResponse(nil, snerror.New(statuscode.SessionNotFoundErrCode,
			queryErr.Error()), startTime)
	}

	return &commonTypes.InstanceResponse{
		InstanceAllocationInfo: commonTypes.InstanceAllocationInfo{
			FuncKey:    funcSpec.FuncKey,
			FuncSig:    funcSpec.FuncMetaSignature,
			InstanceID: instanceID,
		},
		ErrorCode:     constant.InsReqSuccessCode,
		ErrorMessage:  constant.InsReqSuccessMessage,
		SchedulerTime: time.Now().Sub(startTime).Seconds(),
	}
}

func unmarshalExtraData(extraData []byte) (map[string][]byte, error) {
	extraDataMap := make(map[string][]byte, utils.DefaultMapSize)
	if len(extraData) != 0 {
		log.GetLogger().Debugf("acquire libruntime extraData: %s", string(extraData))
		defer func() {
			if r := recover(); r != nil {
				log.GetLogger().Errorf("acquire libruntime unmarshal extraData err: %v", r)
			}
		}()
		jsonErr := json.Unmarshal(extraData, &extraDataMap)
		if jsonErr != nil {
			return nil, jsonErr
		}
	}
	return extraDataMap, nil
}

func parseExtraData(extraData []byte) (*extraDataInfo, snerror.SNError) {
	extraDataMap, err := unmarshalExtraData(extraData)
	if err != nil {
		return nil, snerror.NewWithError(statuscode.StatusInternalServerError,
			fmt.Errorf("unmarshal extraData err: %w", err))
	}
	dataInfo := &extraDataInfo{}
	if instanceName, ok := extraDataMap[constant.RuntimeInstanceName]; ok {
		dataInfo.designateInstanceName = string(instanceName)
	}
	if instanceID, ok := extraDataMap[constant.InstanceRequirementInsIDKey]; ok {
		dataInfo.designateInstanceID = string(instanceID)
	}
	if createEvent, ok := extraDataMap[constant.InstanceCreateEvent]; ok {
		dataInfo.createEvent = createEvent
	}
	if resourceDataByte, ok := extraDataMap[constant.InstanceRequirementResourcesKey]; ok {
		dataInfo.resourceData = resourceDataByte
	}
	if callerPodNameByte, ok := extraDataMap[constant.InstanceCallerPodName]; ok {
		dataInfo.callerPodName = string(callerPodNameByte)
	}
	if poolLabelBytes, ok := extraDataMap[instanceRequirementPoolLabel]; ok {
		dataInfo.poolLabel = string(poolLabelBytes)
	}
	if trafficLimitedByte, ok := extraDataMap[constant.InstanceTrafficLimited]; ok {
		if trafficLimited, err := strconv.ParseBool(string(trafficLimitedByte)); err != nil {
			dataInfo.trafficLimited = trafficLimited
		}
	}
	if sessionConfigData, ok := extraDataMap[constant.InstanceSessionConfig]; ok {
		insSessConfig := commonTypes.InstanceSessionConfig{}
		err := json.Unmarshal(sessionConfigData, &insSessConfig)
		if err != nil {
			return nil, snerror.NewWithError(statuscode.StatusInternalServerError, err)
		}
		if insSessConfig.Concurrency == 0 || insSessConfig.Concurrency < -1 {
			return nil, snerror.New(statuscode.InstanceSessionInvalidErrCode,
				fmt.Sprintf("invalid session concurrency %d, only -1 or positive integer is allowed",
					insSessConfig.Concurrency))
		}
		if !utils.CheckInstanceSessionValid(insSessConfig) {
			return nil, snerror.New(statuscode.InstanceSessionInvalidErrCode, "session config invalid")
		}
		dataInfo.instanceSession = insSessConfig
	}
	if invokeLabel, ok := extraDataMap[constant.InstanceRequirementInvokeLabel]; ok {
		dataInfo.invokeLabel = invokeLabel
	}
	return dataInfo, nil
}

type extraDataInfo struct {
	designateInstanceName string
	designateInstanceID   string
	createEvent           []byte
	resourceData          []byte
	callerPodName         string
	poolLabel             string
	invokeLabel           []byte
	trafficLimited        bool
	instanceSession       commonTypes.InstanceSessionConfig
}

func judgeForwardToOtherCluster(funcURN string, logger api.FormatLogger) (bool, string, snerror.SNError) {
	functionAvailableRegistry := registry.GlobalRegistry.FunctionAvailableRegistry
	frontendRegistry := registry.GlobalRegistry.FaaSFrontendRegistry
	clusters := functionAvailableRegistry.GeClusters(funcURN)
	if len(clusters) == 0 {
		return false, "", nil
	}

	if commonUtils.IsStringInArray(os.Getenv(constant.ClusterIDKey), clusters) {
		return false, "", nil
	}

	for _, cluster := range clusters {
		frontends := frontendRegistry.GetFrontends(cluster)
		if len(frontends) == 0 {
			continue
		}
		endpoint := fmt.Sprintf("%s:%s", frontends[0], frontendNodePort)
		return true, endpoint, nil
	}
	logger.Errorf("func:%s need forward to other cluster, but no available frontend found", funcURN)
	return false, "", snerror.New(statuscode.StatusInternalServerError, "no available frontend found")
}

func (fs *FaaSScheduler) handleInstanceRelease(targetName string, metricsData []byte,
	traceID string,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	items := strings.Split(targetName, stateSplitStr)
	if len(items) == stateFuncKeyLen { // funcKey;stateID
		targetName = items[0]
		stateID := items[1]
		if stateID != "" {
			return fs.deleteState(stateID, targetName, logger)
		}
	}
	insAlloc, err := fs.loadInsAlloc(targetName, logger)
	if err != nil {
		return generateInstanceResponse(nil, err, startTime)
	}
	logger.Infof("handling instance release %s for function %s", insAlloc.AllocationID, insAlloc.Instance.FuncKey)
	if strings.Contains(insAlloc.AllocationID, "stateThread") { // %s-stateThread%d
		fs.allocRecord.Delete(insAlloc.AllocationID)
		err := fs.PoolManager.ReleaseStateThread(insAlloc)
		if err != nil {
			logger.Errorf("release thread %s fail, err: %v", targetName, err)
			return generateInstanceResponse(nil, snerror.New(statuscode.StatusInternalServerError,
				statuscode.InternalErrorMessage), startTime)
		}
		return generateInstanceResponse(insAlloc, nil, startTime)
	}
	data := fs.getInstanceThreadMetrics(insAlloc.AllocationID, metricsData)
	insThdMetrics := fs.buildMetrics(data)
	fs.reportMetrics(insAlloc.Instance.FuncKey, insAlloc.Instance.ResKey, insThdMetrics)
	fs.allocRecord.Delete(insAlloc.AllocationID)

	// If the arg:isAbnormal that received from fronted is true, the instance of this lease will be unusable
	// for user. Then the instance will be removed from instance queue and be clean.
	if data.IsAbnormal == true {
		fs.PoolManager.ReleaseAbnormalInstance(insAlloc.Instance, logger)
	}
	if !commonUtils.IsNil(insAlloc.Lease) {
		err := insAlloc.Lease.Release()
		if err != nil {
			// 正常情况下，通过insAlloc.Lease.Release()中的callback完成release
			// 这里用来防止实例被删除，pool中的sessionrecord中仍然残留sessioninfo的情况
			if err == lease.ErrInstanceNotFound {
				fs.PoolManager.ReleaseInstanceThread(insAlloc)
			}
			logger.Errorf("failed to release instance %s of function %s traceID %s error %s", insAlloc.AllocationID,
				insAlloc.Instance.FuncKey, traceID, err.Error())
		} else {
			logger.Infof("succeed to release instance %s of function %s traceID %s", insAlloc.AllocationID,
				insAlloc.Instance.FuncKey, traceID)
		}
	}
	return generateInstanceResponse(insAlloc, nil, startTime)
}

func (fs *FaaSScheduler) loadInsAlloc(targetName string, logger api.FormatLogger) (*types.InstanceAllocation,
	snerror.SNError,
) {
	rawData, exist := fs.allocRecord.Load(targetName)
	if !exist {
		logger.Errorf("allocation of instance thread %s not found", targetName)
		return nil, snerror.New(statuscode.InstanceNotFoundErrCode, statuscode.InstanceNotFoundErrMsg)
	}
	insAlloc, ok := rawData.(*types.InstanceAllocation)
	if !ok {
		logger.Errorf("instance thread allocation type error")
		return nil, snerror.New(statuscode.StatusInternalServerError, statuscode.InternalErrorMessage)
	}
	return insAlloc, nil
}

func (fs *FaaSScheduler) deleteState(stateID string, funcKey string,
	logger api.FormatLogger,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
	if funcSpec == nil {
		logger.Errorf("failed to get instance, function %s doesn't exist", funcKey)
		return generateInstanceResponse(nil, snerror.New(statuscode.FuncMetaNotFoundErrCode,
			statuscode.FuncMetaNotFoundErrMsg), startTime)
	}
	exist := fs.PoolManager.GetAndDeleteState(stateID, funcKey, funcSpec, logger)
	if !exist {
		return generateInstanceResponse(nil, snerror.New(statuscode.StateNotExistedErrCode,
			statuscode.StateNotExistedErrMsg), startTime)
	}
	return generateInstanceResponse(&types.InstanceAllocation{Instance: &types.Instance{}}, nil, startTime)
}

func (fs *FaaSScheduler) handleInstanceBatchRetain(target string, metricsData []byte,
	traceID string,
) *commonTypes.BatchInstanceResponse {
	startTime := time.Now()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	targetNames := strings.Split(target, ",")
	insThdMetrics := map[string]*types.InstanceThreadMetrics{}
	err := json.Unmarshal(metricsData, &insThdMetrics)
	if err != nil {
		logger.Errorf("failed to unmarshal metrics from data %s, err %s, trace %s", string(metricsData),
			err.Error(), traceID)
	}
	batchInstanceResp := &commonTypes.BatchInstanceResponse{
		InstanceAllocSucceed: map[string]commonTypes.InstanceAllocationSucceedInfo{},
		InstanceAllocFailed:  map[string]commonTypes.InstanceAllocationFailedInfo{},
		LeaseInterval:        fs.leaseInterval.Milliseconds(),
	}
	for _, name := range targetNames {
		if _, ok := insThdMetrics[name]; !ok {
			continue
		}
		insAlloc, err := fs.retainInstance(name, traceID, insThdMetrics[name], logger)
		if err != nil {
			batchInstanceResp.InstanceAllocFailed[name] = commonTypes.InstanceAllocationFailedInfo{
				ErrorCode:    err.Code(),
				ErrorMessage: err.Error(),
			}
			continue
		}
		batchInstanceResp.InstanceAllocSucceed[name] = commonTypes.InstanceAllocationSucceedInfo{
			FuncKey:    insAlloc.Instance.FuncKey,
			FuncSig:    insAlloc.Instance.FuncSig,
			InstanceID: insAlloc.Instance.InstanceID,
			ThreadID:   insAlloc.AllocationID,
		}
	}
	batchInstanceResp.SchedulerTime = time.Now().Sub(startTime).Seconds()
	return batchInstanceResp
}

func (fs *FaaSScheduler) handleInstanceRetain(targetName string, metricsData []byte,
	traceID string,
) *commonTypes.InstanceResponse {
	startTime := time.Now()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	insThdMetrics := &types.InstanceThreadMetrics{}
	err := json.Unmarshal(metricsData, insThdMetrics)
	if err != nil {
		logger.Errorf("failed to unmarshal metrics from data %s for instance %s", string(metricsData), targetName)
	}
	insAlloc, retainErr := fs.retainInstance(targetName, traceID, insThdMetrics, logger)
	return generateInstanceResponse(insAlloc, retainErr, startTime)
}

func (fs *FaaSScheduler) retainInstance(targetName, traceID string, insThdMetrics *types.InstanceThreadMetrics,
	logger api.FormatLogger,
) (*types.InstanceAllocation, snerror.SNError) {
	rawData, exist := fs.allocRecord.Load(targetName)
	if !exist && len(insThdMetrics.ReacquireData) == 0 {
		logger.Errorf("allocation of instance thread %s not found", targetName)
		return nil, snerror.New(statuscode.LeaseIDNotFoundCode,
			statuscode.LeaseIDNotFoundMsg)
	}
	if !exist {
		insAlloc, err := fs.reacquireLease(targetName, traceID, insThdMetrics, logger)
		if err != nil {
			logger.Errorf("reacquire lease failed, %s", err.Error())
			return nil, err
		}
		return insAlloc, err
	}
	insAlloc, ok := rawData.(*types.InstanceAllocation)
	if !ok {
		logger.Errorf("instance thread allocation type error")
		return nil, snerror.New(statuscode.StatusInternalServerError,
			statuscode.InternalErrorMessage)
	}
	if ownerSchedulerInstanceId, ok := selfregister.GlobalSchedulerProxy.CheckFuncOwner(
		insAlloc.Instance.FuncKey); !ok {
		logger.Errorf("non-owner faasscheduler, return owner faasscheduelr: %s", ownerSchedulerInstanceId)
		return nil, snerror.New(statuscode.AcquireNonOwnerSchedulerErrorCode, ownerSchedulerInstanceId)
	}
	if strings.Contains(insAlloc.AllocationID, "stateThread") { // %s-stateThread%d
		return fs.retainStateInstance(targetName, insAlloc, logger)
	}
	if insThdMetrics != nil {
		insThdMetrics.InsThdID = insAlloc.AllocationID
		fs.reportMetrics(insAlloc.Instance.FuncKey, insAlloc.Instance.ResKey, insThdMetrics)
	}
	if insAlloc.Instance.InstanceStatus.Code == int32(constant.KernelInstanceStatusSubHealth) {
		fs.allocRecord.Delete(insAlloc.AllocationID)
		if !commonUtils.IsNil(insAlloc.Lease) {
			err := insAlloc.Lease.Release()
			if err != nil {
				logger.Errorf("failed to delete abnormal thread %s of function %s error %s",
					insAlloc.AllocationID, insAlloc.Instance.FuncKey, err.Error())
			}
		}
		return nil, snerror.New(statuscode.InstanceStatusAbnormalCode, constant.LeaseErrorInstanceIsAbnormalMessage)
	}
	if !commonUtils.IsNil(insAlloc.Lease) {
		err := insAlloc.Lease.Extend()
		if err != nil {
			fs.allocRecord.Delete(insAlloc.AllocationID)
			logger.Errorf("failed to retain instance %s of function %s error %s", insAlloc.AllocationID,
				insAlloc.Instance.FuncKey, err.Error())
			return nil, snerror.New(constant.LeaseExpireOrDeletedErrorCode, constant.LeaseExpireOrDeletedErrorMessage)
		}
		logger.Infof("succeed to retain instance %s of function %s ", insAlloc.AllocationID,
			insAlloc.Instance.FuncKey)
	}
	return insAlloc, nil
}

func (fs *FaaSScheduler) reacquireLease(targetName, traceID string, insThdMetrics *types.InstanceThreadMetrics,
	logger api.FormatLogger,
) (*types.InstanceAllocation, snerror.SNError) {
	instanceId, _, parseErr := parseRetainTargetName(targetName)
	if parseErr != nil {
		return nil, snerror.New(statuscode.LeaseIDIllegalCode, statuscode.LeaseIDIllegalMsg)
	}
	dataInfo, err := parseExtraData(insThdMetrics.ReacquireData)
	if err != nil {
		return nil, err
	}
	funcSpec := registry.GlobalRegistry.GetFuncSpec(insThdMetrics.FunctionKey)

	resSpec, err := getResourceSpecification(dataInfo.resourceData, dataInfo.invokeLabel, funcSpec)
	if err != nil {
		return nil, err
	}
	logger.Infof("handling instance reacquire for resSpec %v instanceID %s instanceSession %v", resSpec,
		dataInfo.designateInstanceID, dataInfo.instanceSession)
	poolLabel := getPoolLabel(dataInfo.poolLabel, funcSpec.InstanceMetaData.PoolLabel)
	insAlloc, err := fs.PoolManager.AcquireInstanceThread(&types.InstanceAcquireRequest{
		FuncSpec:            funcSpec, // etcd
		ResSpec:             resSpec,  // args
		TraceID:             traceID,
		PoolLabel:           poolLabel,
		DesignateInstanceID: instanceId,
		DesignateThreadID:   targetName,
		InstanceSession:     dataInfo.instanceSession,
	})
	if err != nil {
		logger.Errorf("failed to reacquire instance of function %s traceID %s error %s", funcSpec.FuncKey, traceID,
			err.Error())
		return nil, err
	}
	if insAlloc.Lease != nil {
		fs.allocRecord.Store(insAlloc.AllocationID, insAlloc)
	}
	logger.Infof("succeed to reacquire instance %s of function %s traceID %s", insAlloc.AllocationID, funcSpec.FuncKey,
		traceID)
	return insAlloc, nil
}

func (fs *FaaSScheduler) retainStateInstance(targetName string, insAlloc *types.InstanceAllocation,
	logger api.FormatLogger,
) (*types.InstanceAllocation, snerror.SNError) {
	if insAlloc.Instance.InstanceStatus.Code == int32(constant.KernelInstanceStatusSubHealth) {
		err := fs.PoolManager.ReleaseStateThread(insAlloc)
		if err != nil {
			logger.Errorf("release thread %s fail", targetName)
		}
		return nil, snerror.New(statuscode.InstanceStatusAbnormalCode,
			constant.LeaseErrorInstanceIsAbnormalMessage)
	}
	err := fs.PoolManager.RetainStateThread(insAlloc)
	if err != nil {
		logger.Errorf("handleInstanceRetain err %v", err)
		return nil, snerror.New(constant.LeaseExpireOrDeletedErrorCode, constant.LeaseExpireOrDeletedErrorMessage)
	}
	return insAlloc, nil
}

func (fs *FaaSScheduler) handleRollout(targetName, traceID string) *commonTypes.RolloutResponse {
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	logger.Infof("received rollout request from %s, start to gray update", targetName)
	if !config.GlobalConfig.EnableRollout {
		return generateRolloutErrorResponse("", nil, errors.New("rollout is not enable"))
	}
	discoveryConfig := config.GlobalConfig.SchedulerDiscovery
	if discoveryConfig == nil || discoveryConfig.RegisterMode != types.RegisterTypeContend {
		return generateRolloutErrorResponse("", nil, errors.New("incompatible register mode"))
	}
	if !selfregister.Registered || len(selfregister.RegisterKey) == 0 {
		return generateRolloutErrorResponse("", nil, errors.New("scheduler not registered"))
	}
	rollout.GetGlobalRolloutHandler().IsGaryUpdating = true
	selfregister.IsRollingOut = true
	rollout.GetGlobalRolloutHandler().UpdateForwardInstance(targetName)
	allocRecord := make(map[string][]string)
	fs.allocRecord.Range(func(key, value any) bool {
		allocLease, ok := key.(string)
		if !ok {
			logger.Warnf("allocRecord key is invalid")
			return true
		}
		insAlloc, ok := value.(*types.InstanceAllocation)
		if !ok {
			logger.Warnf("allocRecord value is invalid")
			return true
		}
		funcKey := insAlloc.Instance.FuncKey
		allocRecord[funcKey] = append(allocRecord[funcKey], allocLease)
		return true
	})
	return generateRolloutErrorResponse(selfregister.RegisterKey, allocRecord, nil)
}

func (fs *FaaSScheduler) syncAllocRecordDuringRollout() {
	syncCh := rollout.GetGlobalRolloutHandler().GetAllocRecordSyncChan()
	for {
		select {
		case allocRecord, ok := <-syncCh:
			if !ok {
				log.GetLogger().Warnf("stop syncing allocation record")
				return
			}
			fs.syncAllocRecord(allocRecord)
		}
	}
}

func (fs *FaaSScheduler) syncAllocRecord(allocRecord map[string][]string) {
	log.GetLogger().Infof("start ot sync allocRecord")
	for funcKey, record := range allocRecord {
		funcSpec := registry.GlobalRegistry.GetFuncSpec(funcKey)
		if funcSpec == nil {
			log.GetLogger().Errorf("failed to sync allocRecord for function %s, function doesn't exist", funcKey)
			continue
		}
		resSpec := &resspeckey.ResourceSpecification{
			CPU:              funcSpec.ResourceMetaData.CPU,
			Memory:           funcSpec.ResourceMetaData.Memory,
			EphemeralStorage: funcSpec.ResourceMetaData.EphemeralStorage,
		}
		for _, allocation := range record {
			items := strings.Split(allocation, "-")
			insAcqReq := &types.InstanceAcquireRequest{
				FuncSpec:     funcSpec,
				ResSpec:      resSpec,
				InstanceName: items[0],
			}
			insAlloc, err := fs.PoolManager.AcquireInstanceThread(insAcqReq)
			if err != nil {
				log.GetLogger().Errorf("failed to sync allocation %s, acquire instance error %s", allocation,
					err.Error())
				continue
			}
			fs.allocRecord.Store(insAlloc.AllocationID, insAlloc)
		}
	}
}

func (fs *FaaSScheduler) reportMetrics(funcKey string, resKey resspeckey.ResSpecKey,
	insThdMetrics *types.InstanceThreadMetrics,
) {
	if len(funcKey) == 0 {
		return
	}
	fs.PoolManager.ReportMetrics(funcKey, resKey, insThdMetrics)
}

func (fs *FaaSScheduler) getInstanceThreadMetrics(threadID string, metricsData []byte) *types.InstanceThreadMetrics {
	metrics := &types.InstanceThreadMetrics{}
	err := json.Unmarshal(metricsData, metrics)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal metrics from data %s for instance %s", string(metricsData),
			threadID)
		return nil
	}
	metrics.InsThdID = threadID
	return metrics
}

func (fs *FaaSScheduler) buildMetrics(extraData *types.InstanceThreadMetrics) *types.InstanceThreadMetrics {
	if extraData == nil {
		return &types.InstanceThreadMetrics{}
	}
	return &types.InstanceThreadMetrics{
		ProcReqNum:  extraData.ProcReqNum,
		AvgProcTime: extraData.AvgProcTime,
		MaxProcTime: extraData.MaxProcTime,
	}
}

func parseInstanceOperation(args []api.Arg, traceID string) (InstanceOperation, string, []byte, []byte) {
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	insOp := insOpUnknown
	if len(args) < minArgsNum {
		logger.Errorf("argument number is smaller than %d check args %+v", minArgsNum, args)
		return insOp, "", nil, nil
	}
	operationArg := args[0]
	if operationArg.Type != api.Value {
		logger.Errorf("invalid argument type for args[0]")
		return insOp, "", nil, nil
	}
	items := strings.SplitN(string(operationArg.Data), insOpSeparator, validInsOpLen)
	if len(items) != validInsOpLen {
		logger.Errorf("failed to parse operation and target from %s", string(operationArg.Data))
		return insOp, "", nil, nil
	}
	insOp = InstanceOperation(items[0])
	target := items[1]
	if len(args) == minArgsNum {
		return insOp, target, nil, nil
	}
	extraDataArg := args[1]
	if extraDataArg.Type != api.Value {
		logger.Errorf("invalid argument type for args[1]")
		return insOp, target, nil, nil
	}
	eventDataArg := api.Arg{}
	// temporary process for forward compatible, remove this in future
	if len(args) == libruntimeValidArgsNum {
		eventDataArg = args[2]
	}
	return insOp, target, extraDataArg.Data, eventDataArg.Data
}

func getPoolLabel(poolLabelFromReq, poolLabelFromMeta string) string {
	if poolLabelFromReq != "" {
		return poolLabelFromReq
	}
	return poolLabelFromMeta
}

func parseStateOperation(ops string) (string, string) {
	targetName := ops
	items := strings.Split(ops, stateSplitStr)
	if len(items) != validArgsNum {
		return targetName, ""
	}

	targetName = items[0]
	stateID := items[1]

	return targetName, stateID
}

func parseRetainTargetName(targetName string) (string, string, error) {
	// targetName: f49a9bc8-bddd-4e0c-8000-00000000b90d-thread21
	items := strings.Split(targetName, "-thread")
	if len(items) != validArgsNum {
		return "", "", fmt.Errorf("target name fmt error. %s", targetName)
	}
	instanceId := items[0]
	threadId := items[1]
	return instanceId, threadId, nil
}

func getResourceSpecification(resData, labelData []byte, funcSpec *types.FunctionSpecification) (
	*resspeckey.ResourceSpecification, snerror.SNError,
) {
	resSpec := &resspeckey.ResourceSpecification{
		CustomResources: make(map[string]int64, constant.DefaultMapSize),
	}
	resMap := map[string]types.IntOrString{}
	if len(resData) != 0 {
		err := json.Unmarshal(resData, &resMap)
		if err != nil {
			return nil, snerror.NewWithError(statuscode.StatusInternalServerError, err)
		}
	}
	for k, v := range resMap {
		if v.Type != types.Int {
			continue
		}
		if k == constant.ResourceCPUName {
			resSpec.CPU = v.IntVal
			continue
		}
		if k == constant.ResourceMemoryName {
			resSpec.Memory = v.IntVal
			continue
		}
		resSpec.CustomResources[k] = v.IntVal
	}
	if resSpec.CPU == 0 {
		resSpec.CPU = funcSpec.ResourceMetaData.CPU
	}
	if resSpec.Memory == 0 {
		resSpec.Memory = funcSpec.ResourceMetaData.Memory
	}
	if resSpec.EphemeralStorage == 0 {
		resSpec.EphemeralStorage = funcSpec.ResourceMetaData.EphemeralStorage
	}
	if len(labelData) > 0 {
		labelMap := map[string]string{}
		err := json.Unmarshal(labelData, &labelMap)
		if err != nil {
			return nil, snerror.NewWithError(statuscode.StatusInternalServerError, err)
		}
		resSpec.InvokeLabel = labelMap[types.HeaderInstanceLabel]
	}
	return resSpec, nil
}

func generateInstanceResponse(insAlloc *types.InstanceAllocation, snErr snerror.SNError,
	startTime time.Time,
) *commonTypes.InstanceResponse {
	if snErr != nil {
		return &commonTypes.InstanceResponse{
			InstanceAllocationInfo: commonTypes.InstanceAllocationInfo{
				InstanceID:    "",
				LeaseInterval: 0,
			},
			ErrorCode:     snErr.Code(),
			ErrorMessage:  snErr.Error(),
			SchedulerTime: time.Now().Sub(startTime).Seconds(),
		}
	}
	leaseInterval := time.Duration(0)
	if insAlloc.Lease != nil {
		leaseInterval = insAlloc.Lease.GetInterval()
	}
	forceInvoke := false
	if insAlloc.Instance.InstanceStatus.Code == int32(constant.KernelInstanceStatusEvicting) {
		forceInvoke = true
	}
	return &commonTypes.InstanceResponse{
		InstanceAllocationInfo: commonTypes.InstanceAllocationInfo{
			FuncKey:       insAlloc.Instance.FuncKey,
			FuncSig:       insAlloc.Instance.FuncSig,
			InstanceID:    insAlloc.Instance.InstanceID,
			InstanceIP:    insAlloc.Instance.InstanceIP,
			InstancePort:  insAlloc.Instance.InstancePort,
			NodeIP:        insAlloc.Instance.NodeIP,
			NodePort:      insAlloc.Instance.NodePort,
			ThreadID:      insAlloc.AllocationID,
			LeaseInterval: leaseInterval.Milliseconds(),
			CPU:           insAlloc.Instance.ResKey.CPU,
			Memory:        insAlloc.Instance.ResKey.Memory,
			ForceInvoke:   forceInvoke,
		},
		ErrorCode:     constant.InsReqSuccessCode,
		ErrorMessage:  constant.InsReqSuccessMessage,
		SchedulerTime: time.Now().Sub(startTime).Seconds(),
	}
}

func generateRolloutErrorResponse(registerKey string, allocRecord map[string][]string,
	err error,
) *commonTypes.RolloutResponse {
	errorCode := constant.InsReqSuccessCode
	errorMessage := constant.InsReqSuccessMessage
	if err != nil {
		errorCode = statuscode.InternalErrorCode
		errorMessage = err.Error()
	}
	return &commonTypes.RolloutResponse{
		RegisterKey:  registerKey,
		AllocRecord:  allocRecord,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

func printInputLog() {
	log.GetLogger().Infof("%s is alive.", logFileName)
}
