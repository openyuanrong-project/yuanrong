/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package crprocessor -
package crprocessor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/functioncr"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

var manager *AgentCRsManager

const (
	minRetryInterval = 100 * time.Millisecond
	maxRetryInterval = 5 * time.Minute
	defaultChanSize  = 1000
)

// NewAgentCRsManager -
func NewAgentCRsManager(stopCh <-chan struct{}) *AgentCRsManager {
	manager = &AgentCRsManager{
		AgentCRCh:                  make(chan registry.SubEvent, defaultChanSize),
		AgentCRInsCh:               make(chan registry.SubEvent, defaultChanSize),
		instanceListCh:             make(chan struct{}, 1),
		instanceConfigChs:          make([]chan registry.SubEvent, 0, 1),
		functionChs:                make([]chan registry.SubEvent, 0, 1),
		FaaSSchedulerProxyCh:       make(chan registry.SubEvent, defaultChanSize),
		RWMutex:                    sync.RWMutex{},
		logger:                     log.GetLogger(),
		stopCh:                     stopCh,
		crNameToAgentCRReconciler:  make(map[string]*AgentCRReconciler, constant.DefaultMapSize),
		funcKeyToAgentCRReconciler: make(map[string]*AgentCRReconciler, constant.DefaultMapSize),
		instanceMap:                make(map[string]*commonType.InstanceSpecification, constant.DefaultMapSize),
		agentEventInfoMap:          make(map[string]*types.AgentEventInfo, constant.DefaultMapSize),
		synced:                     false,
	}
	return manager
}

// AgentCRsManager -
type AgentCRsManager struct {
	AgentCRCh            chan registry.SubEvent
	AgentCRInsCh         chan registry.SubEvent
	FaaSSchedulerProxyCh chan registry.SubEvent
	instanceListCh       chan struct{}

	instanceConfigChs []chan registry.SubEvent
	functionChs       []chan registry.SubEvent

	stopCh <-chan struct{}

	sync.RWMutex
	logger                     api.FormatLogger
	crNameToAgentCRReconciler  map[string]*AgentCRReconciler // key: crName, value: AgentCRReconciler
	funcKeyToAgentCRReconciler map[string]*AgentCRReconciler
	instanceMap                map[string]*commonType.InstanceSpecification
	agentEventInfoMap          map[string]*types.AgentEventInfo
	synced                     bool
}

// AgentCRReconciler -
type AgentCRReconciler struct {
	funcKey            string
	crKey              string
	info               *types.AgentEventInfo
	logger             api.FormatLogger
	funcSign           string
	insConfigSign      string
	instances          map[string]*commonType.InstanceSpecification
	enable             bool
	stopCh             chan struct{}
	reconcileTriggerCh chan struct{}
	sync.RWMutex
}

// AgentCRInfo -
type AgentCRInfo struct {
	crName   string
	funcSpec *types.FunctionSpecification
	status   *types.AgentCRStatus
}

// StartLoop -
func (arm *AgentCRsManager) StartLoop() {
	go arm.ProcessCREventLoop()
	go arm.ProcessInstanceEventLoop()
	go arm.ProcessFaaSSchedulerProxyEventLoop()
}

// ProcessCREventLoop -
func (arm *AgentCRsManager) ProcessCREventLoop() {
	handleSyncEventFunc := func(eventMsg interface{}) {
		ch, ok := eventMsg.(chan struct{})
		if !ok {
			return
		}
		arm.logger.Infof("recv cr synced event")
		if ch != nil {
			ch <- struct{}{}
		}
		return
	}
	for {
		select {
		case event, ok := <-arm.AgentCRCh:
			if !ok {
				return
			}
			if event.EventType == registry.SubEventTypeSynced {
				handleSyncEventFunc(event.EventMsg)
				continue
			}
			agentRunEventInfo, ok := event.EventMsg.(*types.AgentEventInfo)
			if !ok {
				continue
			}
			arm.processCRInfoEvent(event.EventType, agentRunEventInfo)
		case <-arm.stopCh:
			arm.logger.Infof("process cr event loop exited")
			return
		}
	}
}

// AddInstanceConfigSubscriberChan -
func (arm *AgentCRsManager) AddInstanceConfigSubscriberChan(subChan chan registry.SubEvent) {
	arm.Lock()
	arm.instanceConfigChs = append(arm.instanceConfigChs, subChan)
	arm.Unlock()
}

// AddFunctionSubscriberChan -
func (arm *AgentCRsManager) AddFunctionSubscriberChan(subChan chan registry.SubEvent) {
	arm.Lock()
	arm.functionChs = append(arm.functionChs, subChan)
	arm.Unlock()
}

// ProcessInstanceEventLoop -
func (arm *AgentCRsManager) ProcessInstanceEventLoop() {
	for {
		select {
		case event, ok := <-arm.AgentCRInsCh:
			if !ok {
				return
			}
			if event.EventType == registry.SubEventTypeSynced {
				arm.logger.Infof("process instance sync event")
				arm.processInstanceEvent(registry.SubEventTypeSynced, nil)
				arm.logger.Infof("process instance sync event over")
				continue
			}

			info, ok := event.EventMsg.(*commonType.InstanceSpecification)
			if !ok {
				continue
			}

			if !utils.IsFaaSInstance(info.Function) {
				continue
			}
			arm.processInstanceEvent(event.EventType, info)
		case <-arm.stopCh:
			arm.logger.Infof("process instance event loop exited")
			return
		}
	}
}

// ProcessFaaSSchedulerProxyEventLoop -
func (arm *AgentCRsManager) ProcessFaaSSchedulerProxyEventLoop() {
	for {
		select {
		case _, ok := <-arm.FaaSSchedulerProxyCh:
			if !ok {
				return
			}
			arm.processSchedulerProxyChanged()
		case <-arm.stopCh:
			arm.logger.Infof("process faasscheduler proxy event loop exited")
			return
		}
	}
}

func isCrReconcilerOwner(synced bool, funcKey string) bool {
	if !synced {
		return false
	}
	return selfregister.GlobalSchedulerProxy.IsFuncOwner(funcKey)
}

func (arm *AgentCRsManager) processCRInfoEvent(eventType registry.EventType, info *types.AgentEventInfo) {
	arm.Lock()
	defer arm.Unlock()
	logger := arm.logger.With(zap.Any("crKey", info.CrKey), zap.Any("eventType", eventType))
	logger.Infof("process cr info event")
	defer logger.Infof("process cr info event over")
	switch eventType {
	case registry.SubEventTypeUpdate:
		arm.agentEventInfoMap[info.CrKey] = info
		reconciler, ok := arm.crNameToAgentCRReconciler[info.CrKey]
		if !ok {
			reconciler = newAgentRunCRReconciler(info, isCrReconcilerOwner(arm.synced, info.FuncSpec.FuncKey))
			arm.crNameToAgentCRReconciler[info.CrKey] = reconciler
			arm.funcKeyToAgentCRReconciler[info.FuncSpec.FuncKey] = reconciler
		}
		if !ok || reconciler.IsFuncSpecChanged(info.FuncSpec) {
			logger.Infof("publish funcSpec event, funcKey: %s", info.FuncSpec.FuncKey)
			arm.publishFunctionEvent(eventType, info.FuncSpec)
		}
		if !ok || reconciler.IsInstanceConfigChanged(info.FuncSpec) {
			if arm.synced {
				logger.Infof("publish instance config event, funcKey: %s", info.FuncSpec.FuncKey)
				arm.publishInstanceConfigEvent(eventType, info.FuncSpec)
			}
		}
		reconciler.processCRUpdate(info)
	case registry.SubEventTypeDelete:
		delete(arm.agentEventInfoMap, info.CrKey)
		reconciler, ok := arm.crNameToAgentCRReconciler[info.CrKey]
		if !ok {
			return
		}
		delete(arm.crNameToAgentCRReconciler, reconciler.crKey)
		delete(arm.funcKeyToAgentCRReconciler, reconciler.funcKey)
		arm.publishFunctionEvent(eventType, reconciler.info.FuncSpec)
		arm.publishInstanceConfigEvent(eventType, reconciler.info.FuncSpec)
		reconciler.exit()
	default:
	}
}

func (arm *AgentCRsManager) publishFunctionEvent(eventType registry.EventType,
	funcSpec *types.FunctionSpecification) {
	for _, ch := range arm.functionChs {
		if ch == nil {
			continue
		}
		ch <- registry.SubEvent{
			EventType: eventType,
			EventMsg:  funcSpec,
		}
	}
}

func (arm *AgentCRsManager) publishInstanceConfigEvent(event registry.EventType,
	info *types.FunctionSpecification) {
	if info == nil {
		return
	}
	insConfig := &instanceconfig.Configuration{
		FuncKey:          info.FuncKey,
		InstanceLabel:    "",
		InstanceMetaData: info.InstanceMetaData,
	}
	for _, ch := range arm.instanceConfigChs {
		if ch == nil {
			continue
		}
		ch <- registry.SubEvent{
			EventType: event,
			EventMsg:  insConfig,
		}
	}
}

func (arm *AgentCRsManager) processSchedulerProxyChanged() {
	arm.RLock()
	defer arm.RUnlock()
	for _, reconciler := range arm.crNameToAgentCRReconciler {
		reconciler.setEnable(isCrReconcilerOwner(arm.synced, reconciler.funcKey))
	}
}

func (arm *AgentCRsManager) processInstanceEvent(eventType registry.EventType,
	info *commonType.InstanceSpecification) {
	arm.RLock()
	defer arm.RUnlock()
	if eventType == registry.SubEventTypeSynced {
		arm.sync()
		return
	}

	funckey := info.CreateOptions[types.FunctionKeyNote]
	logger := arm.logger.With(zap.Any("funcKey", funckey), zap.Any("instanceId", info.InstanceID),
		zap.Any("statusCode", info.InstanceStatus.Code), zap.Any("eventType", eventType))
	logger.Infof("process instance event")
	defer logger.Infof("process instance event over")
	ar, ok := arm.funcKeyToAgentCRReconciler[info.CreateOptions[types.FunctionKeyNote]]
	if !ok {
		logger.Infof("no reconciler, funckey: %s, instanceId: %s", funckey, info.InstanceID)
		return
	}
	switch eventType {
	case registry.SubEventTypeUpdate:
		if info.InstanceStatus.Code != int32(constant.KernelInstanceStatusRunning) {
			ar.processInstanceDelete(info)
		} else {
			ar.processInstanceUpdate(info, logger)
		}
	case registry.SubEventTypeDelete:
		ar.processInstanceDelete(info)
	default:
	}
}

func (arm *AgentCRsManager) sync() {
	if arm.synced {
		return
	}
	arm.synced = true

	for _, info := range arm.agentEventInfoMap {
		arm.publishInstanceConfigEvent(registry.SubEventTypeUpdate, info.FuncSpec)
	}
	arm.publishInstanceConfigEvent(registry.SubEventTypeCRSynced, &types.FunctionSpecification{})

	for _, reconciler := range arm.crNameToAgentCRReconciler {
		reconciler.setEnable(isCrReconcilerOwner(arm.synced, reconciler.funcKey))
	}
	arm.logger.Infof("sync over")
}

func newAgentRunCRReconciler(info *types.AgentEventInfo, enable bool) *AgentCRReconciler {
	reconciler := &AgentCRReconciler{
		funcKey:   info.FuncSpec.FuncKey,
		crKey:     info.CrKey,
		info:      info,
		instances: make(map[string]*commonType.InstanceSpecification),
		logger: log.GetLogger().With(zap.Any("crKey", info.CrKey),
			zap.Any("funcKey", info.FuncSpec.FuncKey)),
		enable:             enable,
		stopCh:             make(chan struct{}),
		reconcileTriggerCh: make(chan struct{}, 1),
		RWMutex:            sync.RWMutex{},
	}
	go reconciler.reconcileLoop()
	return reconciler
}

func (ar *AgentCRReconciler) exit() {
	commonUtils.SafeCloseChannel(ar.stopCh)
}

func (ar *AgentCRReconciler) processCRUpdate(info *types.AgentEventInfo) {
	ar.Lock()
	defer ar.Unlock()
	if ar.funcSign != info.FuncSpec.FuncMetaSignature {
		ar.funcSign = info.FuncSpec.FuncMetaSignature
		ar.info.FuncSpec = info.FuncSpec
	}
	ar.insConfigSign = ""
	if info.FuncSpec != nil {
		ar.insConfigSign = getInstanceConfigSign(info.FuncSpec.InstanceMetaData)
	}
	if isStatusDiff(ar.info.Status, info.Status) {
		ar.info.Status = info.Status
	}
	if ar.enable && !isExpectedStatus(ar.info.FuncSpec.FuncMetaSignature, info.Status, ar.instances) {
		ar.triggerReconcile()
	}
}

// IsFuncSpecChanged -
func (ar *AgentCRReconciler) IsFuncSpecChanged(funcSpec *types.FunctionSpecification) bool {
	ar.RLock()
	defer ar.RUnlock()
	return ar.funcSign != funcSpec.FuncMetaSignature
}

func getInstanceConfigSign(insConfig commonType.InstanceMetaData) string {
	instanceConfigBytes, err := json.Marshal(insConfig)
	if err != nil {
		return ""
	}
	return commonUtils.FnvHash(string(instanceConfigBytes))
}

// IsInstanceConfigChanged -
func (ar *AgentCRReconciler) IsInstanceConfigChanged(funcSpec *types.FunctionSpecification) bool {
	ar.RLock()
	defer ar.RUnlock()
	return ar.insConfigSign != getInstanceConfigSign(funcSpec.InstanceMetaData)
}

func isStatusDiff(statusold, statusNew *types.AgentCRStatus) bool {
	if statusold == nil && statusNew == nil {
		return false
	}
	if statusold == nil {
		statusold = &types.AgentCRStatus{}
	}
	if statusNew == nil {
		statusNew = &types.AgentCRStatus{}
	}

	if statusold.ReadyReplicas != statusNew.ReadyReplicas || len(statusold.Conditions) != len(statusNew.Conditions) {
		return true
	}

	for i, _ := range statusold.Conditions {
		if statusold.Conditions[i].Status != statusNew.Conditions[i].Status ||
			statusold.Conditions[i].LastTransitionTime != statusNew.Conditions[i].LastTransitionTime ||
			statusold.Conditions[i].Message != statusNew.Conditions[i].Message ||
			statusold.Conditions[i].Type != statusNew.Conditions[i].Type ||
			statusold.Conditions[i].Reason != statusNew.Conditions[i].Reason {
			return true
		}
	}
	return false
}

func isExpectedStatus(expectedSignature string, status *types.AgentCRStatus,
	instancesMap map[string]*commonType.InstanceSpecification) bool {
	expectInstances := make(map[string]*commonType.InstanceSpecification)
	for _, ins := range instancesMap {
		if ins.CreateOptions[types.FunctionSign] == expectedSignature {
			expectInstances[ins.InstanceID] = ins
		}
	}
	if status == nil {
		status = &types.AgentCRStatus{}
	}
	if len(status.Conditions) != len(expectInstances) {
		return false
	}
	for _, condition := range status.Conditions {
		if condition.Type != "ReadyPod" || condition.Reason != "ReadyPod" || condition.Status != "true" {
			return false
		}
		_, ok := expectInstances[condition.Message]
		if !ok {
			return false
		}
	}
	return true
}

func (ar *AgentCRReconciler) setEnable(enable bool) {
	ar.Lock()
	defer ar.Unlock()
	oldEnable := ar.enable
	ar.enable = enable
	if enable && !oldEnable {
		ar.triggerReconcile()
	}
}

func (ar *AgentCRReconciler) processInstanceUpdate(info *commonType.InstanceSpecification, logger api.FormatLogger) {
	ar.Lock()
	defer ar.Unlock()
	_, ok := ar.instances[info.InstanceID]
	ar.instances[info.InstanceID] = info
	if ar.info.FuncSpec.FuncMetaSignature != info.CreateOptions[types.FunctionSign] {
		logger.Infof("signature not match, func signature: %s, instance signature: %s",
			ar.info.FuncSpec.FuncMetaSignature, info.CreateOptions[types.FunctionSign])
		return
	}
	if !ok && ar.enable {
		ar.triggerReconcile()
	}
}

func (ar *AgentCRReconciler) processInstanceDelete(info *commonType.InstanceSpecification) {
	ar.Lock()
	defer ar.Unlock()
	instance, ok := ar.instances[info.InstanceID]
	if !ok {
		return
	}
	delete(ar.instances, info.InstanceID)
	if instance.CreateOptions[types.FunctionSign] != ar.info.FuncSpec.FuncMetaSignature {
		return
	}
	if ar.enable {
		ar.triggerReconcile()
	}
}

func (ar *AgentCRReconciler) triggerReconcile() {
	select {
	case ar.reconcileTriggerCh <- struct{}{}:
	default:
		ar.logger.Warnf("reconcile triggerCh is blocked")
	}
}

func (ar *AgentCRReconciler) reconcileLoop() {
	timer := time.NewTimer(1 * time.Second)
	timer.Stop()
	retryInterval := minRetryInterval
	f := func() {
		ar.RLock()
		if !ar.enable {
			ar.RUnlock()
			return
		}
		if isExpectedStatus(ar.info.FuncSpec.FuncMetaSignature, ar.info.Status, ar.instances) {
			ar.RUnlock()
			timer.Stop()
			retryInterval = minRetryInterval
			return
		}
		ar.RUnlock()
		ar.logger.Infof("status is not expected, begin reconcile")
		if ar.reconcileStatus() != nil {
			ar.logger.Infof("reconcile status failed")
			retryInterval *= 2

			if retryInterval >= maxRetryInterval {
				retryInterval = maxRetryInterval
			}
			timer.Reset(retryInterval)
		} else {
			ar.logger.Infof("reconcile status over")
			timer.Stop()
			retryInterval = minRetryInterval
		}
	}
	for {
		select {
		case _, ok := <-ar.reconcileTriggerCh:
			if !ok {
				ar.logger.Errorf("reconcile triggerCh is closed")
				timer.Stop()
				return
			}
			ar.logger.Infof("recv ch from reconcileTriggerCh")
			f()
		case <-timer.C:
			ar.logger.Infof("triggered by timer.C")
			f()
		case <-ar.stopCh:
			timer.Stop()
			ar.logger.Infof("reconcile loop exited")
			return
		}
	}

}

func (ar *AgentCRReconciler) reconcileStatus() error {
	resource := k8sclient.GetDynamicClient().Resource(functioncr.GetCrdGVR())
	namespace, crName, err := utils.ParseFromCrKey(ar.crKey)
	if err != nil {
		log.GetLogger().Warnf("%s, and will use the default namespace", err.Error())
		namespace = constant.DefaultNameSpace
		crName = ar.crKey
	}
	crObj, err := resource.Namespace(namespace).Get(context.Background(), crName, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		ar.logger.Infof("cr is not found, no need reconcile status")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get cr, error: %v", err)
	}

	crStatus, err := ar.getCrStatus(crObj)
	if err != nil {
		return fmt.Errorf("failed to get cr's status, error: %v", err)
	}

	ar.RLock()
	crStatus.Conditions = make([]*metav1.Condition, 0)
	for _, ins := range ar.instances {
		if ins.CreateOptions[types.FunctionSign] != ar.info.FuncSpec.FuncMetaSignature {
			continue
		}
		crStatus.Conditions = append(crStatus.Conditions, &metav1.Condition{
			LastTransitionTime: metav1.Now(),
			Message:            ins.InstanceID,
			Reason:             "Ready",
			Status:             metav1.ConditionTrue,
			Type:               "ReadyPod",
		})
	}
	ar.RUnlock()
	crStatus.ReadyReplicas = len(crStatus.Conditions)
	err = ar.updateCRStatus(resource, crStatus, crObj, namespace)
	if err != nil {
		return fmt.Errorf("failed to update CR Status, error: %s", err.Error())
	}
	return nil
}

func (ar *AgentCRReconciler) getCrStatus(crObj *unstructured.Unstructured) (*types.AgentCRStatus, error) {
	crStatus := &types.AgentCRStatus{
		Conditions:    []*metav1.Condition{},
		ReadyReplicas: 0,
	}

	if statusObj, ok := crObj.UnstructuredContent()["status"]; ok {
		status, ok := statusObj.(map[string]interface{})
		if !ok {
			ar.logger.Errorf("cr status format error, status: %v", statusObj)
			return nil, errors.New("transfer status format error")
		}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(status, crStatus)
		if err != nil {
			ar.logger.Errorf("cr status format is not AgentRunStatus, error: %v", err)
			return nil, errors.New("status format is not AgentRunStatus")
		}
	}
	return crStatus, nil
}

func (ar *AgentCRReconciler) updateCRStatus(resource dynamic.NamespaceableResourceInterface,
	crStatus *types.AgentCRStatus, crObj *unstructured.Unstructured, namespace string) error {
	toUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(crStatus)
	if err != nil {
		return fmt.Errorf("failed to transfer status to Unstructured, error: %v", err)
	}
	err = unstructured.SetNestedField(crObj.Object, toUnstructured, "status")
	if err != nil {
		return fmt.Errorf("failed to set status, error: %v", err)
	}
	_, err = resource.Namespace(namespace).UpdateStatus(context.Background(), crObj.DeepCopy(), metav1.UpdateOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to update cr, error: %v", err)
		}
		log.GetLogger().Warnf("CR not found when update status, error: %v", err)
	}
	return nil
}
