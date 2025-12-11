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

// Package functionmanager -
package functionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"
	"time"

	k8serror "k8s.io/apimachinery/pkg/api/errors"

	"yuanrong.org/kernel/runtime/libruntime/api"

	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionmanager/config"
	"yuanrong.org/kernel/pkg/functionmanager/constant"
	"yuanrong.org/kernel/pkg/functionmanager/state"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
	"yuanrong.org/kernel/pkg/functionmanager/vpcmanager"
)

// Manager manages functions for faas pattern
type Manager struct {
	patKeyList        map[string]map[string]struct{}
	VpcPlugin         vpcmanager.PluginVPC
	vpcMutex          sync.Mutex
	remoteClientLease map[string]*leaseTimer
	remoteClientList  map[string]struct{}
	clientMutex       sync.Mutex
	stopCh            chan struct{}
	queue             *queue
	leaseRenewMinute  time.Duration
}

type leaseTimer struct {
	Timer  *time.Timer
	Cancel context.CancelFunc
	Ctx    context.Context
}

// RequestOperation defines request operations
type RequestOperation string

var (
	// requestOpCreate stands for create pat-service operation
	requestOpCreate RequestOperation = "CreatePATService"
	// requestOpCreateTrigger stands for create pullTrigger operation
	requestOpCreateTrigger RequestOperation = "CreatePullTrigger"
	// requestOpDeleteTrigger stands for delete pullTrigger operation
	requestOpDeleteTrigger RequestOperation = "DeletePullTrigger"
	// requestOpReport stands for instance report operation
	requestOpReport RequestOperation = "ReportInstanceID"
	// requestOpDelete stands for instance delete operation
	requestOpDelete RequestOperation = "DeleteInstanceID"
	// requestOpUnknown stands for unknown operation
	requestOpUnknown RequestOperation = "Unknown"
	// requestOpNewLease stands for create lease operation
	requestOpNewLease RequestOperation = commonconstant.NewLease
	// requestOpKeepAlive stands for keep-alive lease operation
	requestOpKeepAlive RequestOperation = commonconstant.KeepAlive
	// requestOpDelLease stands for delete lease operation
	requestOpDelLease RequestOperation = commonconstant.DelLease
	libruntimeClient  api.LibruntimeAPI
)

const (
	validArgsNum                     = 3
	validArgsNumLibruntime           = 4
	faasManagerOpIndex               = 0
	faasManagerOpDataIndex           = 1
	faasManagerOpIndexLibruntime     = 1
	faasManagerOpDataIndexLibruntime = 2

	leaseEtcdKeyLen = 3

	defaultLeaseRenewMinute = 5
)

// NewFaaSManagerLibruntime will create a new faas functions manager
func NewFaaSManagerLibruntime(libruntimeAPI api.LibruntimeAPI, stopCh chan struct{}) (*Manager, error) {
	libruntimeClient = libruntimeAPI
	return MakeFaasManager(stopCh)
}

// MakeFaasManager will create a new faas functions manager
func MakeFaasManager(stopCh chan struct{}) (*Manager, error) {
	functionLibPath := os.Getenv(constant.FunctionLibPath)
	filePath := filepath.Join(functionLibPath, constant.PluginFile)
	log.GetLogger().Infof("plugin file path is %s", filePath)
	pluginFile, err := plugin.Open(filePath)
	if err != nil {
		log.GetLogger().Errorf("failed to open vpc plugin file: %s", err.Error())
		return nil, err
	}
	cfg := config.GetConfig()
	leaseRenewMinute := cfg.LeaseRenewMinute
	if leaseRenewMinute == 0 {
		leaseRenewMinute = defaultLeaseRenewMinute
	}
	faaSManager := &Manager{
		patKeyList:        make(map[string]map[string]struct{}, 1),
		vpcMutex:          sync.Mutex{},
		remoteClientLease: make(map[string]*leaseTimer),
		remoteClientList:  map[string]struct{}{},
		clientMutex:       sync.Mutex{},
		stopCh:            stopCh,
		queue: &queue{
			cond:       sync.NewCond(&sync.RWMutex{}),
			dirty:      map[interface{}]struct{}{},
			processing: map[interface{}]struct{}{},
		},
		leaseRenewMinute: time.Duration(cfg.LeaseRenewMinute),
	}
	faaSManager.VpcPlugin.Plugin = pluginFile
	err = faaSManager.VpcPlugin.InitVpcPlugin()
	if err != nil {
		return nil, err
	}
	err = utils.InitKubeClient()
	if err != nil {
		return nil, err
	}
	go faaSManager.saveStateLoop()
	return faaSManager, nil
}

// RecoverData -
func (m *Manager) RecoverData() {
	patKeyList := state.GetState().PatKeyList
	log.GetLogger().Infof("now recover faaSManager patKeyList")
	m.vpcMutex.Lock()
	for patPodName, val := range patKeyList {
		instanceIDMap := val
		m.patKeyList[patPodName] = instanceIDMap
	}
	log.GetLogger().Infof("recovered patKeyList: %v", m.patKeyList)
	m.vpcMutex.Unlock()
	rcm := map[string]struct{}{}
	err := commonUtils.DeepCopyObj(state.GetState().RemoteClientList, &rcm)
	if err != nil {
		log.GetLogger().Errorf("recovered remoteClientList error: %v", err)
		return
	}
	log.GetLogger().Infof("now recover faaSManager remoteClientList")
	m.clientMutex.Lock()
	for client := range rcm {
		lease := newLeaseTimer(m.leaseRenewMinute * time.Minute)
		m.remoteClientLease[client] = lease
		m.remoteClientList[client] = struct{}{}
		go startLeaseTimeOutWatcher(lease.Ctx, m, lease.Timer, client, client)
	}
	log.GetLogger().Infof("recovered remoteClientList: %v", m.remoteClientList)
	m.clientMutex.Unlock()
}
func newLeaseTimer(timeout time.Duration) *leaseTimer {
	timer := time.NewTimer(timeout)
	ctx, cancel := context.WithCancel(context.Background())
	return &leaseTimer{
		Timer:  timer,
		Cancel: cancel,
		Ctx:    ctx,
	}
}

func startLeaseTimeOutWatcher(ctx context.Context, fm *Manager, timer *time.Timer, clientID string, traceID string) {
	select {
	case <-timer.C:
		log.GetLogger().Infof("lease timeout, traceID: %s, remoteClientID: %s", traceID, clientID)
		if fm != nil {
			fm.clearLease(clientID, traceID)
		}
	case <-ctx.Done():
		log.GetLogger().Infof("lease stopped before timeout, traceID: %s, remoteClientID: %s", traceID,
			clientID)
		timer.Stop()
	}
}

// HandlerRequest -
func (m *Manager) HandlerRequest(requestOp RequestOperation, requestData []byte,
	traceID string) *commonType.CallHandlerResponse {
	switch requestOp {
	case requestOpCreate:
		return m.handleRequestOpCreate(requestData, traceID)
	case requestOpCreateTrigger:
		return m.handleRequestOpCreateTrigger(requestData, traceID)
	case requestOpReport:
		return m.handleRequestOpReport(requestData, traceID)
	case requestOpDelete:
		return m.handleRequestOpDelete(requestData, traceID)
	case requestOpDeleteTrigger:
		return m.handleRequestOpDeleteTrigger(requestData, traceID)
	case requestOpNewLease:
		return m.handleNewLease(requestData, traceID, 0)
	case requestOpKeepAlive:
		return m.handleKeepAlive(requestData, traceID)
	case requestOpDelLease:
		return m.handleDelLease(requestData, traceID)
	default:
		log.GetLogger().Warnf("unknown request operation %s, traceID %s", requestOp, traceID)
	}
	return utils.GenerateErrorResponse(commonconstant.UnsupportedOperationErrorCode,
		commonconstant.UnsupportedOperationErrorMessage)
}

// ProcessSchedulerRequestLibruntime will handle create, report and delete of instance
func (m *Manager) ProcessSchedulerRequestLibruntime(args []api.Arg, traceID string) *commonType.CallHandlerResponse {
	requestOp, requestData := parseRequestOperationLibruntime(args)
	return m.HandlerRequest(requestOp, requestData, traceID)
}

// ProcessSchedulerRequest will handle create, report and delete of instance
func (m *Manager) ProcessSchedulerRequest(args []*api.Arg, traceID string) *commonType.CallHandlerResponse {
	requestOp, requestData := parseRequestOperation(args)
	return m.HandlerRequest(requestOp, requestData, traceID)
}

func (m *Manager) handleRequestOpCreate(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	newPatPod, err := m.VpcPlugin.CreateVpcResource(requestData)
	if err != nil {
		log.GetLogger().Errorf("failed to create vpc pat pod, traceID: %s, error: %s", traceID, err.Error())
		return nil
	}
	if !m.checkExists(newPatPod.PatPodName) {
		m.newPatListKey(newPatPod.PatPodName)
	}
	log.GetLogger().Infof("succeed to create pat service pod: %s, ip: %s, traceID %s",
		newPatPod.PatPodName, newPatPod.PatContainerIP, traceID)
	patInfo, err := json.Marshal(newPatPod)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal vpc pat pod info, traceID: %s, error: %s", traceID, err.Error())
		return nil
	}
	return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, string(patInfo))
}

func (m *Manager) handleRequestOpCreateTrigger(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	log.GetLogger().Infof("succeed receive pullTrigger create request, traceID: %s", traceID)
	requestInfo := types.PullTriggerRequestInfo{}
	err := json.Unmarshal(requestData, &requestInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to Unmarshal pullTrigger requestInfo, traceID: %s, err: %s",
			traceID, err.Error())
		return nil
	}
	podName := utils.HandlePullTriggerName(requestInfo.PodName)
	if podName == "" {
		log.GetLogger().Errorf("invalid pull trigger name, traceID: %s", traceID)
		return nil
	}
	requestInfo.PodName = podName
	newPatPod, err := m.VpcPlugin.CreateVpcResource(requestData)
	if err != nil {
		log.GetLogger().Errorf("failed to create pullTrigger vpc pat pod, traceID: %s, error: %s",
			traceID, err.Error())
		return nil
	}
	if !m.checkExists(newPatPod.PatPodName) {
		m.newPatListKey(newPatPod.PatPodName)
	}
	log.GetLogger().Infof("succeed to create pullTrigger pat pod: %s, ip: %s, traceID: %s",
		newPatPod.PatPodName, newPatPod.PatContainerIP, traceID)
	vpcNatConfByte, err := json.Marshal(newPatPod)
	if err != nil {
		log.GetLogger().Errorf("failed to Marshal PatPod info for %s, traceID: %s", newPatPod.PatPodName, traceID)
		return nil
	}
	triggerInfo := vpcmanager.ParseFunctionMeta(requestInfo)
	if triggerInfo == nil {
		return nil
	}
	// get pull trigger deploy
	triggerDeploy, errs := utils.GetDeployByK8S(utils.GetKubeClient(), triggerInfo.PodName)
	if errs != nil {
		log.GetLogger().Errorf("failed to get deploy %s, traceID %s", errs.Error(), traceID)
	}
	if k8serror.IsNotFound(errs) {
		// if trigger deploy is not found, it will create pull trigger deploy
		triggerDeploy = vpcmanager.MakePullTriggerDeploy(triggerInfo, vpcNatConfByte)
		if err = utils.CreateDeployByK8S(utils.GetKubeClient(), triggerDeploy); err != nil {
			log.GetLogger().Errorf("failed to create deploy %s by k8s, traceID: %s, error: %s",
				triggerDeploy.Name, traceID, err.Error())
			return nil
		}
		m.addPatList(newPatPod.PatPodName, requestInfo.PodName)
		log.GetLogger().Infof("succeed add Trigger to patList, patPod %s, name %s, traceID: %s",
			newPatPod.PatPodName, requestInfo.PodName, traceID)
		return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "succeed create pull-trigger")
	}
	log.GetLogger().Infof("trigger %s has already exist, traceID %s", triggerDeploy.Name, traceID)
	return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "skip create pull-trigger")
}

func (m *Manager) handleRequestOpReport(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	reportInfo := types.ReportInfo{}
	err := json.Unmarshal(requestData, &reportInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal report info, traceID: %s, error: %s", traceID, err.Error())
		return nil
	}
	if m.checkExists(reportInfo.PatPodName) {
		m.addPatList(reportInfo.PatPodName, reportInfo.InstanceID)
		log.GetLogger().Infof("succeed add pat list, patPodName %s, instanceID %s, traceID: %s",
			reportInfo.PatPodName, reportInfo.InstanceID, traceID)
		return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "succeed add pat list")
	}
	log.GetLogger().Infof("patPodName %s is not exist, skip, traceID: %s", reportInfo.PatPodName, traceID)
	return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "skip add pat list")
}

func (m *Manager) handleRequestOpDelete(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	deleteInfo := types.DeleteInfo{}
	err := json.Unmarshal(requestData, &deleteInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal delete info, traceID: %s, error: %s", traceID, err.Error())
		return nil
	}
	patPodName, needDeletePat := m.deletePatList(deleteInfo.InstanceID)
	if needDeletePat {
		err := m.VpcPlugin.DeleteVpcResource(patPodName)
		if err != nil {
			log.GetLogger().Errorf("failed to delete vpc pat pod, traceID: %s, error: %s", traceID, err.Error())
			return nil
		}
		log.GetLogger().Infof("succeed to delete pat-service %s, traceID: %s", patPodName, traceID)
	}
	log.GetLogger().Infof("succeed to delete instance %s, traceID: %s", deleteInfo.InstanceID, traceID)
	return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "succeed delete instance list")
}

func (m *Manager) handleRequestOpDeleteTrigger(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	deleteInfo := types.PullTriggerDeleteInfo{}
	err := json.Unmarshal(requestData, &deleteInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal pullTrigger delete info, traceID: %s, error: %s", traceID,
			err.Error())
		return nil
	}
	podName := utils.HandlePullTriggerName(deleteInfo.PodName)
	if podName == "" {
		log.GetLogger().Errorf("invalid pull trigger name, traceID: %s", traceID)
		return nil
	}
	deleteInfo.PodName = podName
	// get pull trigger deploy
	triggerDeploy, errs := utils.GetDeployByK8S(utils.GetKubeClient(), deleteInfo.PodName)
	if errs != nil {
		log.GetLogger().Errorf("failed to get deploy, traceID: %s error: %s", traceID, errs.Error())
	}
	if k8serror.IsNotFound(errs) {
		log.GetLogger().Infof("trigger %s has already deleted, traceID: %s", triggerDeploy.Name, traceID)
		return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "skip delete pull-trigger")
	}
	if err = utils.DeleteDeployByK8S(utils.GetKubeClient(), triggerDeploy.Name); err != nil {
		log.GetLogger().Errorf("failed to delete deploy %s by k8s, traceID: %s, error: %s",
			triggerDeploy.Name, traceID, err.Error())
		return nil
	}
	patPodName, needDeletePat := m.deletePatList(deleteInfo.PodName)
	if needDeletePat {
		err := m.VpcPlugin.DeleteVpcResource(patPodName)
		if err != nil {
			log.GetLogger().Errorf("failed to delete vpc pat pod, traceID: %s, error: %s", traceID, err.Error())
			return nil
		}
		log.GetLogger().Infof("succeed to delete pat-service %s, traceID: %s", patPodName, traceID)
	}
	log.GetLogger().Infof("succeed delete Trigger, patPod %s, name %s, traceID: %s",
		patPodName, deleteInfo.PodName, traceID)
	return utils.GenerateSuccessResponse(commonconstant.InsReqSuccessCode, "succeed delete pull-trigger")
}

func (m *Manager) handleNewLease(requestData []byte, traceID string, timeout int64) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive new lease from faas-frontend, traceID: %s", traceID)
	m.clientMutex.Lock()
	if _, value := m.remoteClientLease[remoteClientID]; value {
		m.clientMutex.Unlock()
		log.GetLogger().Infof("lease already existed, traceID: %s", traceID)
		return &commonType.CallHandlerResponse{Code: commonconstant.InsReqSuccessCode, Message: "lease existed"}
	}
	t := time.Minute * m.leaseRenewMinute
	if timeout > 0 {
		t = time.Duration(timeout) * time.Second
	}
	lease := newLeaseTimer(t)
	m.remoteClientLease[remoteClientID] = lease
	m.remoteClientList[remoteClientID] = struct{}{}
	m.clientMutex.Unlock()
	m.saveStateData(traceID)

	go startLeaseTimeOutWatcher(lease.Ctx, m, lease.Timer, remoteClientID, traceID)

	log.GetLogger().Infof("succeed to create lease, traceID: %s", traceID)
	return &commonType.CallHandlerResponse{Code: commonconstant.InsReqSuccessCode, Message: "Succeed to create lease"}
}

func (m *Manager) clearLease(clientID string, traceID string) {
	m.clientMutex.Lock()
	lease, ok := m.remoteClientLease[clientID]
	if !ok {
		log.GetLogger().Warnf("client id not existed, %s %s ", clientID, traceID)
		m.clientMutex.Unlock()
		killInstanceOuter(clientID, traceID)
		return
	}
	lease.Timer.Stop()
	lease.Cancel()
	delete(m.remoteClientLease, clientID)
	delete(m.remoteClientList, clientID)
	m.clientMutex.Unlock()
	m.saveStateData(traceID)
	killInstanceOuter(clientID, traceID)
}

type stateItem struct {
	ch      chan error
	traceID string
}

// String -
func (s *stateItem) String() string {
	return s.traceID
}

// Done -
func (s *stateItem) Done(err error) {
	s.ch <- err
}

func (m *Manager) saveStateData(traceID string) {
	i := &stateItem{
		ch:      make(chan error, 1),
		traceID: traceID,
	}
	if !m.queue.add(i) {
		log.GetLogger().Warnf("add save state req to queue failed, traceID: %s", traceID)
		return
	}
	err := <-i.ch
	if err != nil {
		log.GetLogger().Warnf("save state data failed, err:%v, traceID:%s", err, traceID)
	}
}

func (m *Manager) saveStateLoop() {
	for {
		all, shutdown := m.queue.getAll()
		if shutdown {
			return
		}
		m.clientMutex.Lock()
		managerState := state.GetState()
		state.ManagerStateLock.Lock()
		managerState.RemoteClientList = m.remoteClientList
		marshal, err := json.Marshal(managerState)
		state.ManagerStateLock.Unlock()
		if err != nil {
			m.queue.doneAll(all)
			for _, it := range all {
				it.Done(err)
			}
			continue
		}
		m.clientMutex.Unlock()
		client := etcd3.GetRouterEtcdClient()
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), etcd3.DurationContextTimeout)
			_, err = client.Client.Put(ctx, "/faas/state/recover/faasmanager", string(marshal))
			cancel()
		} else {
			err = fmt.Errorf("router etcd client is nil")
		}
		m.queue.doneAll(all)
		for _, it := range all {
			it.Done(err)
		}
	}
}

func killInstanceOuter(clientID string, traceID string) {

}

func (m *Manager) handleKeepAlive(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive keep-alive lease from faas-frontend, traceID: %s", traceID)
	m.clientMutex.Lock()
	if _, value := m.remoteClientLease[remoteClientID]; value {
		lease := m.remoteClientLease[remoteClientID]
		lease.Timer.Reset(m.leaseRenewMinute * time.Minute)
		m.clientMutex.Unlock()
		log.GetLogger().Infof("succeed to renew lease, traceID: %s", traceID)
		return &commonType.CallHandlerResponse{
			Code:    commonconstant.InsReqSuccessCode,
			Message: "Succeed to renew lease",
		}
	}
	m.clientMutex.Unlock()
	log.GetLogger().Infof("failed to renew unknown lease, traceID: %s", traceID)
	return &commonType.CallHandlerResponse{
		Code:    commonconstant.UnsupportedOperationErrorCode,
		Message: fmt.Sprintf("%s remote client id not exist", remoteClientID),
	}
}

func (m *Manager) handleDelLease(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive delete lease from faas-frontend, traceID: %s", traceID)
	m.clearLease(remoteClientID, traceID)

	return &commonType.CallHandlerResponse{
		Code:    commonconstant.InsReqSuccessCode,
		Message: "Succeed to delete lease",
	}
}

func (m *Manager) checkExists(patPodName string) bool {
	m.vpcMutex.Lock()
	if m.patKeyList == nil {
		m.vpcMutex.Unlock()
		return false
	}
	_, ok := m.patKeyList[patPodName]
	m.vpcMutex.Unlock()
	return ok
}

func (m *Manager) newPatListKey(patPodName string) {
	m.vpcMutex.Lock()
	m.patKeyList[patPodName] = map[string]struct{}{}
	savePatKeyList(m.patKeyList)
	m.vpcMutex.Unlock()
	return
}

func (m *Manager) addPatList(patPodName, instanceID string) {
	m.vpcMutex.Lock()
	if m.patKeyList == nil {
		m.vpcMutex.Unlock()
		return
	}
	m.patKeyList[patPodName][instanceID] = struct{}{}
	savePatKeyList(m.patKeyList)
	m.vpcMutex.Unlock()
	return
}

func (m *Manager) deletePatList(instanceID string) (string, bool) {
	patPodName := ""
	needDeletePat := false
	m.vpcMutex.Lock()
	if m.patKeyList == nil {
		m.vpcMutex.Unlock()
		return "", false
	}
	check := false
	for patName, instanceList := range m.patKeyList {
		for id := range instanceList {
			if id == instanceID {
				delete(m.patKeyList[patName], instanceID)
				savePatKeyList(m.patKeyList)
				check = true
				break
			}
		}
		if check {
			if len(m.patKeyList[patName]) == 0 {
				delete(m.patKeyList, patName)
				savePatKeyList(m.patKeyList)
				patPodName = patName
				needDeletePat = true
			}
			break
		}
	}
	m.vpcMutex.Unlock()
	return patPodName, needDeletePat
}

func savePatKeyList(m map[string]map[string]struct{}) {
	dst := map[string]map[string]struct{}{}
	err := commonUtils.DeepCopyObj(m, &dst)
	if err != nil {
		log.GetLogger().Errorf("deep copy patKeyList failed, err: %v", err)
		return
	}
	state.Update(dst)
}

func parseRequestOperationLibruntime(args []api.Arg) (RequestOperation, []byte) {
	requestOp := requestOpUnknown
	if len(args) != validArgsNumLibruntime {
		log.GetLogger().Errorf("invalid argument number")
		return requestOp, []byte{}
	}
	requestOp = RequestOperation(args[faasManagerOpIndexLibruntime].Data)
	requestData := args[faasManagerOpDataIndexLibruntime].Data
	return requestOp, requestData
}

func parseRequestOperation(args []*api.Arg) (RequestOperation, []byte) {
	requestOp := requestOpUnknown
	if len(args) != validArgsNum {
		log.GetLogger().Errorf("invalid argument number")
		return requestOp, []byte{}
	}
	requestOp = RequestOperation(args[faasManagerOpIndex].Data)
	requestData := args[faasManagerOpDataIndex].Data
	return requestOp, requestData
}

// WatchLeaseEvent -
func (m *Manager) WatchLeaseEvent() {
	etcdClient := etcd3.GetRouterEtcdClient()
	watcher := etcd3.NewEtcdWatcher(commonconstant.LeasePrefix, func(event *etcd3.Event) bool {
		etcdKey := event.Key
		keyParts := strings.Split(etcdKey, commonconstant.ETCDEventKeySeparator)
		return len(keyParts) == leaseEtcdKeyLen
	}, func(event *etcd3.Event) {
		log.GetLogger().Infof("handling lease event type %d, key:%s, remoteClientLease in use len: %d", event.Type,
			event.Key, len(m.remoteClientLease))
		switch event.Type {
		case etcd3.PUT:
			m.handleLeaseEvent(event)
			go func() {
				_, err := etcdClient.Client.Delete(context.Background(), event.Key)
				if err != nil {
					log.GetLogger().Errorf("delete lease event failed, key: %s, err: %v", event.Key, err)
				}
			}()
		case etcd3.DELETE:
		case etcd3.ERROR:
			log.GetLogger().Warnf("etcd error event: %s", event.Value)
		default:
			log.GetLogger().Warnf("unsupported event, key: %s", event.Key)
		}
	}, m.stopCh, etcdClient)
	watcher.StartWatch()
	select {
	case <-m.stopCh:
		m.queue.shutDown()
		return
	}
}

func (m *Manager) handleLeaseEvent(event *etcd3.Event) {
	if event == nil || len(event.Value) == 0 {
		log.GetLogger().Warnf("event is nil or value is empty")
		return
	}
	log.GetLogger().Infof("handling lease put event key:%s", event.Key)
	e := &commonType.LeaseEvent{}
	err := json.Unmarshal(event.Value, e)
	if err != nil {
		log.GetLogger().Errorf("error unmarshalling lease event: %s, err: %v", string(event.Value), err)
		return
	}
	switch e.Type {
	case commonconstant.NewLease:
		m.handleNewLease([]byte(e.RemoteClientID), e.TraceID, 0)
	case commonconstant.DelLease:
		m.handleDelLease([]byte(e.RemoteClientID), e.TraceID)
	case commonconstant.KeepAlive:
		// 判断是否是过期心跳
		timeout := int64((m.leaseRenewMinute * time.Minute).Seconds()) - (time.Now().Unix() - e.Timestamp)
		if timeout <= 0 {
			log.GetLogger().Infof("lease is expired, traceID: %s", e.TraceID)
			m.handleDelLease([]byte(e.RemoteClientID), e.TraceID)
			return
		}
		m.clientMutex.Lock()
		if lease, value := m.remoteClientLease[e.RemoteClientID]; value {
			lease.Timer.Reset(time.Duration(timeout) * time.Second)
			m.remoteClientLease[e.RemoteClientID] = lease
			m.clientMutex.Unlock()
			log.GetLogger().Infof("succeed to renew lease, traceID: %s", e.TraceID)
			return
		}
		m.clientMutex.Unlock()
		m.handleNewLease([]byte(e.RemoteClientID), e.TraceID, timeout)
	default:
		log.GetLogger().Errorf("unexpected lease type: %s, traceId: %s", e.Type, e.TraceID)
	}
}
